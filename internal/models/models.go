package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VacancyStatus string

const (
	VacancyStatusDraft    VacancyStatus = "draft"
	VacancyStatusPending  VacancyStatus = "pending"
	VacancyStatusApproved VacancyStatus = "approved"
	VacancyStatusRejected VacancyStatus = "rejected"
)

type ApplicationStatus string

const (
	ApplicationStatusSubmitted ApplicationStatus = "submitted"
	ApplicationStatusReviewed  ApplicationStatus = "reviewed"
	ApplicationStatusAccepted  ApplicationStatus = "accepted"
	ApplicationStatusRejected  ApplicationStatus = "rejected"
)

type UserRole string

const (
	UserRoleApplicant UserRole = "applicant"
	UserRoleUnit      UserRole = "unit"
	UserRoleCentral   UserRole = "central"
)

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type User struct {
	Base
	Name         string    `json:"name"`
	Email        string    `gorm:"uniqueIndex" json:"email"`
	Password     string    `json:"-"`
	Role         UserRole  `json:"role"`
	UnitKerjaID  *uuid.UUID `json:"unitKerjaId,omitempty"`
	UnitKerja    *UnitKerja `json:"unitKerja,omitempty"`
}

type UnitKerja struct {
	Base
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Vacancy struct {
	Base
	Title         string        `json:"title"`
	UnitKerjaID   uuid.UUID     `json:"unitKerjaId"`
	UnitKerja     UnitKerja     `json:"unitKerja"`
	Description   string        `json:"description"`
	Requirements  []string      `gorm:"type:text[]" json:"requirements"`
	Duration      string        `json:"duration"`
	Location      string        `json:"location"`
	Quota         int           `json:"quota"`
	Deadline      time.Time     `json:"deadline"`
	Status        VacancyStatus `json:"status"`
	CreatedBy     uuid.UUID     `json:"createdBy"`
	RejectionNote string        `json:"rejectionNote,omitempty"`
}

type Application struct {
	Base
	UserID       uuid.UUID         `json:"userId"`
	User         User              `json:"user"`
	VacancyID    uuid.UUID         `json:"vacancyId"`
	Vacancy      Vacancy           `json:"vacancy"`
	Phone        string            `json:"phone"`
	University   string            `json:"university"`
	Major        string            `json:"major"`
	Semester     int               `json:"semester"`
	Motivation   string            `json:"motivation"`
	CVFileName   string            `json:"cvFileName"`
	Status       ApplicationStatus `json:"status"`
	AppliedAt    time.Time         `json:"appliedAt"`
}
