package handlers

import (
	"net/http"

	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/dr15/internship-hub-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserRequest struct {
	Name        string          `json:"name" binding:"required"`
	Email       string          `json:"email" binding:"required,email"`
	Password    string          `json:"password" binding:"required,min=6"`
	Role        models.UserRole `json:"role" binding:"required"`
	UnitKerjaID *uuid.UUID      `json:"unitKerjaId"`
}

type UpdateUserRequest struct {
	Name        string          `json:"name"`
	Role        models.UserRole `json:"role"`
	UnitKerjaID *uuid.UUID      `json:"unitKerjaId"`
}

// GetUsers for superadmin
func (h *Handler) GetUsers(c *gin.Context) {
	role := c.Query("role")
	search := c.Query("search")
	pagination := utils.GetPaginationRequest(c)

	users, total, err := h.UserRepo.FindAll(role, search, pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: users,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}

// CreateUser for superadmin
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	user := models.User{
		Name:        req.Name,
		Email:       req.Email,
		Password:    string(hashedPassword),
		Role:        req.Role,
		UnitKerjaID: req.UnitKerjaID,
	}

	if err := h.UserRepo.Create(&user); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser for superadmin
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.UserRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	user.UnitKerjaID = req.UnitKerjaID

	if err := h.UserRepo.Update(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser for superadmin
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.UserRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
