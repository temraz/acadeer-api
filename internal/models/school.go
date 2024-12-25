package models

import (
	"time"
)

type School struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	NameEn    string    `gorm:"not null" json:"name_en"`
	NameAr    string    `gorm:"not null" json:"name_ar"`
	Logo      string    `gorm:"not null" json:"logo"`
	CityID    uint      `gorm:"not null" json:"city_id"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SchoolSignupRequest struct {
	NameEn        string `form:"name_en" binding:"required"`
	NameAr        string `form:"name_ar" binding:"required"`
	AdminFullName string `form:"admin_full_name" binding:"required"`
	AdminEmail    string `form:"admin_email" binding:"required,email"`
	AdminPhone    string `form:"admin_phone" binding:"required"`
	Password      string `form:"password" binding:"required,min=6"`
	CityID        uint   `form:"city_id" binding:"required"`
	// Logo will be handled separately as multipart file
}
