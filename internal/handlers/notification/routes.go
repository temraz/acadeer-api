package notification

import (
	"subteacher/backend/config"
	"subteacher/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, h *Handler, cfg *config.Config) {
	// Protected routes (require authentication)
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	// Notification routes
	protected.GET("", h.GetUserNotifications) // Get user's notifications
	protected.PUT("/:id/read", h.MarkAsRead)  // Mark notification as read
}
