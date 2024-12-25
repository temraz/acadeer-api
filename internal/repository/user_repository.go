package repository

import (
	"subteacher/backend/internal/models"
	"time"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByPhoneNumber(phone string) (*models.User, error) {
	var user models.User
	err := r.db.Where("phone_number = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdateLastLogin(userID uint) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		UpdateColumn("last_login_at", time.Now()).Error
}

func (r *userRepository) BeginTx() (*gorm.DB, error) {
	return r.db.Begin(), nil
}

func (r *userRepository) UpdateProfile(userID uint, data *models.UpdateProfileRequest) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Update user basic info
	updates := map[string]interface{}{}
	if data.FullName != nil {
		updates["full_name"] = *data.FullName
	}
	if data.Email != nil {
		updates["email"] = *data.Email
	}
	if data.PhoneNumber != nil {
		updates["phone_number"] = *data.PhoneNumber
	}
	if data.CityID != nil {
		updates["city_id"] = *data.CityID
	}
	if data.CostPerDay != nil {
		updates["cost_per_day"] = *data.CostPerDay
	}
	if data.Gender != nil {
		updates["gender"] = *data.Gender
	}
	if data.SubjectID != nil {
		updates["subject_id"] = *data.SubjectID
	}
	updates["background_check_status"] = 1
	if data.Birthday != nil {
		// Parse the date string
		birthday, err := time.Parse("2006-01-02", *data.Birthday)
		if err != nil {
			tx.Rollback()
			return err
		}
		updates["birthday"] = birthday
	}

	if err := tx.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update teaching styles if provided
	if len(data.TeachingStyles) > 0 {
		// Remove existing teaching styles
		if err := tx.Where("user_id = ?", userID).Delete(&models.TeacherTeachingStyle{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Add new teaching styles
		for _, styleID := range data.TeachingStyles {
			if err := tx.Create(&models.TeacherTeachingStyle{
				UserID:          userID,
				TeachingStyleID: styleID,
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

func (r *userRepository) UpdateReceivedApplication(userID uint) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("received_application", true).Error
}

// GetTeachersWithReceivedApplications returns teachers with received applications and pending background checks
func (r *userRepository) GetTeachersWithReceivedApplications() ([]models.User, error) {
	var users []models.User
	err := r.db.Preload("School").
		Preload("TeachingStyles").
		Where("user_type = ? AND received_application = ? AND background_check_status = ?", 2, true, 1).
		Find(&users).Error
	return users, err
}
