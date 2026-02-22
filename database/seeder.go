package database

import (
	"fmt"

	"github.com/dr15/internship-hub-api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedData(db *gorm.DB) {
	// 1. Seed Unit Kerja
	units := []models.UnitKerja{
		{Name: "Bagian Teknologi Informasi", Description: "Mengelola sistem informasi dan infrastruktur IT"},
		{Name: "Bagian Keuangan", Description: "Mengelola keuangan dan akuntansi"},
		{Name: "Bagian SDM", Description: "Mengelola sumber daya manusia"},
		{Name: "Bagian Humas", Description: "Mengelola hubungan masyarakat dan komunikasi"},
	}

	for _, unit := range units {
		var existing models.UnitKerja
		if err := db.Where("name = ?", unit.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				db.Create(&unit)
			}
		}
	}

	// 2. Seed Admin Users
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	// Central Admin
	adminCentral := models.User{
		Name:     "Super Admin",
		Email:    "admin@instansi.go.id",
		Password: string(hashedPassword),
		Role:     models.UserRoleCentral,
	}
	var existingCentral models.User
	if err := db.Where("email = ?", adminCentral.Email).First(&existingCentral).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			db.Create(&adminCentral)
		}
	}

	// Unit Admins
	unitNames := []string{"Bagian Teknologi Informasi", "Bagian Keuangan", "Bagian SDM", "Bagian Humas"}
	unitEmails := []string{"admin.it@instansi.go.id", "admin.keu@instansi.go.id", "admin.sdm@instansi.go.id", "admin.humas@instansi.go.id"}

	for i, name := range unitNames {
		var unit models.UnitKerja
		db.Where("name = ?", name).First(&unit)

		adminUnit := models.User{
			Name:        fmt.Sprintf("Admin %s", name),
			Email:       unitEmails[i],
			Password:    string(hashedPassword),
			Role:        models.UserRoleUnit,
			UnitKerjaID: &unit.ID,
		}

		var existingUnit models.User
		if err := db.Where("email = ?", adminUnit.Email).First(&existingUnit).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				db.Create(&adminUnit)
			}
		}
	}

	fmt.Println("Seeding completed")
}
