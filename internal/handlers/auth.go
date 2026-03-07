package handlers

import (
	"net/http"
	"time"

	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/dr15/internship-hub-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register as an Applicant
// @Summary Register a new applicant
// @Description Create a new account for an internship applicant
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     models.UserRoleApplicant,
	}

	if err := h.UserRepo.Create(&user); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	token, err := utils.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"token":   token,
		"user":    user,
	})
}

// Login for all roles
// @Summary User login
// @Description Login to get access token for all user roles
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.UserRepo.FindByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := utils.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}

type UpdateProfileRequest struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	KTP        string `json:"ktp"`
	University string `json:"university"`
	Major      string `json:"major"`
	Semester   int    `json:"semester"`
}

// Me returns current user profile
// @Summary Get current user profile
// @Description Fetch profile of the currently logged-in user
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /me [get]
func (h *Handler) Me(c *gin.Context) {
	userId, _ := c.Get("userId")
	user, err := h.UserRepo.FindByID(userId.(uuid.UUID).String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates current user profile
// @Summary Update current user profile
// @Description Update profile details of the currently logged-in user
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Update profile request"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /me [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userId, _ := c.Get("userId")
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.UserRepo.FindByID(userId.(uuid.UUID).String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	user.Phone = req.Phone
	user.Address = req.Address
	user.KTP = req.KTP
	user.University = req.University
	user.Major = req.Major
	user.Semester = req.Semester

	if err := h.UserRepo.Update(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// ForgotPassword handles forgot password request
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.UserRepo.FindByEmail(req.Email)
	if err != nil {
		// For security reasons, don't reveal if email exists or not
		c.JSON(http.StatusOK, gin.H{"message": "Jika email terdaftar, instruksi reset password akan dikirimkan."})
		return
	}

	// Generate reset token
	token := uuid.New().String()
	expiry := time.Now().Add(1 * time.Hour) // Token valid for 1 hour

	user.ResetToken = token
	user.ResetExpiry = &expiry

	if err := h.UserRepo.Update(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses permintaan"})
		return
	}

	// Send email
	if err := utils.SendResetPasswordEmail(user.Email, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim email reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instruksi reset password telah dikirim ke email Anda."})
}

// ResetPassword handles password reset
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.UserRepo.FindByResetToken(req.Token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Token tidak valid atau sudah kadaluarsa"})
		return
	}

	// Check expiry
	if user.ResetExpiry == nil || time.Now().After(*user.ResetExpiry) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token sudah kadaluarsa"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses password baru"})
		return
	}

	user.Password = string(hashedPassword)
	user.ResetToken = "" // Clear token
	user.ResetExpiry = nil

	if err := h.UserRepo.Update(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mereset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diperbarui. Silakan login kembali."})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// ChangePassword handles password change for authenticated users
func (h *Handler) ChangePassword(c *gin.Context) {
	userId, _ := c.Get("userId")
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.UserRepo.FindByID(userId.(uuid.UUID).String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password lama salah"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses password baru"})
		return
	}

	user.Password = string(hashedPassword)
	if err := h.UserRepo.Update(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diperbarui"})
}
