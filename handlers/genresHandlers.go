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

type GenreHandlers struct {
	repo *repositories.GenresRepository
}

func NewGenreHandlers(repo *repositories.GenresRepository) *GenreHandlers {
	return &GenreHandlers{
		repo: repo,
	}
}

// FindById godoc
// @Summary  Find by id
// @Tags  genres
// @Accept  json
// @Produce  json
// @Param id path int true "Genre id"
// @Success 200 {object} models.Genre "OK"
// @Failure 400 {object} models.ApiError "Invalid Genre id"
// @Failure 404 {object} models.ApiError "Genre not found"
// @Failure 500 {object} models.ApiError
// @Router /genres/{id} [get]
func (h *GenreHandlers) FindById(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Could not parse id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Genre Id"))
		return
	}

	genre, err := h.repo.FindById(c, id)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, genre)
}

// FindAll godoc
// @Summary     Find all genres
// @Tags        genres
// @Accept      json
// @Produce     json
// @Success     200 {array} models.Genre
// @Failure     500 {object} models.ApiError
// @Router      /genres [get]
func (h *GenreHandlers) FindAll(c *gin.Context) {
	genres, err := h.repo.FindAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, genres)
}

// Create godoc
// @Summary     Create a new genre
// @Tags        genres
// @Accept      json
// @Produce     json
// @Param       genre body models.Genre true "Genre to create"
// @Success     200 {object} map[string]int
// @Failure     400 {object} models.ApiError
// @Router      /genres [post]
func (h *GenreHandlers) Create(c *gin.Context) {
	logger := logger.GetLogger()
	var g models.Genre
	err := c.BindJSON(&g)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	id, err := h.repo.Create(c, g)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

// Update godoc
// @Summary Update a user
// @Tags genres
// @Accept json
// @Produce json
// @Param id path int true "Genre id"
// @Param genre body models.Genre true "Updated genre data"
// @Success 200
// @Failure 400 {object} models.ApiError
// @Router /genres/{id} [put]
func (h *GenreHandlers) Update(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Could not parse id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Genre Id"))
		return
	}

	_, err = h.repo.FindById(c, id)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var updatedGenre models.Genre
	err = c.BindJSON(&updatedGenre)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.repo.Update(c, id, updatedGenre)
	if err != nil {
		logger.Error("Could not update Genre", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary Delete a user
// @Tags genres
// @Accept json
// @Produce json
// @Param id path int true "Genre id"
// @Success 200
// @Failure 400 {object} models.ApiError
// @Router /genres/{id} [delete]
func (h *GenreHandlers) Delete(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Could not find genre", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Genre Id"))
		return
	}

	_, err = h.repo.FindById(c, id)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.repo.Delete(c, id)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
