package repository

import (
	"subteacher/backend/internal/models"
	"time"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *models.Notification) error
	FindByUserID(userID uint) ([]models.Notification, error)
	MarkAsRead(notificationID uint) error
	GetUserNotifications(userID int, page, perPage int) (*PaginationResponse, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) FindByUserID(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) MarkAsRead(notificationID uint) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND read_at IS NULL", notificationID).
		Update("read_at", time.Now()).Error
}

type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

func (r *notificationRepository) GetUserNotifications(userID int, page, perPage int) (*PaginationResponse, error) {
	var notifications []models.Notification
	var total int64

	// Calculate offset
	offset := (page - 1) * perPage

	// Get total count
	if err := r.db.Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	// Get paginated notifications
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&notifications).Error

	if err != nil {
		return nil, err
	}

	// For notifications with school_id, get school names
	for i := range notifications {
		if notifications[i].SchoolID != nil {
			var result struct {
				NameEn string
				NameAr string
			}
			err := r.db.Table("schools").
				Select("name_en, name_ar").
				Where("id = ?", *notifications[i].SchoolID).
				Scan(&result).Error

			if err == nil {
				notifications[i].SchoolNameEn = result.NameEn
				notifications[i].SchoolNameAr = result.NameAr
			}
		}
	}

	return &PaginationResponse{
		Data:       notifications,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}
