package repository

import (
	"github.com/dr15/internship-hub-api/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (models.User, error)
	FindByID(id string) (models.User, error)
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
