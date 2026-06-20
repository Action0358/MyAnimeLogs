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
	Processed          int // 実際に処理したループ件数（中断時は Total 未満になる）
	PendingTranslation int // 翻訳候補（英語あらすじ）の件数。DryRun時はコスト見積もりの基礎
	Translated         int // 英語→日本語へ翻訳しDB更新した件数
	FormattedOnly      int // 書式のみ更新した件数（日本語、または翻訳されず整形のみ）
	AlreadyClean       int // 整形・翻訳とも変更不要だった件数
	Failed             int // DB更新に失敗した件数
}

type backfillRecord struct {
	id   int
	desc string
}

// BackfillDescriptions は既存作品のあらすじを一括で整形・日本語化する。
// - 日本語のあらすじ → 書式整形のみ（翻訳しない＝コスト0）
// - 英語のあらすじ   → Claudeで翻訳＋整形（ANTHROPIC_API_KEY未設定時は整形のみ）
// 翻訳結果はDBに保存され、再実行時は日本語化済みの行が整形のみになるため冪等・再開可能。
// ctx がキャンセルされた場合は速やかにループを抜け、ここまでの集計結果を返す。
func (s *AnimeService) BackfillDescriptions(ctx context.Context, opts BackfillOptions) (*BackfillResult, error) {
	records, err := s.fetchNonEmptyDescriptions(ctx, opts.Limit)
	if err != nil {
		return nil, err
	}

	result := &BackfillResult{Total: len(records)}

	claudeEnabled := s.claudeClient != nil && s.claudeClient.IsEnabled()
	if !opts.DryRun && !claudeEnabled {
		s.logger.Warn("ANTHROPIC_API_KEY未設定のため、英語あらすじは翻訳されず整形のみが適用されます")
	}

	for _, r := range records {
		// 中断（Ctrl+C/タイムアウト）に速やかに反応する
		if err := ctx.Err(); err != nil {
			s.logger.Warn("コンテキストがキャンセルされました。ここまでの結果を返します",
				"processed", result.Processed, "total", result.Total)
			break
		}
		result.Processed++

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
		}

		if final == r.desc {
			result.AlreadyClean++
		} else if err := s.updateDescriptionByID(ctx, r.id, final); err != nil {
			// DB更新失敗時は即中断する。継続するとDB障害時に翻訳APIの予算を
			// 保存できない更新に浪費してしまうため。再実行で続きから再開可能。
			result.Failed++
			s.logger.Error("DB更新に失敗したためバックフィルを中断します", "anime_id", r.id, "error", err)
			return result, fmt.Errorf("failed to update description (anime_id=%d): %w", r.id, err)
		} else if !japanese && isJapaneseText(final) {
			result.Translated++
		} else {
			result.FormattedOnly++
		}

		if result.Processed%50 == 0 {
			s.logger.Info("Backfill progress",
				"processed", result.Processed, "total", result.Total,
				"translated", result.Translated, "formatted", result.FormattedOnly)
		}

		// 翻訳API呼び出しを行った場合のみスロットリング（Claude無効＝整形のみの時は不要）
		if !japanese && claudeEnabled && opts.Throttle > 0 {
			if err := sleepWithContext(ctx, opts.Throttle); err != nil {
				s.logger.Warn("スロットリング中に中断されました。ここまでの結果を返します",
					"processed", result.Processed, "total", result.Total)
				break
			}
		}
	}

	return result, nil
}

// fetchNonEmptyDescriptions は非空descriptionの作品を取得する。
func (s *AnimeService) fetchNonEmptyDescriptions(ctx context.Context, limit int) ([]backfillRecord, error) {
	query := `SELECT id, description FROM animes WHERE description IS NOT NULL AND description != '' ORDER BY id`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query animes: %w", err)
	}
	defer rows.Close()

	var records []backfillRecord
	for rows.Next() {
		var r backfillRecord
		if err := rows.Scan(&r.id, &r.desc); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return records, nil
}

// updateDescriptionByID は指定アニメの description を更新する。
func (s *AnimeService) updateDescriptionByID(ctx context.Context, animeID int, description string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE animes SET description = $1 WHERE id = $2`, description, animeID)
	return err
}

// sleepWithContext は d だけ待機するが、ctx がキャンセルされた場合は即座に返す。
func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
