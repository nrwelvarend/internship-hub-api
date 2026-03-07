package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
	ApplicationStatusCompleted ApplicationStatus = "completed"
)

type AttendanceStatus string

const (
	AttendanceStatusPresent AttendanceStatus = "present"
	AttendanceStatusSick    AttendanceStatus = "sick"
	AttendanceStatusLeave   AttendanceStatus = "leave"
	AttendanceStatusAlpha   AttendanceStatus = "alpha"
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
	Name        string     `json:"name"`
	Email       string     `gorm:"uniqueIndex" json:"email"`
	Password    string     `json:"-"`
	Role        UserRole   `json:"role"`
	UnitKerjaID *uuid.UUID `json:"unitKerjaId,omitempty"`
	UnitKerja   *UnitKerja `json:"unitKerja,omitempty"`
	Phone       string     `json:"phone"`
	Address     string     `json:"address"`
	KTP         string     `json:"ktp"`
	University  string     `json:"university"`
	Major       string     `json:"major"`
	Semester    int        `json:"semester"`
	ResetToken  string     `json:"-"`
	ResetExpiry *time.Time `json:"-"`
}

type UnitKerja struct {
	Base
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Vacancy struct {
	Base
	Title          string         `json:"title"`
	UnitKerjaID    uuid.UUID      `json:"unitKerjaId"`
	UnitKerja      UnitKerja      `json:"unitKerja"`
	Description    string         `json:"description"`
	Requirements   pq.StringArray `gorm:"type:text[]" json:"requirements"`
	Duration       string         `json:"duration"`
	DurationMonths int            `json:"durationMonths"`
	Location       string         `json:"location"`
	Quota          int            `json:"quota"`
	Deadline       time.Time      `json:"deadline"`
	Status         VacancyStatus  `json:"status"`
	CreatedBy      uuid.UUID      `json:"createdBy"`
	RejectionNote  string         `json:"rejectionNote,omitempty"`
}

type Application struct {
	Base
	UserID        uuid.UUID         `json:"userId"`
	User          User              `json:"user"`
	VacancyID     uuid.UUID         `json:"vacancyId"`
	Vacancy       Vacancy           `json:"vacancy"`
	Phone         string            `json:"phone"`
	University    string            `json:"university"`
	Major         string            `json:"major"`
	Semester      int               `json:"semester"`
	Motivation    string            `json:"motivation"`
	CVFileName    string            `json:"cvFileName"`
	Status        ApplicationStatus `json:"status"`
	AppliedAt     time.Time         `json:"appliedAt"`
	RejectionNote string            `json:"rejectionNote,omitempty"`
}

type Attendance struct {
	Base
	UserID        uuid.UUID        `json:"userId"`
	User          User             `json:"user"`
	ApplicationID uuid.UUID        `json:"applicationId"`
	Application   Application      `json:"application"`
	Date          time.Time        `gorm:"type:date;index:idx_attendance_user_date,unique" json:"date"`
	CheckIn       *time.Time       `json:"checkIn"`
	CheckOut      *time.Time       `json:"checkOut"`
	Status        AttendanceStatus `json:"status"`
	Notes         string           `json:"notes"`
}

type InternshipResult struct {
	Base
	ApplicationID        uuid.UUID   `gorm:"uniqueIndex" json:"applicationId"`
	Application          Application `json:"application"`
	UserID               uuid.UUID   `json:"userId"`
	User                 User        `json:"user"`
	AttendanceScore      float64     `json:"attendanceScore"`
	PerformanceScore     float64     `json:"performanceScore"`
	ReportScore          float64     `json:"reportScore"`
	DisciplineScore      float64     `json:"disciplineScore"`
	OtherScore           float64     `json:"otherScore"`
	FinalScore           float64     `json:"finalScore"`
	ReportFileName       string      `json:"reportFileName"`
	CertificatePath      string      `json:"certificatePath"`
	CompletionLetterPath string      `json:"completionLetterPath"`
	ReviewNotes          string      `json:"reviewNotes"`
	ReviewedBy           uuid.UUID   `json:"reviewedBy"`
	ReviewedAt           *time.Time  `json:"reviewedAt"`
}
