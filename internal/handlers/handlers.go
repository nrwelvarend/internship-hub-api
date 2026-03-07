package handlers

import (
	"github.com/dr15/internship-hub-api/internal/repository"
	"github.com/dr15/internship-hub-api/internal/services"
)

type Handler struct {
	UserRepo             repository.UserRepository
	VacancyRepo          repository.VacancyRepository
	ApplicationRepo      repository.ApplicationRepository
	AttendanceRepo       repository.AttendanceRepository
	UnitKerjaRepo        repository.UnitKerjaRepository
	InternshipResultRepo repository.InternshipResultRepository
	PDFService           *services.PDFService
}

func NewHandler(userRepo repository.UserRepository, vacancyRepo repository.VacancyRepository, appRepo repository.ApplicationRepository, attendanceRepo repository.AttendanceRepository, unitRepo repository.UnitKerjaRepository, resultRepo repository.InternshipResultRepository, pdfService *services.PDFService) *Handler {
	return &Handler{
		UserRepo:             userRepo,
		VacancyRepo:          vacancyRepo,
		ApplicationRepo:      appRepo,
		AttendanceRepo:       attendanceRepo,
		UnitKerjaRepo:        unitRepo,
		InternshipResultRepo: resultRepo,
		PDFService:           pdfService,
	}
}
