package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/constants"
	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
	"myanimelogs/internal/utils"
)

var (
	ErrAnimeNotFound         = errors.New("anime not found")
	ErrEpisodeNotFound       = errors.New("episode not found")
	ErrSearchHistoryNotFound = errors.New("search history item not found")
)

// GetCurrentSeasonAnimes returns animes for the current season from the database
func (s *AnimeService) GetCurrentSeasonAnimes(limit int) ([]models.Anime, error) {
	if limit <= 0 {
		limit = 12
	}

	// Get current season string (e.g., "2026-winter")
	currentSeason := getSeasonStringFromTime(time.Now())

	// キャッシュキーを生成
	cacheKey := fmt.Sprintf("%s%s:%d", constants.CacheKeyCurrentSeason, currentSeason, limit)

	// キャッシュから取得を試みる
	var cachedAnimes []models.Anime
	if found, err := s.cache.Get(cacheKey, &cachedAnimes); found && err == nil {
		s.logger.Debug("Cache hit for current season animes", "season", currentSeason, "limit", limit)
		return cachedAnimes, nil
	}

	// DBから取得
	animes, err := s.getAnimesByQuery(
		`SELECT id, title, image_url, season,
		        CASE 
		            WHEN released_on IS NOT NULL AND released_on != '' AND released_on::date > CURRENT_DATE THEN 'upcoming'
		            WHEN released_on IS NOT NULL AND released_on != '' AND released_on::date <= CURRENT_DATE AND status = 'not_yet_aired' THEN 'airing'
		            ELSE status
		        END as status,
		        mal_anime_id, annict_id, created_at, updated_at, watchers_count,
		        COALESCE(cover_image_extra_large, '') as cover_image_extra_large,
		        COALESCE(cover_image_large, '') as cover_image_large,
		        COALESCE(cover_image_medium, '') as cover_image_medium,
		        released_on
		FROM animes
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

	// DBから取得
	animes, err := s.getAnimesByQuery(
		`SELECT id, title, image_url, season,
		        CASE 
		            WHEN released_on IS NOT NULL AND released_on != '' AND released_on::date > CURRENT_DATE THEN 'upcoming'
		            WHEN released_on IS NOT NULL AND released_on != '' AND released_on::date <= CURRENT_DATE AND status = 'not_yet_aired' THEN 'airing'
		            ELSE status
		        END as status,
		        mal_anime_id, annict_id, created_at, updated_at, watchers_count,
		        COALESCE(cover_image_extra_large, '') as cover_image_extra_large,
		        COALESCE(cover_image_large, '') as cover_image_large,
		        COALESCE(cover_image_medium, '') as cover_image_medium,
		        released_on
		FROM animes
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
func (s *AnimeService) getAnimesByQuery(query string, args []interface{}, queryName string) ([]models.Anime, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", queryName, err)
	}
	defer rows.Close()

	animes := []models.Anime{}
	for rows.Next() {
		var anime models.Anime
		err := rows.Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount, &anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium, &anime.ReleasedOn)
		if err != nil {
			s.logger.Warn("Failed to scan anime row", "query", queryName, "error", err)
			continue
		}
		animes = append(animes, anime)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating anime rows", "query", queryName, "error", err)
		return nil, fmt.Errorf("error iterating %s: %w", queryName, err)
	}

	// DBキャッシュがないアニメのみ外部APIから画像を取得
	s.enrichAnimesWithCoverImagesIfNeeded(animes)

	return animes, nil
}

// getSeasonStringFromTime returns the season string in format "YYYY-season" for a given time
// This function accepts time.Time as a parameter to improve testability
func getSeasonStringFromTime(t time.Time) string {
	year := t.Year()
	month := t.Month()

	var season string
	switch {
	case month >= 1 && month <= 3:
		season = "winter"
	case month >= 4 && month <= 6:
		season = "spring"
	case month >= 7 && month <= 9:
		season = "summer"
	default:
		season = "autumn"
	}

	return fmt.Sprintf("%d-%s", year, season)
}

type AnimeService struct {
	db                  *database.DB
	logger              *common.Logger
	anilistClient       *AniListClient
	jikanClient         *JikanClient
	tmdbClient          *TMDBClient  // TMDB APIクライアント
	backgroundTaskQueue chan func()  // バックグラウンドタスク用ワーカープールのキュー
	cache               *MemoryCache // レスポンスキャッシュ
	redisCache          *RedisCache  // Redisキャッシュ（オプション）
}

func NewAnimeService(db *database.DB, logger *common.Logger) *AnimeService {
	// キャッシュを初期化
	cacheConfig := CacheConfig{
		DefaultTTL:      constants.SearchResultCacheTTL,
		CleanupInterval: constants.CacheCleanupInterval,
		MaxItems:        constants.CacheMaxItems,
	}

	// Redisキャッシュを初期化（オプション - 接続失敗時はnilのまま）
	var redisCache *RedisCache
	if rc, err := NewRedisCache(logger); err != nil {
		logger.Warn("Redis cache not available, using memory cache only", "error", err)
	} else {
		redisCache = rc
	}

	s := &AnimeService{
		db:                  db,
		logger:              logger,
		anilistClient:       NewAniListClient(),
		jikanClient:         NewJikanClient(logger),
		tmdbClient:          NewTMDBClient(logger),
		backgroundTaskQueue: make(chan func(), constants.BackgroundTaskQueueSize),
		cache:               NewMemoryCache(cacheConfig),
		redisCache:          redisCache,
	}

	// ワーカープールを起動
	s.startBackgroundTaskWorkers()

	return s
}

// startBackgroundTaskWorkers - バックグラウンドタスク用ワーカープールを起動
// 高負荷時のgoroutine爆発を防ぐため、固定数のワーカーでタスクを処理
func (s *AnimeService) startBackgroundTaskWorkers() {
	for i := 0; i < constants.BackgroundTaskWorkerPoolSize; i++ {
		go func() {
			for task := range s.backgroundTaskQueue {
				// パニックからの復帰（アプリケーション全体のクラッシュを防止）
				func() {
					defer func() {
						if r := recover(); r != nil {
							s.logger.Error("Panic recovered in image update worker",
								"panic", r)
						}
					}()
					task()
				}()
			}
		}()
	}
}

// enqueueBackgroundTask - バックグラウンドタスクをキューに追加
// キューが満杯の場合はタスクをドロップ（ノンブロッキング）
func (s *AnimeService) enqueueBackgroundTask(task func()) {
	select {
	case s.backgroundTaskQueue <- task:
		// タスクをキューに追加成功
	default:
		// キューが満杯の場合はドロップ（次回リクエスト時に再試行される）
		s.logger.Debug("Background task queue is full, task dropped")
	}
}

// Close - AnimeServiceのリソースを解放
// ワーカープールを graceful にシャットダウンし、goroutineリークを防止
func (s *AnimeService) Close() {
	if s.backgroundTaskQueue != nil {
		close(s.backgroundTaskQueue)
	}
	if s.cache != nil {
		s.cache.Close()
	}
	if s.redisCache != nil {
		s.redisCache.Close()
	}
}

// ClearAnimeCache - アニメ関連のキャッシュをクリア
// データ同期後に呼び出して、最新データを反映させる
// メモリキャッシュとRedisキャッシュを並行してクリアする
func (s *AnimeService) ClearAnimeCache() {
	s.logger.Info("Clearing anime cache after data sync")

	// クリア対象のキー（完全一致）
	exactKeys := []string{
		constants.CacheKeyFilterOptions,
	}

	// クリア対象のプレフィックス
	prefixKeys := []string{
		constants.CacheKeyPopularSearches,
		constants.CacheKeyCurrentSeason,
		constants.CacheKeyPopular,
		constants.CacheKeySearch,
		constants.CacheKeyAdvSearch,
	}

	var wg sync.WaitGroup

	// メモリキャッシュをクリア（goroutine）
	if s.cache != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, key := range exactKeys {
				s.cache.Delete(key)
			}
			for _, prefix := range prefixKeys {
				s.cache.DeleteByPrefix(prefix)
			}
			s.logger.Debug("Memory cache cleared")
		}()
	}

	// Redisキャッシュをクリア（goroutine）
	if s.redisCache != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			for _, key := range exactKeys {
				if err := s.redisCache.Delete(ctx, key); err != nil {
					s.logger.Warn("Failed to clear key from Redis", "key", key, "error", err)
				}
			}
			for _, prefix := range prefixKeys {
				if err := s.redisCache.DeleteByPrefix(ctx, prefix); err != nil {
					s.logger.Warn("Failed to clear prefix from Redis", "prefix", prefix, "error", err)
				}
			}
			s.logger.Debug("Redis cache cleared")
		}()
	}

	// 両方の完了を待つ
	wg.Wait()

	s.logger.Info("Anime cache cleared successfully")
}

// GetEpisodeSyncStatus returns the sync status for a specific anime's episodes
func (s *AnimeService) GetEpisodeSyncStatus(animeID int) (*models.EpisodeSyncStatus, error) {
	// Get anime info and episode statistics in a single query
	var animeTitle string
	var annictID *int
	var totalEpisodes, defaultTitles, customTitles int

	err := s.db.QueryRow(`
		SELECT 
			a.title,
			a.annict_id,
			COUNT(e.id) as total,
			COUNT(CASE WHEN e.title LIKE '第%話' THEN 1 END) as default_titles,
			COUNT(CASE WHEN e.title NOT LIKE '第%話' THEN 1 END) as custom_titles
		FROM animes a
		LEFT JOIN episodes e ON a.id = e.anime_id
		WHERE a.id = $1
		GROUP BY a.id, a.title, a.annict_id
	`, animeID).Scan(&animeTitle, &annictID, &totalEpisodes, &defaultTitles, &customTitles)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAnimeNotFound
		}
		return nil, fmt.Errorf("failed to get anime info and episode statistics: %w", err)
	}

	// Get sample episodes
	rows, err := s.db.Query(`
		SELECT episode_number, title
		FROM episodes
		WHERE anime_id = $1
		ORDER BY episode_number
		LIMIT 5
	`, animeID)

	if err != nil {
		return nil, fmt.Errorf("failed to get sample episodes: %w", err)
	}
	defer rows.Close()

	var sampleEpisodes []string
	for rows.Next() {
		var episodeNumber int
		var title string
		if err := rows.Scan(&episodeNumber, &title); err != nil {
			s.logger.Warn("Failed to scan sample episode", "anime_id", animeID, "error", err)
			continue
		}
		sampleEpisodes = append(sampleEpisodes, fmt.Sprintf("Ep%d: %s", episodeNumber, title))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sample episodes: %w", err)
	}

	syncPercentage := 0.0
	if totalEpisodes > 0 {
		syncPercentage = float64(customTitles) / float64(totalEpisodes) * 100
	}

	return &models.EpisodeSyncStatus{
		AnimeID:        animeID,
		AnimeTitle:     animeTitle,
		AnnictID:       annictID,
		TotalEpisodes:  totalEpisodes,
		DefaultTitles:  defaultTitles,
		CustomTitles:   customTitles,
		SyncPercentage: syncPercentage,
		NeedSync:       defaultTitles > 0,
		SampleEpisodes: sampleEpisodes,
	}, nil
}

// buildKeywordSearchClause - キーワード検索用のWHERE句を構築するヘルパー関数
// tableAlias: テーブルエイリアス（例: "a." または ""）
// 戻り値: WHERE句の条件文字列、引数スライス、次の引数インデックス
// 注: PostgreSQLのFTS（全文検索）は日本語（カタカナ・ひらがな・漢字）を適切にトークン化できないため、
// ILIKE検索のみを使用する。pg_trgmインデックスによりパフォーマンスは維持される。
// また、title_kanaはひらがなで保存されているため、ひらがな変換後のキーワードで検索する。
func buildKeywordSearchClause(keyword string, argIndex int, tableAlias string) (string, []interface{}, int) {
	if keyword == "" {
		return "", nil, argIndex
	}

	likePattern := "%" + keyword + "%"
	hiraganaKeyword := utils.KatakanaToHiragana(keyword)
	katakanaKeyword := utils.HiraganaToKatakana(keyword)

	// キーワードにカタカナもひらがなも含まれない場合（漢字・英語のみ）
	// → 変換しても同じなので引数を1つに最適化
	if keyword == hiraganaKeyword && keyword == katakanaKeyword {
		clause := fmt.Sprintf(`
			(%stitle ILIKE $%d 
			 OR %stitle_kana ILIKE $%d
			 OR %stitle_en ILIKE $%d)
		`, tableAlias, argIndex,
			tableAlias, argIndex,
			tableAlias, argIndex)

		return clause, []interface{}{likePattern}, argIndex + 1
	}

	// カタカナまたはひらがなが含まれる場合
	// - title: 元のキーワード + カタカナ変換版で検索（ひらがな入力でもカタカナタイトルにヒット）
	// - title_kana: ひらがな変換版で検索（カタカナ入力でもひらがなのtitle_kanaにヒット）
	// - title_en: 元のキーワードで検索
	hiraganaPattern := "%" + hiraganaKeyword + "%"
	katakanaPattern := "%" + katakanaKeyword + "%"

	clause := fmt.Sprintf(`
		(%stitle ILIKE $%d 
		 OR %stitle ILIKE $%d
		 OR %stitle_kana ILIKE $%d
		 OR %stitle_en ILIKE $%d)
	`, tableAlias, argIndex,
		tableAlias, argIndex+1,
		tableAlias, argIndex+2,
		tableAlias, argIndex)

	return clause, []interface{}{likePattern, katakanaPattern, hiraganaPattern}, argIndex + 3
}

// generateSearchCacheKey - 検索パラメータからキャッシュキーを生成
func (s *AnimeService) generateSearchCacheKey(params models.AnimeSearchParams) string {
	return fmt.Sprintf("%skw:%s:s:%s:st:%s:g:%s:so:%s:p:%d:pp:%d",
		constants.CacheKeySearch,
		params.Keyword, params.Season, params.Status, params.Genre, params.Sort,
		params.Page, params.PerPage)
}

func (s *AnimeService) SearchAnimes(params models.AnimeSearchParams, userID *int) (*models.PaginatedResponse, error) {
	// 未認証ユーザーの場合のみキャッシュを使用
	// （認証ユーザーはお気に入り・視聴進捗などユーザー固有データがあるため）
	canUseCache := userID == nil
	cacheKey := s.generateSearchCacheKey(params)

	if canUseCache {
		var cachedResponse models.PaginatedResponse
		if found, err := s.cache.Get(cacheKey, &cachedResponse); found && err == nil {
			s.logger.Debug("Cache hit for search", "keyword", params.Keyword, "page", params.Page)
			return &cachedResponse, nil
		}
	}

	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	// キーワード検索（ヘルパー関数を使用）
	if params.Keyword != "" {
		clause, keywordArgs, newArgIndex := buildKeywordSearchClause(params.Keyword, argIndex, "")
		whereClauses = append(whereClauses, clause)
		args = append(args, keywordArgs...)
		argIndex = newArgIndex
	}

	if params.Season != "" {
		dbSeason := utils.ConvertSeasonToDBFormat(params.Season)
		if dbSeason != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("season = $%d", argIndex))
			args = append(args, dbSeason)
			argIndex++
		}
	}

	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	if params.Genre != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("id IN (SELECT anime_id FROM anime_genres WHERE genre = $%d)", argIndex))
		args = append(args, params.Genre)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY title"
	switch params.Sort {
	case "rating":
		orderBy = "ORDER BY (SELECT AVG(rating) FROM user_ratings WHERE anime_id = animes.id) DESC NULLS LAST"
	case "year":
		orderBy = "ORDER BY season DESC"
	}

	// Count total results
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM animes %s", whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count animes: %w", err)
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PerPage
	totalPages := (total + params.PerPage - 1) / params.PerPage

	// Get animes
	query := fmt.Sprintf(`
		SELECT id, title, image_url, season,
		       CASE 
		           WHEN released_on IS NOT NULL AND released_on != '' AND released_on::date > CURRENT_DATE THEN 'upcoming'
		           WHEN released_on IS NOT NULL AND released_on != '' AND released_on::date <= CURRENT_DATE AND status = 'not_yet_aired' THEN 'airing'
		           ELSE status
		       END as status,
		       mal_anime_id, annict_id, created_at, updated_at, watchers_count,
		       COALESCE(cover_image_extra_large, '') as cover_image_extra_large,
		       COALESCE(cover_image_large, '') as cover_image_large,
		       COALESCE(cover_image_medium, '') as cover_image_medium
		FROM animes %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, params.PerPage, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search animes: %w", err)
	}
	defer rows.Close()

	animes := []models.AnimeWithUserData{}
	for rows.Next() {
		var anime models.Anime
		err := rows.Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount, &anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium)
		if err != nil {
			continue
		}

		animeWithData := models.AnimeWithUserData{Anime: anime}
		animes = append(animes, animeWithData)
	}

	// ユーザーデータを一括取得（N+1クエリ解消）
	if userID != nil {
		s.enrichAnimesWithUserData(animes, *userID)
	}

	// DBキャッシュがないアニメのみ外部APIから画像を取得
	s.enrichAnimesWithUserDataCoverImagesIfNeeded(animes)

	// 検索ログを記録
	s.logBasicSearch(params, userID, total)

	response := &models.PaginatedResponse{
		Data: animes,
		Meta: models.PaginationMeta{
			TotalPages:  totalPages,
			CurrentPage: params.Page,
			Total:       total,
		},
	}

	// 未認証ユーザーの検索結果をキャッシュ
	if canUseCache {
		if cacheErr := s.cache.SetWithTTL(cacheKey, response, constants.SearchResultCacheTTL); cacheErr != nil {
			s.logger.Warn("Failed to cache search results", "error", cacheErr)
		}
	}

	return response, nil
}

// generateAdvancedSearchCacheKey - 高度検索パラメータからキャッシュキーを生成
func (s *AnimeService) generateAdvancedSearchCacheKey(params models.AdvancedSearchParams) string {
	genresKey := strings.Join(params.Genres, ",")
	return fmt.Sprintf("%sq:%s:g:%s:y:%d:s:%s:mr:%v:xr:%v:st:%s:so:%s:o:%s:p:%d:pp:%d",
		constants.CacheKeyAdvSearch,
		params.Query, genresKey, params.Year, params.Season,
		params.MinRating, params.MaxRating, params.Status,
		params.Sort, params.Order, params.Page, params.PerPage)
}

// AdvancedSearchAnimes - 高度検索機能
func (s *AnimeService) AdvancedSearchAnimes(params models.AdvancedSearchParams, userID *int) (*models.PaginatedResponse, error) {
	// Set defaults
	s.setSearchDefaults(&params)

	// 未認証ユーザーかつUserStatusフィルターなしの場合のみキャッシュを使用
	// （UserStatusはユーザー固有のフィルターのため）
	canUseCache := userID == nil && params.UserStatus == ""
	if canUseCache {
		cacheKey := s.generateAdvancedSearchCacheKey(params)
		var cachedResponse models.PaginatedResponse
		if found, err := s.cache.Get(cacheKey, &cachedResponse); found && err == nil {
			s.logger.Debug("Cache hit for advanced search", "query", params.Query, "page", params.Page)
			return &cachedResponse, nil
		}
	}

	// Build query components
	baseQuery := s.buildBaseQuery(params, userID)
	whereClause, args, argIndex := s.buildWhereClause(params, userID)
	orderBy := s.buildOrderByClause(params)

	// Count total results (distinct to handle JOINs)
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT a.id) %s %s", baseQuery, whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count animes: %w", err)
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PerPage
	totalPages := (total + params.PerPage - 1) / params.PerPage

	// Get animes
	query := fmt.Sprintf(`
		SELECT a.id, a.title, a.image_url, a.season,
		       CASE 
		           WHEN a.released_on IS NOT NULL AND a.released_on != '' AND a.released_on::date > CURRENT_DATE THEN 'upcoming'
		           WHEN a.released_on IS NOT NULL AND a.released_on != '' AND a.released_on::date <= CURRENT_DATE AND a.status = 'not_yet_aired' THEN 'airing'
		           ELSE a.status
		       END as status,
		       a.mal_anime_id, a.annict_id, a.created_at, a.updated_at, a.watchers_count,
		       COALESCE(a.cover_image_extra_large, '') as cover_image_extra_large,
		       COALESCE(a.cover_image_large, '') as cover_image_large,
		       COALESCE(a.cover_image_medium, '') as cover_image_medium
		%s %s
		GROUP BY a.id, a.title, a.image_url, a.season, a.status, a.released_on, a.mal_anime_id, a.annict_id, a.created_at, a.updated_at, a.watchers_count, a.cover_image_extra_large, a.cover_image_large, a.cover_image_medium
		%s
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, params.PerPage, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search animes: %w", err)
	}
	defer rows.Close()

	animes := []models.AnimeWithUserData{}
	for rows.Next() {
		var anime models.Anime
		err := rows.Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount, &anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium)
		if err != nil {
			continue
		}

		animeWithData := models.AnimeWithUserData{Anime: anime}
		animes = append(animes, animeWithData)
	}

	// ユーザーデータを一括取得（N+1クエリ解消）
	if userID != nil {
		s.enrichAnimesWithUserData(animes, *userID)
	}

	// DBキャッシュがないアニメのみ外部APIから画像を取得
	s.enrichAnimesWithUserDataCoverImagesIfNeeded(animes)

	// 検索ログを記録
	s.logAdvancedSearch(params, userID, total)

	response := &models.PaginatedResponse{
		Data: animes,
		Meta: models.PaginationMeta{
			TotalPages:  totalPages,
			CurrentPage: params.Page,
			Total:       total,
		},
	}

	// 未認証ユーザーの検索結果をキャッシュ
	if canUseCache {
		cacheKey := s.generateAdvancedSearchCacheKey(params)
		if cacheErr := s.cache.SetWithTTL(cacheKey, response, constants.SearchResultCacheTTL); cacheErr != nil {
			s.logger.Warn("Failed to cache advanced search results", "error", cacheErr)
		}
	}

	return response, nil
}

func (s *AnimeService) GetAnimeByID(animeID int, userID *int) (*models.AnimeWithUserData, error) {
	var anime models.Anime
	var description, media, officialSiteURL, titleKana, titleEn, releasedOn, seriesTitle sql.NullString
	var broadcastDay, broadcastTime, broadcastTimezone sql.NullString
	var episodesCount, seriesTMDBID, durationMinutes sql.NullInt64

	err := s.db.QueryRow(`
		SELECT id, title, image_url, season,
		       CASE 
		           WHEN released_on IS NOT NULL AND released_on != '' AND released_on::date > CURRENT_DATE THEN 'upcoming'
		           WHEN released_on IS NOT NULL AND released_on != '' AND released_on::date <= CURRENT_DATE AND status = 'not_yet_aired' THEN 'airing'
		           ELSE status
		       END as status,
		       mal_anime_id, annict_id, created_at, updated_at, watchers_count,
		       COALESCE(cover_image_extra_large, '') as cover_image_extra_large,
		       COALESCE(cover_image_large, '') as cover_image_large,
		       COALESCE(cover_image_medium, '') as cover_image_medium,
		       description, media, released_on, official_site_url, title_kana, title_en, episodes_count,
		       series_title, series_tmdb_id,
		       broadcast_day, broadcast_time, broadcast_timezone, duration_minutes
		FROM animes WHERE id = $1
	`, animeID).Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount, &anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium,
		&description, &media, &releasedOn, &officialSiteURL, &titleKana, &titleEn, &episodesCount, &seriesTitle, &seriesTMDBID,
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
	if episodesCount.Valid {
		count := int(episodesCount.Int64)
		anime.EpisodesCount = &count
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

	// DBキャッシュがない場合、またはextraLargeがない場合はAniListから画像を取得
	if anime.CoverImageExtraLarge == "" || anime.CoverImageLarge == "" {
		s.enrichAnimeWithCoverImage(&anime)
	}

	// descriptionがない場合はAniList APIから取得
	if anime.Description == nil || *anime.Description == "" {
		s.enrichAnimeWithDescription(&anime)
	}

	// series_titleがない場合は設定（シリーズフォールバック用）
	s.ensureSeriesTitle(&anime)

	// released_onがない場合はJikan APIから取得
	if anime.ReleasedOn == nil || *anime.ReleasedOn == "" {
		s.enrichAnimeWithReleasedOn(&anime)
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

	// Get episodes
	episodes, err := s.GetAnimeEpisodes(animeID, userID)
	if err != nil {
		return nil, err
	}
	animeWithData.Episodes = episodes

	// Get rating statistics (average and count)
	// Using database aggregation for better performance and scalability
	var avgRating sql.NullFloat64
	var ratingCount int
	err = s.db.QueryRow(`
		SELECT AVG(rating), COUNT(*)
		FROM user_ratings
		WHERE anime_id = $1
	`, animeID).Scan(&avgRating, &ratingCount)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Warn("Failed to get rating statistics",
			"anime_id", animeID,
			"error", err)
	} else if err == nil && avgRating.Valid {
		animeWithData.AverageRating = &avgRating.Float64
		animeWithData.RatingCount = ratingCount
	}

	// Get rating distribution
	// Using database GROUP BY for efficient aggregation
	rows, err := s.db.Query(`
		SELECT FLOOR(rating) as rating_floor, COUNT(*) as count
		FROM user_ratings
		WHERE anime_id = $1
		GROUP BY FLOOR(rating)
		ORDER BY rating_floor
	`, animeID)
	if err != nil {
		s.logger.Warn("Failed to get rating distribution",
			"anime_id", animeID,
			"error", err)
	} else {
		defer rows.Close()
		distribution := make(map[int]int)
		for rows.Next() {
			var rating int
			var count int
			if err := rows.Scan(&rating, &count); err != nil {
				s.logger.Warn("Failed to scan rating distribution row",
					"anime_id", animeID,
					"error", err)
				continue
			}
			distribution[rating] = count
		}

		// Check for errors during iteration
		if err := rows.Err(); err != nil {
			s.logger.Warn("Error during rating distribution iteration",
				"anime_id", animeID,
				"error", err)
		}

		if len(distribution) > 0 {
			animeWithData.RatingDistribution = distribution
		}
	}

	// Add user data if authenticated
	if userID != nil {
		// Get user favorite status
		var favorite models.UserFavorite
		err := s.db.QueryRow(`
			SELECT id, user_id, anime_id, status, created_at, updated_at
			FROM user_favorites WHERE user_id = $1 AND anime_id = $2
		`, *userID, animeID).Scan(&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.Status, &favorite.CreatedAt, &favorite.UpdatedAt)
		if err != nil && err != sql.ErrNoRows {
			// Log error but don't fail the entire request
			s.logger.Warn("Failed to get user favorite status",
				"user_id", *userID,
				"anime_id", animeID,
				"error", err)
		} else if err == nil {
			animeWithData.UserFavorite = &favorite
		}

		// Get user rating
		var userRating models.UserRating
		err = s.db.QueryRow(`
			SELECT id, user_id, anime_id, rating, comment, created_at, updated_at
			FROM user_ratings WHERE user_id = $1 AND anime_id = $2
		`, *userID, animeID).Scan(&userRating.ID, &userRating.UserID, &userRating.AnimeID, &userRating.Rating, &userRating.Comment, &userRating.CreatedAt, &userRating.UpdatedAt)
		if err != nil && err != sql.ErrNoRows {
			// Log error but don't fail the entire request
			s.logger.Warn("Failed to get user rating",
				"user_id", *userID,
				"anime_id", animeID,
				"error", err)
		} else if err == nil {
			animeWithData.UserRating = &userRating
		}

		// Calculate watch progress
		animeWithData.WatchProgress = s.calculateWatchProgress(*userID, animeID)
	}

	// Get watch providers from TMDB (if TMDB ID is available or can be found)
	s.enrichAnimeWithWatchProviders(animeWithData)

	return animeWithData, nil
}

// SyncAnimeStatus - TMDBからステータスを取得して更新
func (s *AnimeService) SyncAnimeStatus(animeID int, clearCache bool) error {
	// アニメ情報を取得（released_on, mediaも含める）
	var anime models.Anime
	var releasedOn, media sql.NullString
	var seriesTMDBID sql.NullInt64
	err := s.db.QueryRow(`
		SELECT id, series_tmdb_id, season, released_on, media
		FROM animes
		WHERE id = $1
	`, animeID).Scan(&anime.ID, &seriesTMDBID, &anime.Season, &releasedOn, &media)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAnimeNotFound
		}
		return fmt.Errorf("failed to get anime for status sync: %w", err)
	}

	// NullableフィールドをGoポインタに変換
	if seriesTMDBID.Valid {
		id := int(seriesTMDBID.Int64)
		anime.SeriesTMDBID = &id
	}
	var releasedOnPtr *string
	if releasedOn.Valid {
		releasedOnPtr = &releasedOn.String
	}

	var status string
	var seasonStr string
	if anime.Season != nil {
		seasonStr = *anime.Season
	}

	// TMDBからステータスを取得
	// 映画（movie）の場合はTMDB TV APIが使えないため、released_onベースで判定
	isMovie := media.Valid && media.String == "movie"
	if isMovie {
		// 映画はreleased_on + シーズンベースで判定
		status = determineStatusBySeason(seasonStr, releasedOnPtr)
	} else if anime.SeriesTMDBID != nil && s.tmdbClient != nil && s.tmdbClient.IsEnabled() {
		ctx := context.Background()
		details, err := s.tmdbClient.GetTVDetails(ctx, *anime.SeriesTMDBID, "ja-JP")
		if err == nil {
			status = mapTMDBStatus(details.Status)
			// TMDBがairingを返しても、released_onが未来ならupcomingに上書き
			if releasedOnPtr != nil && *releasedOnPtr != "" {
				// UTCでパースして比較（サーバーのタイムゾーンに依存しないようにする）
				releaseDate, parseErr := time.ParseInLocation("2006-01-02", *releasedOnPtr, time.UTC)
				if parseErr == nil {
					today := time.Now().UTC().Truncate(24 * time.Hour)
					if releaseDate.After(today) {
						status = "upcoming"
					}
				}
			}
		} else {
			s.logger.Warn("Failed to get TMDB TV details for status sync, falling back to season-based status",
				"anime_id", animeID,
				"series_tmdb_id", *anime.SeriesTMDBID,
				"error", err)
			// フォールバック
			status = determineStatusBySeason(seasonStr, releasedOnPtr)
		}
	} else {
		// TMDB IDがない、またはTMDBクライアントが利用できない場合はシーズンベース
		status = determineStatusBySeason(seasonStr, releasedOnPtr)
	}

	// 既存のステータスと比較して変更がなければ更新しない
	var currentStatus sql.NullString
	err = s.db.QueryRow("SELECT status FROM animes WHERE id = $1", animeID).Scan(&currentStatus)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get current status for anime %d: %w", animeID, err)
	}

	if currentStatus.Valid && currentStatus.String == status {
		// ステータスに変更がなければ何もしない
		return nil
	}

	// ステータスを更新
	_, err = s.db.Exec(`
		UPDATE animes
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, status, animeID)
	if err != nil {
		return fmt.Errorf("failed to update anime status for anime %d to %s: %w", animeID, status, err)
	}

	// キャッシュをクリア（バッチ処理の場合は最後にまとめてクリアするため false を指定可能）
	if clearCache {
		s.ClearAnimeCache()
	}
	s.logger.Debug("Anime status updated", "anime_id", animeID, "new_status", status)

	return nil
}

// SyncAllAnimeStatuses - 全アニメのステータスを更新
// 処理に時間がかかるため、バックグラウンドで実行することを推奨
func (s *AnimeService) SyncAllAnimeStatuses() (int, int, error) {
	rows, err := s.db.Query(`
        SELECT id, season FROM animes
        WHERE series_tmdb_id IS NOT NULL OR season IS NOT NULL
    `)
	if err != nil {
		s.logger.Error("Failed to query animes for status sync", "error", err)
		return 0, 0, fmt.Errorf("failed to query animes for status sync: %w", err)
	}
	defer rows.Close()

	successCount := 0
	failureCount := 0

	var animeIDs []int
	for rows.Next() {
		var animeID int
		var season sql.NullString // seasonがNULLの場合もあるのでsql.NullStringで受け取る
		if err := rows.Scan(&animeID, &season); err != nil {
			s.logger.Warn("Failed to scan anime ID or season for status sync", "error", err)
			failureCount++
			continue
		}
		// seasonが有効でなければ、determineStatusBySeasonが"finished"を返すため問題なし
		animeIDs = append(animeIDs, animeID)
	}

	totalAnimes := len(animeIDs)
	if totalAnimes == 0 {
		s.logger.Info("No animes found with TMDB ID or season for status sync.")
		return 0, 0, nil
	}

	s.logger.Info("Starting status sync for all animes", "total_animes", totalAnimes)

	// レート制限対策 (TMDB API呼び出しを考慮)
	// TMDBの制限は 40 req / 10 sec なので、余裕を持って 250ms (4 req / sec) とする
	const delay = 250 * time.Millisecond

	for i, animeID := range animeIDs {
		// 進行状況のロギング
		if i > 0 && i%50 == 0 { // 50件ごとにログ出力
			s.logger.Info("Processing anime status sync", "progress", fmt.Sprintf("%d/%d", i, totalAnimes))
		}

		// バッチ処理内では個別のキャッシュクリアを無効化 (false)
		if err := s.SyncAnimeStatus(animeID, false); err != nil {
			s.logger.Warn("Failed to sync anime status", "anime_id", animeID, "error", err)
			failureCount++
		} else {
			successCount++
		}

		time.Sleep(delay)
	}

	// 全ての更新が完了した後に一度だけキャッシュをクリア
	if successCount > 0 {
		s.ClearAnimeCache()
	}

	s.logger.Info("Anime status sync completed", "success", successCount, "failure", failureCount)
	return successCount, failureCount, nil
}

func (s *AnimeService) GetAnimeEpisodes(animeID int, userID *int) ([]models.Episode, error) {
	// Check if anime exists and get MAL ID
	var exists bool
	var malAnimeID *string
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1), mal_anime_id FROM animes WHERE id = $1", animeID).Scan(&exists, &malAnimeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	rows, err := s.db.Query(`
		SELECT id, anime_id, title, episode_number, duration_minutes, air_date, 
		       COALESCE(is_special, FALSE) as is_special, 
		       COALESCE(filler, FALSE) as filler, 
		       COALESCE(recap, FALSE) as recap,
		       created_at, updated_at
		FROM episodes WHERE anime_id = $1
		ORDER BY episode_number ASC, is_special ASC
	`, animeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", err)
	}
	defer rows.Close()

	episodes := []models.Episode{}
	hasMissingDurations := false
	for rows.Next() {
		var episode models.Episode
		err := rows.Scan(&episode.ID, &episode.AnimeID, &episode.Title, &episode.EpisodeNumber,
			&episode.DurationMinutes, &episode.AirDate, &episode.IsSpecial, &episode.Filler,
			&episode.Recap, &episode.CreatedAt, &episode.UpdatedAt)
		if err != nil {
			continue
		}

		// Check if duration or air_date is missing
		if episode.DurationMinutes == nil || episode.AirDate == nil {
			hasMissingDurations = true
		}

		// Add watch status if user is authenticated
		if userID != nil {
			var viewed bool
			err := s.db.QueryRow(`
				SELECT viewed FROM watch_statuses 
				WHERE user_id = $1 AND episode_id = $2
			`, *userID, episode.ID).Scan(&viewed)
			if err == nil {
				episode.Viewed = viewed
			}
		}

		episodes = append(episodes, episode)
	}

	// If there are missing durations and we have a MAL ID, try to fetch them in background
	if hasMissingDurations && malAnimeID != nil && *malAnimeID != "" {
		go func() {
			err := s.enrichEpisodesWithDuration(animeID, *malAnimeID)
			if err != nil {
				s.logger.Warn("Failed to enrich episodes with duration",
					"anime_id", animeID,
					"mal_anime_id", *malAnimeID,
					"error", err)
			}
		}()
	}

	return episodes, nil
}

func (s *AnimeService) GetUserFavorites(userID int, status string) ([]models.AnimeWithUserData, error) {
	whereClause := "WHERE uf.user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if status != "" {
		whereClause += fmt.Sprintf(" AND uf.status = $%d", argIndex)
		args = append(args, status)
	}

	query := fmt.Sprintf(`
		SELECT a.id, a.title, a.image_url, a.season, a.status, a.mal_anime_id, a.created_at, a.updated_at, a.watchers_count,
		       COALESCE(a.cover_image_extra_large, '') as cover_image_extra_large,
		       COALESCE(a.cover_image_large, '') as cover_image_large,
		       COALESCE(a.cover_image_medium, '') as cover_image_medium,
		       uf.id, uf.user_id, uf.anime_id, uf.status, uf.created_at, uf.updated_at
		FROM animes a
		JOIN user_favorites uf ON a.id = uf.anime_id
		%s
		ORDER BY uf.updated_at DESC
	`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user favorites: %w", err)
	}
	defer rows.Close()

	favorites := []models.AnimeWithUserData{}
	for rows.Next() {
		var anime models.Anime
		var favorite models.UserFavorite
		err := rows.Scan(
			&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.MalAnimeID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount,
			&anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium,
			&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.Status, &favorite.CreatedAt, &favorite.UpdatedAt,
		)
		if err != nil {
			continue
		}

		animeWithData := models.AnimeWithUserData{
			Anime:        anime,
			UserFavorite: &favorite,
		}

		// Calculate watch progress
		animeWithData.WatchProgress = s.calculateWatchProgress(userID, anime.ID)

		favorites = append(favorites, animeWithData)
	}

	// DBキャッシュがないアニメのみ外部APIから画像を取得
	s.enrichAnimesWithUserDataCoverImagesIfNeeded(favorites)

	return favorites, nil
}

func (s *AnimeService) UpdateFavorite(userID, animeID int, status string) (*models.UserFavorite, error) {
	// Check if anime exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1)", animeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	// Upsert favorite
	var favorite models.UserFavorite
	err = s.db.QueryRow(`
		INSERT INTO user_favorites (user_id, anime_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, anime_id)
		DO UPDATE SET status = $3, updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, anime_id, status, created_at, updated_at
	`, userID, animeID, status).Scan(&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.Status, &favorite.CreatedAt, &favorite.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update favorite: %w", err)
	}

	return &favorite, nil
}

func (s *AnimeService) DeleteFavorite(userID, animeID int) error {
	_, err := s.db.Exec(`
		DELETE FROM user_favorites WHERE user_id = $1 AND anime_id = $2
	`, userID, animeID)
	if err != nil {
		return fmt.Errorf("failed to delete favorite: %w", err)
	}

	return nil
}

func (s *AnimeService) UpdateWatchStatus(userID, episodeID int, viewed bool) (*models.WatchStatus, error) {
	// Check if episode exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM episodes WHERE id = $1)", episodeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check episode existence: %w", err)
	}
	if !exists {
		return nil, ErrEpisodeNotFound
	}

	// Upsert watch status
	var watchStatus models.WatchStatus
	err = s.db.QueryRow(`
		INSERT INTO watch_statuses (user_id, episode_id, viewed, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, episode_id)
		DO UPDATE SET viewed = $3, updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, episode_id, viewed, updated_at
	`, userID, episodeID, viewed).Scan(&watchStatus.ID, &watchStatus.UserID, &watchStatus.EpisodeID, &watchStatus.Viewed, &watchStatus.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update watch status: %w", err)
	}

	// Add to watch history if marked as viewed
	if viewed {
		_, err = s.db.Exec(`
			INSERT INTO watch_history (user_id, episode_id, watched_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
		`, userID, episodeID)
		if err != nil {
			// Log error but don't fail the request
			s.logger.Warn("Failed to add to watch history", "error", err, "user_id", userID, "episode_id", episodeID)
		}
	}

	return &watchStatus, nil
}

func (s *AnimeService) UpdateRating(userID, animeID int, rating float64, comment *string) (*models.UserRating, error) {
	// Check if anime exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1)", animeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	// Upsert rating
	var userRating models.UserRating
	err = s.db.QueryRow(`
		INSERT INTO user_ratings (user_id, anime_id, rating, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, anime_id)
		DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, anime_id, rating, comment, created_at, updated_at
	`, userID, animeID, rating, comment).Scan(&userRating.ID, &userRating.UserID, &userRating.AnimeID, &userRating.Rating, &userRating.Comment, &userRating.CreatedAt, &userRating.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update rating: %w", err)
	}

	return &userRating, nil
}

func (s *AnimeService) DeleteRating(userID, animeID int) error {
	_, err := s.db.Exec(`
		DELETE FROM user_ratings WHERE user_id = $1 AND anime_id = $2
	`, userID, animeID)
	if err != nil {
		return fmt.Errorf("failed to delete rating: %w", err)
	}

	return nil
}

func (s *AnimeService) GetRecentlyWatched(userID, limit int) ([]models.WatchHistory, error) {
	rows, err := s.db.Query(`
		SELECT wh.id, wh.user_id, wh.episode_id, wh.watched_at,
		       a.title as anime_title, e.title as episode_title
		FROM watch_history wh
		JOIN episodes e ON wh.episode_id = e.id
		JOIN animes a ON e.anime_id = a.id
		WHERE wh.user_id = $1
		ORDER BY wh.watched_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recently watched: %w", err)
	}
	defer rows.Close()

	history := []models.WatchHistory{}
	for rows.Next() {
		var h models.WatchHistory
		err := rows.Scan(&h.ID, &h.UserID, &h.EpisodeID, &h.WatchedAt, &h.AnimeTitle, &h.EpisodeTitle)
		if err != nil {
			continue
		}
		history = append(history, h)
	}

	return history, nil
}

func (s *AnimeService) calculateWatchProgress(userID, animeID int) float64 {
	var totalEpisodes, watchedEpisodes int

	// Get total episodes
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM episodes WHERE anime_id = $1
	`, animeID).Scan(&totalEpisodes)
	if err != nil || totalEpisodes == 0 {
		return 0.0
	}

	// Get watched episodes
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM watch_statuses ws
		JOIN episodes e ON ws.episode_id = e.id
		WHERE ws.user_id = $1 AND e.anime_id = $2 AND ws.viewed = true
	`, userID, animeID).Scan(&watchedEpisodes)
	if err != nil {
		return 0.0
	}

	return float64(watchedEpisodes) / float64(totalEpisodes) * 100.0
}

// batchGetUserFavorites - 複数アニメのお気に入り情報を一括取得（N+1クエリ解消）
func (s *AnimeService) batchGetUserFavorites(userID int, animeIDs []int) (map[int]*models.UserFavorite, error) {
	if len(animeIDs) == 0 {
		return make(map[int]*models.UserFavorite), nil
	}

	// プレースホルダーを動的に生成
	placeholders := make([]string, len(animeIDs))
	args := make([]interface{}, len(animeIDs)+1)
	args[0] = userID
	for i, id := range animeIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, anime_id, status, created_at, updated_at
		FROM user_favorites
		WHERE user_id = $1 AND anime_id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get user favorites: %w", err)
	}
	defer rows.Close()

	favorites := make(map[int]*models.UserFavorite)
	for rows.Next() {
		var fav models.UserFavorite
		if err := rows.Scan(&fav.ID, &fav.UserID, &fav.AnimeID, &fav.Status, &fav.CreatedAt, &fav.UpdatedAt); err != nil {
			s.logger.Warn("Failed to scan user favorite", "error", err)
			continue
		}
		favorites[fav.AnimeID] = &fav
	}

	return favorites, nil
}

// batchCalculateWatchProgress - 複数アニメの視聴進捗を一括計算（N+1クエリ解消）
func (s *AnimeService) batchCalculateWatchProgress(userID int, animeIDs []int) (map[int]float64, error) {
	if len(animeIDs) == 0 {
		return make(map[int]float64), nil
	}

	// プレースホルダーを動的に生成
	placeholders := make([]string, len(animeIDs))
	args := make([]interface{}, len(animeIDs)+1)
	args[0] = userID
	for i, id := range animeIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	// 1つのクエリで全アニメの視聴進捗を計算
	query := fmt.Sprintf(`
		SELECT 
			e.anime_id,
			COUNT(e.id) as total_episodes,
			COUNT(CASE WHEN ws.viewed = true THEN 1 END) as watched_episodes
		FROM episodes e
		LEFT JOIN watch_statuses ws ON e.id = ws.episode_id AND ws.user_id = $1
		WHERE e.anime_id IN (%s)
		GROUP BY e.anime_id
	`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to batch calculate watch progress: %w", err)
	}
	defer rows.Close()

	progress := make(map[int]float64)
	for rows.Next() {
		var animeID, totalEpisodes, watchedEpisodes int
		if err := rows.Scan(&animeID, &totalEpisodes, &watchedEpisodes); err != nil {
			s.logger.Warn("Failed to scan watch progress", "error", err)
			continue
		}
		if totalEpisodes > 0 {
			progress[animeID] = float64(watchedEpisodes) / float64(totalEpisodes) * 100.0
		} else {
			progress[animeID] = 0.0
		}
	}

	// animeIDsに含まれるがエピソードがないアニメは0%として設定
	for _, id := range animeIDs {
		if _, exists := progress[id]; !exists {
			progress[id] = 0.0
		}
	}

	return progress, nil
}

// enrichAnimesWithUserData - アニメリストにユーザーデータを一括追加（N+1クエリ解消版）
// 注意: この関数はユーザーデータ取得に失敗してもエラーを返さず、警告ログのみ出力します。
// これは意図的な設計で、ユーザーデータ（お気に入り・視聴進捗）が取得できなくても
// アニメ一覧自体は表示できるようにするためです。クライアント側では、
// UserFavoriteやWatchProgressがnilまたは0の場合を適切にハンドリングしてください。
func (s *AnimeService) enrichAnimesWithUserData(animes []models.AnimeWithUserData, userID int) {
	if len(animes) == 0 {
		return
	}

	// アニメIDのリストを作成
	animeIDs := make([]int, len(animes))
	for i, anime := range animes {
		animeIDs[i] = anime.ID
	}

	// お気に入り情報を一括取得
	// エラー時は空のmapを使用し、処理を続行（グレースフルデグラデーション）
	favorites, err := s.batchGetUserFavorites(userID, animeIDs)
	if err != nil {
		s.logger.Warn("Failed to batch get user favorites", "error", err)
		favorites = make(map[int]*models.UserFavorite)
	}

	// 視聴進捗を一括計算
	// エラー時は空のmapを使用し、処理を続行（グレースフルデグラデーション）
	progress, err := s.batchCalculateWatchProgress(userID, animeIDs)
	if err != nil {
		s.logger.Warn("Failed to batch calculate watch progress", "error", err)
		progress = make(map[int]float64)
	}

	// 各アニメにユーザーデータを設定
	for i := range animes {
		animeID := animes[i].ID
		if fav, ok := favorites[animeID]; ok {
			animes[i].UserFavorite = fav
		}
		if prog, ok := progress[animeID]; ok {
			animes[i].WatchProgress = prog
		}
	}
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
		Ratings: models.RatingStats{
			Min:     1.0,
			Max:     5.0,
			Average: 3.0,
			Count:   0,
		},
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

	// 評価統計を取得
	var ratingStats models.RatingStats
	err = s.db.QueryRow(`
		SELECT 
			COALESCE(MIN(rating), 1.0) as min_rating,
			COALESCE(MAX(rating), 5.0) as max_rating,
			COALESCE(AVG(rating), 3.0) as avg_rating,
			COUNT(*) as count
		FROM user_ratings
		WHERE rating IS NOT NULL AND rating >= 1.0 AND rating <= 5.0
	`).Scan(&ratingStats.Min, &ratingStats.Max, &ratingStats.Average, &ratingStats.Count)
	if err != nil {
		s.logger.Error("Failed to get rating stats", "error", err)
		// デフォルト値はすでに設定済み
	} else {
		filterOptions.Ratings = ratingStats
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

// logSearch - 検索ログを記録（非同期）
func (s *AnimeService) logSearch(query string, filters map[string]interface{}, userID *int, resultCount int) {
	// 安全にJSONに変換
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		// JSON変換エラーは無視（メイン機能に影響しない）
		s.logger.Warn("Failed to marshal search filters", "error", err, "filters", filters)
		return
	}

	_, err = s.db.Exec(`
		INSERT INTO search_logs (user_id, query, filters, result_count, searched_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`, userID, query, string(filtersJSON), resultCount)
	if err != nil {
		// ログ記録エラーは無視（メイン機能に影響しない）
		s.logger.Warn("Failed to log search", "error", err, "query", query, "result_count", resultCount)
	}
}

// logAdvancedSearch - 高度検索のログを記録（非同期）
func (s *AnimeService) logAdvancedSearch(params models.AdvancedSearchParams, userID *int, resultCount int) {
	filters := map[string]interface{}{
		"genres":      params.Genres,
		"year":        params.Year,
		"season":      params.Season,
		"min_rating":  params.MinRating,
		"max_rating":  params.MaxRating,
		"status":      params.Status,
		"user_status": params.UserStatus,
		"sort":        params.Sort,
		"order":       params.Order,
	}
	go s.logSearch(params.Query, filters, userID, resultCount)
}

// logBasicSearch - 基本検索のログを記録（非同期）
func (s *AnimeService) logBasicSearch(params models.AnimeSearchParams, userID *int, resultCount int) {
	filters := map[string]interface{}{
		"genre":  params.Genre,
		"season": params.Season,
		"status": params.Status,
		"sort":   params.Sort,
	}
	go s.logSearch(params.Keyword, filters, userID, resultCount)
}

// GetPopularSearches - 人気検索キーワードを取得
func (s *AnimeService) GetPopularSearches(limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10
	}

	// キャッシュキーを生成
	cacheKey := fmt.Sprintf("%s%d", constants.CacheKeyPopularSearches, limit)

	// Redisキャッシュから取得を試みる
	if s.redisCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var cachedSearches []string
		if found, err := s.redisCache.Get(ctx, cacheKey, &cachedSearches); found && err == nil {
			s.logger.Debug("Redis cache hit for popular searches", "limit", limit)
			// Redisから取得したデータでメモリキャッシュも更新（後続リクエストの高速化）
			if cacheErr := s.cache.SetWithTTL(cacheKey, cachedSearches, constants.PopularSearchesCacheTTL); cacheErr != nil {
				s.logger.Warn("Failed to update memory cache from Redis for popular searches", "error", cacheErr)
			}
			return cachedSearches, nil
		}
	}

	// メモリキャッシュから取得を試みる
	var cachedSearches []string
	if found, err := s.cache.Get(cacheKey, &cachedSearches); found && err == nil {
		s.logger.Debug("Memory cache hit for popular searches", "limit", limit)
		return cachedSearches, nil
	}

	// まず検索ログが存在するかチェック
	var logCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM search_logs").Scan(&logCount)
	if err != nil {
		s.logger.Error("Failed to count search logs", "error", err)
		return []string{}, nil
	}

	// ログが少ない場合は期間を拡張
	intervalDays := 30
	if logCount < 10 {
		intervalDays = 365 // 1年に拡張
	}

	rows, err := s.db.Query(`
		SELECT query, COUNT(*) as search_count
		FROM search_logs
		WHERE query IS NOT NULL 
		  AND query != '' 
		  AND LENGTH(TRIM(query)) > 0
		  AND searched_at > CURRENT_TIMESTAMP - ($1 || ' days')::INTERVAL
		GROUP BY query
		ORDER BY search_count DESC, query ASC
		LIMIT $2
	`, intervalDays, limit)
	if err != nil {
		s.logger.Error("Failed to get popular searches", "error", err, "limit", limit)
		return []string{}, nil // エラーでも空配列を返す
	}
	defer rows.Close()

	searches := []string{}
	for rows.Next() {
		var query string
		var count int
		err := rows.Scan(&query, &count)
		if err != nil {
			s.logger.Error("Failed to scan popular search", "error", err)
			continue
		}
		// 空文字列やnullをスキップ
		if strings.TrimSpace(query) != "" {
			searches = append(searches, query)
		}
	}

	// 検索ログが少ない場合はデフォルトの人気検索を返す
	if len(searches) == 0 {
		// グローバルスライスを変更しないようにコピーを作成
		searches = make([]string, len(constants.DefaultPopularSearches))
		copy(searches, constants.DefaultPopularSearches)
		if limit < len(searches) {
			searches = searches[:limit]
		}
	}

	// Redisキャッシュに保存
	if s.redisCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := s.redisCache.Set(ctx, cacheKey, searches, constants.PopularSearchesCacheTTL); err != nil {
			s.logger.Warn("Failed to cache popular searches in Redis", "error", err)
		} else {
			s.logger.Debug("Cached popular searches in Redis", "limit", limit)
		}
	}

	// メモリキャッシュにも保存
	if err := s.cache.SetWithTTL(cacheKey, searches, constants.PopularSearchesCacheTTL); err != nil {
		s.logger.Warn("Failed to cache popular searches in memory", "error", err)
	}

	return searches, nil
}

// GetSearchSuggestions - 検索候補を取得
func (s *AnimeService) GetSearchSuggestions(query string, limit int) ([]string, error) {
	// 空のクエリに対する適切な処理
	if strings.TrimSpace(query) == "" {
		return []string{}, nil
	}

	// limitのデフォルト値設定
	if limit <= 0 {
		limit = 10
	}

	// title, title_kana, title_en のいずれかにマッチするタイトルを取得
	rows, err := s.db.Query(`
		SELECT title
		FROM animes
		WHERE (title ILIKE $1 OR title_kana ILIKE $1 OR title_en ILIKE $1)
		  AND TRIM(title) <> ''
		GROUP BY title
		ORDER BY MAX(watchers_count) DESC NULLS LAST, title
		LIMIT $2
	`, query+"%", limit)

	if err != nil {
		s.logger.Error("Failed to get search suggestions", "error", err, "query", query, "limit", limit)
		return []string{}, nil // エラーでも空配列を返す
	}
	defer rows.Close()

	suggestions := []string{}
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			s.logger.Warn("Failed to scan suggestion title", "error", err)
			continue
		}
		suggestions = append(suggestions, title)
	}

	return suggestions, nil
}

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
			anime.Description = &overview

			// DBに保存（非同期）
			animeID := anime.ID
			seriesTitle := NormalizeAnimeTitle(anime.Title)
			s.enqueueBackgroundTask(func() {
				s.updateAnimeSeriesInfo(context.Background(), animeID, overview, tmdbID, seriesTitle, tmdbID)
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
			anime.Description = &seriesDescription

			// DBに保存（非同期）
			animeID := anime.ID
			s.enqueueBackgroundTask(func() {
				var tmdbID int
				if seriesTMDBID.Valid {
					tmdbID = int(seriesTMDBID.Int64)
				}
				s.updateAnimeSeriesInfo(context.Background(), animeID, seriesDescription, 0, seriesTitle, tmdbID)
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

	// アニメ情報に設定
	anime.Description = &description

	// DBに保存（非同期）
	malID := *anime.MalAnimeID
	s.enqueueBackgroundTask(func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.BackgroundTaskDBTimeout)
		defer cancel()

		_, saveErr := s.db.ExecContext(ctx, `
			UPDATE animes 
			SET description = $1
			WHERE mal_anime_id = $2
		`, description, malID)
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

	// Jikan APIから上映時間を取得
	duration, err := s.jikanClient.GetAnimeDuration(*anime.MalAnimeID)
	if err != nil {
		s.logger.Debug("Failed to get movie duration from Jikan",
			"error", err,
			"mal_anime_id", *anime.MalAnimeID,
			"anime_title", anime.Title)
		return
	}

	if duration <= 0 {
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

// AdvancedSearchAnimes のヘルパー関数

// setSearchDefaults - 検索パラメータのデフォルト値を設定
func (s *AnimeService) setSearchDefaults(params *models.AdvancedSearchParams) {
	const (
		SortByTitle = "title"
		OrderAsc    = "asc"
	)

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}
	if params.Sort == "" {
		params.Sort = SortByTitle
	}
	if params.Order == "" {
		params.Order = OrderAsc
	}
}

// buildBaseQuery - ベースクエリとJOINを構築
func (s *AnimeService) buildBaseQuery(params models.AdvancedSearchParams, userID *int) string {
	const SortByRating = "rating"

	baseQuery := "FROM animes a"

	// Track which JOINs are needed
	needGenreJoin := len(params.Genres) > 0
	needRatingJoin := params.MinRating > 0 || params.MaxRating > 0 || params.Sort == SortByRating
	needUserFavoriteJoin := userID != nil && params.UserStatus != ""

	// Add JOINs only when needed
	if needGenreJoin {
		baseQuery += "\n\t\tLEFT JOIN anime_genres ag ON a.id = ag.anime_id"
	}
	if needRatingJoin {
		baseQuery += `
		LEFT JOIN (
			SELECT anime_id, AVG(rating) as avg_rating
			FROM user_ratings
			GROUP BY anime_id
		) AS avg_ratings ON a.id = avg_ratings.anime_id`
	}
	if needUserFavoriteJoin {
		baseQuery += `
		INNER JOIN user_favorites uf ON a.id = uf.anime_id`
	}

	return baseQuery
}

// buildWhereClause - WHERE句と引数を構築
func (s *AnimeService) buildWhereClause(params models.AdvancedSearchParams, userID *int) (string, []interface{}, int) {
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	// キーワード検索（ヘルパー関数を使用、テーブルエイリアス "a." を指定）
	if params.Query != "" {
		clause, keywordArgs, newArgIndex := buildKeywordSearchClause(params.Query, argIndex, "a.")
		whereClauses = append(whereClauses, clause)
		args = append(args, keywordArgs...)
		argIndex = newArgIndex
	}

	// ジャンル検索（複数対応）
	if len(params.Genres) > 0 {
		genrePlaceholders := make([]string, len(params.Genres))
		for i, genre := range params.Genres {
			genrePlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, genre)
			argIndex++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("ag.genre IN (%s)", strings.Join(genrePlaceholders, ",")))
	}

	// 年検索（LIKE演算子を使用してインデックスを活用）
	if params.Year > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("a.season LIKE $%d", argIndex))
		args = append(args, fmt.Sprintf("%d-%%", params.Year))
		argIndex++
	}

	// シーズン検索（日本語フォーマットをDBフォーマットに変換）
	if params.Season != "" {
		dbSeason := utils.ConvertSeasonToDBFormat(params.Season)
		if dbSeason != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("a.season = $%d", argIndex))
			args = append(args, dbSeason)
			argIndex++
		}
	}

	// ステータス検索
	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	// 評価範囲検索
	if params.MinRating > 0 || params.MaxRating > 0 {
		if params.MinRating > 0 && params.MaxRating > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("avg_ratings.avg_rating BETWEEN $%d AND $%d", argIndex, argIndex+1))
			args = append(args, params.MinRating, params.MaxRating)
			argIndex += 2
		} else if params.MinRating > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("avg_ratings.avg_rating >= $%d", argIndex))
			args = append(args, params.MinRating)
			argIndex++
		} else if params.MaxRating > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("avg_ratings.avg_rating <= $%d", argIndex))
			args = append(args, params.MaxRating)
			argIndex++
		}
	}

	// ユーザーステータス検索（認証時のみ）
	// JOINを使用するため、WHERE句で条件を追加
	if userID != nil && params.UserStatus != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("uf.user_id = $%d", argIndex))
		args = append(args, *userID)
		argIndex++
		whereClauses = append(whereClauses, fmt.Sprintf("uf.status = $%d", argIndex))
		args = append(args, params.UserStatus)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	return whereClause, args, argIndex
}

// buildOrderByClause - ORDER BY句を構築
func (s *AnimeService) buildOrderByClause(params models.AdvancedSearchParams) string {
	const (
		SortByTitle      = "title"
		SortByRating     = "rating"
		SortByYear       = "year"
		SortByPopularity = "popularity"
		SortByUpdated    = "updated"
		OrderDesc        = "desc"
	)

	orderDirection := "ASC"
	if params.Order == OrderDesc {
		orderDirection = "DESC"
	}

	switch params.Sort {
	case SortByTitle:
		return fmt.Sprintf("ORDER BY a.title %s", orderDirection)
	case SortByRating:
		return fmt.Sprintf("ORDER BY avg_ratings.avg_rating %s NULLS LAST", orderDirection)
	case SortByYear:
		// season から年を抽出してソート
		// NOTE: SUBSTRING関数を使用しているため、インデックスが利用されません
		// 大規模データセットでパフォーマンスが問題になる場合は、
		// CREATE INDEX idx_animes_season_year ON animes (SUBSTRING(season FROM 1 FOR 4))
		// のような関数ベースのインデックスを作成することを検討してください
		return fmt.Sprintf("ORDER BY SUBSTRING(a.season FROM 1 FOR 4) %s, a.season %s", orderDirection, orderDirection)
	case SortByPopularity:
		return fmt.Sprintf("ORDER BY a.watchers_count %s", orderDirection)
	case SortByUpdated:
		return fmt.Sprintf("ORDER BY a.updated_at %s", orderDirection)
	default:
		return fmt.Sprintf("ORDER BY a.title %s", orderDirection)
	}
}

// enrichEpisodesWithDuration fetches episode durations from Jikan API and updates the database
func (s *AnimeService) enrichEpisodesWithDuration(animeID int, malAnimeID string) error {
	// Fetch episodes from Jikan API
	jikanEpisodes, err := s.jikanClient.GetAnimeEpisodes(malAnimeID)
	if err != nil {
		return fmt.Errorf("failed to fetch episodes from Jikan API: %w", err)
	}

	if len(jikanEpisodes) == 0 {
		s.logger.Debug("No episodes found in Jikan API", "mal_anime_id", malAnimeID)
		return nil
	}

	// Get existing episodes from database
	rows, err := s.db.Query(`
		SELECT id, episode_number, duration_minutes, air_date
		FROM episodes
		WHERE anime_id = $1
		ORDER BY episode_number
	`, animeID)
	if err != nil {
		return fmt.Errorf("failed to get episodes from database: %w", err)
	}
	defer rows.Close()

	// Create a map of episode number to episode ID
	episodeMap := make(map[int]struct {
		id              int
		durationMinutes *int
		airDate         *time.Time
	})
	for rows.Next() {
		var id, episodeNumber int
		var durationMinutes *int
		var airDate *time.Time
		if err := rows.Scan(&id, &episodeNumber, &durationMinutes, &airDate); err != nil {
			continue
		}
		episodeMap[episodeNumber] = struct {
			id              int
			durationMinutes *int
			airDate         *time.Time
		}{id, durationMinutes, airDate}
	}

	// Update episodes with duration from Jikan API
	updatedCount := 0
	for _, jikanEp := range jikanEpisodes {
		// Use mal_id as episode number (Jikan API's mal_id corresponds to episode number)
		episodeNumber := jikanEp.MalID

		if episodeNumber <= 0 {
			s.logger.Debug("Invalid episode number from Jikan API",
				"mal_id", jikanEp.MalID,
				"mal_anime_id", malAnimeID)
			continue
		}

		// Check if episode exists in database
		dbEpisode, exists := episodeMap[episodeNumber]
		if !exists {
			s.logger.Debug("Episode not found in database",
				"episode_number", episodeNumber,
				"mal_anime_id", malAnimeID)
			continue
		}

		// Check if we need to update this episode
		// Skip only if both duration and air_date already exist
		hasDuration := dbEpisode.durationMinutes != nil
		hasAirDate := dbEpisode.airDate != nil

		if hasDuration && hasAirDate {
			// Both fields are already set, skip this episode
			continue
		}

		// Parse air_date from Jikan API
		// Only update if not already set
		var airDate *time.Time
		if dbEpisode.airDate == nil && jikanEp.Aired != nil && *jikanEp.Aired != "" {
			// Try ISO 8601 format first (e.g., "2025-07-06T00:00:00+00:00")
			parsedDate, err := time.Parse(time.RFC3339, *jikanEp.Aired)
			if err != nil {
				// Try simple date format (e.g., "2025-07-06")
				parsedDate, err = time.Parse("2006-01-02", *jikanEp.Aired)
			}
			if err == nil {
				airDate = &parsedDate
			} else {
				s.logger.Debug("Failed to parse air_date from Jikan API",
					"episode_number", episodeNumber,
					"aired", *jikanEp.Aired,
					"error", err)
			}
		}

		// Validate duration (should be between 1 and 180 minutes)
		// Only update if not already set
		var duration *int
		if dbEpisode.durationMinutes == nil {
			if jikanEp.Duration >= 1 && jikanEp.Duration <= 180 {
				duration = &jikanEp.Duration
			} else if jikanEp.Duration != 0 {
				s.logger.Warn("Invalid duration value from Jikan API",
					"episode_number", episodeNumber,
					"duration", jikanEp.Duration,
					"mal_anime_id", malAnimeID)
			}
		}

		// Update episode duration and air_date
		err = s.updateEpisodeMetadata(dbEpisode.id, duration, airDate)
		if err != nil {
			s.logger.Warn("Failed to update episode metadata",
				"episode_id", dbEpisode.id,
				"episode_number", episodeNumber,
				"error", err)
			continue
		}

		updatedCount++
	}

	s.logger.Info("Successfully enriched episodes with metadata",
		"anime_id", animeID,
		"mal_anime_id", malAnimeID,
		"updated_count", updatedCount,
		"total_jikan_episodes", len(jikanEpisodes))

	return nil
}

// updateEpisodeMetadata updates the duration and air_date of a single episode
func (s *AnimeService) updateEpisodeMetadata(episodeID int, duration *int, airDate *time.Time) error {
	// Build dynamic UPDATE query based on which fields need updating
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if duration != nil {
		updates = append(updates, fmt.Sprintf("duration_minutes = $%d", argIndex))
		args = append(args, *duration)
		argIndex++
	}

	if airDate != nil {
		updates = append(updates, fmt.Sprintf("air_date = $%d", argIndex))
		args = append(args, *airDate)
		argIndex++
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, episodeID)

	query := fmt.Sprintf(`
		UPDATE episodes
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), argIndex)

	_, err := s.db.Exec(query, args...)
	return err
}

// SyncEpisodeDurations synchronizes episode durations for a specific anime or all animes
func (s *AnimeService) SyncEpisodeDurations(animeID *int) (int, int, int, int, []string, error) {
	var animes []struct {
		id         int
		malAnimeID string
	}

	var totalAnimes int
	var skippedCount int

	// Get total anime count and skipped count
	if animeID != nil {
		// Sync specific anime - check if it has MAL ID
		var malAnimeID sql.NullString
		err := s.db.QueryRow(`
			SELECT mal_anime_id
			FROM animes
			WHERE id = $1
		`, *animeID).Scan(&malAnimeID)

		if err != nil {
			if err == sql.ErrNoRows {
				return 0, 0, 0, 0, nil, ErrAnimeNotFound
			}
			return 0, 0, 0, 0, nil, fmt.Errorf("failed to get anime: %w", err)
		}

		totalAnimes = 1

		// If no MAL ID, skip this anime
		if !malAnimeID.Valid || malAnimeID.String == "" {
			skippedCount = 1
			s.logger.Info("Anime skipped (no MAL ID)",
				"anime_id", *animeID)
			return totalAnimes, 0, 0, skippedCount, nil, nil
		}

		// Add to sync list
		animes = append(animes, struct {
			id         int
			malAnimeID string
		}{*animeID, malAnimeID.String})
	} else {
		// Sync all animes - get total count
		err := s.db.QueryRow(`SELECT COUNT(*) FROM animes`).Scan(&totalAnimes)
		if err != nil {
			return 0, 0, 0, 0, nil, fmt.Errorf("failed to count animes: %w", err)
		}

		// Get animes with MAL ID
		rows, err := s.db.Query(`
			SELECT id, mal_anime_id
			FROM animes
			WHERE mal_anime_id IS NOT NULL AND mal_anime_id != ''
			ORDER BY id
		`)
		if err != nil {
			return 0, 0, 0, 0, nil, fmt.Errorf("failed to get animes: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var malAnimeID string
			if err := rows.Scan(&id, &malAnimeID); err != nil {
				continue
			}
			animes = append(animes, struct {
				id         int
				malAnimeID string
			}{id, malAnimeID})
		}

		// Calculate skipped count (animes without MAL ID)
		skippedCount = totalAnimes - len(animes)
	}

	// Sync each anime
	successCount := 0
	failureCount := 0
	errors := []string{}

	for _, anime := range animes {
		err := s.enrichEpisodesWithDuration(anime.id, anime.malAnimeID)
		if err != nil {
			failureCount++
			errorMsg := fmt.Sprintf("Anime ID %d: %s", anime.id, err.Error())
			errors = append(errors, errorMsg)
			s.logger.Warn("Failed to sync episode durations",
				"anime_id", anime.id,
				"mal_anime_id", anime.malAnimeID,
				"error", err)
		} else {
			successCount++
		}
	}

	s.logger.Info("Episode duration sync completed",
		"total_animes", totalAnimes,
		"success_count", successCount,
		"failure_count", failureCount,
		"skipped_count", skippedCount)

	return totalAnimes, successCount, failureCount, skippedCount, errors, nil
}

// CreateEpisodesFromJikan creates episodes for an anime using Jikan API as the primary source
// This provides more accurate episode numbering and filler/recap flags
func (s *AnimeService) CreateEpisodesFromJikan(animeID int, malAnimeID string) (int, error) {
	if malAnimeID == "" {
		return 0, fmt.Errorf("mal_anime_id is required")
	}

	// Fetch episodes from Jikan API
	jikanEpisodes, err := s.jikanClient.GetAnimeEpisodes(malAnimeID)
	if err != nil {
		return 0, fmt.Errorf("failed to get episodes from Jikan: %w", err)
	}

	if len(jikanEpisodes) == 0 {
		s.logger.Debug("No episodes found in Jikan API", "anime_id", animeID, "mal_anime_id", malAnimeID)
		return 0, nil
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing episodes for this anime (to recreate with accurate data)
	_, err = tx.Exec("DELETE FROM episodes WHERE anime_id = $1", animeID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete existing episodes: %w", err)
	}

	// Insert episodes from Jikan API
	// All inserts must succeed for atomicity - if any fails, rollback entire transaction
	insertedCount := 0
	for _, ep := range jikanEpisodes {
		// Skip episodes with invalid episode number (0 or negative)
		// Some anime have episode 0 as a special/preview which violates DB constraint
		if ep.MalID <= 0 {
			s.logger.Debug("Skipping episode with invalid number",
				"anime_id", animeID,
				"episode_number", ep.MalID,
				"title", ep.Title)
			continue
		}

		// Use Japanese title if available, otherwise fall back to English title
		title := ep.Title
		if ep.TitleJapanese != nil && *ep.TitleJapanese != "" {
			title = *ep.TitleJapanese
		}

		// Determine if this is a special episode based on filler/recap flags
		isSpecial := ep.Filler || ep.Recap

		// Parse air date if available
		var airDate *time.Time
		if ep.Aired != nil && *ep.Aired != "" {
			if parsed, err := time.Parse(time.RFC3339, *ep.Aired); err == nil {
				airDate = &parsed
			}
		}

		// Insert episode - fail entire transaction if any insert fails
		_, err := tx.Exec(`
			INSERT INTO episodes (
				anime_id, episode_number, title, 
				is_special, filler, recap, 
				duration_minutes, air_date,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, animeID, ep.MalID, title, isSpecial, ep.Filler, ep.Recap, ep.Duration, airDate)

		if err != nil {
			// Return error to trigger rollback via defer, preserving original episode data
			return 0, fmt.Errorf("failed to insert episode %d: %w", ep.MalID, err)
		}
		insertedCount++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Created episodes from Jikan API",
		"anime_id", animeID,
		"mal_anime_id", malAnimeID,
		"episode_count", insertedCount)

	return insertedCount, nil
}

// SyncEpisodesFromJikan synchronizes episodes for all animes using Jikan API
// This replaces the Annict-based episode creation with more accurate Jikan data
func (s *AnimeService) SyncEpisodesFromJikan(animeID *int) (int, int, []string, error) {
	const (
		progressLogInterval = 50
		rateLimitWait       = 400 * time.Millisecond
	)

	// Get animes with MAL ID
	query := `
		SELECT a.id, a.mal_anime_id, a.title
		FROM animes a
		WHERE a.mal_anime_id IS NOT NULL AND a.mal_anime_id != ''
	`
	args := []interface{}{}

	if animeID != nil {
		query += " AND a.id = $1"
		args = append(args, *animeID)
	}

	query += " ORDER BY a.id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to query animes: %w", err)
	}
	defer rows.Close()

	type animeInfo struct {
		ID         int
		MalAnimeID string
		Title      string
	}
	var animes []animeInfo

	for rows.Next() {
		var anime animeInfo
		if err := rows.Scan(&anime.ID, &anime.MalAnimeID, &anime.Title); err != nil {
			continue
		}
		animes = append(animes, anime)
	}

	s.logger.Info("Starting episode sync from Jikan API", "total_animes", len(animes))

	successCount := 0
	errorCount := 0
	errorMessages := []string{}

	for i, anime := range animes {
		// Log progress
		if i > 0 && i%progressLogInterval == 0 {
			s.logger.Info("Episode sync progress",
				"processed", i,
				"total", len(animes),
				"success", successCount,
				"errors", errorCount)
		}

		// Create episodes from Jikan API
		_, err := s.CreateEpisodesFromJikan(anime.ID, anime.MalAnimeID)
		if err != nil {
			errorCount++
			errMsg := fmt.Sprintf("anime_id=%d (%s): %v", anime.ID, anime.Title, err)
			errorMessages = append(errorMessages, errMsg)
			s.logger.Warn("Failed to sync episodes",
				"anime_id", anime.ID,
				"title", anime.Title,
				"error", err)
		} else {
			successCount++
		}

		// Rate limiting
		time.Sleep(rateLimitWait)
	}

	s.logger.Info("Episode sync from Jikan API completed",
		"total_animes", len(animes),
		"success", successCount,
		"errors", errorCount)

	return successCount, errorCount, errorMessages, nil
}

// SyncGenresForAllAnimes syncs genre information for all animes that have MAL ID
// Only syncs animes that haven't been synced in the last 30 days to reduce API calls
// Returns: success count, failure count, error messages, error
func (s *AnimeService) SyncGenresForAllAnimes() (int, int, []string, error) {
	return s.syncGenresForAnimesWithFilter(30, 0) // 30 days filter, no limit
}

// SyncGenresWithLimit syncs genre information for a limited number of animes
// Useful for batch processing to avoid API rate limits
// Returns: success count, failure count, error messages, error
func (s *AnimeService) SyncGenresWithLimit(limit int) (int, int, []string, error) {
	return s.syncGenresForAnimesWithFilter(30, limit) // 30 days filter, with limit
}

// syncGenresForAnimesWithFilter syncs genres for animes that haven't been synced in the specified number of days
// If daysThreshold is 0, syncs all animes regardless of last sync time
// If limit > 0, only syncs up to that many animes (for batch processing)
func (s *AnimeService) syncGenresForAnimesWithFilter(daysThreshold int, limit int) (int, int, []string, error) {
	s.logger.Info("Starting genre sync for animes",
		"days_threshold", daysThreshold,
		"limit", limit)

	// Build query to get animes with MAL ID that need genre sync
	query := `
		SELECT id, mal_anime_id
		FROM animes
		WHERE mal_anime_id IS NOT NULL AND mal_anime_id != ''
	`
	args := []interface{}{}
	argIndex := 1

	// Add filter for animes that haven't been synced recently
	if daysThreshold > 0 {
		query += fmt.Sprintf(`
		AND (genres_synced_at IS NULL OR genres_synced_at < CURRENT_TIMESTAMP - ($%d || ' days')::INTERVAL)
		`, argIndex)
		args = append(args, daysThreshold)
		argIndex++
	}

	query += " ORDER BY id"

	// Add limit if specified
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to query animes: %w", err)
	}
	defer rows.Close()

	successCount := 0
	failureCount := 0
	var errorMessages []string
	processedCount := 0
	logInterval := 100 // Log progress every 100 animes
	consecutiveErrors := 0
	maxConsecutiveErrors := 10 // Stop if 10 consecutive errors occur

	// Stream processing: process each anime one by one without loading all into memory
	for rows.Next() {
		var animeID int
		var malAnimeID string

		if err := rows.Scan(&animeID, &malAnimeID); err != nil {
			s.logger.Error("Failed to scan anime row", "error", err)
			failureCount++
			consecutiveErrors++
			errorMessages = append(errorMessages, fmt.Sprintf("Anime scan error: %s", err.Error()))
		} else {
			// Sync genres for this anime
			err := s.syncGenresForAnime(animeID, malAnimeID)
			if err != nil {
				failureCount++
				consecutiveErrors++
				errorMsg := fmt.Sprintf("Anime ID %d (MAL ID %s): %s", animeID, malAnimeID, err.Error())
				errorMessages = append(errorMessages, errorMsg)
				s.logger.Warn("Failed to sync genres for anime",
					"anime_id", animeID,
					"mal_anime_id", malAnimeID,
					"error", err,
					"consecutive_errors", consecutiveErrors)
			} else {
				successCount++
				consecutiveErrors = 0 // Reset on success
			}

			processedCount++

			// Rate limiting is handled by AniListClient's internal rate limiter
			// No need for manual sleep here

			// Log progress periodically
			if processedCount%logInterval == 0 {
				s.logger.Info("Genre sync progress",
					"processed", processedCount,
					"success", successCount,
					"failure", failureCount,
					"error_rate", fmt.Sprintf("%.2f%%", float64(failureCount)/float64(processedCount)*100))
			}
		}

		// Check for too many consecutive errors (circuit breaker)
		if consecutiveErrors >= maxConsecutiveErrors {
			s.logger.Error("Too many consecutive errors, stopping genre sync",
				"consecutive_errors", consecutiveErrors,
				"processed", processedCount)
			break
		}
	}

	if err := rows.Err(); err != nil {
		return 0, 0, nil, fmt.Errorf("error iterating anime rows: %w", err)
	}

	s.logger.Info("Genre sync completed",
		"success_count", successCount,
		"failure_count", failureCount,
		"total_processed", processedCount,
		"days_threshold", daysThreshold,
		"limit", limit)

	return successCount, failureCount, errorMessages, nil
}

// syncGenresForAnime syncs genre information for a single anime using transaction
// Uses AniList API for genre information (recommended over Jikan API)
func (s *AnimeService) syncGenresForAnime(animeID int, malAnimeID string) error {
	// Use background context for batch processing
	// In a real-world scenario, you might want to pass context from the caller
	ctx := context.Background()

	// Fetch genres from AniList API (returns Japanese genre names)
	genres, err := s.anilistClient.GetGenresByMALId(ctx, malAnimeID)
	if err != nil {
		return fmt.Errorf("failed to fetch genres from AniList: %w", err)
	}

	// Save genres to database using shared function
	if err := saveGenresToDatabase(s.db, s.logger, animeID, genres); err != nil {
		return err
	}

	s.logger.Debug("Synced genres for anime",
		"anime_id", animeID,
		"mal_anime_id", malAnimeID,
		"genre_count", len(genres))

	return nil
}

// SyncEpisodeDetailsForAllAnimes syncs episode details (air_date and duration) for all animes
// Returns: success count, failure count, error messages, error
func (s *AnimeService) SyncEpisodeDetailsForAllAnimes() (int, int, []string, error) {
	s.logger.Info("Starting episode details sync for all animes")

	// Get total count for progress reporting (only animes with episodes missing duration or air_date)
	countQuery := `
		SELECT COUNT(DISTINCT a.id)
		FROM animes a
		INNER JOIN episodes e ON a.id = e.anime_id
		WHERE a.mal_anime_id IS NOT NULL AND a.mal_anime_id != ''
		AND (e.duration_minutes IS NULL OR e.air_date IS NULL)
	`
	var totalCount int
	err := s.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to count animes: %w", err)
	}

	s.logger.Info("Starting episode details sync", "total_animes", totalCount)

	// Stream processing: get animes with MAL ID that have episodes missing duration or air_date
	query := `
		SELECT DISTINCT a.id, a.mal_anime_id
		FROM animes a
		INNER JOIN episodes e ON a.id = e.anime_id
		WHERE a.mal_anime_id IS NOT NULL AND a.mal_anime_id != ''
		AND (e.duration_minutes IS NULL OR e.air_date IS NULL)
		ORDER BY a.id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to query animes: %w", err)
	}
	defer rows.Close()

	successCount := 0
	failureCount := 0
	var errorMessages []string
	processedCount := 0
	logInterval := 100 // Log progress every 100 animes

	// Stream processing: process each anime one by one without loading all into memory
	for rows.Next() {
		var animeID int
		var malAnimeID string

		if err := rows.Scan(&animeID, &malAnimeID); err != nil {
			s.logger.Error("Failed to scan anime row", "error", err)
			failureCount++
			errorMsg := fmt.Sprintf("Anime scan error: %s", err.Error())
			errorMessages = append(errorMessages, errorMsg)
			processedCount++
			continue
		}

		// Sync episode details for this anime
		err := s.syncEpisodeDetailsForAnime(animeID, malAnimeID)
		if err != nil {
			failureCount++
			errorMsg := fmt.Sprintf("Anime ID %d (MAL ID %s): %s", animeID, malAnimeID, err.Error())
			errorMessages = append(errorMessages, errorMsg)
			s.logger.Warn("Failed to sync episode details for anime",
				"anime_id", animeID,
				"mal_anime_id", malAnimeID,
				"error", err)
		} else {
			successCount++
		}

		processedCount++

		// Log progress periodically
		if processedCount%logInterval == 0 {
			s.logger.Info("Episode details sync progress",
				"processed", processedCount,
				"total", totalCount,
				"success", successCount,
				"failure", failureCount,
				"progress", fmt.Sprintf("%.1f%%", float64(processedCount)/float64(totalCount)*100))
		}
	}

	if err := rows.Err(); err != nil {
		return 0, 0, nil, fmt.Errorf("error iterating anime rows: %w", err)
	}

	s.logger.Info("Episode details sync completed",
		"success_count", successCount,
		"failure_count", failureCount,
		"total_processed", processedCount,
		"total_animes", totalCount)

	return successCount, failureCount, errorMessages, nil
}

// syncEpisodeDetailsForAnime syncs episode details (air_date and duration) for a single anime
// Uses a single transaction to ensure atomicity and batch updates for better performance
func (s *AnimeService) syncEpisodeDetailsForAnime(animeID int, malAnimeID string) error {
	// Fetch episodes from Jikan API
	episodes, err := s.jikanClient.GetAnimeEpisodes(malAnimeID)
	if err != nil {
		if errors.Is(err, ErrJikanNotFound) {
			s.logger.Debug("Anime not found in Jikan", "anime_id", animeID, "mal_anime_id", malAnimeID)
			return nil // Not an error, just no data available
		}
		return fmt.Errorf("failed to fetch episodes from Jikan: %w", err)
	}

	if len(episodes) == 0 {
		return nil // No episodes to sync
	}

	// Begin transaction for atomic updates
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Prepare update statement for better performance
	updateQuery := `
		UPDATE episodes
		SET duration_minutes = COALESCE($1, duration_minutes),
		    air_date = COALESCE($2, air_date),
		    updated_at = CURRENT_TIMESTAMP
		WHERE anime_id = $3 AND episode_number = $4
	`
	stmt, err := tx.Prepare(updateQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	updatedCount := 0
	for _, jikanEp := range episodes {
		// Use mal_id as episode number (Jikan API's mal_id corresponds to episode number)
		episodeNumber := jikanEp.MalID

		if episodeNumber <= 0 {
			s.logger.Debug("Invalid episode number from Jikan API",
				"anime_id", animeID,
				"mal_id", jikanEp.MalID)
			continue
		}

		// Prepare duration (nullable)
		var duration *int
		if jikanEp.Duration > 0 {
			duration = &jikanEp.Duration
		}

		// Parse air_date to time.Time (nullable)
		var airDate *time.Time
		if jikanEp.Aired != nil && *jikanEp.Aired != "" {
			// Try ISO 8601 format first (e.g., "2025-07-06T00:00:00+00:00")
			parsedDate, err := time.Parse(time.RFC3339, *jikanEp.Aired)
			if err != nil {
				// Try simple date format (e.g., "2025-07-06")
				parsedDate, err = time.Parse("2006-01-02", *jikanEp.Aired)
			}
			if err == nil {
				airDate = &parsedDate
			} else {
				s.logger.Debug("Failed to parse air_date",
					"anime_id", animeID,
					"episode_number", episodeNumber,
					"aired", *jikanEp.Aired,
					"error", err)
			}
		}

		// Execute update with prepared statement
		_, err = stmt.Exec(duration, airDate, animeID, episodeNumber)
		if err != nil {
			s.logger.Warn("Failed to update episode details",
				"anime_id", animeID,
				"episode_number", episodeNumber,
				"error", err)
			continue
		}

		updatedCount++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("Synced episode details for anime",
		"anime_id", animeID,
		"mal_anime_id", malAnimeID,
		"updated_count", updatedCount,
		"total_episodes", len(episodes))

	return nil
}

// SyncCoverImagesForAllAnimes - 全アニメのカバー画像をAniListから事前取得してDBにキャッシュ
// スケジューラーから呼び出され、画像未取得のアニメに対してバッチ処理を行う
// 戻り値: 成功数, 失敗数, エラーメッセージリスト, エラー
func (s *AnimeService) SyncCoverImagesForAllAnimes() (int, int, []string, error) {
	// 画像未取得のアニメを取得（MAL IDがあるもののみ）
	rows, err := s.db.Query(`
		SELECT id, title, mal_anime_id
		FROM animes
		WHERE mal_anime_id IS NOT NULL
		  AND mal_anime_id != ''
		  AND (cover_image_large IS NULL OR cover_image_large = '')
		ORDER BY watchers_count DESC NULLS LAST
	`)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to query animes without cover images: %w", err)
	}
	defer rows.Close()

	type animeToSync struct {
		ID    int
		Title string
		MalID string
	}

	var animesToSync []animeToSync
	for rows.Next() {
		var a animeToSync
		if err := rows.Scan(&a.ID, &a.Title, &a.MalID); err != nil {
			s.logger.Warn("Failed to scan anime for cover image sync", "error", err)
			continue
		}
		animesToSync = append(animesToSync, a)
	}

	if len(animesToSync) == 0 {
		s.logger.Info("No animes need cover image sync")
		return 0, 0, nil, nil
	}

	s.logger.Info("Starting cover image sync", "total_animes", len(animesToSync))

	// 結果を集約するための構造体
	type syncResult struct {
		success bool
		errMsg  string
	}

	results := make(chan syncResult, len(animesToSync))
	ctx := context.Background()

	// ワーカープールで並列処理（AniListClientの内部レートリミッターが制御）
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, constants.CoverImageSyncWorkerPoolSize)

	for i, anime := range animesToSync {
		// 進捗ログ（100件ごと）
		if (i+1)%100 == 0 {
			s.logger.Info("Cover image sync progress",
				"queued", i+1,
				"total", len(animesToSync))
		}

		wg.Add(1)
		go func(anime animeToSync) {
			defer wg.Done()
			semaphore <- struct{}{}        // セマフォ取得
			defer func() { <-semaphore }() // セマフォ解放

			// AniListから画像を取得
			coverImage, apiErr := s.anilistClient.GetCoverImageByMALId(ctx, anime.MalID)

			// MAL IDで取得失敗またはnot foundの場合、タイトルでフォールバック
			if apiErr != nil || coverImage == nil {
				coverImage, apiErr = s.anilistClient.GetCoverImageByTitle(ctx, anime.Title)
			}

			if apiErr != nil || coverImage == nil {
				errMsg := fmt.Sprintf("anime_id=%d (%s): failed to get cover image: %v", anime.ID, anime.Title, apiErr)
				results <- syncResult{success: false, errMsg: errMsg}
				return
			}

			// DBに保存
			_, saveErr := s.db.Exec(`
				UPDATE animes 
				SET cover_image_extra_large = $1,
					cover_image_large = $2,
					cover_image_medium = $3,
					cover_image_updated_at = NOW()
				WHERE id = $4
			`, coverImage.ExtraLarge, coverImage.Large, coverImage.Medium, anime.ID)

			if saveErr != nil {
				errMsg := fmt.Sprintf("anime_id=%d (%s): failed to save to DB: %v", anime.ID, anime.Title, saveErr)
				results <- syncResult{success: false, errMsg: errMsg}
				return
			}

			results <- syncResult{success: true}
		}(anime)
	}

	// 全ワーカーの完了を待ってからチャネルを閉じる
	go func() {
		wg.Wait()
		close(results)
	}()

	// 結果を集計
	successCount := 0
	failureCount := 0
	var errors []string

	for result := range results {
		if result.success {
			successCount++
		} else {
			failureCount++
			if len(errors) < constants.BatchSyncMaxErrors {
				errors = append(errors, result.errMsg)
			}
		}
	}

	s.logger.Info("Cover image sync completed",
		"success", successCount,
		"failure", failureCount,
		"total", len(animesToSync))

	return successCount, failureCount, errors, nil
}

// GetSearchHistory - ユーザーの検索履歴を取得
func (s *AnimeService) GetSearchHistory(userID int, limit int) ([]models.SearchLog, error) {
	if limit <= 0 {
		limit = constants.SearchHistoryDefaultLimit
	}
	if limit > constants.SearchHistoryMaxItems {
		limit = constants.SearchHistoryMaxItems
	}

	rows, err := s.db.Query(`
		SELECT id, user_id, query, COALESCE(filters, '{}') as filters, result_count, clicked_anime_id, searched_at
		FROM search_logs
		WHERE user_id = $1
		ORDER BY searched_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get search history: %w", err)
	}
	defer rows.Close()

	history := []models.SearchLog{}
	for rows.Next() {
		var log models.SearchLog
		err := rows.Scan(&log.ID, &log.UserID, &log.Query, &log.Filters, &log.ResultCount, &log.ClickedID, &log.SearchedAt)
		if err != nil {
			s.logger.Warn("Failed to scan search log row", "error", err)
			continue
		}
		history = append(history, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search history: %w", err)
	}

	return history, nil
}

// DeleteSearchHistory - ユーザーの検索履歴を削除
func (s *AnimeService) DeleteSearchHistory(userID int) error {
	_, err := s.db.Exec(`
		DELETE FROM search_logs
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete search history: %w", err)
	}

	return nil
}

// DeleteSearchHistoryItem - 特定の検索履歴を削除
func (s *AnimeService) DeleteSearchHistoryItem(userID int, logID int) error {
	result, err := s.db.Exec(`
		DELETE FROM search_logs
		WHERE id = $1 AND user_id = $2
	`, logID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete search history item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSearchHistoryNotFound
	}

	return nil
}

// CleanupOldSearchHistory - 古い検索履歴を削除（最新N件を残す）
// ROW_NUMBER()ウィンドウ関数を使用して効率的に削除
func (s *AnimeService) CleanupOldSearchHistory(userID int) error {
	_, err := s.db.Exec(`
		DELETE FROM search_logs
		WHERE id IN (
			SELECT id FROM (
				SELECT id, ROW_NUMBER() OVER (ORDER BY searched_at DESC) as rn
				FROM search_logs
				WHERE user_id = $1
			) ranked
			WHERE rn > $2
		)
	`, userID, constants.SearchHistoryMaxItems)
	if err != nil {
		return fmt.Errorf("failed to cleanup old search history: %w", err)
	}

	return nil
}

// GetEpisodesByType returns episodes filtered by type (regular or special)
// N+1クエリ問題を解消: 視聴ステータスを一括取得
func (s *AnimeService) GetEpisodesByType(animeID int, isSpecial bool, userID *int) ([]models.Episode, error) {
	rows, err := s.db.Query(`
		SELECT id, anime_id, episode_number, title, duration_minutes, air_date,
		       COALESCE(is_special, FALSE) as is_special, 
		       COALESCE(filler, FALSE) as filler, 
		       COALESCE(recap, FALSE) as recap,
		       created_at, updated_at
		FROM episodes
		WHERE anime_id = $1 AND COALESCE(is_special, FALSE) = $2
		ORDER BY episode_number ASC, is_special ASC
	`, animeID, isSpecial)
	if err != nil {
		return nil, fmt.Errorf("failed to query episodes: %w", err)
	}
	defer rows.Close()

	var episodes []models.Episode
	var episodeIDs []int
	for rows.Next() {
		var ep models.Episode
		err := rows.Scan(
			&ep.ID, &ep.AnimeID, &ep.EpisodeNumber, &ep.Title,
			&ep.DurationMinutes, &ep.AirDate, &ep.IsSpecial, &ep.Filler, &ep.Recap,
			&ep.CreatedAt, &ep.UpdatedAt,
		)
		if err != nil {
			s.logger.Warn("Failed to scan episode", "error", err)
			continue
		}
		episodes = append(episodes, ep)
		episodeIDs = append(episodeIDs, ep.ID)
	}

	// 視聴ステータスを一括取得（N+1クエリ解消）
	if userID != nil && len(episodeIDs) > 0 {
		viewedMap, err := s.getWatchStatusesBatch(*userID, episodeIDs)
		if err != nil {
			s.logger.Warn("Failed to get watch statuses batch", "error", err)
		} else {
			for i := range episodes {
				if viewed, ok := viewedMap[episodes[i].ID]; ok {
					episodes[i].Viewed = viewed
				}
			}
		}
	}

	return episodes, nil
}

// getWatchStatusesBatch returns watch statuses for multiple episodes in a single query
func (s *AnimeService) getWatchStatusesBatch(userID int, episodeIDs []int) (map[int]bool, error) {
	if len(episodeIDs) == 0 {
		return make(map[int]bool), nil
	}

	// プレースホルダーを動的に生成
	placeholders := make([]string, len(episodeIDs))
	args := make([]interface{}, len(episodeIDs)+1)
	args[0] = userID
	for i, id := range episodeIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		SELECT episode_id, viewed 
		FROM watch_statuses 
		WHERE user_id = $1 AND episode_id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query watch statuses: %w", err)
	}
	defer rows.Close()

	viewedMap := make(map[int]bool)
	for rows.Next() {
		var episodeID int
		var viewed bool
		if err := rows.Scan(&episodeID, &viewed); err != nil {
			continue
		}
		viewedMap[episodeID] = viewed
	}

	return viewedMap, nil
}

// GetEpisodesSeparated returns episodes separated by type (regular and special)
// 1回のクエリで全エピソードを取得し、メモリ上で分離（効率化）
func (s *AnimeService) GetEpisodesSeparated(animeID int, userID *int) (regular, specials []models.Episode, err error) {
	rows, err := s.db.Query(`
		SELECT id, anime_id, episode_number, title, duration_minutes, air_date,
		       COALESCE(is_special, FALSE) as is_special, 
		       COALESCE(filler, FALSE) as filler, 
		       COALESCE(recap, FALSE) as recap,
		       created_at, updated_at
		FROM episodes
		WHERE anime_id = $1
		ORDER BY episode_number ASC, is_special ASC
	`, animeID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query episodes: %w", err)
	}
	defer rows.Close()

	var allEpisodes []models.Episode
	var episodeIDs []int
	for rows.Next() {
		var ep models.Episode
		err := rows.Scan(
			&ep.ID, &ep.AnimeID, &ep.EpisodeNumber, &ep.Title,
			&ep.DurationMinutes, &ep.AirDate, &ep.IsSpecial, &ep.Filler, &ep.Recap,
			&ep.CreatedAt, &ep.UpdatedAt,
		)
		if err != nil {
			s.logger.Warn("Failed to scan episode", "error", err)
			continue
		}
		allEpisodes = append(allEpisodes, ep)
		episodeIDs = append(episodeIDs, ep.ID)
	}

	// 視聴ステータスを一括取得（N+1クエリ解消）
	if userID != nil && len(episodeIDs) > 0 {
		viewedMap, err := s.getWatchStatusesBatch(*userID, episodeIDs)
		if err != nil {
			s.logger.Warn("Failed to get watch statuses batch", "error", err)
		} else {
			for i := range allEpisodes {
				if viewed, ok := viewedMap[allEpisodes[i].ID]; ok {
					allEpisodes[i].Viewed = viewed
				}
			}
		}
	}

	// メモリ上で通常エピソードと特別編に分離
	for _, ep := range allEpisodes {
		if ep.IsSpecial {
			specials = append(specials, ep)
		} else {
			regular = append(regular, ep)
		}
	}

	return regular, specials, nil
}

// SyncDescriptionsFromTMDB syncs Japanese descriptions from TMDB for all animes without description
// Also updates animes with English descriptions (detected by HTML tags or lack of Japanese characters)
// Returns: success count, not found count, failure count, error
func (s *AnimeService) SyncDescriptionsFromTMDB() (int, int, int, error) {
	s.logger.Info("Starting TMDB description sync")

	// Check if TMDB client is enabled
	if s.tmdbClient == nil || !s.tmdbClient.IsEnabled() {
		return 0, 0, 0, fmt.Errorf("TMDB client is not configured")
	}

	// Get animes that need description update:
	// 1. No description (NULL or empty)
	// 2. English description (contains HTML tags like <i>, <br>, or starts with ASCII and has no Japanese punctuation)
	rows, err := s.db.Query(`
		SELECT id, title, mal_anime_id
		FROM animes
		WHERE description IS NULL
		   OR description = ''
		   OR description LIKE '%<i>%'
		   OR description LIKE '%<br>%'
		   OR (description ~ '^[A-Za-z]' AND description NOT LIKE '%。%' AND description NOT LIKE '%、%')
		ORDER BY watchers_count DESC NULLS LAST, id
	`)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get animes: %w", err)
	}

	type animeInfo struct {
		ID    int
		Title string
		MalID *string
	}
	var animes []animeInfo

	for rows.Next() {
		var a animeInfo
		if err := rows.Scan(&a.ID, &a.Title, &a.MalID); err != nil {
			s.logger.Warn("Failed to scan anime", "error", err)
			continue
		}
		animes = append(animes, a)
	}
	rows.Close()

	totalAnimes := len(animes)
	if totalAnimes == 0 {
		s.logger.Info("All animes already have descriptions")
		return 0, 0, 0, nil
	}

	s.logger.Info("Found animes without description", "count", totalAnimes)

	// TMDB rate limit: ~40 requests/10 seconds
	const batchSize = 40
	const batchInterval = 11 * time.Second

	successCount := 0
	notFoundCount := 0
	failureCount := 0

	// Process in batches
	for i := 0; i < totalAnimes; i += batchSize {
		end := i + batchSize
		if end > totalAnimes {
			end = totalAnimes
		}
		batch := animes[i:end]

		s.logger.Info("Processing TMDB batch",
			"progress", fmt.Sprintf("%d/%d", end, totalAnimes))

		for _, anime := range batch {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			malID := ""
			if anime.MalID != nil {
				malID = *anime.MalID
			}

			// AniListからメディアタイプを取得（MOVIE, TV等）
			mediaTypeHint := ""
			if malID != "" {
				format, formatErr := s.anilistClient.GetMediaFormatByMALId(ctx, malID)
				if formatErr == nil {
					mediaTypeHint = format
				}
			}

			overview, tmdbID, err := s.tmdbClient.GetOverviewByMALIDOrTitle(ctx, malID, anime.Title, mediaTypeHint)
			cancel()

			if err != nil || overview == "" {
				// TMDBから直接取得できなかった場合、シリーズフォールバックを試行
				seriesTitle := NormalizeAnimeTitle(anime.Title)
				var seriesDesc string
				var seriesTMDBID sql.NullInt64
				seriesErr := s.db.QueryRow(`
					SELECT description, series_tmdb_id
					FROM animes
					WHERE series_title = $1
					  AND description IS NOT NULL
					  AND description != ''
					  AND description NOT LIKE '%<i>%'
					  AND description NOT LIKE '%<br>%'
					ORDER BY series_tmdb_id IS NOT NULL DESC, id ASC
					LIMIT 1
				`, seriesTitle).Scan(&seriesDesc, &seriesTMDBID)

				if seriesErr == nil && seriesDesc != "" {
					// シリーズフォールバックで取得成功
					var seriesTMDBIDVal int
					if seriesTMDBID.Valid {
						seriesTMDBIDVal = int(seriesTMDBID.Int64)
					}
					s.updateAnimeSeriesInfo(context.Background(), anime.ID, seriesDesc, 0, seriesTitle, seriesTMDBIDVal)
					successCount++
					continue
				}

				notFoundCount++
				continue
			}

			// Save to database with series_title and series_tmdb_id
			seriesTitle := NormalizeAnimeTitle(anime.Title)
			s.updateAnimeSeriesInfo(context.Background(), anime.ID, overview, tmdbID, seriesTitle, tmdbID)
			successCount++
		}

		// Wait before next batch (except for last batch)
		if end < totalAnimes {
			time.Sleep(batchInterval)
		}
	}

	s.logger.Info("TMDB description sync completed",
		"success", successCount,
		"not_found", notFoundCount,
		"failure", failureCount)

	return successCount, notFoundCount, failureCount, nil
}

// SyncReleasedOnFromJikan syncs release dates from Jikan API for animes without released_on
// Returns: success count, not found count, failure count, error
func (s *AnimeService) SyncReleasedOnFromJikan() (int, int, int, error) {
	s.logger.Info("Starting Jikan released_on sync")

	// Get animes that need released_on update:
	// 1. No released_on (NULL only - DATE type cannot be empty string)
	// 2. Has MAL ID (required for Jikan API)
	rows, err := s.db.Query(`
		SELECT id, title, mal_anime_id 
		FROM animes 
		WHERE released_on IS NULL
		  AND mal_anime_id IS NOT NULL AND mal_anime_id != ''
		ORDER BY watchers_count DESC NULLS LAST, id
	`)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get animes: %w", err)
	}

	type animeInfo struct {
		ID    int
		Title string
		MalID string
	}
	var animes []animeInfo

	for rows.Next() {
		var a animeInfo
		// SQL query already filters for mal_anime_id IS NOT NULL AND mal_anime_id != ''
		// so we can scan directly into the string field
		if err := rows.Scan(&a.ID, &a.Title, &a.MalID); err != nil {
			s.logger.Warn("Failed to scan anime", "error", err)
			continue
		}
		animes = append(animes, a)
	}
	rows.Close()

	totalAnimes := len(animes)
	if totalAnimes == 0 {
		s.logger.Info("All animes already have released_on")
		return 0, 0, 0, nil
	}

	s.logger.Info("Found animes without released_on", "count", totalAnimes)

	// Jikan rate limit: 3 requests/second
	// Process in batches of 30 with 11 second intervals for safety
	const batchSize = 30
	const batchInterval = 11 * time.Second

	successCount := 0
	notFoundCount := 0
	failureCount := 0

	// Process in batches
	for i := 0; i < totalAnimes; i += batchSize {
		end := i + batchSize
		if end > totalAnimes {
			end = totalAnimes
		}
		batch := animes[i:end]

		s.logger.Info("Processing Jikan batch",
			"progress", fmt.Sprintf("%d/%d", end, totalAnimes))

		for _, anime := range batch {
			releasedOn, err := s.jikanClient.GetAnimeAiredDate(anime.MalID)

			if err != nil {
				if errors.Is(err, ErrJikanNotFound) {
					s.logger.Debug("Anime not found in Jikan",
						"anime_id", anime.ID,
						"mal_id", anime.MalID,
						"title", anime.Title)
					notFoundCount++
				} else {
					s.logger.Debug("Failed to get aired date from Jikan",
						"anime_id", anime.ID,
						"mal_id", anime.MalID,
						"title", anime.Title,
						"error", err)
					failureCount++
				}
				continue
			}

			if releasedOn == "" {
				notFoundCount++
				continue
			}

			// Save to database
			result, err := s.db.Exec(`
				UPDATE animes 
				SET released_on = $1
				WHERE id = $2
			`, releasedOn, anime.ID)

			if err != nil {
				s.logger.Warn("Failed to save released_on",
					"anime_id", anime.ID,
					"error", err)
				failureCount++
				continue
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				successCount++
			}
		}

		// Wait before next batch (except for last batch)
		if end < totalAnimes {
			time.Sleep(batchInterval)
		}
	}

	s.logger.Info("Jikan released_on sync completed",
		"success", successCount,
		"not_found", notFoundCount,
		"failure", failureCount)

	return successCount, notFoundCount, failureCount, nil
}

// SyncPriorityAnimeData syncs episodes, descriptions, and released_on for specific anime IDs.
// Used by priority sync mode to enrich only current season + popular animes.
func (s *AnimeService) SyncPriorityAnimeData(animeIDs []int) error {
	if len(animeIDs) == 0 {
		return nil
	}

	s.logger.Info("Starting priority anime data sync", "anime_count", len(animeIDs))

	// バッチクエリで全アニメ情報を一括取得（N+1問題の回避）
	type animeInfo struct {
		ID          int
		Title       string
		Description *string
		ReleasedOn  *string
		MalAnimeID  *string
	}
	animeMap := make(map[int]*animeInfo, len(animeIDs))

	// IN句用のプレースホルダーを構築
	placeholders := ""
	args := make([]interface{}, len(animeIDs))
	for i, id := range animeIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	rows, err := s.db.Query(fmt.Sprintf(`
		SELECT id, title, description, released_on, mal_anime_id
		FROM animes WHERE id IN (%s)
	`, placeholders), args...)
	if err != nil {
		return fmt.Errorf("failed to batch fetch anime info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var info animeInfo
		var desc, relOn, malID sql.NullString
		if err := rows.Scan(&info.ID, &info.Title, &desc, &relOn, &malID); err != nil {
			s.logger.Warn("Failed to scan anime info", "error", err)
			continue
		}
		if desc.Valid && desc.String != "" {
			info.Description = &desc.String
		}
		if relOn.Valid && relOn.String != "" {
			info.ReleasedOn = &relOn.String
		}
		if malID.Valid && malID.String != "" {
			info.MalAnimeID = &malID.String
		}
		animeMap[info.ID] = &info
	}

	s.logger.Info("Batch fetched anime info", "fetched", len(animeMap))

	// Step 1: Sync episodes from Jikan
	s.logger.Info("Priority sync: episodes from Jikan...")
	epSuccess, epFail := 0, 0
	for _, id := range animeIDs {
		info, ok := animeMap[id]
		if !ok || info.MalAnimeID == nil {
			continue
		}
		_, err := s.CreateEpisodesFromJikan(id, *info.MalAnimeID)
		if err != nil {
			s.logger.Debug("Failed to sync episodes", "anime_id", id, "error", err)
			epFail++
		} else {
			epSuccess++
		}
		time.Sleep(400 * time.Millisecond)
	}
	s.logger.Info("Priority sync: episodes completed", "success", epSuccess, "failure", epFail)

	// Step 2: Sync episode details (duration, air_date)
	s.logger.Info("Priority sync: episode details from Jikan...")
	detailSuccess, detailFail := 0, 0
	for _, id := range animeIDs {
		info, ok := animeMap[id]
		if !ok || info.MalAnimeID == nil {
			continue
		}
		err := s.syncEpisodeDetailsForAnime(id, *info.MalAnimeID)
		if err != nil {
			s.logger.Debug("Failed to sync episode details", "anime_id", id, "error", err)
			detailFail++
		} else {
			detailSuccess++
		}
	}
	s.logger.Info("Priority sync: episode details completed", "success", detailSuccess, "failure", detailFail)

	// Step 3: Sync descriptions from TMDB
	s.logger.Info("Priority sync: descriptions from TMDB...")
	descSuccess, descSkip := 0, 0
	for _, id := range animeIDs {
		info, ok := animeMap[id]
		if !ok {
			continue
		}
		if info.Description != nil && *info.Description != "" {
			descSkip++
			continue
		}
		anime := models.Anime{
			ID:         info.ID,
			Title:      info.Title,
			MalAnimeID: info.MalAnimeID,
		}
		s.enrichAnimeWithDescription(&anime)
		if anime.Description != nil && *anime.Description != "" {
			descSuccess++
			// メモリ上のマップも更新
			info.Description = anime.Description
		}
	}
	s.logger.Info("Priority sync: descriptions completed", "success", descSuccess, "skipped", descSkip)

	// Step 4: Sync released_on from Jikan
	s.logger.Info("Priority sync: released_on from Jikan...")
	relSuccess, relSkip := 0, 0
	for _, id := range animeIDs {
		info, ok := animeMap[id]
		if !ok {
			continue
		}
		if info.ReleasedOn != nil && *info.ReleasedOn != "" {
			relSkip++
			continue
		}
		anime := models.Anime{
			ID:         info.ID,
			Title:      info.Title,
			MalAnimeID: info.MalAnimeID,
		}
		s.enrichAnimeWithReleasedOn(&anime)
		if anime.ReleasedOn != nil && *anime.ReleasedOn != "" {
			relSuccess++
		}
	}
	s.logger.Info("Priority sync: released_on completed", "success", relSuccess, "skipped", relSkip)

	s.logger.Info("Priority anime data sync completed")
	return nil
}

// enrichAnimeWithWatchProviders - TMDBから配信サービス情報を取得してアニメ情報を拡張
// series_tmdb_idがある場合はそれを優先的に使用（シリーズ作品の統合管理）
func (s *AnimeService) enrichAnimeWithWatchProviders(animeWithData *models.AnimeWithUserData) {
	if s.tmdbClient == nil || !s.tmdbClient.IsEnabled() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, try to get TMDB ID and series_tmdb_id from database
	var tmdbID int
	var seriesTMDBID sql.NullInt64
	var dbMediaType sql.NullString
	err := s.db.QueryRow(`SELECT tmdb_id, series_tmdb_id, media FROM animes WHERE id = $1`, animeWithData.ID).Scan(&tmdbID, &seriesTMDBID, &dbMediaType)

	// Determine media type (movie or tv)
	mediaType := "tv" // Default to TV
	malID := ""
	if animeWithData.MalAnimeID != nil {
		malID = *animeWithData.MalAnimeID
	}

	// Use series_tmdb_id if available (for series works)
	if err == nil && seriesTMDBID.Valid && seriesTMDBID.Int64 > 0 {
		tmdbID = int(seriesTMDBID.Int64)
		// series_tmdb_id is always for TV series
		mediaType = "tv"
	} else if err == nil && tmdbID > 0 {
		// Check DB media field first
		if dbMediaType.Valid && dbMediaType.String == "movie" {
			mediaType = "movie"
		} else if animeWithData.Media != nil && *animeWithData.Media == "movie" {
			mediaType = "movie"
		} else if malID != "" {
			// Fall back to AniList
			format, formatErr := s.anilistClient.GetMediaFormatByMALId(ctx, malID)
			if formatErr == nil && format == "MOVIE" {
				mediaType = "movie"
			}
		}

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
		mediaType = foundMediaType

		// Save TMDB ID and series_tmdb_id to database for future use (async)
		animeID := animeWithData.ID
		animeTitle := animeWithData.Title
		s.enqueueBackgroundTask(func() {
			seriesTitle := NormalizeAnimeTitle(animeTitle)
			s.updateAnimeSeriesInfo(context.Background(), animeID, "", tmdbID, seriesTitle, tmdbID)
		})
	}

	// Get watch providers from TMDB (use appropriate endpoint for TV or movie)
	providers, err := s.tmdbClient.GetJapaneseWatchProviders(ctx, tmdbID, mediaType)
	if err != nil {
		s.logger.Debug("Failed to get watch providers from TMDB",
			"anime_id", animeWithData.ID,
			"tmdb_id", tmdbID,
			"media_type", mediaType,
			"error", err)
		return
	}

	if len(providers) == 0 {
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

	animeWithData.WatchProviders = watchProviders
}

// determineStatusBySeason はシーズンと放送開始日を考慮してアニメの放送状態を判定します。
// TMDBの情報が利用できない場合のフォールバックとして使用されます。
func determineStatusBySeason(seasonName string, releasedOn *string) string {
	// released_onが未来の日付なら放送予定
	if releasedOn != nil && *releasedOn != "" {
		// UTCで比較してサーバータイムゾーンの影響を排除
		releaseDate, err := time.ParseInLocation("2006-01-02", *releasedOn, time.UTC)
		if err == nil {
			today := time.Now().UTC().Truncate(24 * time.Hour)
			if releaseDate.After(today) {
				return "upcoming"
			}
		}
	}

	if seasonName == "" {
		return "finished"
	}

	// シーズン名（例: "YYYY-season"）をパース
	var seasonYear int
	var seasonStr string
	_, err := fmt.Sscanf(seasonName, "%d-%s", &seasonYear, &seasonStr)
	if err != nil {
		return "finished" // パース失敗時は完結とみなす
	}

	now := time.Now().UTC()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// シーズンの開始月と終了月を判定
	var seasonStartMonth, seasonEndMonth int
	switch seasonStr {
	case "winter": // 1月, 2月, 3月
		seasonStartMonth = 1
		seasonEndMonth = 3
	case "spring": // 4月, 5月, 6月
		seasonStartMonth = 4
		seasonEndMonth = 6
	case "summer": // 7月, 8月, 9月
		seasonStartMonth = 7
		seasonEndMonth = 9
	case "autumn", "fall": // 10月, 11月, 12月
		seasonStartMonth = 10
		seasonEndMonth = 12
	default:
		return "finished" // 未知のシーズン名は完結とみなす
	}

	// 判定ロジック
	if seasonYear > currentYear {
		// 未来のシーズン → 放送予定
		return "upcoming"
	} else if seasonYear == currentYear {
		// 同じ年のシーズン
		if currentMonth < seasonStartMonth {
			// シーズン開始月より前 → 放送予定
			return "upcoming"
		} else if currentMonth <= seasonEndMonth {
			// シーズン期間中 → 放送中
			return "airing"
		} else {
			// シーズン終了月以後 → 完結
			return "finished"
		}
	} else {
		// 過去のシーズン → 完結
		return "finished"
	}
}

// mapTMDBStatus はTMDBのステータス文字列を内部の放送状態文字列にマッピングします。
func mapTMDBStatus(tmdbStatus string) string {
	switch tmdbStatus {
	case "Returning Series", "In Production":
		return "airing"
	case "Planned":
		return "upcoming"
	case "Ended", "Canceled":
		return "finished"
	default:
		return "finished" // デフォルトは完結とみなす
	}
}
