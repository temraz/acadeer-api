package repository

import (
	"subteacher/backend/internal/models"

	"gorm.io/gorm"
)

type schoolRepository struct {
	db *gorm.DB
}

func NewSchoolRepository(db *gorm.DB) SchoolRepository {
	return &schoolRepository{db: db}
}

func (r *schoolRepository) FindByID(id uint) (*models.School, error) {
	var school models.School
	err := r.db.First(&school, id).Error
	if err != nil {
		return nil, err
	}
	return &school, nil
}

func (r *schoolRepository) Create(school *models.School) error {
	return r.db.Create(school).Error
}

func (r *schoolRepository) BeginTx() (*gorm.DB, error) {
	return r.db.Begin(), nil
}
