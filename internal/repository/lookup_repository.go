package repository

import (
	"subteacher/backend/internal/models"

	"gorm.io/gorm"
)

type LookupRepository interface {
	GetTeachingStyles() ([]models.TeachingStyle, error)
	GetSubjects() ([]models.Subject, error)
	GetCertifications() ([]models.Certification, error)
	GetCities() ([]models.City, error)
}

type lookupRepository struct {
	db *gorm.DB
}

func NewLookupRepository(db *gorm.DB) LookupRepository {
	return &lookupRepository{db: db}
}

func (r *lookupRepository) GetTeachingStyles() ([]models.TeachingStyle, error) {
	var styles []models.TeachingStyle
	err := r.db.Find(&styles).Error
	return styles, err
}

func (r *lookupRepository) GetSubjects() ([]models.Subject, error) {
	var subjects []models.Subject
	err := r.db.Find(&subjects).Error
	return subjects, err
}

func (r *lookupRepository) GetCertifications() ([]models.Certification, error) {
	var certifications []models.Certification
	err := r.db.Find(&certifications).Error
	return certifications, err
}

func (r *lookupRepository) GetCities() ([]models.City, error) {
	var cities []models.City
	err := r.db.Find(&cities).Error
	return cities, err
}
