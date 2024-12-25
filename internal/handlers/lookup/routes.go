package lookup

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	router.GET("", h.GetLookups)
}
