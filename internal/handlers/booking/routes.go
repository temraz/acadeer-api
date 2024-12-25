package booking

import (
	"subteacher/backend/config"
	"subteacher/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, h *Handler, cfg *config.Config) {
	// Protected routes (require authentication)
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	// Booking routes
	protected.POST("", h.CreateBooking)                                    // Create a new booking (school admin only)
	protected.GET("/school-admin", h.GetSchoolAdminBookings)               // Get bookings for school admin
	protected.GET("/teacher", h.GetTeacherBookings)                        // Get bookings for teacher
	protected.GET("/teacher-future-bookings", h.GetTeacherFutureBookings)  // Get future accepted bookings for a teacher
	protected.GET("/teacher-schedule", h.GetTeacherScheduleStats)          // Get teacher schedule statistics
	protected.GET("/teacher-bookings-by-date", h.GetTeacherBookingsByDate) // Get teacher bookings for a specific date
	protected.GET("/:id", h.GetBookingDetails)                             // Get detailed booking information by ID
	protected.PUT("/:id/status", h.UpdateBookingStatus)                    // Update booking status (accept/reject)
	protected.GET("/school-schedule", h.GetSchoolScheduleStats)            // Get school schedule statistics
	protected.GET("/school-bookings-by-date", h.GetSchoolBookingsByDate)   // Get school bookings for a specific date
}
