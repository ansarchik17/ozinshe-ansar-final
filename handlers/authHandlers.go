package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"goozinshe/config"
	"goozinshe/logger"
	"goozinshe/models"
	"goozinshe/repositories"
	"net/http"
	"strconv"
	"time"
)

type AuthHandlers struct {
	userRepo *repositories.UsersRepository
}

func NewAuthHandlers(userRepo *repositories.UsersRepository) *AuthHandlers {
	return &AuthHandlers{userRepo: userRepo}
}

type signInRequest struct {
	Email    string
	Password string
}

// SignIn godoc
// @Summary Sign in
// @Tags authorization
// @Accept json
// @Produce json
// @Success 200
// @Failure 500 {object} models.ApiError
// @Router /auth/{signIn} [post]
func (h *AuthHandlers) SignIn(c *gin.Context) {
	logger := logger.GetLogger()
	var request signInRequest
	err := c.BindJSON(&request)
	if err != nil {
		logger.Error("Could not sign in to the account", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request parameters"))
		return
	}
	user, err := h.userRepo.FindByEmail(c, request.Email)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusUnauthorized, models.NewApiError("Invalid credials"))
		return
	}
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(user.Id),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.Config.JwtExpiresIn)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Config.JwtSecretKey))
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't sign JWT"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// SignOut godoc
// @Summary Sign out
// @Tags authorization
// @Accept json
// @Produce json
// @Success 200
// @Failure 500 {object} models.ApiError
// @Router /auth/signOut [post]
func (h *AuthHandlers) SignOut(c *gin.Context) {
	c.Status(http.StatusOK)
}

// GetUserInfo godoc
// @Summary Get user info
// @Tags authorization
// @Accept json
// @Produce json
// @Param userId path int true "User id"
// @Success 200
// @Failure 500 {object} models.ApiError
// @Router /auth/userInfo [get]
func (h *AuthHandlers) GetUserInfo(c *gin.Context) {
	logger := logger.GetLogger()
	userId := c.GetInt("userId")
	user, err := h.userRepo.FindById(c, userId)
	if err != nil {
		logger.Error("Could not find user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("user not found"))
		return
	}
	c.JSON(http.StatusOK, userResponse{
		Id:    userId,
		Email: user.Email,
		Name:  user.Name,
	})
}
