package handlers

import (
	"net/http"
	"time"

	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/dr15/internship-hub-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type VacancyRequest struct {
	Title        string         `json:"title" binding:"required"`
	UnitKerjaID  uuid.UUID      `json:"unitKerjaId" binding:"required"`
	Description  string         `json:"description" binding:"required"`
	Requirements pq.StringArray `json:"requirements" binding:"required"`
	Duration     string         `json:"duration" binding:"required"`
	Location     string         `json:"location" binding:"required"`
	Quota        int            `json:"quota" binding:"required"`
	Deadline     string         `json:"deadline" binding:"required"`
}

type ApprovalRequest struct {
	Status        models.VacancyStatus `json:"status" binding:"required"`
	RejectionNote string               `json:"rejectionNote"`
}

// GetVacancies returns all approved vacancies for public/applicants
// @Summary List approved vacancies
// @Description Fetch all vacancies with 'approved' status. Can be filtered by unitId.
// @Tags Vacancies
// @Produce json
// @Param unitId query string false "Filter by Unit Kerja ID"
// @Success 200 {array} models.Vacancy
// @Failure 500 {object} map[string]string
// @Router /vacancies [get]
func (h *Handler) GetVacancies(c *gin.Context) {
	unitId := c.Query("unitId")
	search := c.Query("search")
	pagination := utils.GetPaginationRequest(c)

	vacancies, total, err := h.VacancyRepo.FindAll(unitId, search, pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vacancies"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: vacancies,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}

// GetVacancy returns single vacancy detail
// @Summary Get vacancy detail
// @Description Fetch detailed information about a specific vacancy
// @Tags Vacancies
// @Produce json
// @Param id path string true "Vacancy ID"
// @Success 200 {object} models.Vacancy
// @Failure 404 {object} map[string]string
// @Router /vacancies/{id} [get]
func (h *Handler) GetVacancy(c *gin.Context) {
	id := c.Param("id")
	vacancy, err := h.VacancyRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vacancy not found"})
		return
	}

	c.JSON(http.StatusOK, vacancy)
}

// CreateVacancy for unit admin
// @Summary Create a new vacancy
// @Description Create a new internship vacancy (for Unit Admins). Initial status will be 'pending'.
// @Tags Vacancies
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body VacancyRequest true "Vacancy creation request"
// @Success 201 {object} models.Vacancy
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vacancies [post]
func (h *Handler) CreateVacancy(c *gin.Context) {
	var req VacancyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, _ := c.Get("userId")
	role, _ := c.Get("role")
	unitId, _ := c.Get("unitKerjaId")

	// Verify unit admin is creating for their own unit
	if role == models.UserRoleUnit && unitId != nil && (*unitId.(*uuid.UUID)).String() != req.UnitKerjaID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: can only create vacancy for your own unit"})
		return
	}

	deadline, err := time.Parse("2006-01-02", req.Deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deadline format, use YYYY-MM-DD"})
		return
	}

	vacancy := models.Vacancy{
		Title:        req.Title,
		UnitKerjaID:  req.UnitKerjaID,
		Description:  req.Description,
		Requirements: req.Requirements,
		Duration:     req.Duration,
		Location:     req.Location,
		Quota:        req.Quota,
		Deadline:     deadline,
		Status:       models.VacancyStatusPending,
		CreatedBy:    userId.(uuid.UUID),
	}

	if err := h.VacancyRepo.Create(&vacancy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vacancy"})
		return
	}

	c.JSON(http.StatusCreated, vacancy)
}

// GetAllVacanciesAdmin for central admin to see all vacancies for approval
// @Summary List all vacancies (Admin)
// @Description Fetch all vacancies. Unit Admins see their unit's vacancies, Central Admins see all.
// @Tags Vacancies
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Vacancy
// @Failure 500 {object} map[string]string
// @Router /vacancies/admin [get]
// @Router /vacancies/all [get]
func (h *Handler) GetAllVacanciesAdmin(c *gin.Context) {
	role, _ := c.Get("role")
	unitIdCtx, _ := c.Get("unitKerjaId")
	status := c.Query("status")
	search := c.Query("search")
	unitIdQuery := c.Query("unitId")
	pagination := utils.GetPaginationRequest(c)

	var uid *uuid.UUID
	if role == models.UserRoleUnit && unitIdCtx != nil {
		uid = unitIdCtx.(*uuid.UUID)
	} else if role == models.UserRoleCentral && unitIdQuery != "" {
		parsedUid, err := uuid.Parse(unitIdQuery)
		if err == nil {
			uid = &parsedUid
		}
	}

	vacancies, total, err := h.VacancyRepo.FindAllAdmin(role.(models.UserRole), uid, status, search, pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vacancies"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: vacancies,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}

// ApproveVacancy for central admin
// @Summary Approve or reject a vacancy
// @Description Update the status of a vacancy (for Central Admins).
// @Tags Vacancies
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Vacancy ID"
// @Param request body ApprovalRequest true "Approval request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vacancies/{id}/approve [patch]
func (h *Handler) ApproveVacancy(c *gin.Context) {
	id := c.Param("id")
	var req ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.VacancyRepo.UpdateStatus(id, req.Status, req.RejectionNote); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vacancy status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vacancy status updated successfully"})
}
