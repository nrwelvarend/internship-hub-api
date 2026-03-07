package repository

import (
	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VacancyRepository interface {
	FindAll(unitID string, search string, page, limit int) ([]models.Vacancy, int64, error)
	FindByID(id string) (models.Vacancy, error)
	Create(vacancy *models.Vacancy) error
	UpdateStatus(id string, status models.VacancyStatus, rejectionNote string) error
	FindAllAdmin(role models.UserRole, unitID *uuid.UUID, status string, search string, page, limit int) ([]models.Vacancy, int64, error)
}

type vacancyRepository struct {
	db *gorm.DB
}

func NewVacancyRepository(db *gorm.DB) VacancyRepository {
	return &vacancyRepository{db: db}
}

func (r *vacancyRepository) FindAll(unitID string, search string, page, limit int) ([]models.Vacancy, int64, error) {
	var vacancies []models.Vacancy
	var total int64

	query := r.db.Model(&models.Vacancy{}).Preload("UnitKerja").Where("status = ?", models.VacancyStatusApproved)
	if unitID != "" {
		query = query.Where("unit_kerja_id = ?", unitID)
	}
	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	err := query.Offset((page - 1) * limit).Limit(limit).Find(&vacancies).Error
	return vacancies, total, err
}

func (r *vacancyRepository) FindByID(id string) (models.Vacancy, error) {
	var vacancy models.Vacancy
	err := r.db.Preload("UnitKerja").First(&vacancy, "id = ?", id).Error
	return vacancy, err
}

func (r *vacancyRepository) Create(vacancy *models.Vacancy) error {
	return r.db.Create(vacancy).Error
}

func (r *vacancyRepository) UpdateStatus(id string, status models.VacancyStatus, rejectionNote string) error {
	updates := map[string]interface{}{"status": status}
	if status == models.VacancyStatusRejected {
		updates["rejection_note"] = rejectionNote
	}
	return r.db.Model(&models.Vacancy{}).Where("id = ?", id).Updates(updates).Error
}

func (r *vacancyRepository) FindAllAdmin(role models.UserRole, unitID *uuid.UUID, status string, search string, page, limit int) ([]models.Vacancy, int64, error) {
	var vacancies []models.Vacancy
	var total int64

	query := r.db.Model(&models.Vacancy{}).Preload("UnitKerja")
	if unitID != nil {
		query = query.Where("unit_kerja_id = ?", unitID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	err := query.Order("created_at desc").Offset((page - 1) * limit).Limit(limit).Find(&vacancies).Error
	return vacancies, total, err
}
