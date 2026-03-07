package repository

import (
	"github.com/dr15/internship-hub-api/internal/models"
	"gorm.io/gorm"
)

type UnitKerjaRepository interface {
	FindAll(search string, page, limit int) ([]models.UnitKerja, int64, error)
	FindByID(id string) (models.UnitKerja, error)
	Create(unit *models.UnitKerja) error
	Update(unit *models.UnitKerja) error
	Delete(id string) error
	CountUsersByUnit(unitID string) (int64, error)
}

type unitKerjaRepository struct {
	db *gorm.DB
}

func NewUnitKerjaRepository(db *gorm.DB) UnitKerjaRepository {
	return &unitKerjaRepository{db: db}
}

func (r *unitKerjaRepository) FindAll(search string, page, limit int) ([]models.UnitKerja, int64, error) {
	var units []models.UnitKerja
	var total int64

	query := r.db.Model(&models.UnitKerja{})
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	err := query.Offset((page - 1) * limit).Limit(limit).Find(&units).Error
	return units, total, err
}

func (r *unitKerjaRepository) FindByID(id string) (models.UnitKerja, error) {
	var unit models.UnitKerja
	err := r.db.First(&unit, "id = ?", id).Error
	return unit, err
}

func (r *unitKerjaRepository) Create(unit *models.UnitKerja) error {
	return r.db.Create(unit).Error
}

func (r *unitKerjaRepository) Update(unit *models.UnitKerja) error {
	return r.db.Save(unit).Error
}

func (r *unitKerjaRepository) Delete(id string) error {
	return r.db.Delete(&models.UnitKerja{}, "id = ?", id).Error
}

func (r *unitKerjaRepository) CountUsersByUnit(unitID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("unit_kerja_id = ?", unitID).Count(&count).Error
	return count, err
}
