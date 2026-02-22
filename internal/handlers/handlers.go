package handlers

import (
	"github.com/dr15/internship-hub-api/internal/repository"
)

type Handler struct {
	UserRepo        repository.UserRepository
	VacancyRepo     repository.VacancyRepository
	ApplicationRepo repository.ApplicationRepository
}

func NewHandler(userRepo repository.UserRepository, vacancyRepo repository.VacancyRepository, appRepo repository.ApplicationRepository) *Handler {
	return &Handler{
		UserRepo:        userRepo,
		VacancyRepo:     vacancyRepo,
		ApplicationRepo: appRepo,
	}
}
