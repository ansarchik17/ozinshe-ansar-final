package handlers

import (
	"go.uber.org/zap"
	"goozinshe/logger"
	"goozinshe/models"
	"goozinshe/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WatchlistHandler struct {
	moviesRepo    *repositories.MoviesRepository
	watchlistRepo *repositories.WatchlistRepository
}

func NewWatchlistHandler(moviesRepo *repositories.MoviesRepository, watchlistRepo *repositories.WatchlistRepository) *WatchlistHandler {
	return &WatchlistHandler{moviesRepo: moviesRepo, watchlistRepo: watchlistRepo}
}

// HandleGetMovies godoc
// @Summary Get movies from watchlist
// @Tags watchlist
// @Accept json
// @Produce json
// @Success 200
// @Failure 500 {object} models.ApiError
// @Router /watchlist [get]
func (h *WatchlistHandler) HandleGetMovies(c *gin.Context) {
	logger := logger.GetLogger()
	movies, err := h.watchlistRepo.GetMoviesFromWatchlist(c)
	if err != nil {
		logger.Error("Could not get movies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, movies)
}

// HandleAddMovie godoc
// @Summary Add movie to the watchlist
// @Tags watchlist
// @Accept json
// @Produce json
// @Param movieId path int true "Movie id"
// @Success 200
// @Failure 500 {object} models.ApiError
// @Router /watchlist/{movieId} [post]
func (h *WatchlistHandler) HandleAddMovie(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("movieId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Could not parse movie id", zap.String("movieId", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}
	_, err = h.moviesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Could not find movie", zap.String("movieId", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	err = h.watchlistRepo.AddToWatchlist(c, id)
	if err != nil {
		logger.Error("Could not add movie", zap.String("movieId", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

// HandleRemoveMovie godoc
// @Summary Remove the movie from watchlist
// @Tags watchlist
// @Accept json
// @Produce json
// @Param movieId path int true "Movie id"
// @Success 200
// @Failure 500 {object} models.ApiError
// @Router /watchlist/{movieId} [delete]
func (h *WatchlistHandler) HandleRemoveMovie(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("movieId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Could not parse movie id", zap.String("movieId", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}

	_, err = h.moviesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Could not find movie", zap.String("movieId", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	err = h.watchlistRepo.RemoveFromWatchlist(c, id)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
