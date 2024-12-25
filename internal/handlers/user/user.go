package user

import (
	"log"
	"math"
	"net/http"
	"subteacher/backend/config"
	"subteacher/backend/internal/i18n"
	"subteacher/backend/internal/models"
	"subteacher/backend/internal/repository"
	"subteacher/backend/internal/storage"
	"subteacher/backend/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	userRepo      repository.UserRepository
	schoolRepo    repository.SchoolRepository
	educationRepo repository.EducationRepository
	config        *config.Config
	db            *gorm.DB
	r2Client      *storage.R2Client
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
		userRepo:      repository.NewUserRepository(db),
		schoolRepo:    repository.NewSchoolRepository(db),
		educationRepo: repository.NewEducationRepository(db),
		config:        config,
		db:            db,
		r2Client:      r2Client,
	}
}

func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Validate email uniqueness if provided
	if req.Email != nil {
		existingUser, err := h.userRepo.FindByEmail(*req.Email)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			utils.ErrorResponse(c, http.StatusConflict, i18n.GetMessage(lang, "email_exists"))
			return
		}
	}

	// Validate phone number uniqueness if provided
	if req.PhoneNumber != nil {
		existingUser, err := h.userRepo.FindByPhoneNumber(*req.PhoneNumber)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			utils.ErrorResponse(c, http.StatusConflict, i18n.GetMessage(lang, "phone_exists"))
			return
		}
	}

	if err := h.userRepo.UpdateProfile(userID, &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "update_profile_error"))
		return
	}

	// Get updated user data
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "profile_updated"), user)
}

func (h *Handler) GetUserDetails(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")

	var user models.User
	// Preload both School and TeachingStyles
	result := h.db.Preload("School").
		Preload("TeachingStyles").
		Where("id = ?", userID).
		First(&user)

	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusNotFound, i18n.GetMessage(lang, "user_not_found"))
		return
	}

	// Get subject names if subject_id exists
	var subjectNameEn, subjectNameAr string
	if user.SubjectID != nil {
		var subject struct {
			NameEn string `json:"name_en"`
			NameAr string `json:"name_ar"`
		}
		if err := h.db.Table("subjects").
			Select("name_en, name_ar").
			Where("id = ?", *user.SubjectID).
			First(&subject).Error; err == nil {
			subjectNameEn = subject.NameEn
			subjectNameAr = subject.NameAr
		}
	}

	// Add full URLs for profile picture and CV
	userResponse := map[string]interface{}{
		"id":                      user.ID,
		"school_id":               user.SchoolID,
		"school":                  user.School,
		"email":                   user.Email,
		"user_type":               user.UserType,
		"is_active":               user.IsActive,
		"email_verified_at":       user.EmailVerifiedAt,
		"remember_token":          user.RememberToken,
		"full_name":               user.FullName,
		"phone_number":            user.PhoneNumber,
		"city_id":                 user.CityID,
		"subject_id":              user.SubjectID,
		"subject_name_en":         subjectNameEn,
		"subject_name_ar":         subjectNameAr,
		"price_per_day":           user.PricePerDay,
		"availability_status":     user.AvailabilityStatus,
		"status_id":               user.StatusID,
		"bio":                     user.Bio,
		"teaching_style":          user.TeachingStyle,
		"background_check_status": user.BackgroundCheckStatus,
		"background_check_date":   user.BackgroundCheckDate,
		"last_login_at":           user.LastLoginAt,
		"created_at":              user.CreatedAt,
		"updated_at":              user.UpdatedAt,
		"gender":                  user.Gender,
		"birthday":                user.Birthday,
		"cost_per_day":            user.CostPerDay,
		"teaching_styles":         user.TeachingStyles,
		"profile_picture":         user.ProfilePicture,
		"profile_picture_url":     h.r2Client.GetFileURL(user.ProfilePicture),
		"cv_file":                 user.CVFile,
		"cv_file_url":             h.r2Client.GetFileURL(user.CVFile),
		"received_application":    user.ReceivedApplication,
	}

	// Get education history
	history, err := h.educationRepo.GetEducationHistory(c.Request.Context(), userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Add file URL to history if it exists
	var historyResponse map[string]interface{}
	if history != nil {
		historyResponse = map[string]interface{}{
			"id":              history.ID,
			"user_id":         history.UserID,
			"degree":          history.Degree,
			"institution":     history.Institution,
			"major":           history.Major,
			"graduation_year": history.GraduationYear,
			"gpa":             history.GPA,
			"file":            history.File,
			"file_url":        h.r2Client.GetFileURL(history.File),
			"created_at":      history.CreatedAt,
			"updated_at":      history.UpdatedAt,
		}
	}

	// Get teaching license
	license, err := h.educationRepo.GetTeachingLicense(c.Request.Context(), userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Add file URL to license if it exists
	var licenseResponse map[string]interface{}
	if license != nil {
		licenseResponse = map[string]interface{}{
			"id":                license.ID,
			"user_id":           license.UserID,
			"license_number":    license.LicenseNumber,
			"issuing_authority": license.IssuingAuthority,
			"issue_date":        license.IssueDate,
			"expiry_date":       license.ExpiryDate,
			"file":              license.File,
			"file_url":          h.r2Client.GetFileURL(license.File),
			"status":            license.Status,
			"created_at":        license.CreatedAt,
			"updated_at":        license.UpdatedAt,
		}
	}

	// Get certifications
	certifications, err := h.educationRepo.GetCertifications(c.Request.Context(), userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Add file URLs to certifications
	certificationsResponse := make([]map[string]interface{}, len(certifications))
	for i, cert := range certifications {
		certificationsResponse[i] = map[string]interface{}{
			"id":                cert.ID,
			"user_id":           cert.UserID,
			"certification_id":  cert.CertificationID,
			"issue_date":        cert.IssueDate,
			"expiry_date":       cert.ExpiryDate,
			"file":              cert.File,
			"file_url":          h.r2Client.GetFileURL(cert.File),
			"issuing_authority": cert.IssuingAuthority,
			"created_at":        cert.CreatedAt,
			"updated_at":        cert.UpdatedAt,
			"certification":     cert.Certification,
		}
	}

	response := gin.H{
		"user": userResponse,
		"education": gin.H{
			"history":        historyResponse,
			"license":        licenseResponse,
			"certifications": certificationsResponse,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), response)
}

// UpdateUserFiles handles the update of user profile picture and CV
func (h *Handler) UpdateUserFiles(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")
	log.Printf("Updating files for user ID: %d", userID)

	var req models.UpdateUserFilesRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("Error binding form: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get existing user
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		log.Printf("Error finding user: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Handle profile picture upload
	if req.ProfilePicture != nil {
		// Delete old profile picture if exists
		if user.ProfilePicture != "" {
			if err := h.r2Client.DeleteFile(c.Request.Context(), user.ProfilePicture); err != nil {
				log.Printf("Warning: Failed to delete old profile picture: %v", err)
			}
		}

		// Upload new profile picture
		fileURL, err := h.r2Client.UploadFile(c.Request.Context(), req.ProfilePicture, "profile-pictures")
		if err != nil {
			log.Printf("Error uploading profile picture: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Error uploading profile picture")
			return
		}
		user.ProfilePicture = fileURL
	}

	// Handle CV upload
	if req.CV != nil {
		// Delete old CV if exists
		if user.CVFile != "" {
			if err := h.r2Client.DeleteFile(c.Request.Context(), user.CVFile); err != nil {
				log.Printf("Warning: Failed to delete old CV: %v", err)
			}
		}

		// Upload new CV
		fileURL, err := h.r2Client.UploadFile(c.Request.Context(), req.CV, "cvs")
		if err != nil {
			log.Printf("Error uploading CV: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Error uploading CV")
			return
		}
		user.CVFile = fileURL
	}

	// Update user in database
	if err := h.userRepo.Update(user); err != nil {
		log.Printf("Error updating user: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "update_profile_error"))
		return
	}

	// Check if all required files are present and update received_application status
	if user.ProfilePicture != "" && user.CVFile != "" {
		// Get education history, license, and certifications
		history, _ := h.educationRepo.GetEducationHistory(c.Request.Context(), userID)
		license, _ := h.educationRepo.GetTeachingLicense(c.Request.Context(), userID)
		certifications, _ := h.educationRepo.GetCertifications(c.Request.Context(), userID)

		// Check if all required documents are present
		if history != nil && license != nil && len(certifications) > 0 {
			if err := h.userRepo.UpdateReceivedApplication(userID); err != nil {
				log.Printf("Warning: Failed to update received_application status: %v", err)
			} else {
				user.ReceivedApplication = true
			}
		}
	}

	response := map[string]interface{}{
		"profile_picture":      user.ProfilePicture,
		"profile_picture_url":  h.r2Client.GetFileURL(user.ProfilePicture),
		"cv_file":              user.CVFile,
		"cv_file_url":          h.r2Client.GetFileURL(user.CVFile),
		"received_application": user.ReceivedApplication,
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "profile_updated"), response)
}

// GetPendingTeachers returns a list of teachers with received applications and pending background checks
func (h *Handler) GetPendingTeachers(c *gin.Context) {
	lang := c.GetString("language")

	// Check if user is superadmin
	userType := c.GetInt("user_type")
	if userType != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "unauthorized"),
			"data":    nil,
		})
		return
	}

	// Parse pagination parameters
	var pagination models.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// Calculate offset
	offset := (pagination.Page - 1) * pagination.PageSize

	// Get total count
	var totalRecords int64
	countResult := h.db.Table("users").
		Where("users.user_type = ? AND users.received_application = ? AND users.background_check_status = ?", 2, true, 1).
		Count(&totalRecords)

	if countResult.Error != nil {
		log.Printf("Error counting pending teachers: %v", countResult.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "server_error"),
			"data":    nil,
		})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pagination.PageSize)))

	// Define a temporary struct to hold the joined data
	type UserWithCity struct {
		models.User
		CityNameEn string `json:"city_name_en"`
		CityNameAr string `json:"city_name_ar"`
	}

	// Get teachers with received applications and pending background checks
	var users []UserWithCity
	result := h.db.Table("users").
		Select("users.*, cities.name_en as city_name_en, cities.name_ar as city_name_ar").
		Joins("LEFT JOIN cities ON users.city_id = cities.id").
		Where("users.user_type = ? AND users.received_application = ? AND users.background_check_status = ?", 2, true, 1).
		Offset(offset).
		Limit(pagination.PageSize).
		Find(&users)

	if result.Error != nil {
		log.Printf("Error fetching pending teachers: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "server_error"),
			"data":    nil,
		})
		return
	}

	// Create response with file URLs
	teachersResponse := make([]map[string]interface{}, len(users))
	for i, user := range users {
		teachersResponse[i] = map[string]interface{}{
			"id":                      user.ID,
			"full_name":               user.FullName,
			"email":                   user.Email,
			"phone_number":            user.PhoneNumber,
			"city_id":                 user.CityID,
			"city_name_en":            user.CityNameEn,
			"city_name_ar":            user.CityNameAr,
			"cost_per_day":            user.CostPerDay,
			"gender":                  user.Gender,
			"birthday":                user.Birthday,
			"profile_picture":         user.ProfilePicture,
			"profile_picture_url":     h.r2Client.GetFileURL(user.ProfilePicture),
			"cv_file":                 user.CVFile,
			"cv_file_url":             h.r2Client.GetFileURL(user.CVFile),
			"received_application":    user.ReceivedApplication,
			"background_check_status": user.BackgroundCheckStatus,
			"created_at":              user.CreatedAt,
			"updated_at":              user.UpdatedAt,
		}
	}

	// Create paginated response
	paginatedResponse := models.PaginationResponse{
		CurrentPage:  pagination.Page,
		PageSize:     pagination.PageSize,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
		Records:      teachersResponse,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.GetMessage(lang, "success"),
		"data":    paginatedResponse,
	})
}

// ApproveTeacher approves a teacher's application by updating their background check status
func (h *Handler) ApproveTeacher(c *gin.Context) {
	lang := c.GetString("language")

	// Check if user is superadmin
	userType := c.GetInt("user_type")
	if userType != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "unauthorized"),
			"data":    nil,
		})
		return
	}

	// Get teacher ID from URL parameter
	teacherID := c.Param("id")
	if teacherID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "invalid_teacher_id"),
			"data":    nil,
		})
		return
	}

	// Find the teacher
	var teacher models.User
	result := h.db.Where("id = ? AND user_type = ? AND received_application = ?", teacherID, 2, true).First(&teacher)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": i18n.GetMessage(lang, "teacher_not_found"),
				"data":    nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "server_error"),
			"data":    nil,
		})
		return
	}

	// Update background check status and date
	now := time.Now()
	updates := map[string]interface{}{
		"background_check_status": 2, // verified
		"background_check_date":   now,
	}

	if err := h.db.Model(&teacher).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "server_error"),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.GetMessage(lang, "teacher_approved"),
		"data": gin.H{
			"id":                      teacher.ID,
			"background_check_status": 2,
			"background_check_date":   now,
		},
	})
}

// GetUserDetailsAdmin returns complete user details, only accessible by superadmins and school admins
func (h *Handler) GetUserDetailsAdmin(c *gin.Context) {
	lang := c.GetString("language")

	// Check if user is superadmin or school admin
	userType := c.GetInt("user_type")
	// if userType != 1 && userType != 3 {
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"success": false,
	// 		"message": i18n.GetMessage(lang, "unauthorized"),
	// 		"data":    nil,
	// 	})
	// 	return
	// }

	// Get user ID from URL parameter
	userID := utils.StringToUint(c.Param("id"))
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "invalid_user_id"),
			"data":    nil,
		})
		return
	}

	var user models.User
	// Preload all related data
	result := h.db.Preload("School").
		Preload("TeachingStyles").
		Where("id = ?", userID).
		First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": i18n.GetMessage(lang, "user_not_found"),
				"data":    nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "server_error"),
			"data":    nil,
		})
		return
	}

	// If user is school admin, only allow viewing teacher details
	if userType == 3 && user.UserType != 2 {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "unauthorized"),
			"data":    nil,
		})
		return
	}

	// Get city names
	var city struct {
		NameEn string `json:"name_en"`
		NameAr string `json:"name_ar"`
	}
	if err := h.db.Table("cities").
		Select("name_en, name_ar").
		Where("id = ?", user.CityID).
		First(&city).Error; err != nil {
		log.Printf("Error fetching city: %v", err)
	}

	// Get subject names if subject_id exists
	var subjectNameEn, subjectNameAr string
	if user.SubjectID != nil {
		var subject struct {
			NameEn string `json:"name_en"`
			NameAr string `json:"name_ar"`
		}
		if err := h.db.Table("subjects").
			Select("name_en, name_ar").
			Where("id = ?", *user.SubjectID).
			First(&subject).Error; err == nil {
			subjectNameEn = subject.NameEn
			subjectNameAr = subject.NameAr
		}
	}

	// Add full URLs for profile picture and CV
	userResponse := map[string]interface{}{
		"id":                      user.ID,
		"school_id":               user.SchoolID,
		"school":                  user.School,
		"email":                   user.Email,
		"user_type":               user.UserType,
		"is_active":               user.IsActive,
		"email_verified_at":       user.EmailVerifiedAt,
		"remember_token":          user.RememberToken,
		"full_name":               user.FullName,
		"phone_number":            user.PhoneNumber,
		"city_id":                 user.CityID,
		"city_name_en":            city.NameEn,
		"city_name_ar":            city.NameAr,
		"subject_id":              user.SubjectID,
		"subject_name_en":         subjectNameEn,
		"subject_name_ar":         subjectNameAr,
		"price_per_day":           user.PricePerDay,
		"availability_status":     user.AvailabilityStatus,
		"status_id":               user.StatusID,
		"bio":                     user.Bio,
		"teaching_style":          user.TeachingStyle,
		"background_check_status": user.BackgroundCheckStatus,
		"background_check_date":   user.BackgroundCheckDate,
		"last_login_at":           user.LastLoginAt,
		"created_at":              user.CreatedAt,
		"updated_at":              user.UpdatedAt,
		"gender":                  user.Gender,
		"birthday":                user.Birthday,
		"cost_per_day":            user.CostPerDay,
		"teaching_styles":         user.TeachingStyles,
		"profile_picture":         user.ProfilePicture,
		"profile_picture_url":     h.r2Client.GetFileURL(user.ProfilePicture),
		"cv_file":                 user.CVFile,
		"cv_file_url":             h.r2Client.GetFileURL(user.CVFile),
		"received_application":    user.ReceivedApplication,
	}

	// Get education history
	history, err := h.educationRepo.GetEducationHistory(c.Request.Context(), user.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "server_error"),
			"data":    nil,
		})
		return
	}

	// Add file URL to history if it exists
	var historyResponse map[string]interface{}
	if history != nil {
		historyResponse = map[string]interface{}{
			"id":              history.ID,
			"user_id":         history.UserID,
			"degree":          history.Degree,
			"institution":     history.Institution,
			"major":           history.Major,
			"graduation_year": history.GraduationYear,
			"gpa":             history.GPA,
			"file":            history.File,
			"file_url":        h.r2Client.GetFileURL(history.File),
			"created_at":      history.CreatedAt,
			"updated_at":      history.UpdatedAt,
		}
	}

	// Get teaching license
	license, err := h.educationRepo.GetTeachingLicense(c.Request.Context(), user.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "server_error"),
			"data":    nil,
		})
		return
	}

	// Add file URL to license if it exists
	var licenseResponse map[string]interface{}
	if license != nil {
		licenseResponse = map[string]interface{}{
			"id":                license.ID,
			"user_id":           license.UserID,
			"license_number":    license.LicenseNumber,
			"issuing_authority": license.IssuingAuthority,
			"issue_date":        license.IssueDate,
			"expiry_date":       license.ExpiryDate,
			"file":              license.File,
			"file_url":          h.r2Client.GetFileURL(license.File),
			"status":            license.Status,
			"created_at":        license.CreatedAt,
			"updated_at":        license.UpdatedAt,
		}
	}

	// Get certifications
	certifications, err := h.educationRepo.GetCertifications(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": i18n.GetMessage(lang, "server_error"),
			"data":    nil,
		})
		return
	}

	// Add file URLs to certifications
	certificationsResponse := make([]map[string]interface{}, len(certifications))
	for i, cert := range certifications {
		certificationsResponse[i] = map[string]interface{}{
			"id":                cert.ID,
			"user_id":           cert.UserID,
			"certification_id":  cert.CertificationID,
			"issue_date":        cert.IssueDate,
			"expiry_date":       cert.ExpiryDate,
			"file":              cert.File,
			"file_url":          h.r2Client.GetFileURL(cert.File),
			"issuing_authority": cert.IssuingAuthority,
			"created_at":        cert.CreatedAt,
			"updated_at":        cert.UpdatedAt,
			"certification":     cert.Certification,
		}
	}

	response := gin.H{
		"user": userResponse,
		"education": gin.H{
			"history":        historyResponse,
			"license":        licenseResponse,
			"certifications": certificationsResponse,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.GetMessage(lang, "success"),
		"data":    response,
	})
}

// GetActiveTeachers returns a list of active and verified teachers with search parameters
func (h *Handler) GetActiveTeachers(c *gin.Context) {
	lang := c.GetString("language")

	var req models.TeacherSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Start building the query
	query := h.db.Model(&models.User{}).
		Joins("INNER JOIN cities ON users.city_id = cities.id").
		Joins("LEFT JOIN subjects ON users.subject_id = subjects.id").
		Preload("TeachingStyles").
		Where("users.user_type = ? AND users.status_id = ?", 2, 1)

	// Apply search filters if provided
	if req.TeacherName != nil && *req.TeacherName != "" {
		query = query.Where("users.full_name LIKE ?", "%"+*req.TeacherName+"%")
	}
	if req.CityID != nil {
		query = query.Where("users.city_id = ?", *req.CityID)
	}
	if req.SubjectID != nil {
		query = query.Where("users.subject_id = ?", *req.SubjectID)
	}
	if req.Gender != nil {
		query = query.Where("users.gender = ?", *req.Gender)
	}
	if req.PricePerDay != nil {
		query = query.Where("users.price_per_day <= ?", *req.PricePerDay)
	}

	// Get total count
	var totalRecords int64
	if err := query.Count(&totalRecords).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(req.PageSize)))

	// Get paginated results with city names
	var users []models.User
	if err := query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Get city names for the users
	type CityNames struct {
		ID     uint   `json:"id"`
		NameEn string `json:"name_en"`
		NameAr string `json:"name_ar"`
	}
	cityMap := make(map[uint]CityNames)
	var cityIDs []uint
	for _, user := range users {
		cityIDs = append(cityIDs, user.CityID)
	}

	var cities []CityNames
	if err := h.db.Table("cities").
		Select("id, name_en, name_ar").
		Where("id IN ?", cityIDs).
		Find(&cities).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	for _, city := range cities {
		cityMap[city.ID] = city
	}

	// Get subject names for the users
	type SubjectNames struct {
		ID     uint   `json:"id"`
		NameEn string `json:"name_en"`
		NameAr string `json:"name_ar"`
	}
	subjectMap := make(map[uint]SubjectNames)
	var subjectIDs []uint
	for _, user := range users {
		if user.SubjectID != nil {
			subjectIDs = append(subjectIDs, *user.SubjectID)
		}
	}

	if len(subjectIDs) > 0 {
		var subjects []SubjectNames
		if err := h.db.Table("subjects").
			Select("id, name_en, name_ar").
			Where("id IN ?", subjectIDs).
			Find(&subjects).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
			return
		}

		for _, subject := range subjects {
			subjectMap[subject.ID] = subject
		}
	}

	// Prepare response with calculated age and file URLs
	teachersResponse := make([]map[string]interface{}, len(users))
	for i, teacher := range users {
		var age *int
		if teacher.Birthday != nil {
			calculatedAge := time.Now().Year() - teacher.Birthday.Year()
			age = &calculatedAge
		}

		city := cityMap[teacher.CityID]
		var subjectNameEn, subjectNameAr string
		if teacher.SubjectID != nil {
			if subject, ok := subjectMap[*teacher.SubjectID]; ok {
				subjectNameEn = subject.NameEn
				subjectNameAr = subject.NameAr
			}
		}

		teachersResponse[i] = map[string]interface{}{
			"id":                      teacher.ID,
			"full_name":               teacher.FullName,
			"email":                   teacher.Email,
			"phone_number":            teacher.PhoneNumber,
			"city_id":                 teacher.CityID,
			"city_name_en":            city.NameEn,
			"city_name_ar":            city.NameAr,
			"subject_id":              teacher.SubjectID,
			"subject_name_en":         subjectNameEn,
			"subject_name_ar":         subjectNameAr,
			"price_per_day":           teacher.PricePerDay,
			"gender":                  teacher.Gender,
			"age":                     age,
			"teaching_styles":         teacher.TeachingStyles,
			"profile_picture":         teacher.ProfilePicture,
			"profile_picture_url":     h.r2Client.GetFileURL(teacher.ProfilePicture),
			"background_check_status": teacher.BackgroundCheckStatus,
			"created_at":              teacher.CreatedAt,
			"updated_at":              teacher.UpdatedAt,
		}
	}

	response := models.PaginationResponse{
		CurrentPage:  req.Page,
		PageSize:     req.PageSize,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
		Records:      teachersResponse,
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), response)
}
