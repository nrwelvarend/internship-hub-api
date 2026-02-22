package database

import (
	"fmt"
	"log"

	"github.com/dr15/internship-hub-api/config"
	"github.com/dr15/internship-hub-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	c := config.AppConfig

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		c.DBHost,
		c.DBUser,
		c.DBPass,
		c.DBName,
		c.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Database connection established")

	// Auto Migration
	err = db.AutoMigrate(
		&models.UnitKerja{},
		&models.User{},
		&models.Vacancy{},
		&models.Application{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("Database migration completed")
	DB = db
}
