package handlers

import (
	"fmt"
	"go.uber.org/zap"
	"goozinshe/logger"
	"goozinshe/models"
	"goozinshe/repositories"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MoviesHandler struct {
	moviesRepo *repositories.MoviesRepository
	genresRepo *repositories.GenresRepository
}

type createMovieRequest struct {
	Title       string                `form:"title"`
	Description string                `form:"description"`
	ReleaseYear int                   `form:"releaseYear"`
	Director    string                `form:"director"`
	TrailerUrl  string                `form:"trailerUrl"`
	GenreIds    []int                 `form:"genreIds"`
	Poster      *multipart.FileHeader `form:"poster"`
}

type updateMovieRequest struct {
	Title       string                `form:"title"`
	Description string                `form:"description"`
	ReleaseYear int                   `form:"releaseYear"`
	Director    string                `form:"director"`
	TrailerUrl  string                `form:"trailerUrl"`
	GenreIds    []int                 `form:"genreIds"`
	Poster      *multipart.FileHeader `form:"poster"`
}

func NewMoviesHandler(
	moviesRepo *repositories.MoviesRepository,
	genreRepo *repositories.GenresRepository) *MoviesHandler {
	return &MoviesHandler{
		moviesRepo: moviesRepo,
		genresRepo: genreRepo,
	}
}

func (h *MoviesHandler) FindById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Movie Id"))
		return
	}

	movie, err := h.moviesRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, movie)
}

func (h *MoviesHandler) FindAll(c *gin.Context) {
	filters := models.MovieFilters{
		SearchTerm: c.Query("search"),
		IsWatched:  c.Query("iswatched"),
		GenreId:    c.Query("genreids"),
		Sort:       c.Query("sort"),
	}

	movies, err := h.moviesRepo.FindAll(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, movies)
}

func (h *MoviesHandler) Create(c *gin.Context) {
	var request createMovieRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Could not bind payload"))
		return
	}

	genres, err := h.genresRepo.FindAllByIds(c, request.GenreIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	filename, err := h.saveMoviePoster(c, request.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	movie := models.Movie{
		Title:       request.Title,
		Description: request.Description,
		ReleaseYear: request.ReleaseYear,
		Director:    request.Director,
		TrailerUrl:  request.TrailerUrl,
		PosterUrl:   filename,
		Genres:      genres,
	}

	id, err := h.moviesRepo.Create(c, movie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	logger := logger.GetLogger()
	logger.Info("Movie has been created", zap.Int("movie_id", id))

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *MoviesHandler) saveMoviePoster(c *gin.Context, poster *multipart.FileHeader) (string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(poster.Filename))
	filepath := fmt.Sprintf("images/%s", filename)
	err := c.SaveUploadedFile(poster, filepath)

	return filename, err
}

func (h *MoviesHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Movie Id"))
		return
	}

	_, err = h.moviesRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request updateMovieRequest
	err = c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Could not bind payload"))
		return
	}

	genres, err := h.genresRepo.FindAllByIds(c, request.GenreIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	filename, err := h.saveMoviePoster(c, request.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	movie := models.Movie{
		Title:       request.Title,
		Description: request.Description,
		ReleaseYear: request.ReleaseYear,
		Director:    request.Director,
		TrailerUrl:  request.TrailerUrl,
		PosterUrl:   filename,
		Genres:      genres,
	}

	err = h.moviesRepo.Update(c, id, movie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	logger := logger.GetLogger()
	logger.Info("Movie has been updated", zap.Int("movie_id", id))

	c.Status(http.StatusOK)
}

func (h *MoviesHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Movie Id"))
		return
	}

	_, err = h.moviesRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.moviesRepo.Delete(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	logger := logger.GetLogger()
	logger.Info("Movie has been deleted", zap.Int("movie_id", id))
	c.Status(http.StatusOK)
}

func (h *MoviesHandler) HandleSetRating(c *gin.Context) {
	idStr := c.Param("movieId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Movie Id"))
		return
	}

	ratingStr := c.Query("rating")
	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid rating value"))
		return
	}

	err = h.moviesRepo.SetRating(c, id, rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	logger := logger.GetLogger()
	logger.Info("Could not set rating", zap.Int("movie_id", id))
	c.Status(http.StatusOK)
}

func (h *MoviesHandler) HandleSetWatched(c *gin.Context) {
	idStr := c.Param("movieId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Movie Id"))
		return
	}

	isWatchedStr := c.Query("isWatched")
	isWatched, err := strconv.ParseBool(isWatchedStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid isWatched value"))
		return
	}

	err = h.moviesRepo.SetWatched(c, id, isWatched)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
