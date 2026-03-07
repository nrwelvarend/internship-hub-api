package repository

import (
	"strings"

	"github.com/dr15/internship-hub-api/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (models.User, error)
	FindByID(id string) (models.User, error)
	FindByResetToken(token string) (models.User, error)
	Update(user *models.User) error
	FindAll(role string, search string, page, limit int) ([]models.User, int64, error)
	Delete(id string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).Preload("UnitKerja").First(&user).Error
	return user, err
}

func (r *userRepository) FindByID(id string) (models.User, error) {
	var user models.User
	err := r.db.Preload("UnitKerja").First(&user, "id = ?", id).Error
	return user, err
}

func (r *userRepository) FindByResetToken(token string) (models.User, error) {
	var user models.User
	err := r.db.Where("reset_token = ?", token).First(&user).Error
	return user, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) FindAll(role string, search string, page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{}).Preload("UnitKerja")
	if role != "" {
		if strings.Contains(role, ",") {
			roles := strings.Split(role, ",")
			query = query.Where("role IN ?", roles)
		} else {
			query = query.Where("role = ?", role)
		}
	}
	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	err := query.Offset((page - 1) * limit).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}
