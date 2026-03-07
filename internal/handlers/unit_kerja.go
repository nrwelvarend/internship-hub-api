package handlers

import (
	"net/http"

	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/dr15/internship-hub-api/internal/utils"
	"github.com/gin-gonic/gin"
)

type UnitKerjaRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// GetUnits returns all Unit Kerja
// @Summary List all unit kerja
// @Description Fetch all available unit kerja departments
// @Tags Unit Kerja
// @Produce json
// @Success 200 {array} models.UnitKerja
// @Failure 500 {object} map[string]string
// @Router /units [get]
func (h *Handler) GetUnits(c *gin.Context) {
	search := c.Query("search")
	pagination := utils.GetPaginationRequest(c)

	units, total, err := h.UnitKerjaRepo.FindAll(search, pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch unit kerja"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: units,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}

// CreateUnit for superadmin
func (h *Handler) CreateUnit(c *gin.Context) {
	var req UnitKerjaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unit := models.UnitKerja{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.UnitKerjaRepo.Create(&unit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create unit kerja"})
		return
	}

	c.JSON(http.StatusCreated, unit)
}

// UpdateUnit for superadmin
func (h *Handler) UpdateUnit(c *gin.Context) {
	id := c.Param("id")
	var req UnitKerjaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unit, err := h.UnitKerjaRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit kerja not found"})
		return
	}

	unit.Name = req.Name
	unit.Description = req.Description

	if err := h.UnitKerjaRepo.Update(&unit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update unit kerja"})
		return
	}

	c.JSON(http.StatusOK, unit)
}

// DeleteUnit for superadmin
func (h *Handler) DeleteUnit(c *gin.Context) {
	id := c.Param("id")

	// Check if unit is in use
	count, err := h.UnitKerjaRepo.CountUsersByUnit(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check unit usage"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Unit kerja tidak dapat dihapus karena masih digunakan oleh user/staf."})
		return
	}

	if err := h.UnitKerjaRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete unit kerja"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unit kerja deleted successfully"})
}
