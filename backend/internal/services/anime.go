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
	ErrSearchHistoryNotFound = errors.New("search history item not found")
)

// animeSelectColumns - アニメ取得用の共通SELECT句
// animes_with_computed_status VIEWを使用してステータスを動的に計算
// VIEWでcover_image_*カラムに対してCOALESCEが適用済み
const animeSelectColumns = `
	id, title, image_url, season, computed_status as status,
	mal_anime_id, annict_id, created_at, updated_at, watchers_count,
	cover_image_extra_large, cover_image_large, cover_image_medium,
	released_on, media
`

// aliasedAnimeSelectColumns - テーブルエイリアス付きアニメ取得用SELECT句
// AdvancedSearchAnimesなどのJOINを含むクエリで使用
// VIEWでcover_image_*カラムに対してCOALESCEが適用済み
const aliasedAnimeSelectColumns = `
	a.id, a.title, a.image_url, a.season, a.computed_status as status,
	a.mal_anime_id, a.annict_id, a.created_at, a.updated_at, a.watchers_count,
	a.cover_image_extra_large, a.cover_image_large, a.cover_image_medium,
	a.released_on, a.media
`

// aliasedAnimeGroupByColumns - テーブルエイリアス付きGROUP BY句
// AdvancedSearchAnimesなどのJOINを含むクエリで使用
const aliasedAnimeGroupByColumns = `
	a.id, a.title, a.image_url, a.season, a.computed_status,
	a.mal_anime_id, a.annict_id, a.created_at, a.updated_at, a.watchers_count,
	a.cover_image_extra_large, a.cover_image_large, a.cover_image_medium, a.released_on, a.media
`

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
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", queryName, err)
	}
	defer rows.Close()

	animes := []models.Anime{}
	for rows.Next() {
		anime, err := scanAnime(rows)
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
	tmdbClient          *TMDBClient   // TMDB APIクライアント
	claudeClient        *ClaudeClient // Claude APIクライアント（あらすじの日本語翻訳・整形）
	annictClient        *AnnictClient // Annict APIクライアント（循環参照回避のため後から設定）
	backgroundTaskQueue chan func()   // バックグラウンドタスク用ワーカープールのキュー
	cache               *MemoryCache  // レスポンスキャッシュ
	redisCache          *RedisCache   // Redisキャッシュ（オプション）
}

// SetAnnictClient sets the AnnictClient on AnimeService (called after both are initialized to avoid circular dependency)
func (s *AnimeService) SetAnnictClient(ac *AnnictClient) {
	s.annictClient = ac
}

// GetAniListClient returns the shared AniListClient instance.
// Used by the scheduler to share the same rate-limited client instead of creating a separate instance.
func (s *AnimeService) GetAniListClient() *AniListClient {
	return s.anilistClient
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
		claudeClient:        NewClaudeClient(logger),
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
		whereClauses = append(whereClauses, fmt.Sprintf("computed_status = $%d", argIndex))
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM animes_with_computed_status %s", whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count animes: %w", err)
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PerPage
	totalPages := (total + params.PerPage - 1) / params.PerPage

	// Get animes（VIEWを使用してステータスを動的に計算）
	query := fmt.Sprintf(`
		SELECT `+animeSelectColumns+`
		FROM animes_with_computed_status %s %s
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
		anime, err := scanAnime(rows)
		if err != nil {
			s.logger.Warn("Failed to scan anime row in SearchAnimes", "error", err)
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

	// Get animes（VIEWを使用してステータスを動的に計算）
	query := fmt.Sprintf(`
		SELECT %s
		%s %s
		GROUP BY %s
		%s
		LIMIT $%d OFFSET $%d
	`, aliasedAnimeSelectColumns, baseQuery, whereClause, aliasedAnimeGroupByColumns, orderBy, argIndex, argIndex+1)

	args = append(args, params.PerPage, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search animes: %w", err)
	}
	defer rows.Close()

	animes := []models.AnimeWithUserData{}
	for rows.Next() {
		anime, err := scanAnime(rows)
		if err != nil {
			s.logger.Warn("Failed to scan anime row in AdvancedSearchAnimes", "error", err)
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

	}

	// Get watch providers from TMDB (if TMDB ID is available or can be found)
	s.enrichAnimeWithWatchProviders(animeWithData)

	return animeWithData, nil
}

// SyncAnimeStatus - TMDBからステータスを取得して更新
func (s *AnimeService) SyncAnimeStatus(ctx context.Context, animeID int, clearCache bool) error {
	// アニメ情報を取得（status, season, released_onを含める）
	var currentStatus, releasedOn sql.NullString
	var season *string
	err := s.db.QueryRowContext(ctx, `
		SELECT season, released_on, status
		FROM animes
		WHERE id = $1
	`, animeID).Scan(&season, &releasedOn, &currentStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAnimeNotFound
		}
		return fmt.Errorf("failed to get anime for status sync: %w", err)
	}

	var releasedOnPtr *string
	if releasedOn.Valid {
		releasedOnPtr = &releasedOn.String
	}

	// season + released_on ベースでステータスを判定
	status := determineStatusBySeason(season, releasedOnPtr)

	// 既存のステータスと比較して変更がなければ更新しない
	if currentStatus.Valid && currentStatus.String == status {
		return nil
	}

	// ステータスを更新
	_, err = s.db.ExecContext(ctx, `
		UPDATE animes
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, status, animeID)
	if err != nil {
		return fmt.Errorf("failed to update anime status for anime %d to %s: %w", animeID, status, err)
	}

	if clearCache {
		s.ClearAnimeCache()
	}
	s.logger.Debug("Anime status updated", "anime_id", animeID, "new_status", status)

	return nil
}

// SyncAllAnimeStatuses - 全アニメのステータスを更新
// season + released_on ベースで判定するため、外部API呼び出しなしで高速に処理可能
func (s *AnimeService) SyncAllAnimeStatuses(ctx context.Context) (int, int, error) {
	// 差分フィルタ: 放送中・未放送・ステータス未設定の作品のみ対象
	// 完了作品はステータスが変わらないため除外
	rows, err := s.db.QueryContext(ctx, `
        SELECT id FROM animes
        WHERE season IS NOT NULL
        AND (
            status IN ($1, $2)
            OR status IS NULL
        )
    `, constants.AnimeStatusAiring.String(), constants.AnimeStatusUpcoming.String())
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
		if err := rows.Scan(&animeID); err != nil {
			s.logger.Warn("Failed to scan anime ID for status sync", "error", err)
			failureCount++
			continue
		}
		animeIDs = append(animeIDs, animeID)
	}

	totalAnimes := len(animeIDs)
	if totalAnimes == 0 {
		s.logger.Info("No animes found requiring status sync.")
		return 0, 0, nil
	}

	s.logger.Info("Starting status sync for animes", "total_animes", totalAnimes)

	for i, animeID := range animeIDs {
		// コンテキストキャンセルチェック
		if ctx.Err() != nil {
			s.logger.Warn("Status sync cancelled", "progress", fmt.Sprintf("%d/%d", i+1, totalAnimes), "reason", ctx.Err())
			return successCount, failureCount, ctx.Err()
		}

		// 進行状況のロギング
		if i > 0 && i%100 == 0 {
			s.logger.Info("Processing anime status sync", "progress", fmt.Sprintf("%d/%d", i+1, totalAnimes))
		}

		if err := s.SyncAnimeStatus(ctx, animeID, false); err != nil {
			s.logger.Warn("Failed to sync anime status", "anime_id", animeID, "error", err)
			failureCount++
		} else {
			successCount++
		}
	}

	// 全ての更新が完了した後に一度だけキャッシュをクリア
	if successCount > 0 {
		s.ClearAnimeCache()
	}

	s.logger.Info("Anime status sync completed", "success", successCount, "failure", failureCount)
	return successCount, failureCount, nil
}

// GetUserFavorites はお気に入り（ハート）一覧を返す。視聴ステータスは持たない。
func (s *AnimeService) GetUserFavorites(userID int) ([]models.AnimeWithUserData, error) {
	// VIEWを使用してステータスを動的に計算（GetAnimeByID / SearchAnimes と一貫性を保つ）
	query := `
		SELECT ` + aliasedAnimeSelectColumns + `,
		       uf.id, uf.user_id, uf.anime_id, uf.created_at, uf.updated_at
		FROM animes_with_computed_status a
		JOIN user_favorites uf ON a.id = uf.anime_id
		WHERE uf.user_id = $1
		ORDER BY uf.updated_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user favorites: %w", err)
	}
	defer rows.Close()

	favorites := []models.AnimeWithUserData{}
	for rows.Next() {
		var anime models.Anime
		var favorite models.UserFavorite
		err := rows.Scan(
			&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status,
			&anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount,
			&anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium,
			&anime.ReleasedOn, &anime.Media,
			&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.CreatedAt, &favorite.UpdatedAt,
		)
		if err != nil {
			s.logger.Warn("Failed to scan favorite anime row", "error", err)
			continue
		}

		animeWithData := models.AnimeWithUserData{
			Anime:        anime,
			UserFavorite: &favorite,
		}

		favorites = append(favorites, animeWithData)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user favorites: %w", err)
	}

	// DBキャッシュがないアニメのみ外部APIから画像を取得
	s.enrichAnimesWithUserDataCoverImagesIfNeeded(favorites)

	return favorites, nil
}

// UpdateFavorite はアニメをお気に入り（ハート）に追加する。既に存在する場合はそのまま返す。
func (s *AnimeService) UpdateFavorite(userID, animeID int) (*models.UserFavorite, error) {
	// Check if anime exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1)", animeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	// Upsert favorite（ステータスを持たないため、既存行は updated_at のみ更新）
	var favorite models.UserFavorite
	err = s.db.QueryRow(`
		INSERT INTO user_favorites (user_id, anime_id, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, anime_id)
		DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, anime_id, created_at, updated_at
	`, userID, animeID).Scan(&favorite.ID, &favorite.UserID, &favorite.AnimeID, &favorite.CreatedAt, &favorite.UpdatedAt)
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

// GetUserWatchStatuses は視聴ステータス一覧を返す。status を指定するとそのステータスで絞り込む。
func (s *AnimeService) GetUserWatchStatuses(userID int, status string) ([]models.AnimeWithUserData, error) {
	whereClause := "WHERE ws.user_id = $1"
	args := []interface{}{userID}

	if status != "" {
		whereClause += " AND ws.status = $2"
		args = append(args, status)
	}

	// VIEWを使用してステータスを動的に計算（GetAnimeByID / SearchAnimes と一貫性を保つ）
	query := fmt.Sprintf(`
		SELECT `+aliasedAnimeSelectColumns+`,
		       ws.id, ws.user_id, ws.anime_id, ws.status, ws.created_at, ws.updated_at
		FROM animes_with_computed_status a
		JOIN user_watch_status ws ON a.id = ws.anime_id
		%s
		ORDER BY ws.updated_at DESC
	`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user watch statuses: %w", err)
	}
	defer rows.Close()

	statuses := []models.AnimeWithUserData{}
	for rows.Next() {
		var anime models.Anime
		var watchStatus models.UserWatchStatus
		err := rows.Scan(
			&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status,
			&anime.MalAnimeID, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.WatchersCount,
			&anime.CoverImageExtraLarge, &anime.CoverImageLarge, &anime.CoverImageMedium,
			&anime.ReleasedOn, &anime.Media,
			&watchStatus.ID, &watchStatus.UserID, &watchStatus.AnimeID, &watchStatus.Status, &watchStatus.CreatedAt, &watchStatus.UpdatedAt,
		)
		if err != nil {
			s.logger.Warn("Failed to scan watch status anime row", "error", err)
			continue
		}

		statuses = append(statuses, models.AnimeWithUserData{
			Anime:           anime,
			UserWatchStatus: &watchStatus,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user watch statuses: %w", err)
	}

	// DBキャッシュがないアニメのみ外部APIから画像を取得
	s.enrichAnimesWithUserDataCoverImagesIfNeeded(statuses)

	return statuses, nil
}

// UpdateWatchStatus は視聴ステータスを設定・更新する。
func (s *AnimeService) UpdateWatchStatus(userID, animeID int, status string) (*models.UserWatchStatus, error) {
	// Check if anime exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM animes WHERE id = $1)", animeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}
	if !exists {
		return nil, ErrAnimeNotFound
	}

	// Upsert watch status
	var watchStatus models.UserWatchStatus
	err = s.db.QueryRow(`
		INSERT INTO user_watch_status (user_id, anime_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, anime_id)
		DO UPDATE SET status = $3, updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, anime_id, status, created_at, updated_at
	`, userID, animeID, status).Scan(&watchStatus.ID, &watchStatus.UserID, &watchStatus.AnimeID, &watchStatus.Status, &watchStatus.CreatedAt, &watchStatus.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update watch status: %w", err)
	}

	return &watchStatus, nil
}

func (s *AnimeService) DeleteWatchStatus(userID, animeID int) error {
	_, err := s.db.Exec(`
		DELETE FROM user_watch_status WHERE user_id = $1 AND anime_id = $2
	`, userID, animeID)
	if err != nil {
		return fmt.Errorf("failed to delete watch status: %w", err)
	}

	return nil
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
		SELECT id, user_id, anime_id, created_at, updated_at
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
		if err := rows.Scan(&fav.ID, &fav.UserID, &fav.AnimeID, &fav.CreatedAt, &fav.UpdatedAt); err != nil {
			s.logger.Warn("Failed to scan user favorite", "error", err)
			continue
		}
		favorites[fav.AnimeID] = &fav
	}

	return favorites, nil
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

	// 各アニメにユーザーデータを設定
	for i := range animes {
		animeID := animes[i].ID
		if fav, ok := favorites[animeID]; ok {
			animes[i].UserFavorite = fav
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
// finalizeDescriptionText はあらすじの書式を統一し、日本語でない場合は Claude で日本語へ翻訳する。
// Claude が無効、または翻訳失敗時は整形済みの原文（英語など）にフォールバックする。
func (s *AnimeService) finalizeDescriptionText(ctx context.Context, raw string) string {
	sanitized := sanitizeDescription(raw)
	if sanitized == "" {
		return ""
	}
	if containsJapanese(sanitized) {
		return sanitized
	}
	if s.claudeClient == nil || !s.claudeClient.IsEnabled() {
		return sanitized
	}

	tctx, cancel := context.WithTimeout(ctx, 15*time.Second)
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

	baseQuery := "FROM animes_with_computed_status a"

	// Track which JOINs are needed
	needGenreJoin := len(params.Genres) > 0
	needRatingJoin := params.MinRating > 0 || params.MaxRating > 0 || params.Sort == SortByRating
	needWatchStatusJoin := userID != nil && params.UserStatus != ""

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
	if needWatchStatusJoin {
		baseQuery += `
		INNER JOIN user_watch_status ws ON a.id = ws.anime_id`
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

	// ステータス検索（VIEWのcomputed_statusを使用）
	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.computed_status = $%d", argIndex))
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
		whereClauses = append(whereClauses, fmt.Sprintf("ws.user_id = $%d", argIndex))
		args = append(args, *userID)
		argIndex++
		whereClauses = append(whereClauses, fmt.Sprintf("ws.status = $%d", argIndex))
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
		AND (genres_synced_at IS NULL OR genres_synced_at < CURRENT_TIMESTAMP - INTERVAL '1 day' * $%d)
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
// Uses AniList API for genre information
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
	// 1. No released_on (NULL or empty string)
	// 2. Has MAL ID (required for Jikan API)
	rows, err := s.db.Query(`
		SELECT id, title, mal_anime_id 
		FROM animes 
		WHERE (released_on IS NULL OR released_on = '')
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

// SyncReleasedOnForCurrentSeason syncs released_on from Jikan API for current and previous season animes
// that have released_on=NULL, then re-evaluates their status.
// This ensures that animes within the current season period are correctly labeled as "upcoming" vs "airing".
func (s *AnimeService) SyncReleasedOnForCurrentSeason(ctx context.Context, currentSeason, previousSeason string) (int, int, error) {
	s.logger.Info("Syncing released_on for current season animes without release date",
		"current_season", currentSeason,
		"previous_season", previousSeason)

	// released_onがNULLまたは空文字の今期・前期作品を取得
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, mal_anime_id
		FROM animes
		WHERE season IN ($1, $2)
		AND (released_on IS NULL OR released_on = '')
		AND mal_anime_id IS NOT NULL AND mal_anime_id != ''
	`, currentSeason, previousSeason)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query animes for released_on sync: %w", err)
	}
	defer rows.Close()

	type animeInfo struct {
		ID    int
		Title string
		MalID string
	}
	var animes []animeInfo

	for rows.Next() {
		var a animeInfo
		if err := rows.Scan(&a.ID, &a.Title, &a.MalID); err != nil {
			s.logger.Warn("Failed to scan anime for released_on sync", "error", err)
			continue
		}
		animes = append(animes, a)
	}

	totalAnimes := len(animes)
	if totalAnimes == 0 {
		s.logger.Info("All current/previous season animes already have released_on")
		return 0, 0, nil
	}

	s.logger.Info("Found current/previous season animes without released_on", "count", totalAnimes)

	successCount := 0
	failureCount := 0

	for _, anime := range animes {
		// コンテキストのキャンセル・タイムアウトを確認
		select {
		case <-ctx.Done():
			s.logger.Warn("SyncReleasedOnForCurrentSeason cancelled",
				"success", successCount, "failure", failureCount)
			return successCount, failureCount, ctx.Err()
		default:
		}

		releasedOn, err := s.jikanClient.GetAnimeAiredDate(anime.MalID)
		if err != nil {
			s.logger.Debug("Failed to get aired date from Jikan for current season anime",
				"anime_id", anime.ID,
				"mal_id", anime.MalID,
				"title", anime.Title,
				"error", err)
			failureCount++
			time.Sleep(constants.JikanRateLimitDelay)
			continue
		}

		if releasedOn == "" {
			time.Sleep(constants.JikanRateLimitDelay)
			continue
		}

		// released_onをDBに保存
		_, err = s.db.ExecContext(ctx, `
			UPDATE animes SET released_on = $1 WHERE id = $2
		`, releasedOn, anime.ID)
		if err != nil {
			s.logger.Warn("Failed to save released_on for current season anime",
				"anime_id", anime.ID, "error", err)
			failureCount++
			time.Sleep(constants.JikanRateLimitDelay)
			continue
		}

		s.logger.Debug("Saved released_on for current season anime",
			"anime_id", anime.ID,
			"title", anime.Title,
			"released_on", releasedOn)

		// released_on設定後にステータスを再判定
		if syncErr := s.SyncAnimeStatus(ctx, anime.ID, false); syncErr != nil {
			s.logger.Warn("Failed to re-evaluate status after released_on update",
				"anime_id", anime.ID, "error", syncErr)
		}

		successCount++
		time.Sleep(constants.JikanRateLimitDelay)
	}

	// キャッシュをクリア
	if successCount > 0 {
		s.ClearAnimeCache()
	}

	s.logger.Info("Current season released_on sync completed",
		"success", successCount,
		"failure", failureCount)

	return successCount, failureCount, nil
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
		AnnictID    *int
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
		SELECT id, title, description, released_on, mal_anime_id, annict_id
		FROM animes WHERE id IN (%s)
	`, placeholders), args...)
	if err != nil {
		return fmt.Errorf("failed to batch fetch anime info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var info animeInfo
		var desc, relOn, malID sql.NullString
		var annictID sql.NullInt64
		if err := rows.Scan(&info.ID, &info.Title, &desc, &relOn, &malID, &annictID); err != nil {
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
		if annictID.Valid && annictID.Int64 > 0 {
			aid := int(annictID.Int64)
			info.AnnictID = &aid
		}
		animeMap[info.ID] = &info
	}

	s.logger.Info("Batch fetched anime info", "fetched", len(animeMap))

	// Step 1: Sync descriptions from TMDB
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

	// Step 2: Sync released_on from Jikan
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
// determineMediaType determines the TMDB media type ("tv" or "movie") based on
// DB media field, model data, and AniList format as a fallback.
func (s *AnimeService) determineMediaType(ctx context.Context, dbMediaType sql.NullString, modelMedia *string, malID string) string {
	if dbMediaType.Valid && dbMediaType.String == "movie" {
		return "movie"
	}
	if modelMedia != nil && *modelMedia == "movie" {
		return "movie"
	}
	if malID != "" {
		format, err := s.anilistClient.GetMediaFormatByMALId(ctx, malID)
		if err == nil && format == "MOVIE" {
			return "movie"
		}
	}
	return "tv"
}

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
		mediaType = s.determineMediaType(ctx, dbMediaType, animeWithData.Media, malID)
	} else if err == nil && tmdbID > 0 {
		mediaType = s.determineMediaType(ctx, dbMediaType, animeWithData.Media, malID)

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

	// Get watch providers from TMDB with fallback (try other mediaType if not found)
	providers, err := s.tmdbClient.GetJapaneseWatchProvidersWithFallback(ctx, tmdbID, mediaType)
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
func determineStatusBySeason(season *string, releasedOn *string) string {
	var seasonName string
	if season != nil {
		seasonName = *season
	}
	return determineStatusBySeasonAt(seasonName, releasedOn, time.Now().UTC())
}

// determineStatusBySeasonAt はシーズンと放送開始日を考慮してアニメの放送状態を判定します。
// 基準時刻を引数で受け取ることでテスタビリティを確保しています。
func determineStatusBySeasonAt(seasonName string, releasedOn *string, now time.Time) string {
	// released_onが未来の日付なら放送予定
	if releasedOn != nil && *releasedOn != "" {
		// UTCで比較してサーバータイムゾーンの影響を排除
		releaseDate, err := time.ParseInLocation("2006-01-02", *releasedOn, time.UTC)
		if err == nil {
			today := now.Truncate(24 * time.Hour)
			if releaseDate.After(today) {
				return constants.AnimeStatusUpcoming.String()
			}
		}
	}

	if seasonName == "" {
		return constants.AnimeStatusFinished.String()
	}

	// シーズン名（例: "YYYY-season"）をパース
	var seasonYear int
	var seasonStr string
	_, err := fmt.Sscanf(seasonName, "%d-%s", &seasonYear, &seasonStr)
	if err != nil {
		return constants.AnimeStatusFinished.String() // パース失敗時は完結とみなす
	}

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
		return constants.AnimeStatusFinished.String() // 未知のシーズン名は完結とみなす
	}

	// 判定ロジック
	if seasonYear > currentYear {
		// 未来のシーズン → 放送予定
		return constants.AnimeStatusUpcoming.String()
	} else if seasonYear == currentYear {
		// 同じ年のシーズン
		if currentMonth < seasonStartMonth {
			// シーズン開始月より前 → 放送予定
			return constants.AnimeStatusUpcoming.String()
		} else if currentMonth <= seasonEndMonth {
			// シーズン期間中 → 放送中
			return constants.AnimeStatusAiring.String()
		} else {
			// シーズン終了月以後 → 完結
			return constants.AnimeStatusFinished.String()
		}
	} else {
		// 過去のシーズン → 完結
		return constants.AnimeStatusFinished.String()
	}
}
