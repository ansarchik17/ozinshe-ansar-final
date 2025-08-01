package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
)

type ImageHandlers struct{}

func NewImageHandlers() *ImageHandlers {
	return &ImageHandlers{}
}

// HandleGetImageById godoc
// @Summary Get image by id
// @Tags images
// @Accept json
// @Produce json
// @Param imageId path int true "Image id"
// @Success 200
// @Failure 400 {object} models.ApiError "Invalid image id"
// @Failure 500 {object} models.ApiError
// @Router /images/{imageId} [get]
func (h *ImageHandlers) HandleGetImageById(c *gin.Context) {
	imageId := c.Param("imageId")
	if imageId == "" {
		c.JSON(http.StatusBadRequest, "Invalid image id")
		return
	}

	fileName := filepath.Base(imageId)
	byteFile, err := os.ReadFile(fmt.Sprintf("images/%s", imageId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Data(http.StatusOK, "application/octet-stream", byteFile)
}
