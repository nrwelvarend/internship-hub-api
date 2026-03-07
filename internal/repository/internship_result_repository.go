package repository

import (
	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InternshipResultRepository interface {
	Create(result *models.InternshipResult) error
	Update(result *models.InternshipResult) error
	FindByApplicationID(appID uuid.UUID) (*models.InternshipResult, error)
	FindByUserID(userID uuid.UUID) ([]models.InternshipResult, error)
	FindAllPendingReview(unitKerjaID uuid.UUID, search string, page, limit int) ([]models.InternshipResult, int64, error)
}

type internshipResultRepository struct {
	db *gorm.DB
}

func NewInternshipResultRepository(db *gorm.DB) InternshipResultRepository {
	return &internshipResultRepository{db: db}
}

func (r *internshipResultRepository) Create(result *models.InternshipResult) error {
	return r.db.Create(result).Error
}

func (r *internshipResultRepository) Update(result *models.InternshipResult) error {
	return r.db.Save(result).Error
}

func (r *internshipResultRepository) FindByApplicationID(appID uuid.UUID) (*models.InternshipResult, error) {
	var result models.InternshipResult
	err := r.db.Preload("Application.Vacancy.UnitKerja").Preload("User").First(&result, "application_id = ?", appID).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *internshipResultRepository) FindByUserID(userID uuid.UUID) ([]models.InternshipResult, error) {
	var results []models.InternshipResult
	err := r.db.Preload("Application.Vacancy.UnitKerja").Find(&results, "user_id = ?", userID).Error
	return results, err
}

func (r *internshipResultRepository) FindAllPendingReview(unitKerjaID uuid.UUID, search string, page, limit int) ([]models.InternshipResult, int64, error) {
	var results []models.InternshipResult
	var total int64

	query := r.db.Model(&models.InternshipResult{}).
		Preload("Application.Vacancy.UnitKerja").
		Preload("User").
		Joins("Join applications ON applications.id = internship_results.application_id").
		Joins("Join vacancies ON vacancies.id = applications.vacancy_id")

	if unitKerjaID != uuid.Nil {
		query = query.Where("vacancies.unit_kerja_id = ?", unitKerjaID)
	}

	if search != "" {
		query = query.Where("users.name ILIKE ?", "%"+search+"%")
	}

	query.Count(&total)
	err := query.Offset((page - 1) * limit).Limit(limit).Find(&results).Error
	return results, total, err
}
