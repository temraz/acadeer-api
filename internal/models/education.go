package models

import (
	"fmt"
	"mime/multipart"
	"time"
)

// EducationHistory represents the education history record
type EducationHistory struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id"`
	Degree         string    `json:"degree"`
	Institution    string    `json:"institution"`
	Major          string    `json:"major"`
	GraduationYear int       `json:"graduation_year"`
	GPA            float64   `json:"gpa"`
	File           string    `json:"file"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (EducationHistory) TableName() string {
	return "education_history"
}

// TeachingLicense represents the teaching license record
type TeachingLicense struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id"`
	LicenseNumber    string    `json:"license_number"`
	IssuingAuthority string    `json:"issuing_authority"`
	IssueDate        time.Time `json:"issue_date"`
	ExpiryDate       time.Time `json:"expiry_date"`
	File             string    `json:"file"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (TeachingLicense) TableName() string {
	return "teaching_licenses"
}

// UserCertification represents the certification record
type UserCertification struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id"`
	CertificationID  uint      `json:"certification_id"`
	IssueDate        time.Time `json:"issue_date"`
	ExpiryDate       time.Time `json:"expiry_date"`
	File             string    `json:"file"`
	IssuingAuthority string    `json:"issuing_authority"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relationship
	Certification *Certification `json:"certification,omitempty"`
}

func (UserCertification) TableName() string {
	return "user_certifications"
}

// Request structs for form handling
type UpdateEducationHistoryRequest struct {
	Degree         string                `form:"degree" binding:"required"`
	Institution    string                `form:"institution" binding:"required"`
	Major          string                `form:"major" binding:"required"`
	GraduationYear string                `form:"graduation_year" binding:"required"`
	GPA            string                `form:"gpa" binding:"required"`
	File           *multipart.FileHeader `form:"file" binding:"required"`
}

type BatchUpdateEducationHistoryRequest struct {
	Histories []UpdateEducationHistoryRequest `form:"histories" binding:"required,dive"`
}

type UpdateTeachingLicenseRequest struct {
	LicenseNumber    string                `form:"license_number" binding:"required"`
	IssuingAuthority string                `form:"issuing_authority" binding:"required"`
	IssueDate        string                `form:"issue_date" binding:"required"`
	ExpiryDate       string                `form:"expiry_date" binding:"required"`
	Status           string                `form:"status" binding:"required,oneof=active expired suspended"`
	File             *multipart.FileHeader `form:"file" binding:"required"`
}

type UpdateCertificationRequest struct {
	CertificationID  string                `form:"certification_id" binding:"required"`
	IssueDate        string                `form:"issue_date" binding:"required"`
	ExpiryDate       string                `form:"expiry_date" binding:"required"`
	IssuingAuthority string                `form:"issuing_authority" binding:"required"`
	File             *multipart.FileHeader `form:"file" binding:"required"`
}

type BatchUpdateCertificationRequest struct {
	Certifications []UpdateCertificationRequest `form:"certifications" binding:"required,dive"`
}

// String methods for logging
func (r UpdateEducationHistoryRequest) String() string {
	return fmt.Sprintf("UpdateEducationHistoryRequest{Degree: %s, Institution: %s, Major: %s, GraduationYear: %s, GPA: %s, HasFile: %v}",
		r.Degree, r.Institution, r.Major, r.GraduationYear, r.GPA, r.File != nil)
}

func (r UpdateTeachingLicenseRequest) String() string {
	return fmt.Sprintf("UpdateTeachingLicenseRequest{LicenseNumber: %s, Authority: %s, IssueDate: %s, ExpiryDate: %s, Status: %s, HasFile: %v}",
		r.LicenseNumber, r.IssuingAuthority, r.IssueDate, r.ExpiryDate, r.Status, r.File != nil)
}

func (r UpdateCertificationRequest) String() string {
	return fmt.Sprintf("UpdateCertificationRequest{CertID: %s, IssueDate: %s, ExpiryDate: %s, Authority: %s, HasFile: %v}",
		r.CertificationID, r.IssueDate, r.ExpiryDate, r.IssuingAuthority, r.File != nil)
}
