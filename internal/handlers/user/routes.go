package user

import (
	"subteacher/backend/config"
	"subteacher/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, h *Handler, eh *EducationHandler, cfg *config.Config) {
	// Protected routes (require authentication)
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	// User profile routes
	protected.GET("/profile", h.GetProfile)
	protected.PUT("/profile", h.UpdateProfile)
	protected.PUT("/profile/files", h.UpdateUserFiles)
	protected.GET("/me", h.GetUserDetails)
	protected.GET("/pending-teachers", h.GetPendingTeachers)
	protected.PUT("/teachers/:id/approve", h.ApproveTeacher)
	protected.GET("/users/:id", h.GetUserDetailsAdmin)
	protected.GET("/teachers", h.GetActiveTeachers)

	// Education routes
	education := protected.Group("/profile/education")
	{
		// Education History
		education.PUT("/history", eh.UpdateEducationHistory)
		education.PUT("/histories", eh.BatchUpdateEducationHistory)
		education.GET("/history", eh.GetEducationHistory)

		// Teaching License
		education.PUT("/license", eh.UpdateTeachingLicense)
		education.GET("/license", eh.GetTeachingLicense)

		// Certifications
		education.POST("/certification", eh.UpdateCertification)
		education.POST("/certifications", eh.BatchUpdateCertification)
		education.GET("/certifications", eh.GetCertifications)
	}
}
