package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"
)

type AnimeHandler struct {
	animeService *services.AnimeService
}

func NewAnimeHandler(animeService *services.AnimeService) *AnimeHandler {
	return &AnimeHandler{
		animeService: animeService,
	}
}

func (h *AnimeHandler) GetAnimes(c *gin.Context) {
	var params models.AnimeSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: err.Error(),
			Code:    models.ErrCodeValidation,
		})
		return
	}

	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}

	// Get user ID if authenticated
	var userID *int
	if id, exists := c.Get("user_id"); exists {
		uid := id.(int)
		userID = &uid
	}

	result, err := h.animeService.SearchAnimes(params, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "アニメ検索に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnimeHandler) GetAnime(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: "無効なアニメIDです",
			Code:    models.ErrCodeValidation,
		})
		return
	}

	// Get user ID if authenticated
	var userID *int
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(int)
		userID = &id
	}

	anime, err := h.animeService.GetAnimeByID(id, userID)
	if err != nil {
		if err == services.ErrAnimeNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   models.ErrMsgNotFound,
				Message: "アニメが見つかりません",
				Code:    models.ErrCodeNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "アニメ詳細の取得に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, anime)
}

func (h *AnimeHandler) GetAnimeEpisodes(c *gin.Context) {
	idStr := c.Param("id")
	animeID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: "無効なアニメIDです",
			Code:    models.ErrCodeValidation,
		})
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
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   models.ErrMsgNotFound,
				Message: "アニメが見つかりません",
				Code:    models.ErrCodeNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "エピソード一覧の取得に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"anime_id": animeID,
		"episodes": episodes,
	})
}

func (h *AnimeHandler) GetFavorites(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	status := c.Query("status")
	favorites, err := h.animeService.GetUserFavorites(userID.(int), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "お気に入り一覧の取得に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, favorites)
}

func (h *AnimeHandler) UpdateFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	animeIDStr := c.Param("anime_id")
	animeID, err := strconv.Atoi(animeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: "無効なアニメIDです",
			Code:    models.ErrCodeValidation,
		})
		return
	}

	var req models.UserFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: err.Error(),
			Code:    models.ErrCodeValidation,
		})
		return
	}

	favorite, err := h.animeService.UpdateFavorite(userID.(int), animeID, req.Status)
	if err != nil {
		if err == services.ErrAnimeNotFound {
			c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse{
				Error:   models.ErrMsgUnprocessableEntity,
				Message: "存在しないアニメです",
				Code:    models.ErrCodeUnprocessableEntity,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "お気に入り更新に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "お気に入りを更新しました",
		"favorite": favorite,
	})
}

func (h *AnimeHandler) DeleteFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	animeIDStr := c.Param("anime_id")
	animeID, err := strconv.Atoi(animeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: "無効なアニメIDです",
			Code:    models.ErrCodeValidation,
		})
		return
	}

	err = h.animeService.DeleteFavorite(userID.(int), animeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "お気に入り削除に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "お気に入りから削除しました",
	})
}

func (h *AnimeHandler) UpdateWatchStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	episodeIDStr := c.Param("episode_id")
	episodeID, err := strconv.Atoi(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: "無効なエピソードIDです",
			Code:    models.ErrCodeValidation,
		})
		return
	}

	var req models.WatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: err.Error(),
			Code:    models.ErrCodeValidation,
		})
		return
	}

	watchStatus, err := h.animeService.UpdateWatchStatus(userID.(int), episodeID, req.Viewed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "視聴状態の更新に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "視聴状態を更新しました",
		"watch_status": watchStatus,
	})
}

func (h *AnimeHandler) UpdateRating(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	animeIDStr := c.Param("id")
	animeID, err := strconv.Atoi(animeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: "無効なアニメIDです",
			Code:    models.ErrCodeValidation,
		})
		return
	}

	var req models.UserRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: err.Error(),
			Code:    models.ErrCodeValidation,
		})
		return
	}

	rating, err := h.animeService.UpdateRating(userID.(int), animeID, req.Rating, req.Comment)
	if err != nil {
		if err == services.ErrAnimeNotFound {
			c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse{
				Error:   models.ErrMsgUnprocessableEntity,
				Message: "存在しないアニメです",
				Code:    models.ErrCodeUnprocessableEntity,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "評価の更新に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "評価を登録しました",
		"rating":  rating,
	})
}

func (h *AnimeHandler) DeleteRating(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	animeIDStr := c.Param("id")
	animeID, err := strconv.Atoi(animeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   models.ErrMsgValidation,
			Message: "無効なアニメIDです",
			Code:    models.ErrCodeValidation,
		})
		return
	}

	err = h.animeService.DeleteRating(userID.(int), animeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "評価の削除に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "評価を削除しました",
	})
}

func (h *AnimeHandler) GetRecentlyWatched(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	history, err := h.animeService.GetRecentlyWatched(userID.(int), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "視聴履歴の取得に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *AnimeHandler) GetWatchlist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   models.ErrMsgUnauthorized,
			Message: "認証が必要です",
			Code:    models.ErrCodeUnauthorized,
		})
		return
	}

	watchlist, err := h.animeService.GetUserFavorites(userID.(int), "planned")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   models.ErrMsgInternalServer,
			Message: "視聴予定リストの取得に失敗しました",
			Code:    models.ErrCodeInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, watchlist)
}