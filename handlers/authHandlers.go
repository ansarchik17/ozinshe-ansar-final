package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"goozinshe/config"
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

func (h *AuthHandlers) SignIn(c *gin.Context) {
	var request signInRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiError{"Invalid request parameters"})
		return
	}
	user, err := h.userRepo.FindByEmail(c, request.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiError{err.Error()})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ApiError{"Invalid credials"})
		return
	}
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(user.Id),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.Config.JwtExpiresIn)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Config.JwtSecretKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiError{"Couldn't sign JWT"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (h *AuthHandlers) SignOut(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (h *AuthHandlers) GetUserInfo(c *gin.Context) {
	userId := c.GetInt("userId")
	user, err := h.userRepo.FindById(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("user not found"))
		return
	}
	c.JSON(http.StatusOK, userResponse{
		Id:    userId,
		Email: user.Email,
		Name:  user.Name,
	})
}
