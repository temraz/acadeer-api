package handlers

import (
	"net/http"
	"subteacher/backend/config"
	"subteacher/backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db     *gorm.DB
	config *config.Config
}

func NewAuthHandler(db *gorm.DB, config *config.Config) *AuthHandler {
	return &AuthHandler{
		db:     db,
		config: config,
	}
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	var existingUser models.User
	if result := h.db.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	user := models.User{
		Email:       req.Email,
		Password:    req.Password,
		UserType:    req.UserType,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		CityID:      req.CityID,
		PricePerDay: req.PricePerDay,
		SchoolID:    req.SchoolID,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusCreated, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.config.AccessTokenDuration.Seconds()),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := user.ComparePassword(req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	h.db.Save(&user)

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.config.AccessTokenDuration.Seconds()),
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	// Parse and validate refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token claims"})
		return
	}

	userID := uint(claims["user_id"].(float64))

	// Generate new tokens
	accessToken, newRefreshToken, err := h.generateTokens(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.config.AccessTokenDuration.Seconds()),
	})
}

func (h *AuthHandler) generateTokens(userID uint) (string, string, error) {
	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(h.config.AccessTokenDuration).Unix(),
		"type":    "access",
	})

	accessTokenString, err := accessToken.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(h.config.RefreshTokenDuration).Unix(),
		"type":    "refresh",
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
