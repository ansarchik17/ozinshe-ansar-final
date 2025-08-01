package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"goozinshe/logger"
	"goozinshe/models"
	"goozinshe/repositories"
	"net/http"
	"strconv"
)

type UsersHandlers struct {
	repo *repositories.UsersRepository
}

func NewUsersHandlers(repo *repositories.UsersRepository) *UsersHandlers {
	return &UsersHandlers{repo: repo}
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userResponseById struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type updateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type changePasswordRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Create godoc
// @Summary Create a user
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.User true "User data to create"
// @Success 200
// @Failure 500 {object} models.ApiError
// @Router /users [post]
func (h *UsersHandlers) Create(c *gin.Context) {
	logger := logger.GetLogger()
	var request createUserRequest
	err := c.BindJSON(&request)
	if err != nil {
		logger.Error("Could not parse request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid payload"))
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	user := models.User{
		Name:         request.Name,
		Email:        request.Email,
		PasswordHash: string(passwordHash),
	}

	id, err := h.repo.Create(c, user)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not create user"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// FindAll godoc
// @Summary Get all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200
// @Failure 500 {object} models.ApiError
// @Router /users [get]
func (h *UsersHandlers) FindAll(c *gin.Context) {
	logger := logger.GetLogger()
	users, err := h.repo.FindAll(c)
	if err != nil {
		logger.Error("Could not find users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to find users"))
		return
	}
	dtos := make([]userResponse, 0, len(users))
	for _, u := range users {
		r := userResponse{
			Id:    u.Id,
			Name:  u.Name,
			Email: u.Email,
		}
		dtos = append(dtos, r)
	}
	c.JSON(http.StatusOK, dtos)
}

// FindById godoc
// @Summary Find user by id
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User id"
// @Success 200
// @Failure 404 {object} models.ApiError "Invalid id"
// @Failure 500 {object} models.ApiError
// @Router /users/{id} [get]
func (h *UsersHandlers) FindById(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Could not parse id", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Invalid user id"))
		return
	}
	user, err := h.repo.FindById(c, id)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not find user"))
		return
	}
	r := userResponseById{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}
	c.JSON(http.StatusOK, r)
}

// Update godoc
// @Summary Update user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User id"
// @Success 200
// @Failure 404 {object} models.ApiError "Invalid user id"
// @Failure 500 {object} models.ApiError
// @Router /users/{id} [put]
func (h *UsersHandlers) Update(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Could not parse id", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user Id"))
		return
	}

	var request updateUserRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request payload"))
		return
	}

	user, err := h.repo.FindById(c, id)
	if err != nil {
		logger.Error("Could not find user", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}

	user.Name = request.Name
	user.Email = request.Email

	err = h.repo.Update(c, id, user)
	if err != nil {
		logger.Error("Could not update user", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

// ChangePassword godoc
// @Summary Change password
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User id"
// @Success 200
// @Failure 500 {object} models.ApiError
// @Router /users/{id}/{changePassword} [patch]
func (h *UsersHandlers) ChangePassword(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Could not parse id", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user Id"))
		return
	}

	var request changePasswordRequest
	if err := c.BindJSON(&request); err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request payload"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	user, err := h.repo.FindById(c, id)
	if err != nil {
		logger.Error("Could not find user", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}

	user.PasswordHash = string(passwordHash)

	err = h.repo.ChangePassword(c, id, string(passwordHash))
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary Delete a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User id"
// @Success 200
// @Failure 400 {object} models.ApiError "Invalid id"
// @Failure 500 {object} models.ApiError
// @Router /users/{id} [delete]
func (h *UsersHandlers) Delete(c *gin.Context) {
	logger := logger.GetLogger()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Could not parse id", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user Id"))
		return
	}
	_, err = h.repo.FindById(c, id)
	if err != nil {
		logger.Error("Could not find user", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}
	h.repo.Delete(c, id)
	c.Status(http.StatusOK)
}
