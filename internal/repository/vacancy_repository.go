package repository

import (
	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VacancyRepository interface {
	FindAll(unitID string) ([]models.Vacancy, error)
	FindByID(id string) (models.Vacancy, error)
	Create(vacancy *models.Vacancy) error
	UpdateStatus(id string, status models.VacancyStatus, rejectionNote string) error
	FindAllAdmin(role models.UserRole, unitID *uuid.UUID) ([]models.Vacancy, error)
}

type vacancyRepository struct {
	db *gorm.DB
}

func NewVacancyRepository(db *gorm.DB) VacancyRepository {
	return &vacancyRepository{db: db}
}

func (r *vacancyRepository) FindAll(unitID string) ([]models.Vacancy, error) {
	var vacancies []models.Vacancy
	query := r.db.Preload("UnitKerja").Where("status = ?", models.VacancyStatusApproved)
	if unitID != "" {
		query = query.Where("unit_kerja_id = ?", unitID)
	}
	err := query.Find(&vacancies).Error
	return vacancies, err
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

func (r *vacancyRepository) FindAllAdmin(role models.UserRole, unitID *uuid.UUID) ([]models.Vacancy, error) {
	var vacancies []models.Vacancy
	query := r.db.Preload("UnitKerja")
	if role == models.UserRoleUnit && unitID != nil {
		query = query.Where("unit_kerja_id = ?", unitID)
	}
	err := query.Order("created_at desc").Find(&vacancies).Error
	return vacancies, err
}
