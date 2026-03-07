package handlers

import (
	"net/http"

	"github.com/dr15/internship-hub-api/database"
	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/dr15/internship-hub-api/internal/utils"
	"github.com/gin-gonic/gin"
)

// GetUnits returns all Unit Kerja
// @Summary List all unit kerja
// @Description Fetch all available unit kerja departments
// @Tags Unit Kerja
// @Produce json
// @Success 200 {array} models.UnitKerja
// @Failure 500 {object} map[string]string
// @Router /units [get]
func (h *Handler) GetUnits(c *gin.Context) {
	pagination := utils.GetPaginationRequest(c)
	var units []models.UnitKerja
	var total int64

	database.DB.Model(&models.UnitKerja{}).Count(&total)
	if err := database.DB.Offset(pagination.GetOffset()).Limit(pagination.Limit).Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch unit kerja"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: units,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}
