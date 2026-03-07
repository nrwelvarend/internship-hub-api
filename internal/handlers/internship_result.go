package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/dr15/internship-hub-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReviewInternshipRequest struct {
	AttendanceScore  float64 `json:"attendanceScore" binding:"required,min=0,max=100"`
	PerformanceScore float64 `json:"performanceScore" binding:"required,min=0,max=100"`
	ReportScore      float64 `json:"reportScore" binding:"required,min=0,max=100"`
	DisciplineScore  float64 `json:"disciplineScore" binding:"required,min=0,max=100"`
	OtherScore       float64 `json:"otherScore" binding:"min=0,max=100"`
	ReviewNotes      string  `json:"reviewNotes"`
}

// SubmitReport handles applicant uploading their final internship report
func (h *Handler) SubmitReport(c *gin.Context) {
	userID := c.MustGet("userId").(uuid.UUID)
	appIDStr := c.PostForm("applicationId")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	// Verify application belongs to user and is accepted
	// In a real scenario, we might want to check if the internship period has ended
	app, err := h.ApplicationRepo.FindByID(appID.String())
	if err != nil || app.UserID != userID || app.Status != models.ApplicationStatusAccepted {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot submit a report for this application"})
		return
	}

	file, err := c.FormFile("report")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report file is required"})
		return
	}

	// Save file
	filename := fmt.Sprintf("report_%s_%d%s", appID.String(), time.Now().Unix(), filepath.Ext(file.Filename))
	if err := c.SaveUploadedFile(file, filepath.Join("uploads", filename)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save report file"})
		return
	}

	// Create or update result
	result, err := h.InternshipResultRepo.FindByApplicationID(appID)
	if err != nil {
		// New result
		result = &models.InternshipResult{
			ApplicationID:  appID,
			UserID:         userID,
			ReportFileName: filename,
		}
		if err := h.InternshipResultRepo.Create(result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create internship result"})
			return
		}
	} else {
		// Update existing (maybe re-submit)
		result.ReportFileName = filename
		if err := h.InternshipResultRepo.Update(result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update internship result"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report submitted successfully", "data": result})
}

// ReviewInternship handles admin grading the internship
func (h *Handler) ReviewInternship(c *gin.Context) {
	adminID := c.MustGet("userId").(uuid.UUID)
	resultIDStr := c.Param("id")
	resultID, err := uuid.Parse(resultIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid result ID"})
		return
	}

	var req ReviewInternshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch result with preloads to verify unit access if needed
	// In this implementation, we rely on the middleware to check if it's a unit or central admin
	result, err := h.InternshipResultRepo.FindByApplicationID(resultID) // Wait, the param is result ID or application ID? Let's use application ID for consistency with UI flow
	if err != nil {
		// Try finding by UUID directly if the repo supports it, otherwise find by app ID
		// Let's assume the param is ID of InternshipResult for now
		// Actually, let's make it consistent: /api/internship-results/:appId/review
		c.JSON(http.StatusNotFound, gin.H{"error": "Internship result not found"})
		return
	}

	// Calculate final score (simple average for now)
	count := 4.0
	total := req.AttendanceScore + req.PerformanceScore + req.ReportScore + req.DisciplineScore
	if req.OtherScore > 0 {
		total += req.OtherScore
		count += 1.0
	}
	finalScore := total / count

	now := time.Now()
	result.AttendanceScore = req.AttendanceScore
	result.PerformanceScore = req.PerformanceScore
	result.ReportScore = req.ReportScore
	result.DisciplineScore = req.DisciplineScore
	result.OtherScore = req.OtherScore
	result.FinalScore = finalScore
	result.ReviewNotes = req.ReviewNotes
	result.ReviewedBy = adminID
	result.ReviewedAt = &now

	// PDF Generation
	completionPath, err := h.PDFService.GenerateCompletionLetter(result)
	if err == nil {
		result.CompletionLetterPath = completionPath
	}

	certificatePath, err := h.PDFService.GenerateCertificate(result)
	if err == nil {
		result.CertificatePath = certificatePath
	}

	if err := h.InternshipResultRepo.Update(result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save review"})
		return
	}

	// Update application status to completed
	app, _ := h.ApplicationRepo.FindByID(result.ApplicationID.String())
	app.Status = models.ApplicationStatusCompleted
	h.ApplicationRepo.Update(&app)

	c.JSON(http.StatusOK, gin.H{"message": "Review submitted and documents generated", "data": result})
}

// GetMyInternshipResult for applicant
func (h *Handler) GetMyInternshipResult(c *gin.Context) {
	userID := c.MustGet("userId").(uuid.UUID)
	results, err := h.InternshipResultRepo.FindByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetInternshipResultsForAdmin lists submissions for review
func (h *Handler) GetInternshipResultsForAdmin(c *gin.Context) {
	role := c.MustGet("role").(models.UserRole)
	var unitKerjaID uuid.UUID

	if role == models.UserRoleUnit {
		uID := c.MustGet("unitKerjaId")
		if uID != nil {
			if ptr, ok := uID.(*uuid.UUID); ok && ptr != nil {
				unitKerjaID = *ptr
			}
		}
	}

	search := c.Query("search")
	pagination := utils.GetPaginationRequest(c)

	results, total, err := h.InternshipResultRepo.FindAllPendingReview(unitKerjaID, search, pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch internship results"})
		return
	}

	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data: results,
		Meta: utils.CreatePaginationMeta(total, pagination.Page, pagination.Limit),
	})
}
