package handlers

import (
	"net/http"

	"github.com/dr15/internship-hub-api/database"
	"github.com/dr15/internship-hub-api/internal/models"
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
	var units []models.UnitKerja
	if err := database.DB.Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch unit kerja"})
		return
	}

	c.JSON(http.StatusOK, units)
}
