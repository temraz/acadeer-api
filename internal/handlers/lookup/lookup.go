package lookup

import (
	"log"
	"net/http"
	"subteacher/backend/config"
	"subteacher/backend/internal/i18n"
	"subteacher/backend/internal/models"
	"subteacher/backend/internal/repository"
	"subteacher/backend/internal/utils"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	lookupRepo repository.LookupRepository
	config     *config.Config
}

func NewHandler(db *gorm.DB, config *config.Config) *Handler {
	return &Handler{
		lookupRepo: repository.NewLookupRepository(db),
		config:     config,
	}
}

func (h *Handler) GetLookups(c *gin.Context) {
	lang := c.GetString("language")

	// Get all lookups concurrently using goroutines
	var (
		teachingStyles []models.TeachingStyle
		subjects       []models.Subject
		certifications []models.Certification
		cities         []models.City
		errStyles      error
		errSubjects    error
		errCerts       error
		errCities      error
	)

	// Use wait group to wait for all goroutines
	var wg sync.WaitGroup
	wg.Add(4)

	// Get teaching styles
	go func() {
		defer wg.Done()
		teachingStyles, errStyles = h.lookupRepo.GetTeachingStyles()
	}()

	// Get subjects
	go func() {
		defer wg.Done()
		subjects, errSubjects = h.lookupRepo.GetSubjects()
	}()

	// Get certifications
	go func() {
		defer wg.Done()
		certifications, errCerts = h.lookupRepo.GetCertifications()
	}()

	// Get cities
	go func() {
		defer wg.Done()
		cities, errCities = h.lookupRepo.GetCities()
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for any errors
	if errStyles != nil || errSubjects != nil || errCerts != nil || errCities != nil {
		log.Printf("Error fetching lookups: %v, %v, %v, %v", errStyles, errSubjects, errCerts, errCities)
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	response := models.LookupResponse{
		TeachingStyles: teachingStyles,
		Subjects:       subjects,
		Certifications: certifications,
		Cities:         cities,
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), response)
}
