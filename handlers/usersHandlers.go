package handlers

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
	Name     string
	Email    string
	Password string
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

func (h *UsersHandlers) Create(c *gin.Context) {
	var request createUserRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid payload"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not create user"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *UsersHandlers) FindAll(c *gin.Context) {
	users, err := h.repo.FindAll(c)
	if err != nil {
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

func (h *UsersHandlers) FindById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Invalid user id"))
		return
	}
	user, err := h.repo.FindById(c, id)
	if err != nil {
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

func (h *UsersHandlers) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
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
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}

	user.Name = request.Name
	user.Email = request.Email

	err = h.repo.Update(c, id, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

func (h *UsersHandlers) ChangePassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user Id"))
		return
	}

	var request changePasswordRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request payload"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	user, err := h.repo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}

	user.PasswordHash = string(passwordHash)

	err = h.repo.ChangePassword(c, id, string(passwordHash))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

func (h *UsersHandlers) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user Id"))
		return
	}
	_, err = h.repo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}
	h.repo.Delete(c, id)
	c.Status(http.StatusOK)
}
