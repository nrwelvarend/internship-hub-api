package handlers

import (
	"net/http"
	"strconv"
	"time"

	"fmt"
	"path/filepath"

	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/dr15/internship-hub-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ApplicationRequest struct {
	VacancyID  uuid.UUID `json:"vacancyId" binding:"required"`
	Phone      string    `json:"phone" binding:"required"`
	University string    `json:"university" binding:"required"`
	Major      string    `json:"major" binding:"required"`
	Semester   int       `json:"semester" binding:"required"`
	Motivation string    `json:"motivation" binding:"required"`
	CVFileName string    `json:"cvFileName" binding:"required"`
}

type ApplicationReviewRequest struct {
	Status models.ApplicationStatus `json:"status" binding:"required"`
}

// SubmitApplication for applicant
// @Summary Submit an application
// @Description Submit a new internship application for a specific vacancy.
// @Tags Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ApplicationRequest true "Application submission request"
// @Success 201 {object} models.Application
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /applications [post]
func (h *Handler) SubmitApplication(c *gin.Context) {
	vacancyIDStr := c.PostForm("vacancyId")
	vacancyID, err := uuid.Parse(vacancyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vacancy ID"})
		return
	}

	phone := c.PostForm("phone")
	university := c.PostForm("university")
	major := c.PostForm("major")
	semesterStr := c.PostForm("semester")
	semester, _ := strconv.Atoi(semesterStr)
	motivation := c.PostForm("motivation")

	userId, _ := c.Get("userId")

	// Check if already applied
	_, err = h.ApplicationRepo.FindByUserAndVacancy(userId.(uuid.UUID), vacancyID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already applied for this vacancy"})
		return
	}

	// Handle File Upload
	file, err := c.FormFile("cv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CV file is required"})
		return
	}

	// Create a unique filename
	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s-%d%s", userId.(uuid.UUID).String(), time.Now().Unix(), ext)
	uploadPath := filepath.Join("uploads", "cv", newFileName)

	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save CV file"})
		return
	}

	application := models.Application{
		UserID:     userId.(uuid.UUID),
		VacancyID:  vacancyID,
		Phone:      phone,
		University: university,
		Major:      major,
		Semester:   semester,
		Motivation: motivation,
		CVFileName: newFileName,
		Status:     models.ApplicationStatusSubmitted,
		AppliedAt:  time.Now(),
	}

	if err := h.ApplicationRepo.Create(&application); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit application"})
		return
	}

	c.JSON(http.StatusCreated, application)
}

// GetUserApplications for applicant to see their own applications
// @Summary List my applications
// @Description Fetch all applications submitted by the currently logged-in applicant.
// @Tags Applications
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Application
// @Failure 500 {object} map[string]string
// @Router /applications/my [get]
func (h *Handler) GetUserApplications(c *gin.Context) {
	userId, _ := c.Get("userId")
	pagination := utils.GetPaginationRequest(c)

	apps, total, err := h.ApplicationRepo.FindByUserID(userId.(uuid.UUID), pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: apps,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}

// GetVacancyApplications for unit admin to see applicants for a vacancy
// @Summary List vacancy applications (Admin)
// @Description Fetch all applications for a specific vacancy (for Unit Admins).
// @Tags Applications
// @Security BearerAuth
// @Produce json
// @Param vacancyId path string true "Vacancy ID"
// @Success 200 {array} models.Application
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vacancies/{id}/applications [get]
func (h *Handler) GetVacancyApplications(c *gin.Context) {
	id := c.Param("id")
	role, _ := c.Get("role")
	unitId, _ := c.Get("unitKerjaId")
	pagination := utils.GetPaginationRequest(c)

	vacancy, err := h.VacancyRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vacancy not found"})
		return
	}

	// Permission check: Unit admin can only see applications for their own unit's vacancy
	if role == models.UserRoleUnit && unitId != nil && (*unitId.(*uuid.UUID)).String() != vacancy.UnitKerjaID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: can only view applications for your own unit's vacancy"})
		return
	}

	search := c.Query("search")
	applications, total, err := h.ApplicationRepo.FindByVacancyID(id, search, pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: applications,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}

// ReviewApplication for unit admin
// @Summary Review an application
// @Description Update the status of an application (for Unit Admins).
// @Tags Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param request body ApplicationReviewRequest true "Review request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /applications/{id} [patch]
func (h *Handler) ReviewApplication(c *gin.Context) {
	id := c.Param("id")
	var req ApplicationReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	application, err := h.ApplicationRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	role, _ := c.Get("role")
	unitId, _ := c.Get("unitKerjaId")
	// Verify unit admin ownership
	if role == models.UserRoleUnit && unitId != nil && (*unitId.(*uuid.UUID)).String() != application.Vacancy.UnitKerjaID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: application belongs to another unit's vacancy"})
		return
	}

	if err := h.ApplicationRepo.UpdateStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application status updated successfully"})
}
