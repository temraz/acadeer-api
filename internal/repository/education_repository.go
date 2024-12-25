package repository

import (
	"context"
	"subteacher/backend/internal/models"

	"gorm.io/gorm"
)

type EducationRepository interface {
	// Education History
	CreateEducationHistory(ctx context.Context, history *models.EducationHistory) error
	CreateEducationHistories(ctx context.Context, histories []*models.EducationHistory) error
	GetEducationHistory(ctx context.Context, userID uint) (*models.EducationHistory, error)
	GetEducationHistories(ctx context.Context, userID uint) ([]models.EducationHistory, error)
	UpdateEducationHistory(ctx context.Context, history *models.EducationHistory) error
	DeleteEducationHistory(ctx context.Context, userID uint) error

	// Teaching License
	CreateTeachingLicense(ctx context.Context, license *models.TeachingLicense) error
	GetTeachingLicense(ctx context.Context, userID uint) (*models.TeachingLicense, error)
	UpdateTeachingLicense(ctx context.Context, license *models.TeachingLicense) error
	DeleteTeachingLicense(ctx context.Context, userID uint) error

	// Certifications
	CreateCertification(ctx context.Context, cert *models.UserCertification) error
	CreateCertifications(ctx context.Context, certs []*models.UserCertification) error
	GetCertifications(ctx context.Context, userID uint) ([]models.UserCertification, error)
	UpdateCertification(ctx context.Context, cert *models.UserCertification) error
	DeleteCertification(ctx context.Context, userID uint, certID uint) error

	// DeleteAllCertifications deletes all certifications for a user
	DeleteAllCertifications(ctx context.Context, userID uint) error
}

type educationRepository struct {
	db *gorm.DB
}

func NewEducationRepository(db *gorm.DB) EducationRepository {
	return &educationRepository{db: db}
}

// Education History implementations
func (r *educationRepository) CreateEducationHistory(ctx context.Context, history *models.EducationHistory) error {
	return r.db.Create(history).Error
}

func (r *educationRepository) GetEducationHistory(ctx context.Context, userID uint) (*models.EducationHistory, error) {
	var history models.EducationHistory
	err := r.db.Where("user_id = ?", userID).First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *educationRepository) UpdateEducationHistory(ctx context.Context, history *models.EducationHistory) error {
	return r.db.Model(&models.EducationHistory{}).
		Where("user_id = ?", history.UserID).
		Updates(map[string]interface{}{
			"degree":          history.Degree,
			"institution":     history.Institution,
			"major":           history.Major,
			"graduation_year": history.GraduationYear,
			"gpa":             history.GPA,
			"file":            history.File,
			"updated_at":      history.UpdatedAt,
		}).Error
}

func (r *educationRepository) DeleteEducationHistory(ctx context.Context, userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.EducationHistory{}).Error
}

func (r *educationRepository) CreateEducationHistories(ctx context.Context, histories []*models.EducationHistory) error {
	return r.db.Create(histories).Error
}

func (r *educationRepository) GetEducationHistories(ctx context.Context, userID uint) ([]models.EducationHistory, error) {
	var histories []models.EducationHistory
	err := r.db.Where("user_id = ?", userID).Find(&histories).Error
	return histories, err
}

// Teaching License implementations
func (r *educationRepository) CreateTeachingLicense(ctx context.Context, license *models.TeachingLicense) error {
	return r.db.Create(license).Error
}

func (r *educationRepository) GetTeachingLicense(ctx context.Context, userID uint) (*models.TeachingLicense, error) {
	var license models.TeachingLicense
	err := r.db.Where("user_id = ?", userID).First(&license).Error
	if err != nil {
		return nil, err
	}
	return &license, nil
}

func (r *educationRepository) UpdateTeachingLicense(ctx context.Context, license *models.TeachingLicense) error {
	return r.db.Model(&models.TeachingLicense{}).
		Where("user_id = ?", license.UserID).
		Updates(map[string]interface{}{
			"license_number":    license.LicenseNumber,
			"issuing_authority": license.IssuingAuthority,
			"issue_date":        license.IssueDate,
			"expiry_date":       license.ExpiryDate,
			"file":              license.File,
			"status":            license.Status,
			"updated_at":        license.UpdatedAt,
		}).Error
}

func (r *educationRepository) DeleteTeachingLicense(ctx context.Context, userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.TeachingLicense{}).Error
}

// Certification implementations
func (r *educationRepository) CreateCertification(ctx context.Context, cert *models.UserCertification) error {
	return r.db.Create(cert).Error
}

func (r *educationRepository) GetCertifications(ctx context.Context, userID uint) ([]models.UserCertification, error) {
	var certs []models.UserCertification
	err := r.db.Where("user_id = ?", userID).
		Preload("Certification").
		Find(&certs).Error
	if err != nil {
		return nil, err
	}
	return certs, nil
}

func (r *educationRepository) UpdateCertification(ctx context.Context, cert *models.UserCertification) error {
	return r.db.Model(&models.UserCertification{}).
		Where("user_id = ? AND id = ?", cert.UserID, cert.ID).
		Updates(map[string]interface{}{
			"certification_id":  cert.CertificationID,
			"issue_date":        cert.IssueDate,
			"expiry_date":       cert.ExpiryDate,
			"file":              cert.File,
			"issuing_authority": cert.IssuingAuthority,
			"updated_at":        cert.UpdatedAt,
		}).Error
}

func (r *educationRepository) DeleteCertification(ctx context.Context, userID uint, certID uint) error {
	return r.db.Where("user_id = ? AND id = ?", userID, certID).
		Delete(&models.UserCertification{}).Error
}

func (r *educationRepository) CreateCertifications(ctx context.Context, certs []*models.UserCertification) error {
	return r.db.Create(certs).Error
}

// DeleteAllCertifications deletes all certifications for a user
func (r *educationRepository) DeleteAllCertifications(ctx context.Context, userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.UserCertification{}).Error
}
