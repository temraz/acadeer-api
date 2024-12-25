package booking

import (
	"fmt"
	"net/http"
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
	bookingRepo      repository.BookingRepository
	userRepo         repository.UserRepository
	notificationRepo repository.NotificationRepository
	db               *gorm.DB
	r2Client         *storage.R2Client
}

func NewHandler(db *gorm.DB, r2Client *storage.R2Client) *Handler {
	return &Handler{
		bookingRepo:      repository.NewBookingRepository(db, r2Client),
		userRepo:         repository.NewUserRepository(db),
		notificationRepo: repository.NewNotificationRepository(db),
		db:               db,
		r2Client:         r2Client,
	}
}

// CreateBooking handles the creation of a new booking by a school admin
func (h *Handler) CreateBooking(c *gin.Context) {
	lang := c.GetString("language")
	schoolAdminID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	// Check if user is a school admin
	if userType != 3 {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	// Get school admin details to get school_id
	schoolAdmin, err := h.userRepo.FindByID(schoolAdminID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	if schoolAdmin.SchoolID == nil {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "مدير المدرسة غير مرتبط بأي مدرسة")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "School admin is not associated with any school")
		}
		return
	}

	var req models.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format")
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format")
		return
	}

	// Validate dates
	if startDate.Before(time.Now().Truncate(24 * time.Hour)) {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "لا يمكن أن يكون تاريخ البدء في الماضي")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Start date cannot be in the past")
		}
		return
	}

	if endDate.Before(startDate) {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "يجب أن يكون تاريخ الانتهاء بعد تاريخ البدء")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "End date must be after start date")
		}
		return
	}

	// Check for overlapping bookings
	hasOverlap, err := h.bookingRepo.HasOverlappingBookings(req.TeacherUserID, startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}
	if hasOverlap {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "هذا المعلم محجوز بالفعل في التواريخ المحددة. يجى اختيار تواريخ مختلفة")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "This teacher is already booked for the selected date(s). Please choose different dates.")
		}
		return
	}

	// Get teacher details to get price_per_day
	teacher, err := h.userRepo.FindByID(req.TeacherUserID)
	if err != nil {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusNotFound, "المعلم غير موجود")
		} else {
			utils.ErrorResponse(c, http.StatusNotFound, "Teacher not found")
		}
		return
	}

	// Create booking
	booking := &models.TeacherBooking{
		SchoolAdminUserID: schoolAdminID,
		TeacherUserID:     req.TeacherUserID,
		SchoolID:          *schoolAdmin.SchoolID,
		Status:            1, // pending
		PricePerDay:       teacher.PricePerDay,
		PaymentStatus:     1, // pending
		StartDate:         startDate,
		EndDate:           endDate,
	}

	if err := h.bookingRepo.Create(booking); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Send notification to the teacher
	notification := &models.Notification{
		UserID:           int(req.TeacherUserID),
		SchoolID:         utils.IntPtr(int(*schoolAdmin.SchoolID)),
		Type:             "new_booking",
		RefrenceTable:    "teacher_bookings",
		RefrenceColumnID: booking.ID,
		TitleEn:          "New Booking Request",
		TitleAr:          "طلب حجز جديد",
		MessageEn:        "You have received a new booking request from a school.",
		MessageAr:        "لقد تلقيت طلب حجز جديد من مدرسة",
	}

	if err := h.notificationRepo.Create(notification); err != nil {
		// Log the error but don't return, as the booking was successful
		c.Error(err)
	}

	// Get the created booking with relations
	createdBooking, err := h.bookingRepo.FindByID(booking.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, i18n.GetMessage(lang, "booking_created"), createdBooking)
}

// GetSchoolAdminBookings returns all bookings for a school admin
func (h *Handler) GetSchoolAdminBookings(c *gin.Context) {
	lang := c.GetString("language")
	schoolAdminID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	// Check if user is a school admin
	if userType != 3 {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	bookings, err := h.bookingRepo.FindBySchoolAdminID(schoolAdminID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), bookings)
}

// GetTeacherBookings returns all bookings for a teacher
func (h *Handler) GetTeacherBookings(c *gin.Context) {
	lang := c.GetString("language")
	teacherID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	// Check if user is a teacher
	if userType != 2 {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	bookings, err := h.bookingRepo.FindByTeacherID(teacherID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), bookings)
}

// GetTeacherFutureBookings returns future accepted bookings for a specific teacher
func (h *Handler) GetTeacherFutureBookings(c *gin.Context) {
	lang := c.GetString("language")
	userType := c.GetInt("user_type")

	// Check if user is a school admin
	if userType != 3 {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	// Get teacher ID from query parameter
	teacherID := c.Query("teacher_id")
	if teacherID == "" {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "معرف المعلم مطلوب")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Teacher ID is required")
		}
		return
	}

	// Convert teacher_id to uint
	teacherIDUint := utils.StringToUint(teacherID)
	if teacherIDUint == 0 {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "معرف المعلم غير صالح")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid teacher ID")
		}
		return
	}

	// Get future accepted bookings
	bookings, err := h.bookingRepo.FindTeacherFutureAcceptedBookings(teacherIDUint)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), bookings)
}

// GetTeacherScheduleStats returns schedule statistics for a teacher
func (h *Handler) GetTeacherScheduleStats(c *gin.Context) {
	lang := c.GetString("language")
	teacherID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	// Check if user is a teacher
	if userType != 2 {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	stats, err := h.bookingRepo.GetTeacherScheduleStats(teacherID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), stats)
}

// GetTeacherBookingsByDate returns bookings for a teacher on a specific date
func (h *Handler) GetTeacherBookingsByDate(c *gin.Context) {
	lang := c.GetString("language")
	teacherID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	// Check if user is a teacher
	if userType != 2 {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	// Get date from query parameter
	date := c.Query("date")
	if date == "" {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "التاريخ مطلوب")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Date is required")
		}
		return
	}

	// Validate date format
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "تنسيق التاريخ غير صالح. يجب أن يكون بتنسيق YYYY-MM-DD")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid date format. Must be YYYY-MM-DD")
		}
		return
	}

	bookings, err := h.bookingRepo.GetTeacherBookingsByDate(teacherID, date)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), bookings)
}

// GetBookingDetails returns detailed information about a specific booking
func (h *Handler) GetBookingDetails(c *gin.Context) {
	lang := c.GetString("language")
	teacherID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	// Check if user is a teacher
	if userType != 2 {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	// Get booking ID from path parameter
	bookingID := utils.StringToUint(c.Param("id"))
	if bookingID == 0 {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "معرف الحجز غير صالح")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID")
		}
		return
	}

	// Get booking details
	booking, err := h.bookingRepo.GetBookingDetailsByID(bookingID, teacherID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if lang == "ar" {
				utils.ErrorResponse(c, http.StatusNotFound, "الحجز غير موجود")
			} else {
				utils.ErrorResponse(c, http.StatusNotFound, "Booking not found")
			}
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), booking)
}

// UpdateBookingStatus handles accepting or rejecting a booking by a teacher
func (h *Handler) UpdateBookingStatus(c *gin.Context) {
	lang := c.GetString("language")
	teacherID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	// Check if user is a teacher
	if userType != 2 {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	// Get booking ID from path parameter
	bookingID := utils.StringToUint(c.Param("id"))
	if bookingID == 0 {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "معرف الحجز غير صالح")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID")
		}
		return
	}

	// Parse request body
	var req struct {
		Action string `json:"action" binding:"required,oneof=accept reject"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Determine status based on action
	var status int
	var titleEn, titleAr, messageEn, messageAr string

	if req.Action == "accept" {
		status = 2 // accepted
		titleEn = "Booking Accepted"
		titleAr = "تم قبول الحجز"
		messageEn = "The teacher has accepted your booking request."
		messageAr = "قام المعلم بقبول طلب الحجز الخاص بك"
	} else {
		status = 3 // rejected
		titleEn = "Booking Rejected"
		titleAr = "تم رفض الحجز"
		messageEn = "The teacher has rejected your booking request."
		messageAr = "قام المعلم برفض طلب الحجز الخاص بك"
	}

	// Update booking status
	updatedBooking, err := h.bookingRepo.UpdateBookingStatus(bookingID, teacherID, status)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if lang == "ar" {
				utils.ErrorResponse(c, http.StatusNotFound, "الحجز غير موجود")
			} else {
				utils.ErrorResponse(c, http.StatusNotFound, "Booking not found")
			}
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	// Send notification to school admin
	notification := &models.Notification{
		UserID:           int(updatedBooking.SchoolAdminUserID),
		SchoolID:         utils.IntPtr(int(updatedBooking.SchoolID)),
		Type:             fmt.Sprintf("booking_%s", req.Action),
		RefrenceTable:    "teacher_bookings",
		RefrenceColumnID: updatedBooking.ID,
		TitleEn:          titleEn,
		TitleAr:          titleAr,
		MessageEn:        messageEn,
		MessageAr:        messageAr,
	}

	if err := h.notificationRepo.Create(notification); err != nil {
		// Log the error but don't return, as the status update was successful
		c.Error(err)
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), updatedBooking)
}

// GetSchoolScheduleStats returns schedule statistics for a school
func (h *Handler) GetSchoolScheduleStats(c *gin.Context) {
	lang := c.GetString("language")
	schoolAdminID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	// Check if user is a school admin
	if userType != 3 {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	// Get school admin details to get school_id
	schoolAdmin, err := h.userRepo.FindByID(schoolAdminID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	if schoolAdmin.SchoolID == nil {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "مدير المدرسة غير مرتبط بأي مدرسة")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "School admin is not associated with any school")
		}
		return
	}

	stats, err := h.bookingRepo.GetSchoolScheduleStats(*schoolAdmin.SchoolID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), stats)
}

// GetSchoolBookingsByDate returns bookings for a school on a specific date
func (h *Handler) GetSchoolBookingsByDate(c *gin.Context) {
	lang := c.GetString("language")
	schoolAdminID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	// Check if user is a school admin
	if userType != 3 {
		utils.ErrorResponse(c, http.StatusUnauthorized, i18n.GetMessage(lang, "unauthorized"))
		return
	}

	// Get school admin details to get school_id
	schoolAdmin, err := h.userRepo.FindByID(schoolAdminID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	if schoolAdmin.SchoolID == nil {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "مدير المدرسة غير مرتبط بأي مدرسة")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "School admin is not associated with any school")
		}
		return
	}

	// Get date from query parameter
	date := c.Query("date")
	if date == "" {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "التاريخ مطلوب")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Date is required")
		}
		return
	}

	// Validate date format
	_, err = time.Parse("2006-01-02", date)
	if err != nil {
		if lang == "ar" {
			utils.ErrorResponse(c, http.StatusBadRequest, "تنسيق التاريخ غير صالح. يجب أن يكون بتنسيق YYYY-MM-DD")
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid date format. Must be YYYY-MM-DD")
		}
		return
	}

	bookings, err := h.bookingRepo.GetSchoolBookingsByDate(*schoolAdmin.SchoolID, date)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, i18n.GetMessage(lang, "server_error"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, i18n.GetMessage(lang, "success"), bookings)
}
