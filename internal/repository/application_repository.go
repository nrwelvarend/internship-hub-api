package repository

import (
	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationRepository interface {
	Create(app *models.Application) error
	FindByUserAndVacancy(userID, vacancyID uuid.UUID) (models.Application, error)
	FindByUserID(userID uuid.UUID, page, limit int) ([]models.Application, int64, error)
	FindByVacancyID(vacancyID string, search string, page, limit int) ([]models.Application, int64, error)
	UpdateStatus(id string, status models.ApplicationStatus, rejectionNote string) error
	FindByID(id string) (models.Application, error)
	CountAcceptedByUser(userID uuid.UUID) (int64, error)
	Update(app *models.Application) error
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

func (r *applicationRepository) FindByUserID(userID uuid.UUID, page, limit int) ([]models.Application, int64, error) {
	var apps []models.Application
	var total int64

	query := r.db.Model(&models.Application{}).Preload("Vacancy.UnitKerja").Where("user_id = ?", userID)

	query.Count(&total)
	err := query.Order("applied_at desc").Offset((page - 1) * limit).Limit(limit).Find(&apps).Error
	return apps, total, err
}

func (r *applicationRepository) FindByVacancyID(vacancyID string, search string, page, limit int) ([]models.Application, int64, error) {
	var apps []models.Application
	var total int64

	query := r.db.Model(&models.Application{}).
		Preload("User").
		Joins("Join users ON users.id = applications.user_id").
		Where("vacancy_id = ?", vacancyID)

	if search != "" {
		query = query.Where("users.name ILIKE ? OR users.email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	err := query.Order("applied_at desc").Offset((page - 1) * limit).Limit(limit).Find(&apps).Error
	return apps, total, err
}

func (r *applicationRepository) UpdateStatus(id string, status models.ApplicationStatus, rejectionNote string) error {
	updates := map[string]interface{}{
		"status":         status,
		"rejection_note": rejectionNote,
	}
	return r.db.Model(&models.Application{}).Where("id = ?", id).Updates(updates).Error
}

func (r *applicationRepository) FindByID(id string) (models.Application, error) {
	var app models.Application
	err := r.db.Preload("Vacancy").First(&app, "id = ?", id).Error
	return app, err
}

func (r *applicationRepository) CountAcceptedByUser(userID uuid.UUID) (int64, error) {
	var count int64
	// Check if user has an accepted application for a vacancy that is still ongoing
	// "Ongoing" is defined as: Current time is before (Deadline + DurationMonths)
	err := r.db.Model(&models.Application{}).
		Joins("JOIN vacancies ON vacancies.id = applications.vacancy_id").
		Where("applications.user_id = ? AND applications.status = ?", userID, models.ApplicationStatusAccepted).
		Where("(vacancies.deadline + (vacancies.duration_months * INTERVAL '1 month')) > NOW()").
		Count(&count).Error
	return count, err
}

func (r *applicationRepository) Update(app *models.Application) error {
	return r.db.Save(app).Error
}
