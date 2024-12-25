package models

import "time"

// LookupResponse represents the response structure for lookup data
type LookupResponse struct {
	TeachingStyles []TeachingStyle `json:"teaching_styles"`
	Subjects       []Subject       `json:"subjects"`
	Certifications []Certification `json:"certifications"`
	Cities         []City          `json:"cities"`
}

// Subject represents a teaching subject
type Subject struct {
	ID        uint      `json:"id"`
	NameEn    string    `json:"name_en"`
	NameAr    string    `json:"name_ar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Certification represents a teaching certification
type Certification struct {
	ID        uint      `json:"id"`
	NameEn    string    `json:"name_en"`
	NameAr    string    `json:"name_ar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// City represents a location/city
type City struct {
	ID        uint      `json:"id"`
	NameEn    string    `json:"name_en"`
	NameAr    string    `json:"name_ar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
