package auth

import (
	"log"
	"net/http"
	"subteacher/backend/config"
	"subteacher/backend/internal/i18n"
	"subteacher/backend/internal/models"
	"subteacher/backend/internal/repository"
	"subteacher/backend/internal/storage"
	"subteacher/backend/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Handler struct {
	userRepo   repository.UserRepository
	schoolRepo repository.SchoolRepository
	config     *config.Config
	r2Client   *storage.R2Client
}

func NewHandler(db *gorm.DB, config *config.Config) *Handler {
	r2Client, err := storage.NewR2Client(
		config.R2Config.AccountID,
		config.R2Config.AccessKeyID,
		config.R2Config.AccessKeySecret,
		config.R2Config.BucketName,
	)
	if err != nil {
		log.Fatalf("Failed to initialize R2 client: %v", err)
	}

	return &Handler{
		userRepo:   repository.NewUserRepository(db),
		schoolRepo: repository.NewSchoolRepository(db),
		config:     config,
		r2Client:   r2Client,
	}
}

func (h *Handler) Signup(c *gin.Context) {
	lang := c.GetString("language")
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if email already exists
	existingUser, err := h.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		utils.ErrorResponse(c, http.StatusConflict, i18n.GetMessage(lang, "email_exists"))
		return
	}

	// Check if phone number already exists
	existingUser, err = h.userRepo.FindByPhoneNumber(req.PhoneNumber)
	if err == nil && existingUser != nil {
		utils.ErrorResponse(c, http.StatusConflict, i18n.GetMessage(lang, "phone_exists"))
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
		IsActive:    true,
	}

	// Hash the password before saving
	if err := user.HashPassword(); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "create_user_error"))
		return
	}

	if err := h.userRepo.Create(&user); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "create_user_error"))
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(user.ID, user.UserType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "token_error"))
		return
	}

	// Parse access token duration
	accessTokenDuration, err := time.ParseDuration(h.config.JWTConfig.AccessTokenDuration)
	if err != nil {
		log.Printf("Error parsing access token duration: %v", err)
		accessTokenDuration = 24 * time.Hour // default to 24 hours
	}

	utils.SuccessResponse(c, http.StatusCreated, i18n.GetMessage(lang, "signup_success"), models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessTokenDuration.Seconds()),
	})
}

func (h *Handler) Login(c *gin.Context) {
	lang := c.GetString("language")
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		log.Printf("Login error for email %s: %v", req.Email, err)
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "invalid_credentials"))
		return
	}

	// Check if user is active
	if !user.IsActive {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "account_inactive"))
		return
	}

	if err := user.ComparePassword(req.Password); err != nil {
		// Log the actual error for debugging
		log.Printf("Password comparison error for email %s: %v", req.Email, err)
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "invalid_credentials"))
		return
	}

	// Update last login
	if err := h.userRepo.UpdateLastLogin(user.ID); err != nil {
		log.Printf("Error updating last login: %v", err)
		// Continue anyway as this is not critical
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(user.ID, user.UserType)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "token_error"))
		return
	}

	// Parse access token duration
	accessTokenDuration, err := time.ParseDuration(h.config.JWTConfig.AccessTokenDuration)
	if err != nil {
		log.Printf("Error parsing access token duration: %v", err)
		accessTokenDuration = 24 * time.Hour // default to 24 hours
	}

	response := gin.H{
		"user": user,
		"tokens": models.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(accessTokenDuration.Seconds()),
		},
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "login_success"), response)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	// Parse and validate refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWTConfig.Secret), nil
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
	userType := int(claims["user_type"].(float64))

	// Generate new tokens
	accessToken, newRefreshToken, err := h.generateTokens(userID, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Parse access token duration
	accessTokenDuration, err := time.ParseDuration(h.config.JWTConfig.AccessTokenDuration)
	if err != nil {
		log.Printf("Error parsing access token duration: %v", err)
		accessTokenDuration = 24 * time.Hour // default to 24 hours
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessTokenDuration.Seconds()),
	})
}

func (h *Handler) generateTokens(userID uint, userType int) (string, string, error) {
	// Parse access token duration
	accessTokenDuration, err := time.ParseDuration(h.config.JWTConfig.AccessTokenDuration)
	if err != nil {
		log.Printf("Error parsing access token duration: %v", err)
		accessTokenDuration = 24 * time.Hour // default to 24 hours
	}

	// Parse refresh token duration
	refreshTokenDuration, err := time.ParseDuration(h.config.JWTConfig.RefreshTokenDuration)
	if err != nil {
		log.Printf("Error parsing refresh token duration: %v", err)
		refreshTokenDuration = 720 * time.Hour // default to 30 days
	}

	// Create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID,
		"user_type": userType,
		"exp":       time.Now().Add(accessTokenDuration).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(h.config.JWTConfig.Secret))
	if err != nil {
		return "", "", err
	}

	// Create refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID,
		"user_type": userType,
		"exp":       time.Now().Add(refreshTokenDuration).Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(h.config.JWTConfig.Secret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (h *Handler) SchoolSignup(c *gin.Context) {
	lang := c.GetString("language")

	if err := c.Request.ParseMultipartForm(utils.MaxFileSize); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, i18n.GetMessage(lang, "file_parse_error"))
		return
	}

	var req models.SchoolSignupRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Validate required fields
	if req.NameEn == "" || req.NameAr == "" || req.AdminFullName == "" ||
		req.AdminEmail == "" || req.AdminPhone == "" || req.Password == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, i18n.GetMessage(lang, "all_fields_required"))
		return
	}

	// Check if email already exists
	existingUser, err := h.userRepo.FindByEmail(req.AdminEmail)
	if err == nil && existingUser != nil {
		utils.ErrorResponse(c, http.StatusConflict, i18n.GetMessage(lang, "email_exists"))
		return
	}

	// Check if phone number already exists
	existingUser, err = h.userRepo.FindByPhoneNumber(req.AdminPhone)
	if err == nil && existingUser != nil {
		utils.ErrorResponse(c, http.StatusConflict, i18n.GetMessage(lang, "phone_exists"))
		return
	}

	// Get the logo file after all validations pass
	file, err := c.FormFile("logo")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, i18n.GetMessage(lang, "logo_required"))
		return
	}

	// Upload logo to R2
	logoURL, err := h.r2Client.UploadFile(c.Request.Context(), file, "school-logos")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "logo_upload_error"))
		return
	}

	// Start a transaction
	tx, err := h.userRepo.BeginTx()
	if err != nil {
		// Delete uploaded logo if transaction fails
		h.r2Client.DeleteFile(c.Request.Context(), logoURL)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Create school
	school := models.School{
		NameEn:   req.NameEn,
		NameAr:   req.NameAr,
		Logo:     logoURL,
		CityID:   req.CityID,
		IsActive: true,
	}

	if err := tx.Create(&school).Error; err != nil {
		tx.Rollback()
		// Delete uploaded logo if school creation fails
		h.r2Client.DeleteFile(c.Request.Context(), logoURL)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "create_school_error"))
		return
	}

	// Create school admin user
	admin := models.User{
		Email:       req.AdminEmail,
		Password:    req.Password,
		UserType:    3, // school_admin
		FullName:    req.AdminFullName,
		PhoneNumber: req.AdminPhone,
		SchoolID:    &school.ID,
		CityID:      req.CityID,
		IsActive:    true,
	}

	// Hash the password before saving
	if err := admin.HashPassword(); err != nil {
		tx.Rollback()
		// Delete uploaded logo if password hashing fails
		h.r2Client.DeleteFile(c.Request.Context(), logoURL)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "create_user_error"))
		return
	}

	if err := tx.Create(&admin).Error; err != nil {
		tx.Rollback()
		// Delete uploaded logo if admin creation fails
		h.r2Client.DeleteFile(c.Request.Context(), logoURL)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "create_user_error"))
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		// Delete uploaded logo if commit fails
		h.r2Client.DeleteFile(c.Request.Context(), logoURL)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Generate tokens for the admin user
	accessToken, refreshToken, err := h.generateTokens(admin.ID, admin.UserType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "token_error"))
		return
	}

	// Add logo URL to the response
	schoolResponse := gin.H{
		"id":        school.ID,
		"name_en":   school.NameEn,
		"name_ar":   school.NameAr,
		"logo":      school.Logo,
		"logo_url":  h.r2Client.GetFileURL(school.Logo),
		"is_active": school.IsActive,
	}

	// Parse access token duration
	accessTokenDuration, err := time.ParseDuration(h.config.JWTConfig.AccessTokenDuration)
	if err != nil {
		log.Printf("Error parsing access token duration: %v", err)
		accessTokenDuration = 24 * time.Hour // default to 24 hours
	}

	utils.SuccessResponse(c, http.StatusCreated, i18n.GetMessage(lang, "school_signup_success"), gin.H{
		"school": schoolResponse,
		"tokens": models.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(accessTokenDuration.Seconds()),
		},
	})
}
