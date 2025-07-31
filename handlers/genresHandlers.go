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

func (h *GenreHandlers) FindAll(c *gin.Context) {
	genres, err := h.repo.FindAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, genres)
}

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
