package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Accept-Language header
		lang := c.GetHeader("Accept-Language")

		// Default to English
		if lang == "" {
			lang = "en"
		}

		// Extract primary language
		lang = strings.Split(lang, ",")[0]
		lang = strings.Split(lang, "-")[0]

		// Only support en and ar
		if lang != "ar" {
			lang = "en"
		}

		// Set language in context
		c.Set("language", lang)
		c.Next()
	}
}
