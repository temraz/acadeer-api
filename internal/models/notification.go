package models

import (
	"time"
)

type Notification struct {
	ID               uint       `gorm:"primarykey" json:"id"`
	UserID           int        `json:"user_id"`
	SchoolID         *int       `json:"school_id"`
	Type             string     `json:"type"`
	RefrenceTable    string     `json:"refrence_table"`
	RefrenceColumnID uint       `json:"refrence_column_id"`
	TitleEn          string     `json:"title_en"`
	TitleAr          string     `json:"title_ar"`
	MessageEn        string     `json:"message_en"`
	MessageAr        string     `json:"message_ar"`
	ReadAt           *time.Time `json:"read_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Virtual fields for school names
	SchoolNameEn string `gorm:"-" json:"school_name_en,omitempty"`
	SchoolNameAr string `gorm:"-" json:"school_name_ar,omitempty"`
}
