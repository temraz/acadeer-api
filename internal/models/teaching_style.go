package models

import "time"

// TeachingStyle represents a teaching method or style
type TeachingStyle struct {
	ID          uint      `gorm:"primarykey;type:bigint unsigned" json:"id"`
	StyleNameEn string    `gorm:"type:varchar(50);not null" json:"style_name_en"`
	StyleNameAr string    `gorm:"type:varchar(50);not null" json:"style_name_ar"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TeacherTeachingStyle represents the many-to-many relationship
type TeacherTeachingStyle struct {
	ID              uint      `gorm:"primarykey;type:bigint unsigned" json:"id"`
	UserID          uint      `gorm:"type:bigint unsigned;not null;foreignKey:none" json:"user_id"`
	TeachingStyleID uint      `gorm:"type:bigint unsigned;not null;foreignKey:none" json:"teaching_style_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relationships without foreign keys
	User          *User          `gorm:"-" json:"user,omitempty"`
	TeachingStyle *TeachingStyle `gorm:"-" json:"teaching_style,omitempty"`
}
