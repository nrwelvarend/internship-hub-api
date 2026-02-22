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

	// Initialize Handlers
	h := handlers.NewHandler(userRepo, vacancyRepo, appRepo)

	port := config.AppConfig.ServerPort
	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		api.GET("/units", h.GetUnits)
		api.GET("/vacancies", h.GetVacancies)
		api.GET("/vacancies/:id", h.GetVacancy)
	}

	// Protected Routes
	auth := api.Group("")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/me", h.Me)

		// Applicant Routes
		applicant := auth.Group("")
		applicant.Use(middleware.RoleMiddleware(models.UserRoleApplicant))
		{
			applicant.POST("/applications", h.SubmitApplication)
			applicant.GET("/applications/my", h.GetUserApplications)
		}

		// Unit Admin Routes
		unit := auth.Group("")
		unit.Use(middleware.RoleMiddleware(models.UserRoleUnit))
		{
			unit.POST("/vacancies", h.CreateVacancy)
			unit.GET("/vacancies/admin", h.GetAllVacanciesAdmin)
			unit.GET("/vacancies/:id/applications", h.GetVacancyApplications)
			unit.PATCH("/applications/:id", h.ReviewApplication)
		}

		// Central Admin Routes
		central := auth.Group("")
		central.Use(middleware.RoleMiddleware(models.UserRoleCentral))
		{
			central.GET("/vacancies/all", h.GetAllVacanciesAdmin)
			central.PATCH("/vacancies/:id/approve", h.ApproveVacancy)
		}
	}

	fmt.Printf("Server running on port %s\n", port)
	r.Run(":" + port)
}
