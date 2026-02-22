package repository

import (
	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationRepository interface {
	Create(app *models.Application) error
	FindByUserAndVacancy(userID, vacancyID uuid.UUID) (models.Application, error)
	FindByUserID(userID uuid.UUID) ([]models.Application, error)
	FindByVacancyID(vacancyID string) ([]models.Application, error)
	UpdateStatus(id string, status models.ApplicationStatus) error
	FindByID(id string) (models.Application, error)
}

type applicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

func (r *applicationRepository) Create(app *models.Application) error {
	return r.db.Create(app).Error
}

func (r *applicationRepository) FindByUserAndVacancy(userID, vacancyID uuid.UUID) (models.Application, error) {
	var app models.Application
	err := r.db.Where("user_id = ? AND vacancy_id = ?", userID, vacancyID).First(&app).Error
	return app, err
}

func (r *applicationRepository) FindByUserID(userID uuid.UUID) ([]models.Application, error) {
	var apps []models.Application
	err := r.db.Preload("Vacancy.UnitKerja").Where("user_id = ?", userID).Order("applied_at desc").Find(&apps).Error
	return apps, err
}

func (r *applicationRepository) FindByVacancyID(vacancyID string) ([]models.Application, error) {
	var apps []models.Application
	err := r.db.Preload("User").Where("vacancy_id = ?", vacancyID).Order("applied_at desc").Find(&apps).Error
	return apps, err
}

func (r *applicationRepository) UpdateStatus(id string, status models.ApplicationStatus) error {
	return r.db.Model(&models.Application{}).Where("id = ?", id).Update("status", status).Error
}

func (r *applicationRepository) FindByID(id string) (models.Application, error) {
	var app models.Application
	err := r.db.Preload("Vacancy").First(&app, "id = ?", id).Error
	return app, err
}
