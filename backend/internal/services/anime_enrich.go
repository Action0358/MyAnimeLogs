package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"myanimelogs/internal/constants"
	"myanimelogs/internal/models"
)

// enrichAnimeWithCoverImage - AniListから画像を取得してアニメ情報を拡張
// DBキャッシュを優先し、キャッシュがない場合のみAPIを呼び出す
func (s *AnimeService) enrichAnimeWithCoverImage(anime *models.Anime) {
	// Use background context for image enrichment
	ctx := context.Background()

	var coverImage *CoverImage

	// MAL IDの有効性を一度だけチェック（DRY原則）
	hasMalID := anime.MalAnimeID != nil && *anime.MalAnimeID != ""

	// レイヤー1: データベースキャッシュをチェック（MAL IDがある場合）
	if hasMalID {
		var extraLarge, large, medium string
		// キャッシュTTLをGoで計算してパラメータとして渡す（ポータビリティ向上）
		cacheCutoff := time.Now().Add(-constants.CoverImageCacheTTL)
		cacheErr := s.db.QueryRowContext(ctx, `
			SELECT COALESCE(cover_image_extra_large, ''), cover_image_large, cover_image_medium
			FROM animes 
			WHERE mal_anime_id = $1 
			AND cover_image_large IS NOT NULL
			AND cover_image_updated_at > $2
		`, *anime.MalAnimeID, cacheCutoff).Scan(&extraLarge, &large, &medium)

		if cacheErr == nil && large != "" {
			// DBキャッシュから取得成功
			anime.CoverImageExtraLarge = extraLarge
			anime.CoverImageLarge = large
			anime.CoverImageMedium = medium
			return
		}

		// キャッシュミス以外のDBエラーは警告としてログ出力
		if cacheErr != nil && cacheErr != sql.ErrNoRows {
			s.logger.Warn("Failed to query cover image cache from DB",
				"error", cacheErr,
				"mal_anime_id", *anime.MalAnimeID,
				"anime_title", anime.Title)
		}

		// レイヤー2: AniList APIから取得（MAL IDで検索）
		var apiErr error
		coverImage, apiErr = s.anilistClient.GetCoverImageByMALId(ctx, *anime.MalAnimeID)
		if apiErr != nil {
			s.logger.Debug("Failed to get cover image by MAL ID, trying title fallback",
				"error", apiErr,
				"mal_anime_id", *anime.MalAnimeID,
				"anime_title", anime.Title)
		}
	}

	// MAL IDがないか、MAL IDでの取得に失敗した場合は、タイトルで検索（フォールバック）
	if coverImage == nil && anime.Title != "" {
		var titleErr error
		coverImage, titleErr = s.anilistClient.GetCoverImageByTitle(ctx, anime.Title)
		if titleErr != nil {
			// フォールバックも失敗した場合は、Annictの画像を使用
			s.logger.Debug("Failed to get cover image by title",
				"error", titleErr,
				"anime_title", anime.Title)
			return
		}
	}

	// 画像URLを設定
	if coverImage != nil {
		anime.CoverImageExtraLarge = coverImage.ExtraLarge
		anime.CoverImageLarge = coverImage.Large
		anime.CoverImageMedium = coverImage.Medium

		// DBに保存（MAL IDがある場合のみ）- ワーカープールで非同期実行
		if hasMalID {
			malID := *anime.MalAnimeID
			extraLarge := coverImage.ExtraLarge
			large := coverImage.Large
			medium := coverImage.Medium
			s.enqueueBackgroundTask(func() {
				// タイムアウト付きcontextでDBハングを防止
				ctx, cancel := context.WithTimeout(context.Background(), constants.BackgroundTaskDBTimeout)
				defer cancel()

				_, saveErr := s.db.ExecContext(ctx, `
					UPDATE animes 
					SET cover_image_extra_large = $1,
						cover_image_large = $2,
						cover_image_medium = $3,
						cover_image_updated_at = NOW()
					WHERE mal_anime_id = $4
				`, extraLarge, large, medium, malID)
				if saveErr != nil {
					s.logger.Warn("Failed to save cover image to DB",
						"error", saveErr,
						"mal_anime_id", malID)
				}
			})
		}
	}
}

// enrichAnimesWithCoverImages - 複数のアニメに画像を追加（並列処理）
func (s *AnimeService) enrichAnimesWithCoverImages(animes []models.Anime) {
	var wg sync.WaitGroup

	// 各アニメに対して並列で画像を取得
	for i := range animes {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			s.enrichAnimeWithCoverImage(&animes[index])
		}(i)
	}

	// すべてのゴルーチンが完了するまで待機
	wg.Wait()
}

// enrichAnimeWithDescription - TMDBから日本語あらすじを取得してアニメ情報を拡張
// DBにdescriptionがない場合のみAPIを呼び出す
// TMDBが利用できない場合はAniListにフォールバック
// シリーズ作品の場合、同じシリーズからあらすじを取得するフォールバックも実装
// finalizeDescriptionText はあらすじの書式を統一し、日本語でない場合は Claude で日本語へ翻訳する。
// Claude が無効、または翻訳失敗時は整形済みの原文（英語など）にフォールバックする。
func (s *AnimeService) finalizeDescriptionText(ctx context.Context, raw string) string {
	sanitized := sanitizeDescription(raw)
	if sanitized == "" {
		return ""
	}
	if isJapaneseText(sanitized) {
		return sanitized
	}
	if s.claudeClient == nil || !s.claudeClient.IsEnabled() {
		return sanitized
	}

	// 詳細ページ取得時に同期実行されるため、APIが遅い場合は素早く諦めて
	// 整形済み原文（英語）にフォールバックし、画面遷移をブロックしない。
	tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	translated, err := s.claudeClient.TranslateAndFormatSynopsis(tctx, sanitized)
	if err != nil || translated == "" {
		s.logger.Debug("Claude synopsis translation unavailable, using original text",
			"error", err)
		return sanitized
	}
	return sanitizeDescription(translated)
}

func (s *AnimeService) enrichAnimeWithDescription(anime *models.Anime) {
	ctx := context.Background()

	// まずTMDBから日本語あらすじを取得を試みる（MAL ID経由優先）
	if s.tmdbClient != nil && s.tmdbClient.IsEnabled() {
		malID := ""
		if anime.MalAnimeID != nil {
			malID = *anime.MalAnimeID
		}

		// AniListからメディアタイプを取得（MOVIE, TV等）
		mediaTypeHint := ""
		if malID != "" {
			format, err := s.anilistClient.GetMediaFormatByMALId(ctx, malID)
			if err == nil {
				mediaTypeHint = format // "MOVIE", "TV", "OVA", etc.
			}
		}

		overview, tmdbID, err := s.tmdbClient.GetOverviewByMALIDOrTitle(ctx, malID, anime.Title, mediaTypeHint)
		if err == nil && overview != "" {
			finalized := s.finalizeDescriptionText(ctx, overview)
			anime.Description = &finalized

			// DBに保存（非同期）
			animeID := anime.ID
			seriesTitle := NormalizeAnimeTitle(anime.Title)
			s.enqueueBackgroundTask(func() {
				s.updateAnimeSeriesInfo(context.Background(), animeID, finalized, tmdbID, seriesTitle, tmdbID)
			})
			return
		}

		// TMDBから直接取得できなかった場合、同じシリーズからあらすじを取得
		seriesTitle := NormalizeAnimeTitle(anime.Title)
		var seriesDescription string
		var seriesTMDBID sql.NullInt64
		seriesErr := s.db.QueryRow(`
			SELECT description, series_tmdb_id
			FROM animes 
			WHERE series_title = $1 
			  AND description IS NOT NULL 
			  AND description != ''
			ORDER BY series_tmdb_id IS NOT NULL DESC, id ASC
			LIMIT 1
		`, seriesTitle).Scan(&seriesDescription, &seriesTMDBID)

		if seriesErr == nil && seriesDescription != "" {
			// シリーズDBの値は過去の処理で整形済みだが、念のため整形（翻訳は不要）
			finalized := sanitizeDescription(seriesDescription)
			anime.Description = &finalized

			// DBに保存（非同期）
			animeID := anime.ID
			s.enqueueBackgroundTask(func() {
				var tmdbID int
				if seriesTMDBID.Valid {
					tmdbID = int(seriesTMDBID.Int64)
				}
				s.updateAnimeSeriesInfo(context.Background(), animeID, finalized, 0, seriesTitle, tmdbID)
			})
			return
		}

		s.logger.Debug("Failed to get description from TMDB, falling back to AniList",
			"error", err,
			"mal_id", malID,
			"anime_title", anime.Title)
	}

	// TMDBが利用できない場合はAniListにフォールバック
	// MAL IDがない場合は取得できない
	if anime.MalAnimeID == nil || *anime.MalAnimeID == "" {
		return
	}

	// AniList APIからあらすじを取得
	description, err := s.anilistClient.GetDescriptionByMALId(ctx, *anime.MalAnimeID)
	if err != nil {
		s.logger.Debug("Failed to get description from AniList",
			"error", err,
			"mal_anime_id", *anime.MalAnimeID,
			"anime_title", anime.Title)
		return
	}

	if description == "" {
		return
	}

	// 書式統一＋日本語化（AniListは英語あらすじが多い）
	finalized := s.finalizeDescriptionText(ctx, description)
	if finalized == "" {
		return
	}
	anime.Description = &finalized

	// DBに保存（非同期）
	malID := *anime.MalAnimeID
	s.enqueueBackgroundTask(func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.BackgroundTaskDBTimeout)
		defer cancel()

		_, saveErr := s.db.ExecContext(ctx, `
			UPDATE animes
			SET description = $1
			WHERE mal_anime_id = $2
		`, finalized, malID)
		if saveErr != nil {
			s.logger.Warn("Failed to save description to DB",
				"error", saveErr,
				"mal_anime_id", malID)
		}
	})
}

// updateAnimeSeriesInfo updates anime description and series information
// Only fields with non-empty/non-zero values are included in the update to prevent overwriting existing data.
func (s *AnimeService) updateAnimeSeriesInfo(ctx context.Context, animeID int, description string, tmdbID int, seriesTitle string, seriesTMDBID int) {
	dbCtx, cancel := context.WithTimeout(ctx, constants.BackgroundTaskDBTimeout)
	defer cancel()

	// Build dynamic query
	var updates []string
	var args []interface{}
	argIdx := 1

	if description != "" {
		updates = append(updates, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, description)
		argIdx++
	}
	if tmdbID > 0 {
		updates = append(updates, fmt.Sprintf("tmdb_id = $%d", argIdx))
		args = append(args, tmdbID)
		argIdx++
	}
	if seriesTitle != "" {
		updates = append(updates, fmt.Sprintf("series_title = $%d", argIdx))
		args = append(args, seriesTitle)
		argIdx++
	}
	if seriesTMDBID > 0 {
		updates = append(updates, fmt.Sprintf("series_tmdb_id = $%d", argIdx))
		args = append(args, seriesTMDBID)
		argIdx++
	}

	if len(updates) == 0 {
		return
	}

	query := fmt.Sprintf("UPDATE animes SET %s WHERE id = $%d", strings.Join(updates, ", "), argIdx)
	args = append(args, animeID)

	_, err := s.db.ExecContext(dbCtx, query, args...)
	if err != nil {
		s.logger.Warn("Failed to update anime series info in DB",
			"error", err,
			"anime_id", animeID)
	}
}

// ensureSeriesTitle - series_titleがない場合に設定する
// シリーズフォールバック機能のために、全てのアニメにseries_titleを設定する
// descriptionが既にある場合でも、series_titleがない場合は設定する
func (s *AnimeService) ensureSeriesTitle(anime *models.Anime) {
	// series_titleが既にある場合はスキップ
	if anime.SeriesTitle != nil && *anime.SeriesTitle != "" {
		return
	}

	// タイトルからseries_titleを生成
	seriesTitle := NormalizeAnimeTitle(anime.Title)

	// DBに保存（非同期）
	animeID := anime.ID
	s.enqueueBackgroundTask(func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.BackgroundTaskDBTimeout)
		defer cancel()

		_, saveErr := s.db.ExecContext(ctx, `
			UPDATE animes 
			SET series_title = $1
			WHERE id = $2 AND (series_title IS NULL OR series_title = '')
		`, seriesTitle, animeID)
		if saveErr != nil {
			s.logger.Warn("Failed to save series_title to DB",
				"error", saveErr,
				"anime_id", animeID)
		}
	})

	// メモリ上のanimeにも設定
	anime.SeriesTitle = &seriesTitle
}

// enrichAnimeWithReleasedOn - Jikan APIから放送開始日を取得してアニメ情報を拡張
// DBにreleased_onがない場合のみAPIを呼び出す
// Annictで欠損している放送開始日をJikanから補完する
func (s *AnimeService) enrichAnimeWithReleasedOn(anime *models.Anime) {
	// 既に放送開始日がある場合はスキップ
	if anime.ReleasedOn != nil && *anime.ReleasedOn != "" {
		return
	}

	// MAL IDがない場合は取得できない
	if anime.MalAnimeID == nil || *anime.MalAnimeID == "" {
		return
	}

	// Jikan APIから放送開始日を取得
	releasedOn, err := s.jikanClient.GetAnimeAiredDate(*anime.MalAnimeID)
	if err != nil {
		s.logger.Debug("Failed to get aired date from Jikan",
			"error", err,
			"mal_anime_id", *anime.MalAnimeID,
			"anime_title", anime.Title)
		return
	}

	if releasedOn == "" {
		return
	}

	// アニメ情報に設定
	anime.ReleasedOn = &releasedOn

	// DBに保存（非同期）
	malID := *anime.MalAnimeID
	animeID := anime.ID
	s.enqueueBackgroundTask(func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.BackgroundTaskDBTimeout)
		defer cancel()

		_, saveErr := s.db.ExecContext(ctx, `
			UPDATE animes 
			SET released_on = $1
			WHERE id = $2
		`, releasedOn, animeID)
		if saveErr != nil {
			s.logger.Warn("Failed to save released_on to DB",
				"error", saveErr,
				"mal_anime_id", malID,
				"anime_id", animeID)
		} else {
			s.logger.Debug("Saved released_on from Jikan to DB",
				"anime_id", animeID,
				"released_on", releasedOn)
		}
	})
}

// enrichAnimeWithMovieDuration - Jikan APIから映画の上映時間を取得してアニメ情報を拡張
// 映画作品でduration_minutesがない場合のみAPIを呼び出す
func (s *AnimeService) enrichAnimeWithMovieDuration(anime *models.Anime) {
	// 既に上映時間がある場合はスキップ
	if anime.DurationMinutes != nil {
		return
	}

	// 映画作品でない場合はスキップ
	if anime.Media == nil || *anime.Media != "movie" {
		return
	}

	// MAL IDがない場合は取得できない
	if anime.MalAnimeID == nil || *anime.MalAnimeID == "" {
		return
	}

	// 取得できなかったことを覚えていないと、Jikanにデータが無い作品では
	// 詳細取得のたびに同じ問い合わせを同期的に繰り返し、待ち時間が永久に解消しない。
	// 成功時はDBの duration_minutes に残るため、ここでは失敗だけを記憶する。
	ctx, cancel := context.WithTimeout(context.Background(), constants.BackgroundTaskDBTimeout)
	defer cancel()
	if s.isMovieDurationKnownMissing(ctx, anime.ID) {
		return
	}

	// Jikan APIから上映時間を取得
	duration, err := s.jikanClient.GetAnimeDuration(*anime.MalAnimeID)
	if err != nil {
		s.logger.Debug("Failed to get movie duration from Jikan",
			"error", err,
			"mal_anime_id", *anime.MalAnimeID,
			"anime_title", anime.Title)
		// レート制限や一時的な障害の可能性があるため、短めに抑える
		s.markMovieDurationMissing(ctx, anime.ID, constants.MovieDurationErrorCacheTTL)
		return
	}

	if duration <= 0 {
		// Jikan側に上映時間が無い作品。次に増える見込みは薄いので長めに保持する
		s.markMovieDurationMissing(ctx, anime.ID, constants.MovieDurationMissingCacheTTL)
		return
	}

	// アニメ情報に設定
	anime.DurationMinutes = &duration

	// DBに保存（非同期）
	malID := *anime.MalAnimeID
	animeID := anime.ID
	s.enqueueBackgroundTask(func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.BackgroundTaskDBTimeout)
		defer cancel()

		_, saveErr := s.db.ExecContext(ctx, `
			UPDATE animes 
			SET duration_minutes = $1
			WHERE id = $2
		`, duration, animeID)
		if saveErr != nil {
			s.logger.Warn("Failed to save movie duration to DB",
				"error", saveErr,
				"mal_anime_id", malID,
				"anime_id", animeID)
		} else {
			s.logger.Debug("Saved movie duration from Jikan to DB",
				"anime_id", animeID,
				"duration_minutes", duration)
		}
	})
}

// enrichAnimeWithBroadcastInfo - Jikan APIから放送情報を取得してアニメ情報を拡張
// DBにbroadcast_dayがない場合のみAPIを呼び出す
func (s *AnimeService) enrichAnimeWithBroadcastInfo(anime *models.Anime) {
	// 既に放送情報がある場合はスキップ
	if anime.BroadcastDay != nil {
		return
	}

	// MAL IDがない場合は取得できない
	if anime.MalAnimeID == nil || *anime.MalAnimeID == "" {
		return
	}

	// Jikan APIから放送情報を取得
	broadcast, err := s.jikanClient.GetBroadcastInfo(*anime.MalAnimeID)
	if err != nil {
		s.logger.Debug("Failed to get broadcast info from Jikan",
			"error", err,
			"mal_anime_id", *anime.MalAnimeID,
			"anime_title", anime.Title)
		return
	}

	if broadcast == nil {
		return
	}

	// アニメ情報に設定
	if broadcast.Day != nil && *broadcast.Day != "" {
		anime.BroadcastDay = broadcast.Day
	}
	if broadcast.Time != nil && *broadcast.Time != "" {
		anime.BroadcastTime = broadcast.Time
	}
	if broadcast.Timezone != nil && *broadcast.Timezone != "" {
		anime.BroadcastTimezone = broadcast.Timezone
	}

	// 放送曜日がない場合はDB保存もスキップ
	if anime.BroadcastDay == nil {
		return
	}

	// DBに保存（非同期）
	malID := *anime.MalAnimeID
	animeID := anime.ID
	day := anime.BroadcastDay
	bTime := anime.BroadcastTime
	tz := anime.BroadcastTimezone
	s.enqueueBackgroundTask(func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.BackgroundTaskDBTimeout)
		defer cancel()

		_, saveErr := s.db.ExecContext(ctx, `
			UPDATE animes 
			SET broadcast_day = $1, broadcast_time = $2, broadcast_timezone = $3
			WHERE id = $4
		`, day, bTime, tz, animeID)
		if saveErr != nil {
			s.logger.Warn("Failed to save broadcast info to DB",
				"error", saveErr,
				"mal_anime_id", malID,
				"anime_id", animeID)
		} else {
			s.logger.Debug("Saved broadcast info from Jikan to DB",
				"anime_id", animeID,
				"broadcast_day", day)
		}
	})
}

// enrichAnimesWithCoverImagesIfNeeded - DBキャッシュがないアニメのみ画像を取得（最適化版）
// DBから既に画像が取得されている場合はAPIコールをスキップ
func (s *AnimeService) enrichAnimesWithCoverImagesIfNeeded(animes []models.Anime) {
	var wg sync.WaitGroup
	var needEnrichCount int

	// 画像が必要なアニメのみ処理
	for i := range animes {
		// DBキャッシュがある場合はスキップ
		if animes[i].CoverImageExtraLarge != "" || animes[i].CoverImageLarge != "" {
			continue
		}

		needEnrichCount++
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			s.enrichAnimeWithCoverImage(&animes[index])
		}(i)
	}

	if needEnrichCount > 0 {
		s.logger.Debug("Enriching animes with cover images",
			"total", len(animes),
			"need_enrich", needEnrichCount,
			"cached", len(animes)-needEnrichCount)
	}

	wg.Wait()
}

// enrichAnimesWithUserDataCoverImages - AnimeWithUserDataのスライスに画像を追加（並列処理）
func (s *AnimeService) enrichAnimesWithUserDataCoverImages(animes []models.AnimeWithUserData) {
	var wg sync.WaitGroup

	// 各アニメに対して並列で画像を取得
	for i := range animes {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			s.enrichAnimeWithCoverImage(&animes[index].Anime)
		}(i)
	}

	// すべてのゴルーチンが完了するまで待機
	wg.Wait()
}

// enrichAnimesWithUserDataCoverImagesIfNeeded - DBキャッシュがないアニメのみ画像を取得（最適化版）
func (s *AnimeService) enrichAnimesWithUserDataCoverImagesIfNeeded(animes []models.AnimeWithUserData) {
	var wg sync.WaitGroup
	var needEnrichCount int

	for i := range animes {
		// DBキャッシュがある場合はスキップ
		if animes[i].CoverImageExtraLarge != "" || animes[i].CoverImageLarge != "" {
			continue
		}

		needEnrichCount++
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			s.enrichAnimeWithCoverImage(&animes[index].Anime)
		}(i)
	}

	if needEnrichCount > 0 {
		s.logger.Debug("Enriching animes with user data cover images",
			"total", len(animes),
			"need_enrich", needEnrichCount,
			"cached", len(animes)-needEnrichCount)
	}

	wg.Wait()
}

func (s *AnimeService) enrichAnimeWithWatchProviders(animeWithData *models.AnimeWithUserData) {
	if s.tmdbClient == nil || !s.tmdbClient.IsEnabled() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// キャッシュヒット時はTMDBを一切叩かずに返す。詳細取得はこのenrichが
	// 唯一の同期外部API呼び出しで、ここがレスポンス時間の大半を占めるため。
	if cached, ok := s.getCachedWatchProviders(ctx, animeWithData.ID); ok {
		animeWithData.WatchProviders = cached
		return
	}

	// First, try to get TMDB ID and series_tmdb_id from database
	var tmdbID int
	var seriesTMDBID sql.NullInt64
	var dbMediaType sql.NullString
	err := s.db.QueryRow(`SELECT tmdb_id, series_tmdb_id, media FROM animes WHERE id = $1`, animeWithData.ID).Scan(&tmdbID, &seriesTMDBID, &dbMediaType)

	// Determine media type (movie or tv)
	mediaType := "tv" // Default to TV
	// mediaTypeKnown が false のときだけ、反対の mediaType での再試行に意味がある
	mediaTypeKnown := false
	malID := ""
	if animeWithData.MalAnimeID != nil {
		malID = *animeWithData.MalAnimeID
	}

	// Use series_tmdb_id if available (for series works)
	if err == nil && seriesTMDBID.Valid && seriesTMDBID.Int64 > 0 {
		tmdbID = int(seriesTMDBID.Int64)
		mediaType, mediaTypeKnown = s.determineMediaType(ctx, dbMediaType, animeWithData.Media, malID)
	} else if err == nil && tmdbID > 0 {
		mediaType, mediaTypeKnown = s.determineMediaType(ctx, dbMediaType, animeWithData.Media, malID)

		// series_tmdb_idが未設定の場合、tmdb_idから設定（非同期）
		if !seriesTMDBID.Valid || seriesTMDBID.Int64 == 0 {
			animeID := animeWithData.ID
			animeTitle := animeWithData.Title
			savedTmdbID := tmdbID
			s.enqueueBackgroundTask(func() {
				seriesTitle := NormalizeAnimeTitle(animeTitle)
				s.updateAnimeSeriesInfo(context.Background(), animeID, "", 0, seriesTitle, savedTmdbID)
			})
		}
	} else {
		// TMDB ID not in DB, search by MAL ID first, then by title
		// Use AniList or DB media field for hint
		mediaTypeHint := ""
		if animeWithData.Media != nil && *animeWithData.Media == "movie" {
			mediaTypeHint = "MOVIE"
		} else if malID != "" {
			format, formatErr := s.anilistClient.GetMediaFormatByMALId(ctx, malID)
			if formatErr == nil && format == "MOVIE" {
				mediaTypeHint = "MOVIE"
			}
		}

		// FindTMDBID returns the actual media type from TMDB search
		foundID, foundMediaType, searchErr := s.tmdbClient.FindTMDBID(ctx, malID, animeWithData.Title, mediaTypeHint)
		if searchErr != nil {
			s.logger.Debug("Failed to find TMDB ID for anime",
				"anime_id", animeWithData.ID,
				"mal_id", malID,
				"title", animeWithData.Title,
				"error", searchErr)
			return
		}
		tmdbID = foundID
		// Use the media type returned by FindTMDBID (this is the actual type from TMDB)
		// TMDB自身が返した型なので、反対側を試す必要はない
		mediaType = foundMediaType
		mediaTypeKnown = true

		// Save TMDB ID and series_tmdb_id to database for future use (async)
		animeID := animeWithData.ID
		animeTitle := animeWithData.Title
		s.enqueueBackgroundTask(func() {
			seriesTitle := NormalizeAnimeTitle(animeTitle)
			s.updateAnimeSeriesInfo(context.Background(), animeID, "", tmdbID, seriesTitle, tmdbID)
		})
	}

	// mediaTypeが確定している場合は反対の型を試さない。TVのIDを /movie/ に投げても
	// TMDBは正しい答えを返さず、待ち時間を増やしてレート制限を消費するだけになる。
	var providers []TMDBProvider
	if mediaTypeKnown {
		providers, err = s.tmdbClient.GetJapaneseWatchProviders(ctx, tmdbID, mediaType)
	} else {
		providers, err = s.tmdbClient.GetJapaneseWatchProvidersWithFallback(ctx, tmdbID, mediaType)
	}
	if err != nil {
		s.logger.Debug("Failed to get watch providers from TMDB",
			"anime_id", animeWithData.ID,
			"tmdb_id", tmdbID,
			"media_type", mediaType,
			"error", err)
		return
	}

	// Convert to model
	watchProviders := make([]models.WatchProvider, 0, len(providers))
	for _, p := range providers {
		watchProviders = append(watchProviders, models.WatchProvider{
			ProviderID:   p.ProviderID,
			ProviderName: p.ProviderName,
			LogoURL:      GetProviderLogoURL(p.LogoPath, "w92"),
		})
	}

	// 0件も含めてキャッシュする。配信サービスが無い作品ほどフォールバックまで
	// 走って遅くなるため、そこを繰り返さないことが効果的。
	s.cacheWatchProviders(ctx, animeWithData.ID, watchProviders)

	if len(watchProviders) == 0 {
		return
	}
	animeWithData.WatchProviders = watchProviders
}

// getCachedWatchProviders は作品の配信サービスをRedisから取得する。
// 「0件」も有効なキャッシュ値として扱うため、戻り値のsliceの長さではなく
// 2つ目の戻り値でヒット有無を判定する。
func (s *AnimeService) getCachedWatchProviders(ctx context.Context, animeID int) ([]models.WatchProvider, bool) {
	if s.redisCache == nil {
		return nil, false
	}
	var cached []models.WatchProvider
	found, err := s.redisCache.Get(ctx, watchProvidersCacheKey(animeID), &cached)
	if err != nil {
		s.logger.Debug("Failed to read watch providers from cache", "anime_id", animeID, "error", err)
		return nil, false
	}
	return cached, found
}

func (s *AnimeService) cacheWatchProviders(ctx context.Context, animeID int, providers []models.WatchProvider) {
	if s.redisCache == nil {
		return
	}
	if err := s.redisCache.Set(ctx, watchProvidersCacheKey(animeID), providers, constants.WatchProvidersCacheTTL); err != nil {
		s.logger.Debug("Failed to cache watch providers", "anime_id", animeID, "error", err)
	}
}

func watchProvidersCacheKey(animeID int) string {
	return fmt.Sprintf("%s%d", constants.CacheKeyWatchProviders, animeID)
}

// isMovieDurationKnownMissing は、直近のJikan問い合わせで上映時間を取得できなかった
// 作品かどうかを返す。キーの有無だけが意味を持つため、値は使わない。
func (s *AnimeService) isMovieDurationKnownMissing(ctx context.Context, animeID int) bool {
	if s.redisCache == nil {
		return false
	}
	var placeholder bool
	found, err := s.redisCache.Get(ctx, movieDurationMissingCacheKey(animeID), &placeholder)
	if err != nil {
		s.logger.Debug("Failed to read movie duration cache", "anime_id", animeID, "error", err)
		return false
	}
	return found
}

func (s *AnimeService) markMovieDurationMissing(ctx context.Context, animeID int, ttl time.Duration) {
	if s.redisCache == nil {
		return
	}
	if err := s.redisCache.Set(ctx, movieDurationMissingCacheKey(animeID), true, ttl); err != nil {
		s.logger.Debug("Failed to cache movie duration miss", "anime_id", animeID, "error", err)
	}
}

func movieDurationMissingCacheKey(animeID int) string {
	return fmt.Sprintf("%s%d", constants.CacheKeyMovieDurationMissing, animeID)
}
