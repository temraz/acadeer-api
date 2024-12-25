package models

import (
	"time"
)

type TeacherBooking struct {
	ID                uint       `gorm:"primarykey" json:"id"`
	SchoolAdminUserID uint       `json:"school_admin_user_id"`
	TeacherUserID     uint       `json:"teacher_user_id"`
	SchoolID          uint       `json:"school_id"`
	Status            int        `gorm:"default:1;comment:'1=pending, 2=accepted, 3=rejected'" json:"status"`
	TeacherAcceptedAt *time.Time `json:"teacher_accepted_at"`
	PricePerDay       float64    `gorm:"type:decimal(10,2)" json:"price_per_day"`
	PaymentStatus     int        `gorm:"default:1;comment:'1=pending, 2=paid, 3=failed'" json:"payment_status"`
	StartDate         time.Time  `json:"start_date"`
	EndDate           time.Time  `json:"end_date"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// Relations
	Teacher     *User   `gorm:"foreignKey:TeacherUserID" json:"teacher,omitempty"`
	SchoolAdmin *User   `gorm:"foreignKey:SchoolAdminUserID" json:"school_admin,omitempty"`
	School      *School `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
}

type CreateBookingRequest struct {
	TeacherUserID uint   `json:"teacher_user_id" binding:"required"`
	StartDate     string `json:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate       string `json:"end_date" binding:"required,datetime=2006-01-02"`
}

type BookingResponse struct {
	ID                uint       `json:"id"`
	TeacherUserID     uint       `json:"teacher_user_id"`
	SchoolID          uint       `json:"school_id"`
	Status            int        `json:"status"`
	TeacherAcceptedAt *time.Time `json:"teacher_accepted_at"`
	PricePerDay       float64    `json:"price_per_day"`
	PaymentStatus     int        `json:"payment_status"`
	StartDate         time.Time  `json:"start_date"`
	EndDate           time.Time  `json:"end_date"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	Teacher           *User      `json:"teacher,omitempty"`
	School            *School    `json:"school,omitempty"`
}
