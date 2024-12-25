package repository

import (
	"subteacher/backend/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByPhoneNumber(phone string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	UpdateLastLogin(userID uint) error
	BeginTx() (*gorm.DB, error)
	UpdateProfile(userID uint, data *models.UpdateProfileRequest) error
	UpdateReceivedApplication(userID uint) error
}

type SchoolRepository interface {
	FindByID(id uint) (*models.School, error)
	Create(school *models.School) error
	BeginTx() (*gorm.DB, error)
}
