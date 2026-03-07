package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/dr15/internship-hub-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type CheckInRequest struct {
	Notes string `json:"notes"`
}

type CheckOutRequest struct {
	Notes string `json:"notes"`
}

// CheckIn for intern
func (h *Handler) CheckIn(c *gin.Context) {
	userId, _ := c.Get("userId")
	var req CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Only accepted applicants can check in
	// Check if user has an accepted application
	apps, _, err := h.ApplicationRepo.FindByUserID(userId.(uuid.UUID), 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify application status"})
		return
	}

	var acceptedApp *models.Application
	for _, app := range apps {
		if app.Status == models.ApplicationStatusAccepted {
			acceptedApp = &app
			break
		}
	}

	if acceptedApp == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only accepted interns can perform attendance"})
		return
	}

	// Check if already checked in today
	_, err = h.AttendanceRepo.FindTodayByUser(userId.(uuid.UUID))
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already checked in today"})
		return
	}

	now := time.Now()
	attendance := models.Attendance{
		UserID:        userId.(uuid.UUID),
		ApplicationID: acceptedApp.ID,
		Date:          now,
		CheckIn:       &now,
		Status:        models.AttendanceStatusPresent,
		Notes:         req.Notes,
	}

	if err := h.AttendanceRepo.Create(&attendance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check in"})
		return
	}

	c.JSON(http.StatusCreated, attendance)
}

// CheckOut for intern
func (h *Handler) CheckOut(c *gin.Context) {
	userId, _ := c.Get("userId")
	var req CheckOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	attendance, err := h.AttendanceRepo.FindTodayByUser(userId.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No check-in record found for today"})
		return
	}

	if attendance.CheckOut != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already checked out today"})
		return
	}

	now := time.Now()
	attendance.CheckOut = &now
	if req.Notes != "" {
		if attendance.Notes != "" {
			attendance.Notes += "\nCheckout Notes: " + req.Notes
		} else {
			attendance.Notes = "Checkout Notes: " + req.Notes
		}
	}

	if err := h.AttendanceRepo.Update(&attendance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check out"})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

// GetMyAttendance history for intern
func (h *Handler) GetMyAttendance(c *gin.Context) {
	userId, _ := c.Get("userId")
	pagination := utils.GetPaginationRequest(c)

	attendances, total, err := h.AttendanceRepo.FindByUserID(userId.(uuid.UUID), pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attendance history"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: attendances,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}

// GetAttendanceRecap for admin
func (h *Handler) GetAttendanceRecap(c *gin.Context) {
	role, _ := c.Get("role")
	unitID, _ := c.Get("unitKerjaId")
	search := c.Query("search")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	pagination := utils.GetPaginationRequest(c)

	var unitUUID *uuid.UUID
	if role == models.UserRoleUnit && unitID != nil {
		u := unitID.(*uuid.UUID)
		unitUUID = u
	}

	attendances, total, err := h.AttendanceRepo.FindAllWithFilters(search, unitUUID, startDate, endDate, pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attendance recap"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: attendances,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}

// GetIndividualRecap for admin to see specific intern
func (h *Handler) GetIndividualRecap(c *gin.Context) {
	internIdStr := c.Param("userId")
	internId, err := uuid.Parse(internIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	pagination := utils.GetPaginationRequest(c)
	attendances, total, err := h.AttendanceRepo.FindByUserID(internId, pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch individual recap"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: attendances,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}

// ExportAttendance to Excel
func (h *Handler) ExportAttendance(c *gin.Context) {
	userIdStr := c.Query("userId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	role, _ := c.Get("role")
	unitID, _ := c.Get("unitKerjaId")

	var unitUUID *uuid.UUID
	if role == models.UserRoleUnit && unitID != nil {
		u := unitID.(*uuid.UUID)
		unitUUID = u
	}

	var attendances []models.Attendance
	var err error

	if userIdStr != "" {
		// Individual Export
		internId, err := uuid.Parse(userIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		// We use GetRecap with filters for consistency
		// But if we want individual, we should filter by userID in the Repo.
		// Let's adjust repository if needed. For now let's use a generic approach.
		// Actually I'll use FindAllByUserUnpaginated but without pagination for export.
		// Wait, I didn't add that to Repo. Let's use GetRecap with a trick or add it.
		// Let's add FindAllByUserUnpaginated to repo or just use GetRecap with name filtering.
		// Actually let's just use GetRecap and filter the result here or adjust GetRecap.
		attendances, err = h.AttendanceRepo.GetRecap(unitUUID, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data for export"})
			return
		}
		// Filter by userId if provided
		var filtered []models.Attendance
		for _, a := range attendances {
			if a.UserID == internId {
				filtered = append(filtered, a)
			}
		}
		attendances = filtered
	} else {
		// Overall Export
		attendances, err = h.AttendanceRepo.GetRecap(unitUUID, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data for export"})
			return
		}
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Attendance"
	f.SetSheetName("Sheet1", sheet)

	f.SetCellValue(sheet, "A1", "No")
	f.SetCellValue(sheet, "B1", "Nama")
	f.SetCellValue(sheet, "C1", "Email")
	f.SetCellValue(sheet, "D1", "Unit Kerja")
	f.SetCellValue(sheet, "E1", "Tanggal")
	f.SetCellValue(sheet, "F1", "Check In")
	f.SetCellValue(sheet, "G1", "Check Out")
	f.SetCellValue(sheet, "H1", "Status")
	f.SetCellValue(sheet, "I1", "Catatan")

	for i, a := range attendances {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), a.User.Name)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), a.User.Email)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), a.Application.Vacancy.UnitKerja.Name)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), a.Date.Format("2006-01-02"))
		if a.CheckIn != nil {
			f.SetCellValue(sheet, fmt.Sprintf("F%d", row), a.CheckIn.Format("15:04:05"))
		}
		if a.CheckOut != nil {
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), a.CheckOut.Format("15:04:05"))
		}
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), string(a.Status))
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), a.Notes)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=attendance_recap.xlsx")
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
	}
}
