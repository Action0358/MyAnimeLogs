package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"myanimelogs/internal/constants"
	"myanimelogs/internal/models"
	"myanimelogs/internal/utils"
)

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

// buildGenreFilterClause - ジャンル絞り込みのWHERE句を生成する
//
// 指定されたジャンルの「いずれか」を持つ作品を返す（OR）。
// ANDにすると該当が激減し、複数選択の体験がかえって悪くなるためORを採る。
// これは高度検索側（anime_search.go の buildAdvancedSearchQuery）の挙動とも揃う。
//
// genres が空のときは空文字列を返すので、呼び出し側は clause != "" で判定する。
func buildGenreFilterClause(genres []string, argIndex int) (string, []interface{}, int) {
	if len(genres) == 0 {
		return "", nil, argIndex
	}

	placeholders := make([]string, len(genres))
	args := make([]interface{}, len(genres))
	for i, genre := range genres {
		placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
		args[i] = genre
	}

	clause := fmt.Sprintf(
		"id IN (SELECT anime_id FROM anime_genres WHERE genre IN (%s))",
		strings.Join(placeholders, ", "),
	)

	return clause, args, argIndex + len(genres)
}

// generateSearchCacheKey - 検索パラメータからキャッシュキーを生成
func (s *AnimeService) generateSearchCacheKey(params models.AnimeSearchParams) string {
	// ジャンルは正規化済み（重複除去・昇順）の一覧から組み立てる。
	// "A,B" と "B,A" は同じ検索なので、同じキーに寄せてキャッシュを分断させない。
	genreKey := strings.Join(params.GenreList(), ",")

	return fmt.Sprintf("%skw:%s:s:%s:st:%s:g:%s:m:%s:so:%s:p:%d:pp:%d",
		constants.CacheKeySearch,
		params.Keyword, params.Season, params.Status, genreKey, params.Media, params.Sort,
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

	if clause, genreArgs, newArgIndex := buildGenreFilterClause(params.GenreList(), argIndex); clause != "" {
		whereClauses = append(whereClauses, clause)
		args = append(args, genreArgs...)
		argIndex = newArgIndex
	}

	if params.Media != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(media) = LOWER($%d)", argIndex))
		args = append(args, params.Media)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Build ORDER BY clause
	//
	// タイトル順: title_kana（ひらがな）がある行はそちらを優先する。
	// title は先頭に記号・半角スペースが付くデータが混ざっており、
	// そのままでは五十音順として意味を成さない。
	orderByTitle := "ORDER BY COALESCE(NULLIF(title_kana, ''), title) ASC, id ASC"

	orderBy := orderByTitle
	switch params.Sort {
	case "year":
		// season は "2024-summer" のような文字列で、単純な DESC では
		// 同じ年の中で autumn/spring/summer/winter がアルファベット順に
		// 並んでしまい、実際の放送時期の新しい順にならない。
		// 年→クール内の順序（winter<spring<summer<autumn）の2段階で並べる。
		// 同じシーズン内の作品はこれだけでは順序が確定しない（新規シーズンは
		// watchers_countが同値で並ぶことが多く、タイブレークが無いとページ送りの
		// たびに順序が揺れ、同じ作品が複数ページに重複して現れることがあったため、
		// 一意なidを最終タイブレークにする
		orderBy = `ORDER BY
			split_part(season, '-', 1)::int DESC,
			CASE split_part(season, '-', 2)
				WHEN 'autumn' THEN 4
				WHEN 'summer' THEN 3
				WHEN 'spring' THEN 2
				WHEN 'winter' THEN 1
				ELSE 0
			END DESC,
			id DESC`
	case "popularity":
		// watchers_count が同数の場合にページごとの順序が揺れないよう、
		// タイトル順（さらにidをタイブレークにするorderByTitleを流用）を続ける
		orderBy = "ORDER BY watchers_count DESC NULLS LAST, " + strings.TrimPrefix(orderByTitle, "ORDER BY ")
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
	return fmt.Sprintf("%sq:%s:g:%s:y:%d:s:%s:st:%s:so:%s:o:%s:p:%d:pp:%d",
		constants.CacheKeyAdvSearch,
		params.Query, genresKey, params.Year, params.Season,
		params.Status,
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
	baseQuery := "FROM animes_with_computed_status a"

	// Track which JOINs are needed
	needGenreJoin := len(params.Genres) > 0
	needWatchStatusJoin := userID != nil && params.UserStatus != ""

	// Add JOINs only when needed
	if needGenreJoin {
		baseQuery += "\n\t\tLEFT JOIN anime_genres ag ON a.id = ag.anime_id"
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
		SortByYear       = "year"
		SortByPopularity = "popularity"
		SortByUpdated    = "updated"
		OrderDesc        = "desc"
	)

	orderDirection := "ASC"
	if params.Order == OrderDesc {
		orderDirection = "DESC"
	}

	// 同値の行が並ぶと、OFFSET/LIMITでのページ送りのたびに順序が揺れ、
	// 同じ作品が複数ページに重複して現れたり抜け落ちたりする（一意なidを
	// 最終タイブレークにすることで、どの並び替えでも順序を確定させる）
	switch params.Sort {
	case SortByTitle:
		return fmt.Sprintf("ORDER BY a.title %s, a.id %s", orderDirection, orderDirection)
	case SortByYear:
		// season から年を抽出してソート
		// NOTE: SUBSTRING関数を使用しているため、インデックスが利用されません
		// 大規模データセットでパフォーマンスが問題になる場合は、
		// CREATE INDEX idx_animes_season_year ON animes (SUBSTRING(season FROM 1 FOR 4))
		// のような関数ベースのインデックスを作成することを検討してください
		return fmt.Sprintf("ORDER BY SUBSTRING(a.season FROM 1 FOR 4) %s, a.season %s, a.id %s", orderDirection, orderDirection, orderDirection)
	case SortByPopularity:
		return fmt.Sprintf("ORDER BY a.watchers_count %s, a.id %s", orderDirection, orderDirection)
	case SortByUpdated:
		return fmt.Sprintf("ORDER BY a.updated_at %s, a.id %s", orderDirection, orderDirection)
	default:
		return fmt.Sprintf("ORDER BY a.title %s, a.id %s", orderDirection, orderDirection)
	}
}
