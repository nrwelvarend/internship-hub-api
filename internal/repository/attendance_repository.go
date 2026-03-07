package repository

import (
	"time"

	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceRepository interface {
	Create(attendance *models.Attendance) error
	Update(attendance *models.Attendance) error
	FindByID(id string) (models.Attendance, error)
	FindTodayByUser(userID uuid.UUID) (models.Attendance, error)
	FindByUserID(userID uuid.UUID, page, limit int) ([]models.Attendance, int64, error)
	FindAllWithFilters(search string, unitID *uuid.UUID, startDate, endDate string, page, limit int) ([]models.Attendance, int64, error)
	GetRecap(unitID *uuid.UUID, startDate, endDate string) ([]models.Attendance, error)
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) Create(attendance *models.Attendance) error {
	return r.db.Create(attendance).Error
}

func (r *attendanceRepository) Update(attendance *models.Attendance) error {
	return r.db.Save(attendance).Error
}

func (r *attendanceRepository) FindByID(id string) (models.Attendance, error) {
	var attendance models.Attendance
	err := r.db.Preload("User").First(&attendance, "id = ?", id).Error
	return attendance, err
}

func (r *attendanceRepository) FindTodayByUser(userID uuid.UUID) (models.Attendance, error) {
	var attendance models.Attendance
	today := time.Now().Format("2006-01-02")
	err := r.db.Where("user_id = ? AND date = ?", userID, today).First(&attendance).Error
	return attendance, err
}

func (r *attendanceRepository) FindByUserID(userID uuid.UUID, page, limit int) ([]models.Attendance, int64, error) {
	var attendances []models.Attendance
	var total int64

	query := r.db.Model(&models.Attendance{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.Order("date desc").Offset((page - 1) * limit).Limit(limit).Find(&attendances).Error
	return attendances, total, err
}

func (r *attendanceRepository) FindAllWithFilters(search string, unitID *uuid.UUID, startDate, endDate string, page, limit int) ([]models.Attendance, int64, error) {
	var attendances []models.Attendance
	var total int64

	query := r.db.Model(&models.Attendance{}).
		Preload("User").
		Preload("Application.Vacancy.UnitKerja").
		Joins("JOIN applications ON applications.id = attendances.application_id").
		Joins("JOIN vacancies ON vacancies.id = applications.vacancy_id").
		Joins("JOIN users ON users.id = attendances.user_id")

	if search != "" {
		query = query.Where("users.name ILIKE ? OR users.email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if unitID != nil {
		query = query.Where("vacancies.unit_kerja_id = ?", unitID)
	}

	if startDate != "" && endDate != "" {
		query = query.Where("attendances.date BETWEEN ? AND ?", startDate, endDate)
	}

	query.Count(&total)

	err := query.Order("attendances.date desc, attendances.created_at desc").Offset((page - 1) * limit).Limit(limit).Find(&attendances).Error
	return attendances, total, err
}

func (r *attendanceRepository) GetRecap(unitID *uuid.UUID, startDate, endDate string) ([]models.Attendance, error) {
	var attendances []models.Attendance

	query := r.db.Model(&models.Attendance{}).
		Preload("User").
		Preload("Application.Vacancy.UnitKerja").
		Joins("JOIN applications ON applications.id = attendances.application_id").
		Joins("JOIN vacancies ON vacancies.id = applications.vacancy_id").
		Joins("JOIN users ON users.id = attendances.user_id")

	if unitID != nil {
		query = query.Where("vacancies.unit_kerja_id = ?", unitID)
	}

	if startDate != "" && endDate != "" {
		query = query.Where("attendances.date BETWEEN ? AND ?", startDate, endDate)
	}

	err := query.Order("users.name asc, attendances.date asc").Find(&attendances).Error
	return attendances, err
}
