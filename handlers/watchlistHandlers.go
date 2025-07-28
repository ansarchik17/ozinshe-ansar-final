package handlers

import (
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

func (h *WatchlistHandler) HandleGetMovies(c *gin.Context) {
	movies, err := h.watchlistRepo.GetMoviesFromWatchlist(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, movies)
}

func (h *WatchlistHandler) HandleAddMovie(c *gin.Context) {
	idStr := c.Param("movieId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}
	_, err = h.moviesRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	err = h.watchlistRepo.AddToWatchlist(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

func (h *WatchlistHandler) HandleRemoveMovie(c *gin.Context) {
	idStr := c.Param("movieId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}

	_, err = h.moviesRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	err = h.watchlistRepo.RemoveFromWatchlist(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
