package services

import (
	"context"
	"fmt"
	"time"
)

// BackfillOptions はあらすじ一括バックフィルの実行オプション。
type BackfillOptions struct {
	DryRun   bool          // true ならDB更新・翻訳APIを行わず、対象/翻訳候補件数のみ集計
	Limit    int           // 0=全件。N>0 で先頭N件のみ（検証・少数サンプル用）
	Throttle time.Duration // 翻訳API呼び出し間の待機（負荷・コスト制御）
}

// BackfillResult はバックフィル結果の集計。
type BackfillResult struct {
	Total              int // 走査対象（非空description）の件数
	PendingTranslation int // 翻訳候補（英語あらすじ）の件数。DryRun時はコスト見積もりの基礎
	Translated         int // 英語→日本語へ翻訳しDB更新した件数
	FormattedOnly      int // 書式のみ更新した件数（日本語、または翻訳されず整形のみ）
	AlreadyClean       int // 整形・翻訳とも変更不要だった件数
	Failed             int // DB更新に失敗した件数
}

// BackfillDescriptions は既存作品のあらすじを一括で整形・日本語化する。
// - 日本語のあらすじ → 書式整形のみ（翻訳しない＝コスト0）
// - 英語のあらすじ   → Claudeで翻訳＋整形（ANTHROPIC_API_KEY未設定時は整形のみ）
// 翻訳結果はDBに保存され、再実行時は日本語化済みの行が整形のみになるため冪等・再開可能。
func (s *AnimeService) BackfillDescriptions(ctx context.Context, opts BackfillOptions) (*BackfillResult, error) {
	query := `SELECT id, description FROM animes WHERE description IS NOT NULL AND description != '' ORDER BY id`
	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query animes: %w", err)
	}

	type record struct {
		id   int
		desc string
	}
	var records []record
	for rows.Next() {
		var r record
		if err := rows.Scan(&r.id, &r.desc); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	rows.Close()

	result := &BackfillResult{Total: len(records)}

	if !opts.DryRun && (s.claudeClient == nil || !s.claudeClient.IsEnabled()) {
		s.logger.Warn("ANTHROPIC_API_KEY未設定のため、英語あらすじは翻訳されず整形のみが適用されます")
	}

	for i, r := range records {
		sanitized := sanitizeDescription(r.desc)
		japanese := isJapaneseText(sanitized)

		if !japanese {
			result.PendingTranslation++ // 英語＝翻訳候補
		}

		// dry-run: 集計のみ（DB更新・翻訳APIを行わない）
		if opts.DryRun {
			if japanese && sanitized != r.desc {
				result.FormattedOnly++
			}
			continue
		}

		// 実処理
		var final string
		if japanese {
			final = sanitized // 整形のみ
		} else {
			final = s.finalizeDescriptionText(ctx, r.desc) // 翻訳＋整形（キー無効なら整形のみ）
			if opts.Throttle > 0 {
				time.Sleep(opts.Throttle)
			}
		}

		if final == r.desc {
			result.AlreadyClean++
		} else if err := s.updateDescriptionByID(ctx, r.id, final); err != nil {
			s.logger.Warn("Failed to update description", "anime_id", r.id, "error", err)
			result.Failed++
		} else if !japanese && isJapaneseText(final) {
			result.Translated++
		} else {
			result.FormattedOnly++
		}

		if (i+1)%50 == 0 {
			s.logger.Info("Backfill progress",
				"processed", i+1, "total", result.Total,
				"translated", result.Translated, "formatted", result.FormattedOnly)
		}
	}

	return result, nil
}

// updateDescriptionByID は指定アニメの description を更新する。
func (s *AnimeService) updateDescriptionByID(ctx context.Context, animeID int, description string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE animes SET description = $1 WHERE id = $2`, description, animeID)
	return err
}
