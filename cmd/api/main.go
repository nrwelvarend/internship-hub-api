package main

import (
	"fmt"

	"github.com/dr15/internship-hub-api/config"
	"github.com/dr15/internship-hub-api/database"
	_ "github.com/dr15/internship-hub-api/docs"
	"github.com/dr15/internship-hub-api/internal/handlers"
	"github.com/dr15/internship-hub-api/internal/middleware"
	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/dr15/internship-hub-api/internal/repository"
	"github.com/dr15/internship-hub-api/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Internship Hub API
// @version 1.0
// @description Backend API for Internship Hub Global System
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load Configuration
	config.LoadConfig()

	// Initialize Database
	database.ConnectDB()

	// Seed Data
	database.SeedData(database.DB)

	// Initialize Repositories
	userRepo := repository.NewUserRepository(database.DB)
	vacancyRepo := repository.NewVacancyRepository(database.DB)
	appRepo := repository.NewApplicationRepository(database.DB)
	attendanceRepo := repository.NewAttendanceRepository(database.DB)
	unitKerjaRepo := repository.NewUnitKerjaRepository(database.DB)
	resultRepo := repository.NewInternshipResultRepository(database.DB)
	pdfService := services.NewPDFService("uploads")

	// Initialize Handlers
	h := handlers.NewHandler(userRepo, vacancyRepo, appRepo, attendanceRepo, unitKerjaRepo, resultRepo, pdfService)

	port := config.AppConfig.ServerPort
	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Static files
	r.Static("/uploads", "./uploads")

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public Routes
	api := r.Group("/api")
	{
		api.POST("/register", h.Register)
		api.POST("/login", h.Login)
		api.POST("/forgot-password", h.ForgotPassword)
		api.POST("/reset-password", h.ResetPassword)
		api.GET("/units", h.GetUnits)
		api.GET("/vacancies", h.GetVacancies)
		api.GET("/vacancies/:id", h.GetVacancy)
	}

	// Protected Routes
	auth := api.Group("")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/me", h.Me)
		auth.PUT("/me", h.UpdateProfile)
		auth.POST("/change-password", h.ChangePassword)

		// Applicant Routes
		applicant := auth.Group("")
		applicant.Use(middleware.RoleMiddleware(models.UserRoleApplicant))
		{
			applicant.POST("/applications", h.SubmitApplication)
			applicant.GET("/applications/my", h.GetUserApplications)
			// Attendance for intern
			applicant.POST("/attendance/check-in", h.CheckIn)
			applicant.POST("/attendance/check-out", h.CheckOut)
			applicant.GET("/attendance/my", h.GetMyAttendance)
			// Internship Result
			applicant.POST("/internship/report", h.SubmitReport)
			applicant.GET("/internship/result/my", h.GetMyInternshipResult)
		}

		// Administrative Routes (Both Unit and Central Admins)
		admin := auth.Group("")
		admin.Use(middleware.RoleMiddleware(models.UserRoleUnit, models.UserRoleCentral))
		{
			admin.POST("/vacancies", h.CreateVacancy)
			admin.GET("/vacancies/admin", h.GetAllVacanciesAdmin)
			admin.GET("/vacancies/:id/applications", h.GetVacancyApplications)
			admin.PATCH("/applications/:id", h.ReviewApplication)
			// Attendance recap for admin
			admin.GET("/attendance/recap", h.GetAttendanceRecap)
			admin.GET("/attendance/recap/:userId", h.GetIndividualRecap)
			admin.GET("/attendance/export", h.ExportAttendance)
			// Internship Evaluation
			admin.GET("/internship/results", h.GetInternshipResultsForAdmin)
			admin.POST("/internship/results/:id/review", h.ReviewInternship)
		}

		// Central Admin Only Routes
		central := auth.Group("")
		central.Use(middleware.RoleMiddleware(models.UserRoleCentral))
		{
			central.PATCH("/vacancies/:id/approve", h.ApproveVacancy)

			// User Management
			central.GET("/users", h.GetUsers)
			central.POST("/users", h.CreateUser)
			central.PUT("/users/:id", h.UpdateUser)
			central.DELETE("/users/:id", h.DeleteUser)

			// Unit Kerja Management
			central.POST("/units", h.CreateUnit)
			central.PUT("/units/:id", h.UpdateUnit)
			central.DELETE("/units/:id", h.DeleteUnit)
		}
	}

	fmt.Printf("Server running on port %s\n", port)
	r.Run(":" + port)
}
