package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"subteacher/backend/config"
	"subteacher/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func AuthMiddleware(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header format",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.JWTConfig.Secret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Set user ID in context
			if userIDFloat, ok := claims["user_id"].(float64); ok {
				userID := uint(userIDFloat)
				c.Set("user_id", userID)

				// Get user from database to get user type
				db := c.MustGet("db").(*gorm.DB)
				var user models.User
				if err := db.Select("user_type").First(&user, userID).Error; err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{
						"success": false,
						"message": "User not found",
						"data":    nil,
					})
					c.Abort()
					return
				}

				// Set user type in context
				c.Set("user_type", user.UserType)

				c.Next()
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Invalid token claims",
					"data":    nil,
				})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token",
				"data":    nil,
			})
			c.Abort()
			return
		}
	}
}
