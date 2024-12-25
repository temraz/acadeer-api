package repository

import (
	"sort"
	"subteacher/backend/internal/models"
	"subteacher/backend/internal/storage"
	"time"

	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(booking *models.TeacherBooking) error
	FindByID(id uint) (*models.TeacherBooking, error)
	FindByTeacherID(teacherID uint) ([]models.TeacherBooking, error)
	FindBySchoolAdminID(schoolAdminID uint) ([]models.TeacherBooking, error)
	Update(booking *models.TeacherBooking) error
	HasOverlappingBookings(teacherID uint, startDate, endDate time.Time) (bool, error)
	FindTeacherFutureAcceptedBookings(teacherID uint) ([]models.TeacherBooking, error)
	GetTeacherScheduleStats(teacherID uint) (*TeacherScheduleStats, error)
	GetTeacherBookingsByDate(teacherID uint, date string) ([]BookingWithSchool, error)
	GetBookingDetailsByID(bookingID uint, teacherID uint) (*BookingDetails, error)
	UpdateBookingStatus(bookingID uint, teacherID uint, status int) (*models.TeacherBooking, error)
	GetSchoolScheduleStats(schoolID uint) (*SchoolScheduleStats, error)
	GetSchoolBookingsByDate(schoolID uint, date string) ([]BookingWithTeacher, error)
}

type bookingRepository struct {
	db       *gorm.DB
	r2Client *storage.R2Client
}

func NewBookingRepository(db *gorm.DB, r2Client *storage.R2Client) BookingRepository {
	return &bookingRepository{
		db:       db,
		r2Client: r2Client,
	}
}

func (r *bookingRepository) Create(booking *models.TeacherBooking) error {
	return r.db.Create(booking).Error
}

func (r *bookingRepository) FindByID(id uint) (*models.TeacherBooking, error) {
	var booking models.TeacherBooking
	err := r.db.Preload("Teacher").Preload("School").Where("id = ?", id).First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) FindByTeacherID(teacherID uint) ([]models.TeacherBooking, error) {
	var bookings []models.TeacherBooking
	err := r.db.Preload("Teacher").Preload("School").Where("teacher_user_id = ?", teacherID).Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) FindBySchoolAdminID(schoolAdminID uint) ([]models.TeacherBooking, error) {
	var bookings []models.TeacherBooking
	err := r.db.Preload("Teacher").Preload("School").Where("school_admin_user_id = ?", schoolAdminID).Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) Update(booking *models.TeacherBooking) error {
	return r.db.Save(booking).Error
}

func (r *bookingRepository) HasOverlappingBookings(teacherID uint, startDate, endDate time.Time) (bool, error) {
	var bookings []models.TeacherBooking

	// Debug: Print input parameters
	println("Checking overlaps for teacher:", teacherID)
	println("Start Date:", startDate.Format("2006-01-02"))
	println("End Date:", endDate.Format("2006-01-02"))

	query := r.db.Where("teacher_user_id = ?", teacherID).
		Where("status != ?", 3) // Exclude rejected bookings

	// Check for any booking that overlaps with the given date range
	query = query.Where(
		"(DATE(start_date) <= ? AND DATE(end_date) >= ?) OR "+ // Existing booking spans the new dates
			"(DATE(start_date) = ?) OR "+ // Same start date
			"(DATE(end_date) = ?)", // Same end date
		endDate.Format("2006-01-02"), startDate.Format("2006-01-02"),
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	err := query.Find(&bookings).Error
	if err != nil {
		println("Error checking overlaps:", err.Error())
		return false, err
	}

	// Debug: Print found bookings
	println("Found", len(bookings), "overlapping bookings")
	for _, b := range bookings {
		println("Overlapping booking ID:", b.ID)
		println("- Start Date:", b.StartDate.Format("2006-01-02"))
		println("- End Date:", b.EndDate.Format("2006-01-02"))
		println("- Status:", b.Status)
	}

	return len(bookings) > 0, nil
}

func (r *bookingRepository) FindTeacherFutureAcceptedBookings(teacherID uint) ([]models.TeacherBooking, error) {
	var bookings []models.TeacherBooking
	err := r.db.Preload("Teacher").
		Preload("School").
		Where("teacher_user_id = ? AND status = ? AND start_date >= ?",
			teacherID,
			2, // accepted status
			time.Now().Format("2006-01-02")).
		Order("start_date ASC").
		Find(&bookings).Error
	return bookings, err
}

type TeacherScheduleStats struct {
	TotalAssignments     int64    `json:"total_assignments"`
	ConfirmedAssignments int64    `json:"confirmed_assignments"`
	PendingAssignments   int64    `json:"pending_assignments"`
	TotalSchools         int64    `json:"total_schools"`
	BookedDates          []string `json:"booked_dates"`
}

func (r *bookingRepository) GetTeacherScheduleStats(teacherID uint) (*TeacherScheduleStats, error) {
	var stats TeacherScheduleStats

	// Get total assignments
	if err := r.db.Model(&models.TeacherBooking{}).
		Where("teacher_user_id = ?", teacherID).
		Count(&stats.TotalAssignments).Error; err != nil {
		return nil, err
	}

	// Get confirmed assignments (status = 2)
	if err := r.db.Model(&models.TeacherBooking{}).
		Where("teacher_user_id = ? AND status = ?", teacherID, 2).
		Count(&stats.ConfirmedAssignments).Error; err != nil {
		return nil, err
	}

	// Get pending assignments (status = 1)
	if err := r.db.Model(&models.TeacherBooking{}).
		Where("teacher_user_id = ? AND status = ?", teacherID, 1).
		Count(&stats.PendingAssignments).Error; err != nil {
		return nil, err
	}

	// Get total unique schools
	if err := r.db.Model(&models.TeacherBooking{}).
		Where("teacher_user_id = ?", teacherID).
		Distinct("school_id").
		Count(&stats.TotalSchools).Error; err != nil {
		return nil, err
	}

	// Get all booked dates
	var bookings []models.TeacherBooking
	if err := r.db.Where("teacher_user_id = ? AND status != ?", teacherID, 3). // Exclude rejected bookings
											Select("start_date, end_date").
											Find(&bookings).Error; err != nil {
		return nil, err
	}

	// Create a map to store unique dates
	uniqueDates := make(map[string]bool)
	for _, booking := range bookings {
		current := booking.StartDate
		for !current.After(booking.EndDate) {
			uniqueDates[current.Format("2006-01-02")] = true
			current = current.AddDate(0, 0, 1)
		}
	}

	// Convert unique dates map to slice
	stats.BookedDates = make([]string, 0, len(uniqueDates))
	for date := range uniqueDates {
		stats.BookedDates = append(stats.BookedDates, date)
	}

	// Sort dates
	sort.Strings(stats.BookedDates)

	return &stats, nil
}

type BookingWithSchool struct {
	ID                uint      `json:"id"`
	SchoolID          uint      `json:"school_id"`
	Status            int       `json:"status"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	PricePerDay       float64   `json:"price_per_day"`
	PaymentStatus     int       `json:"payment_status"`
	SchoolNameEn      string    `json:"school_name_en"`
	SchoolNameAr      string    `json:"school_name_ar"`
	CityNameEn        string    `json:"city_name_en"`
	CityNameAr        string    `json:"city_name_ar"`
	AdminPhoneNumber  string    `json:"admin_phone_number"`
	SubjectNameEn     string    `json:"subject_name_en"`
	SubjectNameAr     string    `json:"subject_name_ar"`
	LogoR2Link        string    `json:"logo_r2_link"`
	TeacherProfilePic string    `json:"teacher_profile_pic"`
	TeacherName       string    `json:"teacher_name"`
}

func (r *bookingRepository) GetTeacherBookingsByDate(teacherID uint, date string) ([]BookingWithSchool, error) {
	var bookings []BookingWithSchool

	err := r.db.Table("teacher_bookings").
		Select(`
			teacher_bookings.id,
			teacher_bookings.school_id,
			teacher_bookings.status,
			teacher_bookings.start_date,
			teacher_bookings.end_date,
			teacher_bookings.price_per_day,
			teacher_bookings.payment_status,
			schools.name_en as school_name_en,
			schools.name_ar as school_name_ar,
			schools.logo as logo_r2_link,
			cities.name_en as city_name_en,
			cities.name_ar as city_name_ar,
			admin.phone_number as admin_phone_number,
			subjects.name_en as subject_name_en,
			subjects.name_ar as subject_name_ar,
			teacher.profile_picture as teacher_profile_pic,
			teacher.full_name as teacher_name
		`).
		Joins("LEFT JOIN schools ON teacher_bookings.school_id = schools.id").
		Joins("LEFT JOIN cities ON schools.city_id = cities.id").
		Joins("LEFT JOIN users teacher ON teacher_bookings.teacher_user_id = teacher.id").
		Joins("LEFT JOIN users admin ON teacher_bookings.school_admin_user_id = admin.id").
		Joins("LEFT JOIN subjects ON teacher.subject_id = subjects.id").
		Where("teacher_bookings.teacher_user_id = ? AND ? BETWEEN DATE(teacher_bookings.start_date) AND DATE(teacher_bookings.end_date)",
			teacherID, date).
		Order("teacher_bookings.start_date ASC").
		Scan(&bookings).Error

	// Get presigned URLs for each booking's school logo and teacher profile picture
	for i := range bookings {
		if bookings[i].LogoR2Link != "" {
			bookings[i].LogoR2Link = r.r2Client.GetFileURL(bookings[i].LogoR2Link)
		}
		if bookings[i].TeacherProfilePic != "" {
			bookings[i].TeacherProfilePic = r.r2Client.GetFileURL(bookings[i].TeacherProfilePic)
		}
	}

	return bookings, err
}

type BookingDetails struct {
	ID                 uint      `json:"id"`
	SchoolID           uint      `json:"school_id"`
	Status             int       `json:"status"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	PricePerDay        float64   `json:"price_per_day"`
	PaymentStatus      int       `json:"payment_status"`
	SchoolNameEn       string    `json:"school_name_en"`
	SchoolNameAr       string    `json:"school_name_ar"`
	CityNameEn         string    `json:"city_name_en"`
	CityNameAr         string    `json:"city_name_ar"`
	SubjectNameEn      string    `json:"subject_name_en"`
	SubjectNameAr      string    `json:"subject_name_ar"`
	TeacherName        string    `json:"teacher_name"`
	TeacherEmail       string    `json:"teacher_email"`
	TeacherPhoneNumber string    `json:"teacher_phone_number"`
	LogoR2Link         string    `json:"logo_r2_link"`
	TeacherProfilePic  string    `json:"teacher_profile_pic"`
}

func (r *bookingRepository) GetBookingDetailsByID(bookingID uint, teacherID uint) (*BookingDetails, error) {
	var booking BookingDetails

	err := r.db.Table("teacher_bookings").
		Select(`
			teacher_bookings.id,
			teacher_bookings.school_id,
			teacher_bookings.status,
			teacher_bookings.start_date,
			teacher_bookings.end_date,
			teacher_bookings.price_per_day,
			teacher_bookings.payment_status,
			schools.name_en as school_name_en,
			schools.name_ar as school_name_ar,
			schools.logo as logo_r2_link,
			cities.name_en as city_name_en,
			cities.name_ar as city_name_ar,
			subjects.name_en as subject_name_en,
			subjects.name_ar as subject_name_ar,
			teacher.full_name as teacher_name,
			teacher.email as teacher_email,
			teacher.phone_number as teacher_phone_number,
			teacher.profile_picture as teacher_profile_pic
		`).
		Joins("LEFT JOIN schools ON teacher_bookings.school_id = schools.id").
		Joins("LEFT JOIN cities ON schools.city_id = cities.id").
		Joins("LEFT JOIN users teacher ON teacher_bookings.teacher_user_id = teacher.id").
		Joins("LEFT JOIN subjects ON teacher.subject_id = subjects.id").
		Where("teacher_bookings.id = ? AND teacher_bookings.teacher_user_id = ?", bookingID, teacherID).
		First(&booking).Error

	if err != nil {
		return nil, err
	}

	// Get presigned URL for the school logo
	if booking.LogoR2Link != "" {
		booking.LogoR2Link = r.r2Client.GetFileURL(booking.LogoR2Link)
	}

	// Get presigned URL for the teacher's profile picture
	if booking.TeacherProfilePic != "" {
		booking.TeacherProfilePic = r.r2Client.GetFileURL(booking.TeacherProfilePic)
	}

	return &booking, nil
}

// UpdateBookingStatus updates the booking status and teacher_accepted_at timestamp
func (r *bookingRepository) UpdateBookingStatus(bookingID uint, teacherID uint, status int) (*models.TeacherBooking, error) {
	now := time.Now()
	booking := &models.TeacherBooking{}

	// First find the booking and verify it belongs to the teacher
	err := r.db.Where("id = ? AND teacher_user_id = ?", bookingID, teacherID).First(booking).Error
	if err != nil {
		return nil, err
	}

	// Update the booking status and teacher_accepted_at
	booking.Status = status
	if status == 2 { // accepted
		booking.TeacherAcceptedAt = &now
	}

	err = r.db.Save(booking).Error
	if err != nil {
		return nil, err
	}

	// Reload the booking with relations
	err = r.db.Preload("Teacher").Preload("School").Where("id = ?", bookingID).First(booking).Error
	if err != nil {
		return nil, err
	}

	return booking, nil
}

type SchoolScheduleStats struct {
	TotalAssignments     int64    `json:"total_assignments"`
	ConfirmedAssignments int64    `json:"confirmed_assignments"`
	PendingAssignments   int64    `json:"pending_assignments"`
	TotalTeachers        int64    `json:"total_teachers"`
	BookedDates          []string `json:"booked_dates"`
}

func (r *bookingRepository) GetSchoolScheduleStats(schoolID uint) (*SchoolScheduleStats, error) {
	var stats SchoolScheduleStats

	// Get total assignments
	if err := r.db.Model(&models.TeacherBooking{}).
		Where("school_id = ?", schoolID).
		Count(&stats.TotalAssignments).Error; err != nil {
		return nil, err
	}

	// Get confirmed assignments (status = 2)
	if err := r.db.Model(&models.TeacherBooking{}).
		Where("school_id = ? AND status = ?", schoolID, 2).
		Count(&stats.ConfirmedAssignments).Error; err != nil {
		return nil, err
	}

	// Get pending assignments (status = 1)
	if err := r.db.Model(&models.TeacherBooking{}).
		Where("school_id = ? AND status = ?", schoolID, 1).
		Count(&stats.PendingAssignments).Error; err != nil {
		return nil, err
	}

	// Get total unique teachers
	if err := r.db.Model(&models.TeacherBooking{}).
		Where("school_id = ?", schoolID).
		Distinct("teacher_user_id").
		Count(&stats.TotalTeachers).Error; err != nil {
		return nil, err
	}

	// Get all booked dates
	var bookings []models.TeacherBooking
	if err := r.db.Where("school_id = ? AND status != ?", schoolID, 3). // Exclude rejected bookings
										Select("start_date, end_date").
										Find(&bookings).Error; err != nil {
		return nil, err
	}

	// Create a map to store unique dates
	uniqueDates := make(map[string]bool)
	for _, booking := range bookings {
		current := booking.StartDate
		for !current.After(booking.EndDate) {
			uniqueDates[current.Format("2006-01-02")] = true
			current = current.AddDate(0, 0, 1)
		}
	}

	// Convert unique dates map to slice
	stats.BookedDates = make([]string, 0, len(uniqueDates))
	for date := range uniqueDates {
		stats.BookedDates = append(stats.BookedDates, date)
	}

	// Sort dates
	sort.Strings(stats.BookedDates)

	return &stats, nil
}

type BookingWithTeacher struct {
	ID                uint      `json:"id"`
	Status            int       `json:"status"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	PricePerDay       float64   `json:"price_per_day"`
	PaymentStatus     int       `json:"payment_status"`
	TeacherID         uint      `json:"teacher_id"`
	TeacherName       string    `json:"teacher_name"`
	TeacherEmail      string    `json:"teacher_email"`
	TeacherPhone      string    `json:"teacher_phone"`
	SubjectNameEn     string    `json:"subject_name_en"`
	SubjectNameAr     string    `json:"subject_name_ar"`
	TeacherProfilePic string    `json:"teacher_profile_pic"`
	TeacherCityNameEn string    `json:"teacher_city_name_en"`
	TeacherCityNameAr string    `json:"teacher_city_name_ar"`
}

func (r *bookingRepository) GetSchoolBookingsByDate(schoolID uint, date string) ([]BookingWithTeacher, error) {
	var bookings []BookingWithTeacher

	err := r.db.Table("teacher_bookings").
		Select(`
			teacher_bookings.id,
			teacher_bookings.status,
			teacher_bookings.start_date,
			teacher_bookings.end_date,
			teacher_bookings.price_per_day,
			teacher_bookings.payment_status,
			teacher.id as teacher_id,
			teacher.full_name as teacher_name,
			teacher.email as teacher_email,
			teacher.phone_number as teacher_phone,
			subjects.name_en as subject_name_en,
			subjects.name_ar as subject_name_ar,
			teacher.profile_picture as teacher_profile_pic,
			cities.name_en as teacher_city_name_en,
			cities.name_ar as teacher_city_name_ar
		`).
		Joins("LEFT JOIN users teacher ON teacher_bookings.teacher_user_id = teacher.id").
		Joins("LEFT JOIN subjects ON teacher.subject_id = subjects.id").
		Joins("LEFT JOIN cities ON teacher.city_id = cities.id").
		Where("teacher_bookings.school_id = ? AND ? BETWEEN DATE(teacher_bookings.start_date) AND DATE(teacher_bookings.end_date)",
			schoolID, date).
		Order("teacher_bookings.start_date ASC").
		Scan(&bookings).Error

	// Get presigned URLs for teacher profile pictures
	for i := range bookings {
		if bookings[i].TeacherProfilePic != "" {
			bookings[i].TeacherProfilePic = r.r2Client.GetFileURL(bookings[i].TeacherProfilePic)
		}
	}

	return bookings, err
}
