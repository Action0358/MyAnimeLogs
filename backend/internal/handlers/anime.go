package handlers

import (
	"net/http"
	"strconv"

	"myanimelogs/internal/models"
	"myanimelogs/internal/services"

	"github.com/gin-gonic/gin"
)

type AnimeHandler struct {
	*BaseHandler
	animeService *services.AnimeService
}

func NewAnimeHandler(animeService *services.AnimeService) *AnimeHandler {
	return &AnimeHandler{
		BaseHandler:  NewBaseHandler(),
		animeService: animeService,
	}
}

func (h *AnimeHandler) GetAnimes(c *gin.Context) {
	var params models.AnimeSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	// Set defaults using validator
	params.Page, params.PerPage = h.validator.ValidatePagination(params.Page, params.PerPage)

	// Get user ID if authenticated
	var userID *int
	if id, exists := c.Get("user_id"); exists {
		uid := id.(int)
		userID = &uid
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

	// Get user ID if authenticated
	var userID *int
	if uid, exists := c.Get("user_id"); exists {
		userIDVal := uid.(int)
		userID = &userIDVal
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

	// Get user ID if authenticated
	var userID *int
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(int)
		userID = &id
	}

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
