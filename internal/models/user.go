package models

import (
	"mime/multipart"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID                    uint            `gorm:"primarykey" json:"id"`
	SchoolID              *uint           `json:"school_id"`
	School                *School         `json:"school,omitempty"`
	Email                 string          `gorm:"unique;not null" json:"email"`
	Password              string          `gorm:"not null" json:"-"`
	UserType              int             `gorm:"not null;comment:'1=super_admin, 2=teacher, 3=school_admin'" json:"user_type"`
	IsActive              bool            `gorm:"default:true" json:"is_active"`
	EmailVerifiedAt       *time.Time      `json:"email_verified_at"`
	RememberToken         *string         `json:"remember_token"`
	FullName              string          `gorm:"not null" json:"full_name"`
	PhoneNumber           string          `gorm:"type:varchar(20);not null" json:"phone_number"`
	CityID                uint            `gorm:"not null" json:"city_id"`
	SubjectID             *uint           `gorm:"default:null" json:"subject_id"`
	PricePerDay           float64         `gorm:"type:decimal(10,2)" json:"price_per_day"`
	AvailabilityStatus    int             `gorm:"default:1;comment:'1=immediate, 2=next_week, 3=unavailable'" json:"availability_status"`
	StatusID              int             `gorm:"default:1;comment:'1=active, 2=suspended'" json:"status_id"`
	Bio                   *string         `json:"bio"`
	TeachingStyle         *string         `json:"teaching_style"`
	BackgroundCheckStatus int             `gorm:"default:1;comment:'1=pending, 2=verified, 3=failed'" json:"background_check_status"`
	BackgroundCheckDate   *time.Time      `json:"background_check_date"`
	LastLoginAt           *time.Time      `json:"last_login_at"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	Gender                *int            `gorm:"type:tinyint;comment:'1=male, 2=female'" json:"gender"`
	Birthday              *time.Time      `json:"birthday"`
	CostPerDay            *float64        `gorm:"type:decimal(10,2)" json:"cost_per_day"`
	TeachingStyles        []TeachingStyle `gorm:"many2many:teacher_teaching_styles;" json:"teaching_styles,omitempty"`
	ProfilePicture        string          `gorm:"column:profile_picture" json:"profile_picture"`
	CVFile                string          `gorm:"column:cv_file" json:"cv_file"`
	ReceivedApplication   bool            `gorm:"column:received_application;default:0" json:"received_application"`
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	if tx.Statement.Changed("Password") {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

type SignupRequest struct {
	Email       string  `json:"email" binding:"required,email"`
	Password    string  `json:"password" binding:"required,min=6"`
	UserType    int     `json:"user_type" binding:"required"`
	FullName    string  `json:"full_name" binding:"required"`
	PhoneNumber string  `json:"phone_number" binding:"required"`
	CityID      uint    `json:"city_id" binding:"required"`
	PricePerDay float64 `json:"price_per_day"`
	SchoolID    *uint   `json:"school_id"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // in seconds
}

// HashPassword hashes the password using bcrypt
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

type UpdateProfileRequest struct {
	FullName       *string  `json:"full_name"`
	Email          *string  `json:"email" binding:"omitempty,email"`
	PhoneNumber    *string  `json:"phone_number"`
	CityID         *uint    `json:"city_id"`
	CostPerDay     *float64 `json:"cost_per_day"`
	Gender         *int     `json:"gender" binding:"omitempty,oneof=1 2"`
	Birthday       *string  `json:"birthday" binding:"omitempty,datetime=2006-01-02"`
	TeachingStyles []uint   `json:"teaching_style_ids"`
	SubjectID      *uint    `json:"subject_id"`
}

// Request structs for file uploads
type UpdateUserFilesRequest struct {
	ProfilePicture *multipart.FileHeader `form:"profile_picture"`
	CV             *multipart.FileHeader `form:"cv"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=10" binding:"min=1,max=100"`
}

// PaginationResponse represents paginated response
type PaginationResponse struct {
	CurrentPage  int         `json:"current_page"`
	PageSize     int         `json:"page_size"`
	TotalPages   int         `json:"total_pages"`
	TotalRecords int64       `json:"total_records"`
	Records      interface{} `json:"records"`
}

// TeacherSearchRequest represents search parameters for teachers
type TeacherSearchRequest struct {
	TeacherName *string  `form:"teacher_name"`
	CityID      *uint    `form:"city_id"`
	SubjectID   *uint    `form:"subject_id"`
	Gender      *int     `form:"gender"`
	PricePerDay *float64 `form:"price_per_day"`
	Page        int      `form:"page,default=1" binding:"min=1"`
	PageSize    int      `form:"page_size,default=100" binding:"min=1,max=100"`
}
