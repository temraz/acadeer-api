package user

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"subteacher/backend/internal/i18n"
	"subteacher/backend/internal/models"
	"subteacher/backend/internal/repository"
	"subteacher/backend/internal/storage"
	"subteacher/backend/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type EducationHandler struct {
	educationRepo repository.EducationRepository
	r2Client      *storage.R2Client
}

func NewEducationHandler(educationRepo repository.EducationRepository, r2Client *storage.R2Client) *EducationHandler {
	return &EducationHandler{
		educationRepo: educationRepo,
		r2Client:      r2Client,
	}
}

// Example of updating one file upload function
func (h *EducationHandler) handleFileUpload(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	return h.r2Client.UploadFile(ctx, file, folder)
}

// UpdateEducationHistory handles the education history update
func (h *EducationHandler) UpdateEducationHistory(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")
	log.Printf("Updating education history for user ID: %d", userID)

	var req models.UpdateEducationHistoryRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("Error binding form: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Create upload directory
	uploadDir := "uploads/education"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Error creating upload directory: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error creating upload directory")
		return
	}

	// Handle file upload
	fileURL, err := h.r2Client.UploadFile(c.Request.Context(), req.File, "education")
	if err != nil {
		log.Printf("Error uploading file to R2: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error uploading file")
		return
	}

	// Convert string values
	graduationYear, err := strconv.Atoi(req.GraduationYear)
	if err != nil {
		log.Printf("Invalid graduation year: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid graduation year")
		return
	}

	gpa, err := strconv.ParseFloat(req.GPA, 64)
	if err != nil {
		log.Printf("Invalid GPA: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid GPA")
		return
	}

	// Validate GPA range (0-100)
	if gpa < 0 || gpa > 100 {
		log.Printf("GPA out of range: %v", gpa)
		utils.ErrorResponse(c, http.StatusBadRequest, "GPA must be between 0 and 100")
		return
	}

	history := &models.EducationHistory{
		UserID:         userID,
		Degree:         req.Degree,
		Institution:    req.Institution,
		Major:          req.Major,
		GraduationYear: graduationYear,
		GPA:            gpa,
		File:           fileURL,
		UpdatedAt:      time.Now(),
	}

	// Try to get existing record
	existing, err := h.educationRepo.GetEducationHistory(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Creating new education history")
		err = h.educationRepo.CreateEducationHistory(c.Request.Context(), history)
	} else {
		log.Printf("Updating existing education history")
		// Delete old file from R2
		if err := h.r2Client.DeleteFile(c.Request.Context(), existing.File); err != nil {
			log.Printf("Warning: Failed to delete old file from R2: %v", err)
		}
		history.ID = existing.ID
		history.CreatedAt = existing.CreatedAt
		err = h.educationRepo.UpdateEducationHistory(c.Request.Context(), history)
	}

	if err != nil {
		log.Printf("Error in database operation: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "update_education_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "education_updated"), history)
}

// UpdateTeachingLicense handles the teaching license update
func (h *EducationHandler) UpdateTeachingLicense(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")
	log.Printf("Updating teaching license for user ID: %d", userID)

	var req models.UpdateTeachingLicenseRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("Error binding form: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Handle file upload
	fileURL, err := h.r2Client.UploadFile(c.Request.Context(), req.File, "licenses")
	if err != nil {
		log.Printf("Error uploading file to R2: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error uploading file")
		return
	}

	// Parse dates
	issueDate, err := time.Parse("2006-01-02", req.IssueDate)
	if err != nil {
		log.Printf("Invalid issue date: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid issue date format")
		return
	}

	expiryDate, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		log.Printf("Invalid expiry date: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid expiry date format")
		return
	}

	license := &models.TeachingLicense{
		UserID:           userID,
		LicenseNumber:    req.LicenseNumber,
		IssuingAuthority: req.IssuingAuthority,
		IssueDate:        issueDate,
		ExpiryDate:       expiryDate,
		File:             fileURL,
		Status:           req.Status,
		UpdatedAt:        time.Now(),
	}

	// Try to get existing record
	existing, err := h.educationRepo.GetTeachingLicense(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Creating new teaching license")
		err = h.educationRepo.CreateTeachingLicense(c.Request.Context(), license)
	} else {
		log.Printf("Updating existing teaching license")
		// Delete old file from R2
		if err := h.r2Client.DeleteFile(c.Request.Context(), existing.File); err != nil {
			log.Printf("Warning: Failed to delete old file from R2: %v", err)
		}
		license.ID = existing.ID
		license.CreatedAt = existing.CreatedAt
		err = h.educationRepo.UpdateTeachingLicense(c.Request.Context(), license)
	}

	if err != nil {
		log.Printf("Error in database operation: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "update_license_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "license_updated"), license)
}

// UpdateCertification handles the certification update
func (h *EducationHandler) UpdateCertification(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")
	log.Printf("Updating certification for user ID: %d", userID)

	var req models.UpdateCertificationRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("Error binding form: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Create upload directory
	uploadDir := "uploads/certifications"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Error creating upload directory: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error creating upload directory")
		return
	}

	// Handle file upload
	fileURL, err := h.r2Client.UploadFile(c.Request.Context(), req.File, "certifications")
	if err != nil {
		log.Printf("Error uploading file to R2: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error uploading file")
		return
	}

	// Parse certification ID
	certID, err := strconv.ParseUint(req.CertificationID, 10, 64)
	if err != nil {
		log.Printf("Invalid certification ID: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid certification ID")
		return
	}

	// Parse dates
	issueDate, err := time.Parse("2006-01-02", req.IssueDate)
	if err != nil {
		log.Printf("Invalid issue date: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid issue date format")
		return
	}

	expiryDate, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		log.Printf("Invalid expiry date: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid expiry date format")
		return
	}

	cert := &models.UserCertification{
		UserID:           userID,
		CertificationID:  uint(certID),
		IssueDate:        issueDate,
		ExpiryDate:       expiryDate,
		File:             fileURL,
		IssuingAuthority: req.IssuingAuthority,
		UpdatedAt:        time.Now(),
	}

	// Delete existing certifications for this user
	if err := h.educationRepo.DeleteAllCertifications(c.Request.Context(), userID); err != nil {
		log.Printf("Error deleting existing certifications: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "update_certification_error"))
		return
	}

	// Create new certification
	if err := h.educationRepo.CreateCertification(c.Request.Context(), cert); err != nil {
		log.Printf("Error creating certification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "update_certification_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "certification_updated"), cert)
}

// GetEducationHistory retrieves all education histories for a user
func (h *EducationHandler) GetEducationHistory(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")
	log.Printf("Fetching education histories for user ID: %d", userID)

	histories, err := h.educationRepo.GetEducationHistories(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Error fetching education histories: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Add full URLs to the response
	response := make([]map[string]interface{}, len(histories))
	for i, history := range histories {
		historyMap := map[string]interface{}{
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
		response[i] = historyMap
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), response)
}

// GetTeachingLicense retrieves the teaching license
func (h *EducationHandler) GetTeachingLicense(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")
	log.Printf("Fetching teaching license for user ID: %d", userID)

	license, err := h.educationRepo.GetTeachingLicense(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Error fetching teaching license: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Add full URL to the response
	response := map[string]interface{}{
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

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), response)
}

// GetCertifications retrieves all certifications
func (h *EducationHandler) GetCertifications(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")
	log.Printf("Fetching certifications for user ID: %d", userID)

	certs, err := h.educationRepo.GetCertifications(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Error fetching certifications: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Add full URLs to the response
	response := make([]map[string]interface{}, len(certs))
	for i, cert := range certs {
		certMap := map[string]interface{}{
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
		response[i] = certMap
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), response)
}

// BatchUpdateEducationHistory handles multiple education history updates
func (h *EducationHandler) BatchUpdateEducationHistory(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")
	log.Printf("Batch updating education history for user ID: %d", userID)

	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error getting multipart form: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Create upload directory
	uploadDir := "uploads/education"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Error creating upload directory: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error creating upload directory")
		return
	}

	var histories []*models.EducationHistory
	// Get the number of histories from the form data
	maxIndex := -1
	for key := range form.Value {
		if len(key) > 9 && key[:9] == "histories" { // Check if key starts with "histories"
			// Extract index from key format "histories[N][field]"
			if idx := strings.Index(key[9:], "]"); idx > 0 {
				if index, err := strconv.Atoi(key[10 : 9+idx]); err == nil {
					if index > maxIndex {
						maxIndex = index
					}
				}
			}
		}
	}

	// Process each history entry
	for i := 0; i <= maxIndex; i++ {
		prefix := fmt.Sprintf("histories[%d]", i)

		// Get form values
		degree := c.PostForm(prefix + "[degree]")
		institution := c.PostForm(prefix + "[institution]")
		major := c.PostForm(prefix + "[major]")
		graduationYearStr := c.PostForm(prefix + "[graduation_year]")
		gpaStr := c.PostForm(prefix + "[gpa]")

		// Validate required fields
		if degree == "" || institution == "" || major == "" || graduationYearStr == "" || gpaStr == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Missing required fields")
			return
		}

		// Get file
		file, err := c.FormFile(prefix + "[file]")
		if err != nil {
			log.Printf("Error getting file: %v", err)
			utils.ErrorResponse(c, http.StatusBadRequest, "File is required")
			return
		}

		// Handle file upload
		fileURL, err := h.r2Client.UploadFile(c.Request.Context(), file, "education")
		if err != nil {
			log.Printf("Error uploading file to R2: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Error uploading file")
			return
		}

		// Convert string values
		graduationYear, err := strconv.Atoi(graduationYearStr)
		if err != nil {
			log.Printf("Invalid graduation year: %v", err)
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid graduation year")
			return
		}

		gpa, err := strconv.ParseFloat(gpaStr, 64)
		if err != nil {
			log.Printf("Invalid GPA: %v", err)
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid GPA")
			return
		}

		// Validate GPA range (0-100)
		if gpa < 0 || gpa > 100 {
			log.Printf("GPA out of range: %v", gpa)
			utils.ErrorResponse(c, http.StatusBadRequest, "GPA must be between 0 and 100")
			return
		}

		history := &models.EducationHistory{
			UserID:         userID,
			Degree:         degree,
			Institution:    institution,
			Major:          major,
			GraduationYear: graduationYear,
			GPA:            gpa,
			File:           fileURL,
			UpdatedAt:      time.Now(),
		}
		histories = append(histories, history)
	}

	if len(histories) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No education histories provided")
		return
	}

	// Delete existing records and their files
	if histories, err := h.educationRepo.GetEducationHistories(c.Request.Context(), userID); err == nil {
		for _, history := range histories {
			if err := h.r2Client.DeleteFile(c.Request.Context(), history.File); err != nil {
				log.Printf("Warning: Failed to delete old file from R2: %v", err)
			}
		}
	}
	if err := h.educationRepo.DeleteEducationHistory(c.Request.Context(), userID); err != nil {
		log.Printf("Error deleting existing education history: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "update_education_error"))
		return
	}

	// Create new records
	if err := h.educationRepo.CreateEducationHistories(c.Request.Context(), histories); err != nil {
		log.Printf("Error creating education histories: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "update_education_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "education_updated"), histories)
}

// BatchUpdateCertification handles multiple certification updates
func (h *EducationHandler) BatchUpdateCertification(c *gin.Context) {
	lang := c.GetString("language")
	userID := c.GetUint("user_id")
	log.Printf("Batch updating certifications for user ID: %d", userID)

	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error getting multipart form: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var certifications []*models.UserCertification
	// Get the number of certifications from the form data
	maxIndex := -1
	for key := range form.Value {
		if len(key) > 14 && key[:14] == "certifications" { // Check if key starts with "certifications"
			// Extract index from key format "certifications[N][field]"
			if idx := strings.Index(key[14:], "]"); idx > 0 {
				if index, err := strconv.Atoi(key[15 : 14+idx]); err == nil {
					if index > maxIndex {
						maxIndex = index
					}
				}
			}
		}
	}

	// Process each certification entry
	for i := 0; i <= maxIndex; i++ {
		prefix := fmt.Sprintf("certifications[%d]", i)

		// Get form values
		certificationID := c.PostForm(prefix + "[certification_id]")
		issueDate := c.PostForm(prefix + "[issue_date]")
		expiryDate := c.PostForm(prefix + "[expiry_date]")
		issuingAuthority := c.PostForm(prefix + "[issuing_authority]")

		// Validate required fields
		if certificationID == "" || issueDate == "" || expiryDate == "" || issuingAuthority == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Missing required fields")
			return
		}

		// Get file
		file, err := c.FormFile(prefix + "[file]")
		if err != nil {
			log.Printf("Error getting file: %v", err)
			utils.ErrorResponse(c, http.StatusBadRequest, "File is required")
			return
		}

		// Handle file upload
		fileURL, err := h.r2Client.UploadFile(c.Request.Context(), file, "certifications")
		if err != nil {
			log.Printf("Error uploading file to R2: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Error uploading file")
			return
		}

		// Parse certification ID
		certID, err := strconv.ParseUint(certificationID, 10, 64)
		if err != nil {
			log.Printf("Invalid certification ID: %v", err)
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid certification ID")
			return
		}

		// Parse dates
		issueDateParsed, err := time.Parse("2006-01-02", issueDate)
		if err != nil {
			log.Printf("Invalid issue date: %v", err)
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid issue date format")
			return
		}

		expiryDateParsed, err := time.Parse("2006-01-02", expiryDate)
		if err != nil {
			log.Printf("Invalid expiry date: %v", err)
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid expiry date format")
			return
		}

		cert := &models.UserCertification{
			UserID:           userID,
			CertificationID:  uint(certID),
			IssueDate:        issueDateParsed,
			ExpiryDate:       expiryDateParsed,
			File:             fileURL,
			IssuingAuthority: issuingAuthority,
			UpdatedAt:        time.Now(),
		}
		certifications = append(certifications, cert)
	}

	if len(certifications) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No certifications provided")
		return
	}

	// Delete existing certifications and their files
	if certs, err := h.educationRepo.GetCertifications(c.Request.Context(), userID); err == nil {
		for _, cert := range certs {
			if err := h.r2Client.DeleteFile(c.Request.Context(), cert.File); err != nil {
				log.Printf("Warning: Failed to delete old file from R2: %v", err)
			}
		}
	}
	if err := h.educationRepo.DeleteAllCertifications(c.Request.Context(), userID); err != nil {
		log.Printf("Error deleting existing certifications: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "update_certification_error"))
		return
	}

	// Create new certifications
	if err := h.educationRepo.CreateCertifications(c.Request.Context(), certifications); err != nil {
		log.Printf("Error creating certifications: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "update_certification_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "certification_updated"), certifications)
}
