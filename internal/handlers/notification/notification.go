package notification

import (
	"net/http"
	"subteacher/backend/internal/i18n"
	"subteacher/backend/internal/models"
	"subteacher/backend/internal/repository"
	"subteacher/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	notificationRepo repository.NotificationRepository
	db               *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		notificationRepo: repository.NewNotificationRepository(db),
		db:               db,
	}
}

// GetUserNotifications returns all notifications for the authenticated user
func (h *Handler) GetUserNotifications(c *gin.Context) {
	lang := c.GetString("language")

	// Get user ID from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	// Get pagination parameters
	page := utils.StringToInt(c.DefaultQuery("page", "1"))
	perPage := utils.StringToInt(c.DefaultQuery("per_page", "10"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	// Convert interface to float64 (JWT typically decodes numbers as float64)
	userIDFloat, ok := userIDInterface.(float64)
	if ok {
		userID := int(userIDFloat)
		notifications, err := h.notificationRepo.GetUserNotifications(userID, page, perPage)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
			return
		}

		utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), notifications)
		return
	}

	// Try as uint
	userIDUint, ok := userIDInterface.(uint)
	if ok {
		userID := int(userIDUint)
		notifications, err := h.notificationRepo.GetUserNotifications(userID, page, perPage)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
			return
		}

		utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), notifications)
		return
	}

	utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
}

// MarkAsRead marks a notification as read
func (h *Handler) MarkAsRead(c *gin.Context) {
	lang := c.GetString("language")

	// Get user ID from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	notificationID := utils.StringToUint(c.Param("id"))
	if notificationID == 0 {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "معرف الإشعار غير صالح")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID")
		}
		return
	}

	// Convert userID to int
	var userID int
	if userIDFloat, ok := userIDInterface.(float64); ok {
		userID = int(userIDFloat)
	} else if userIDUint, ok := userIDInterface.(uint); ok {
		userID = int(userIDUint)
	} else {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	// Verify the notification belongs to the user by getting the first page
	response, err := h.notificationRepo.GetUserNotifications(userID, 1, 100)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Check if the notification belongs to the user
	notifications, ok := response.Data.([]models.Notification)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	found := false
	for _, n := range notifications {
		if n.ID == notificationID {
			found = true
			break
		}
	}

	if !found {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusNotFound, "الإشعار غير موجود")
		} else {
			utils.ErrorResponse(c, http.StatusNotFound, "Notification not found")
		}
		return
	}

	if err := h.notificationRepo.MarkAsRead(notificationID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), nil)
}
