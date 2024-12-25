package auth

import "github.com/gin-gonic/gin"

// SetupRoutes sets up all the auth routes
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	router.POST("/signup", h.Signup)
	router.POST("/school/signup", h.SchoolSignup)
	router.POST("/login", h.Login)
	router.POST("/refresh", h.RefreshToken)
}
