package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"myanimelogs/internal/constants"
	"myanimelogs/internal/models"
	"myanimelogs/internal/utils"
)

// GetCurrentSeasonAnimes returns animes for the current season from the database
func (s *AnimeService) GetCurrentSeasonAnimes(limit int) ([]models.Anime, error) {
	if limit <= 0 {
		limit = 12
	}

	// Get current season string (e.g., "2026-winter")
	currentSeason := utils.SeasonForTime(time.Now())

	// キャッシュキーを生成
	cacheKey := fmt.Sprintf("%s%s:%d", constants.CacheKeyCurrentSeason, currentSeason, limit)

	// キャッシュから取得を試みる
	var cachedAnimes []models.Anime
	if found, err := s.cache.Get(cacheKey, &cachedAnimes); found && err == nil {
		s.logger.Debug("Cache hit for current season animes", "season", currentSeason, "limit", limit)
		return cachedAnimes, nil
	}

	// DBから取得（VIEWを使用してステータスを動的に計算）
	animes, err := s.getAnimesByQuery(
		`SELECT `+animeSelectColumns+`
		FROM animes_with_computed_status
		WHERE season = $1
		ORDER BY watchers_count DESC NULLS LAST, title
		LIMIT $2`,
		[]interface{}{currentSeason, limit},
		"current season animes",
	)
	if err != nil {
		return nil, err
	}

	// キャッシュに保存
	if cacheErr := s.cache.SetWithTTL(cacheKey, animes, constants.CurrentSeasonAnimeCacheTTL); cacheErr != nil {
		s.logger.Warn("Failed to cache current season animes", "error", cacheErr)
	}

	return animes, nil
}

// GetPopularAnimes returns popular animes from the database (sorted by watchers_count)
func (s *AnimeService) GetPopularAnimes(limit int) ([]models.Anime, error) {
	if limit <= 0 {
		limit = 12
	}

	// キャッシュキーを生成
	cacheKey := fmt.Sprintf("%s%d", constants.CacheKeyPopular, limit)

	// キャッシュから取得を試みる
	var cachedAnimes []models.Anime
	if found, err := s.cache.Get(cacheKey, &cachedAnimes); found && err == nil {
		s.logger.Debug("Cache hit for popular animes", "limit", limit)
		return cachedAnimes, nil
	}

	// DBから取得（VIEWを使用してステータスを動的に計算）
	animes, err := s.getAnimesByQuery(
		`SELECT `+animeSelectColumns+`
		FROM animes_with_computed_status
		WHERE watchers_count > 0
		ORDER BY watchers_count DESC, title
		LIMIT $1`,
		[]interface{}{limit},
		"popular animes",
	)
	if err != nil {
		return nil, err
	}

	// キャッシュに保存
	if cacheErr := s.cache.SetWithTTL(cacheKey, animes, constants.PopularAnimeCacheTTL); cacheErr != nil {
		s.logger.Warn("Failed to cache popular animes", "error", cacheErr)
	}

	return animes, nil
}

// getAnimesByQuery is a helper function to execute anime queries with proper error handling
// scanAnime - rows.Scanの共通ヘルパー関数
// カラム追加時の修正漏れを防ぐため、スキャン処理を一箇所に集約
func scanAnime(rows *sql.Rows) (models.Anime, error) {
	var anime models.Anime
	err := rows.Scan(
		&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status,
		&anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount,
		&anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium,
		&anime.ReleasedOn, &anime.Media,
	)
	return anime, err
}

func (s *AnimeService) getAnimesByQuery(query string, args []interface{}, queryName string) ([]models.Anime, error) {
	animes, err := queryAnimes(s.db, s.logger, queryName, query, args, scanAnime)
	if err != nil {
		return nil, err
	}

	// DBキャッシュがないアニメのみ外部APIから画像を取得
	s.enrichAnimesWithCoverImagesIfNeeded(animes)

	return animes, nil
}

func (s *AnimeService) GetAnimeByID(animeID int, userID *int) (*models.AnimeWithUserData, error) {
	var anime models.Anime
	var description, media, officialSiteURL, titleKana, titleEn, releasedOn, seriesTitle sql.NullString
	var broadcastDay, broadcastTime, broadcastTimezone sql.NullString
	var seriesTMDBID, durationMinutes sql.NullInt64
	var coverImageUpdatedAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, title, image_url, season, computed_status as status,
		       mal_anime_id, annict_id, created_at, updated_at, watchers_count,
		       cover_image_extra_large, cover_image_large, cover_image_medium,
		       cover_image_updated_at,
		       description, media, released_on, official_site_url, title_kana, title_en,
		       series_title, series_tmdb_id,
		       broadcast_day, broadcast_time, broadcast_timezone, duration_minutes
		FROM animes_with_computed_status WHERE id = $1
	`, animeID).Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount, &anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium,
		&coverImageUpdatedAt,
		&description, &media, &releasedOn, &officialSiteURL, &titleKana, &titleEn, &seriesTitle, &seriesTMDBID,
		&broadcastDay, &broadcastTime, &broadcastTimezone, &durationMinutes)

	// Check error immediately after Scan - accessing scanned variables before error check is undefined behavior
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAnimeNotFound
		}
		return nil, fmt.Errorf("failed to get anime: %w", err)
	}

	// Set nullable fields (safe to access after error check)
	if description.Valid {
		anime.Description = &description.String
	}
	if releasedOn.Valid {
		anime.ReleasedOn = &releasedOn.String
	}
	if media.Valid {
		anime.Media = &media.String
	}
	if officialSiteURL.Valid {
		anime.OfficialSiteURL = &officialSiteURL.String
	}
	if titleKana.Valid {
		anime.TitleKana = &titleKana.String
	}
	if titleEn.Valid {
		anime.TitleEn = &titleEn.String
	}
	if seriesTitle.Valid {
		anime.SeriesTitle = &seriesTitle.String
	}
	if seriesTMDBID.Valid {
		id := int(seriesTMDBID.Int64)
		anime.SeriesTMDBID = &id
	}
	if broadcastDay.Valid {
		anime.BroadcastDay = &broadcastDay.String
	}
	if broadcastTime.Valid {
		anime.BroadcastTime = &broadcastTime.String
	}
	if broadcastTimezone.Valid {
		anime.BroadcastTimezone = &broadcastTimezone.String
	}
	if durationMinutes.Valid {
		duration := int(durationMinutes.Int64)
		anime.DurationMinutes = &duration
	}
	if coverImageUpdatedAt.Valid {
		anime.CoverImageUpdatedAt = &coverImageUpdatedAt.Time
	}

	// カバー画像が未取得、またはTTL期限切れの場合のみAniListから取得
	if anime.CoverImageExtraLarge == "" || anime.CoverImageLarge == "" ||
		anime.CoverImageUpdatedAt == nil || time.Since(*anime.CoverImageUpdatedAt) > constants.CoverImageCacheTTL {
		s.enrichAnimeWithCoverImage(&anime)
	}

	// descriptionがない場合はAniList APIから取得
	if anime.Description == nil || *anime.Description == "" {
		s.enrichAnimeWithDescription(&anime)
	}

	// DBキャッシュ済みのあらすじも含め、表示時に書式を統一する（バックフィル前の既存データ対策）。
	// sanitizeDescription は冪等なため、enrich 済みの値に再適用しても問題ない。
	if anime.Description != nil {
		sanitized := sanitizeDescription(*anime.Description)
		anime.Description = &sanitized
	}

	// series_titleがない場合は設定（シリーズフォールバック用）
	s.ensureSeriesTitle(&anime)

	// released_onがない場合はJikan APIから取得
	if anime.ReleasedOn == nil || *anime.ReleasedOn == "" {
		s.enrichAnimeWithReleasedOn(&anime)
	}

	// released_on取得後にステータスを補正（VIEWはDB時点のreleased_onで計算するため、
	// enrichAnimeWithReleasedOnでオンデマンド取得した場合はstatusが未反映）
	if anime.ReleasedOn != nil && *anime.ReleasedOn != "" {
		releaseDate, parseErr := time.ParseInLocation("2006-01-02", *anime.ReleasedOn, time.UTC)
		if parseErr == nil {
			today := time.Now().UTC().Truncate(24 * time.Hour)
			if releaseDate.After(today) {
				upcomingStatus := constants.AnimeStatusUpcoming.String()
				anime.Status = &upcomingStatus
			} else if anime.Status != nil && (*anime.Status == constants.AnimeStatusUpcoming.String() || *anime.Status == "not_yet_aired") {
				// released_onが今日以前なのにupcoming/not_yet_airedの場合はairingに補正
				airingStatus := constants.AnimeStatusAiring.String()
				anime.Status = &airingStatus
			}
		}
	}

	// 放送情報がない場合はJikan APIから取得
	if anime.BroadcastDay == nil {
		s.enrichAnimeWithBroadcastInfo(&anime)
	}

	// 映画の上映時間がない場合はJikan APIから取得
	if anime.Media != nil && *anime.Media == "movie" && anime.DurationMinutes == nil {
		s.enrichAnimeWithMovieDuration(&anime)
	}

	animeWithData := &models.AnimeWithUserData{Anime: anime}

	// Get genres from anime_genres table
	genreRows, err := s.db.Query(`
		SELECT genre FROM anime_genres WHERE anime_id = $1 ORDER BY genre
	`, animeID)
	if err != nil {
		s.logger.Warn("Failed to get genres", "anime_id", animeID, "error", err)
	} else {
		defer genreRows.Close()
		var genres []string
		for genreRows.Next() {
			var genre string
			if err := genreRows.Scan(&genre); err != nil {
				s.logger.Warn("Failed to scan genre", "anime_id", animeID, "error", err)
				continue
			}
			genres = append(genres, genre)
		}
		if len(genres) > 0 {
			animeWithData.Genres = genres
		}
	}

	// Add user data if authenticated
	if userID != nil {
		// Get user favorite (heart)
		var favorite models.UserFavorite
		err := s.db.QueryRow(`
			SELECT id, user_id, anime_id, created_at, updated_at
			FROM user_favorites WHERE user_id = $1 AND anime_id = $2
		`, *userID, animeID).Scan(&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.CreatedAt, &favorite.UpdatedAt)
		if err != nil && err != sql.ErrNoRows {
			// Log error but don't fail the entire request
			s.logger.Warn("Failed to get user favorite",
				"user_id", *userID,
				"anime_id", animeID,
				"error", err)
		} else if err == nil {
			animeWithData.UserFavorite = &favorite
		}

		// Get user watch status
		var watchStatus models.UserWatchStatus
		err = s.db.QueryRow(`
			SELECT id, user_id, anime_id, status, created_at, updated_at
			FROM user_watch_status WHERE user_id = $1 AND anime_id = $2
		`, *userID, animeID).Scan(&watchStatus.ID, &watchStatus.UserID, &watchStatus.AnimeID, &watchStatus.Status, &watchStatus.CreatedAt, &watchStatus.UpdatedAt)
		if err != nil && err != sql.ErrNoRows {
			s.logger.Warn("Failed to get user watch status",
				"user_id", *userID,
				"anime_id", animeID,
				"error", err)
		} else if err == nil {
			animeWithData.UserWatchStatus = &watchStatus
		}
	}

	// Get watch providers from TMDB (if TMDB ID is available or can be found)
	s.enrichAnimeWithWatchProviders(animeWithData)

	// 同シリーズの他シーズン（続編・前作）。series_titleはensureSeriesTitleで
	// 必ず設定済みのため、常に問い合わせる。
	animeWithData.RelatedSeasons = s.getRelatedSeasons(animeID, anime.Title, anime.SeriesTitle, anime.SeriesTMDBID)

	return animeWithData, nil
}

// getRelatedSeasons - 以下3条件のいずれかを満たす他のアニメ（自分自身を除く）を
// 放送時期順で返す。3条件はOR結合（マージ）され、重複はidの一意性により自然に排除される。
//
//  1. series_title の完全一致（従来ロジック）
//  2. series_tmdb_id の一致（自身のseries_tmdb_idが非nilの場合のみ）
//  3. タイトルの双方向前方一致（双方の文字数がMinRelatedTitlePrefixLength以上の場合のみ）
//
// 条件3はseries_titleの正規化パターンが続編タイトルの命名規則（数字サフィックスや
// サブタイトル型）を網羅しきれず、1・2でも拾えない関連作品を補完する目的。
// 短い文字数のタイトル同士（「F」「K」等）が無関係な作品を大量に巻き込むことが
// 実データ調査で確認されたため、文字数ガードで除外する。
func (s *AnimeService) getRelatedSeasons(animeID int, title string, seriesTitle *string, seriesTMDBID *int) []models.RelatedAnime {
	if (seriesTitle == nil || *seriesTitle == "") && seriesTMDBID == nil && utf8.RuneCountInString(title) < constants.MinRelatedTitlePrefixLength {
		return nil
	}

	// seasonは "YYYY-season" 形式の文字列のため、単純な文字列ソートだと
	// 同一年内で "autumn" < "winter" のように字句順とチクロ順がずれる
	// （例: "2023-autumn" が "2023-winter" より前に来てしまう）。
	// 年 → 季節（冬→春→夏→秋）の順で明示的に並べる。
	rows, err := s.db.Query(`
		SELECT id, title, season, cover_image_extra_large, cover_image_large, cover_image_medium, image_url
		FROM animes
		WHERE id != $2
		AND (
			($1::text IS NOT NULL AND series_title = $1)
			OR ($3::int IS NOT NULL AND series_tmdb_id = $3)
			OR (
				length($4) >= $5 AND length(title) >= $5
				AND (title LIKE $4 || '%' OR $4 LIKE title || '%')
			)
		)
		ORDER BY
			split_part(season, '-', 1) NULLS LAST,
			CASE split_part(season, '-', 2)
				WHEN 'winter' THEN 1
				WHEN 'spring' THEN 2
				WHEN 'summer' THEN 3
				WHEN 'autumn' THEN 4
				ELSE 5
			END,
			id
		LIMIT $6
	`, seriesTitle, animeID, seriesTMDBID, title, constants.MinRelatedTitlePrefixLength, constants.MaxRelatedSeasons)
	if err != nil {
		s.logger.Warn("Failed to get related seasons", "anime_id", animeID, "error", err)
		return nil
	}
	defer rows.Close()

	var related []models.RelatedAnime
	for rows.Next() {
		var r models.RelatedAnime
		var coverImageExtraLarge, coverImageLarge, coverImageMedium sql.NullString
		if err := rows.Scan(&r.ID, &r.Title, &r.Season, &coverImageExtraLarge, &coverImageLarge, &coverImageMedium, &r.ImageURL); err != nil {
			s.logger.Warn("Failed to scan related season", "anime_id", animeID, "error", err)
			continue
		}
		if coverImageExtraLarge.Valid {
			r.CoverImageExtraLarge = coverImageExtraLarge.String
		}
		if coverImageLarge.Valid {
			r.CoverImageLarge = coverImageLarge.String
		}
		if coverImageMedium.Valid {
			r.CoverImageMedium = coverImageMedium.String
		}
		related = append(related, r)
	}
	return related
}

// GetFilterOptions - フィルター情報を取得
func (s *AnimeService) GetFilterOptions() (*models.FilterOptions, error) {
	// Redisキャッシュから取得を試みる
	if s.redisCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var cachedOptions models.FilterOptions
		if found, err := s.redisCache.Get(ctx, constants.CacheKeyFilterOptions, &cachedOptions); found && err == nil {
			s.logger.Debug("Redis cache hit for filter options")
			// Redisから取得したデータでメモリキャッシュも更新（後続リクエストの高速化）
			if cacheErr := s.cache.SetWithTTL(constants.CacheKeyFilterOptions, &cachedOptions, constants.FilterOptionsCacheTTL); cacheErr != nil {
				s.logger.Warn("Failed to update memory cache from Redis", "error", cacheErr)
			}
			return &cachedOptions, nil
		}
	}

	// メモリキャッシュから取得を試みる
	var cachedOptions models.FilterOptions
	if found, err := s.cache.Get(constants.CacheKeyFilterOptions, &cachedOptions); found && err == nil {
		s.logger.Debug("Memory cache hit for filter options")
		return &cachedOptions, nil
	}

	filterOptions := &models.FilterOptions{
		Genres:  []models.GenreInfo{},
		Years:   []int{},
		Seasons: []string{},
	}

	// ジャンル一覧を取得（エラーがあっても他の情報は取得を続行）
	genreRows, err := s.db.Query(`
		SELECT genre, COUNT(*) as count
		FROM anime_genres
		WHERE genre IS NOT NULL AND genre != ''
		GROUP BY genre
		ORDER BY count DESC, genre ASC
	`)
	if err != nil {
		s.logger.Error("Failed to get genres", "error", err)
	} else {
		defer genreRows.Close()
		genres := []models.GenreInfo{}
		for genreRows.Next() {
			var genre models.GenreInfo
			err := genreRows.Scan(&genre.Name, &genre.Count)
			if err != nil {
				s.logger.Error("Failed to scan genre", "error", err)
				continue
			}
			if strings.TrimSpace(genre.Name) != "" {
				genres = append(genres, genre)
			}
		}
		filterOptions.Genres = genres
	}

	// 放送年一覧を取得（season から年を抽出）
	yearRows, err := s.db.Query(`
		SELECT DISTINCT CAST(SUBSTRING(season FROM 1 FOR 4) AS INTEGER) as year
		FROM animes
		WHERE season IS NOT NULL 
		  AND season ~ '^\d{4}-'
		  AND CAST(SUBSTRING(season FROM 1 FOR 4) AS INTEGER) > 1900
		  AND CAST(SUBSTRING(season FROM 1 FOR 4) AS INTEGER) <= EXTRACT(YEAR FROM CURRENT_DATE) + 1
		ORDER BY year DESC
	`)
	if err != nil {
		s.logger.Error("Failed to get years", "error", err)
	} else {
		defer yearRows.Close()
		years := []int{}
		for yearRows.Next() {
			var year int
			err := yearRows.Scan(&year)
			if err != nil {
				s.logger.Error("Failed to scan year", "error", err)
				continue
			}
			years = append(years, year)
		}
		filterOptions.Years = years
	}

	// シーズン一覧を取得
	seasonRows, err := s.db.Query(`
		SELECT season
		FROM animes
		WHERE season IS NOT NULL AND season != ''
		GROUP BY season
		ORDER BY 
			CASE season
				WHEN $1 THEN 1
				WHEN $2 THEN 2
				WHEN $3 THEN 3
				WHEN $4 THEN 4
				ELSE 5
			END
	`, constants.SeasonSpring, constants.SeasonSummer, constants.SeasonAutumn, constants.SeasonWinter)
	if err != nil {
		s.logger.Error("Failed to get seasons", "error", err)
	} else {
		defer seasonRows.Close()
		seasons := []string{}
		for seasonRows.Next() {
			var season string
			err := seasonRows.Scan(&season)
			if err != nil {
				s.logger.Error("Failed to scan season", "error", err)
				continue
			}
			if strings.TrimSpace(season) != "" {
				seasons = append(seasons, season)
			}
		}
		filterOptions.Seasons = seasons
	}

	// Redisキャッシュに保存
	if s.redisCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := s.redisCache.Set(ctx, constants.CacheKeyFilterOptions, filterOptions, constants.FilterOptionsCacheTTL); err != nil {
			s.logger.Warn("Failed to cache filter options in Redis", "error", err)
		} else {
			s.logger.Debug("Cached filter options in Redis")
		}
	}

	// メモリキャッシュにも保存
	if cacheErr := s.cache.SetWithTTL(constants.CacheKeyFilterOptions, filterOptions, constants.FilterOptionsCacheTTL); cacheErr != nil {
		s.logger.Warn("Failed to cache filter options", "error", cacheErr)
	}

	return filterOptions, nil
}
