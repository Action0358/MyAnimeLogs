package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"myanimelogs/internal/common"
	"myanimelogs/internal/constants"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"
	"myanimelogs/internal/utils"

	"github.com/gin-gonic/gin"
)

type AnimeHandler struct {
	*BaseHandler
	animeService services.AnimeServiceInterface
}

func NewAnimeHandler(animeService services.AnimeServiceInterface, logger *common.Logger, errorMonitor *common.ErrorMonitor) *AnimeHandler {
	return &AnimeHandler{
		BaseHandler:  NewBaseHandlerWithLogger(logger, errorMonitor),
		animeService: animeService,
	}
}

func (h *AnimeHandler) GetAnimes(c *gin.Context) {
	var params models.AnimeSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	// バリデーターを使用してデフォルト値を設定
	params.Page, params.PerPage = h.validator.ValidatePagination(params.Page, params.PerPage)

	// 認証されている場合はユーザーIDを取得
	var userID *int
	if id, exists := c.Get("user_id"); exists {
		if uid, ok := id.(int); ok {
			userID = &uid
		}
	}

	result, err := h.animeService.SearchAnimes(params, userID)
	if err != nil {
		h.response.InternalServerError(c, "アニメ検索に失敗しました")
		return
	}

	h.response.Success(c, result)
}

func (h *AnimeHandler) GetAnime(c *gin.Context) {
	id, err := h.validator.ParseIDParam(c, "id")
	if err != nil {
		h.response.ValidationError(c, "無効なアニメIDです")
		return
	}

	// 認証されている場合はユーザーIDを取得
	var userID *int
	if uid, exists := c.Get("user_id"); exists {
		if userIDVal, ok := uid.(int); ok {
			userID = &userIDVal
		}
	}

	anime, err := h.animeService.GetAnimeByID(id, userID)
	if err != nil {
		if err == services.ErrAnimeNotFound {
			h.response.NotFoundError(c, "アニメが見つかりません")
			return
		}
		h.response.InternalServerError(c, "アニメ詳細の取得に失敗しました")
		return
	}

	h.response.Success(c, anime)
}

func (h *AnimeHandler) GetAnimeEpisodes(c *gin.Context) {
	animeID, err := h.validator.ParseIDParam(c, "id")
	if err != nil {
		h.response.ValidationError(c, "無効なアニメIDです")
		return
	}

	// 認証されている場合はユーザーIDを取得
	var userID *int
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(int); ok {
			userID = &id
		}
	}

	// separatedパラメータで分離表示を制御（デフォルトはtrue）
	separated := c.DefaultQuery("separated", "true") == "true"

	if separated {
		// 通常エピソードと特別編を分離して取得
		regularEpisodes, specialEpisodes, err := h.animeService.GetEpisodesSeparated(animeID, userID)
		if err != nil {
			if err == services.ErrAnimeNotFound {
				h.response.NotFoundError(c, "アニメが見つかりません")
				return
			}
			h.response.InternalServerError(c, "エピソード一覧の取得に失敗しました")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"anime_id":         animeID,
			"regular_episodes": regularEpisodes,
			"special_episodes": specialEpisodes,
			"total_regular":    len(regularEpisodes),
			"total_special":    len(specialEpisodes),
		})
		return
	}

	// 従来の形式（全エピソードを一覧で返す）
	episodes, err := h.animeService.GetAnimeEpisodes(animeID, userID)
	if err != nil {
		if err == services.ErrAnimeNotFound {
			h.response.NotFoundError(c, "アニメが見つかりません")
			return
		}
		h.response.InternalServerError(c, "エピソード一覧の取得に失敗しました")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"anime_id": animeID,
		"episodes": episodes,
	})
}

func (h *AnimeHandler) GetFavorites(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	status := c.Query("status")
	favorites, err := h.animeService.GetUserFavorites(userID, status)
	if err != nil {
		h.response.InternalServerError(c, "お気に入り一覧の取得に失敗しました")
		return
	}

	h.response.Success(c, favorites)
}

func (h *AnimeHandler) UpdateFavorite(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	animeID, err := h.validator.ParseIDParam(c, "anime_id")
	if err != nil {
		h.response.ValidationError(c, "無効なアニメIDです")
		return
	}

	var req models.UserFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	favorite, err := h.animeService.UpdateFavorite(userID, animeID, req.Status)
	if err != nil {
		if err == services.ErrAnimeNotFound {
			h.response.UnprocessableEntityError(c, "存在しないアニメです")
			return
		}
		h.response.InternalServerError(c, "お気に入り更新に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"message":  "お気に入りを更新しました",
		"favorite": favorite,
	})
}

func (h *AnimeHandler) DeleteFavorite(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	animeID, err := h.validator.ParseIDParam(c, "anime_id")
	if err != nil {
		h.response.ValidationError(c, "無効なアニメIDです")
		return
	}

	err = h.animeService.DeleteFavorite(userID, animeID)
	if err != nil {
		h.response.InternalServerError(c, "お気に入り削除に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"message": "お気に入りから削除しました",
	})
}

func (h *AnimeHandler) UpdateWatchStatus(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	episodeID, err := h.validator.ParseIDParam(c, "episode_id")
	if err != nil {
		h.response.ValidationError(c, "無効なエピソードIDです")
		return
	}

	var req models.WatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	watchStatus, err := h.animeService.UpdateWatchStatus(userID, episodeID, req.Viewed)
	if err != nil {
		h.response.InternalServerError(c, "視聴状態の更新に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"message":      "視聴状態を更新しました",
		"watch_status": watchStatus,
	})
}

func (h *AnimeHandler) UpdateRating(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	animeID, err := h.validator.ParseIDParam(c, "id")
	if err != nil {
		h.response.ValidationError(c, "無効なアニメIDです")
		return
	}

	var req models.UserRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	rating, err := h.animeService.UpdateRating(userID, animeID, req.Rating, req.Comment)
	if err != nil {
		if err == services.ErrAnimeNotFound {
			h.response.UnprocessableEntityError(c, "存在しないアニメです")
			return
		}
		h.response.InternalServerError(c, "評価の更新に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"message": "評価を登録しました",
		"rating":  rating,
	})
}

func (h *AnimeHandler) DeleteRating(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	animeID, err := h.validator.ParseIDParam(c, "id")
	if err != nil {
		h.response.ValidationError(c, "無効なアニメIDです")
		return
	}

	err = h.animeService.DeleteRating(userID, animeID)
	if err != nil {
		h.response.InternalServerError(c, "評価の削除に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"message": "評価を削除しました",
	})
}

func (h *AnimeHandler) GetRecentlyWatched(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	history, err := h.animeService.GetRecentlyWatched(userID, limit)
	if err != nil {
		h.response.InternalServerError(c, "視聴履歴の取得に失敗しました")
		return
	}

	h.response.Success(c, history)
}

func (h *AnimeHandler) GetWatchlist(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	watchlist, err := h.animeService.GetUserFavorites(userID, "planned")
	if err != nil {
		h.response.InternalServerError(c, "視聴予定リストの取得に失敗しました")
		return
	}

	h.response.Success(c, watchlist)
}

// AdvancedSearchAnimes 高度検索API
func (h *AnimeHandler) AdvancedSearchAnimes(c *gin.Context) {
	var params models.AdvancedSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	// バリデーターを使用してデフォルト値を設定
	params.Page, params.PerPage = h.validator.ValidatePagination(params.Page, params.PerPage)

	// 認証されている場合はユーザーIDを取得
	var userID *int
	if id, exists := c.Get("user_id"); exists {
		if uid, ok := id.(int); ok {
			userID = &uid
		}
	}

	result, err := h.animeService.AdvancedSearchAnimes(params, userID)
	if err != nil {
		h.response.InternalServerError(c, "高度検索に失敗しました")
		return
	}

	h.response.Success(c, result)
}

// GetFilterOptions フィルター情報取得API
func (h *AnimeHandler) GetFilterOptions(c *gin.Context) {
	filterOptions, err := h.animeService.GetFilterOptions()
	if err != nil {
		h.logger.Error("Failed to get filter options", "error", err)
		// エラーでもデフォルト値を返す
		defaultOptions := &models.FilterOptions{
			Genres:  []models.GenreInfo{},
			Years:   []int{},
			Seasons: constants.DefaultSeasons,
			Ratings: models.RatingStats{
				Min:     1.0,
				Max:     5.0,
				Average: 3.0,
				Count:   0,
			},
		}
		h.response.Success(c, defaultOptions)
		return
	}

	h.response.Success(c, filterOptions)
}

// GetPopularSearches 人気検索キーワード取得API
func (h *AnimeHandler) GetPopularSearches(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit := utils.ParseLimitQuery(limitStr, 10, 50)

	searches, err := h.animeService.GetPopularSearches(limit)
	if err != nil {
		h.logger.Error("Failed to get popular searches", "error", err)
		// エラーでもデフォルト値を返す
		// グローバルスライスを変更しないようにコピーを作成
		defaultSearches := make([]string, len(constants.DefaultPopularSearches))
		copy(defaultSearches, constants.DefaultPopularSearches)
		if limit < len(defaultSearches) {
			defaultSearches = defaultSearches[:limit]
		}
		h.response.Success(c, gin.H{
			"popular_searches": defaultSearches,
		})
		return
	}

	h.response.Success(c, gin.H{
		"popular_searches": searches,
	})
}

// GetSearchSuggestions 検索候補取得API
func (h *AnimeHandler) GetSearchSuggestions(c *gin.Context) {
	query := c.Query("q")
	if strings.TrimSpace(query) == "" {
		h.response.Success(c, gin.H{
			"suggestions": []string{},
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit := utils.ParseLimitQuery(limitStr, 10, 50)

	suggestions, err := h.animeService.GetSearchSuggestions(query, limit)
	if err != nil {
		h.logger.Error("Failed to get search suggestions", "error", err, "query", query)
		// エラーでも空配列を返す
		h.response.Success(c, gin.H{
			"suggestions": []string{},
		})
		return
	}

	h.response.Success(c, gin.H{
		"suggestions": suggestions,
	})
}

// GetCurrentSeasonAnimes 今期アニメ取得API
func (h *AnimeHandler) GetCurrentSeasonAnimes(c *gin.Context) {
	h.getAnimesWithLimit(c, h.animeService.GetCurrentSeasonAnimes, "今期アニメ")
}

// GetPopularAnimes 人気アニメ取得API
func (h *AnimeHandler) GetPopularAnimes(c *gin.Context) {
	h.getAnimesWithLimit(c, h.animeService.GetPopularAnimes, "人気アニメ")
}

// getAnimesWithLimit is a helper method to handle anime list endpoints with limit parameter
func (h *AnimeHandler) getAnimesWithLimit(c *gin.Context, fetchFunc func(int) ([]models.Anime, error), resourceName string) {
	limitStr := c.DefaultQuery("limit", "12")
	limit := utils.ParseLimitQuery(limitStr, 12, constants.PopularAnimeLimit) // 今期・人気アニメ専用ページ対応

	animes, err := fetchFunc(limit)
	if err != nil {
		h.logger.Error("Failed to get "+resourceName, "error", err)
		h.response.InternalServerError(c, resourceName+"の取得に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"animes": animes,
		// NOTE: totalは現在のレスポンスに含まれる件数（limit適用後）
		// DBの総件数が必要な場合は別途COUNT(*)クエリが必要
		"total": len(animes),
	})
}

// GetSearchHistory 検索履歴取得API
func (h *AnimeHandler) GetSearchHistory(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit := utils.ParseLimitQuery(limitStr, constants.SearchHistoryDefaultLimit, constants.SearchHistoryMaxItems)

	history, err := h.animeService.GetSearchHistory(userID, limit)
	if err != nil {
		h.logger.Error("Failed to get search history", "error", err, "user_id", userID)
		h.response.InternalServerError(c, "検索履歴の取得に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"history": history,
		"total":   len(history),
	})
}

// DeleteSearchHistory 検索履歴削除API（全件）
func (h *AnimeHandler) DeleteSearchHistory(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	err = h.animeService.DeleteSearchHistory(userID)
	if err != nil {
		h.logger.Error("Failed to delete search history", "error", err, "user_id", userID)
		h.response.InternalServerError(c, "検索履歴の削除に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"message": "検索履歴を削除しました",
	})
}

// DeleteSearchHistoryItem 検索履歴削除API（個別）
func (h *AnimeHandler) DeleteSearchHistoryItem(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	logID, err := h.validator.ParseIDParam(c, "id")
	if err != nil {
		h.response.ValidationError(c, "無効な履歴IDです")
		return
	}

	err = h.animeService.DeleteSearchHistoryItem(userID, logID)
	if err != nil {
		if err == services.ErrSearchHistoryNotFound {
			h.response.NotFoundError(c, "検索履歴が見つかりません")
			return
		}
		h.logger.Error("Failed to delete search history item", "error", err, "user_id", userID, "log_id", logID)
		h.response.InternalServerError(c, "検索履歴の削除に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"message": "検索履歴を削除しました",
	})
}
