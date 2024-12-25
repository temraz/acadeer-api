package i18n

var Messages = map[string]map[string]string{
	"en": {
		"email_exists":          "Email already exists",
		"phone_exists":          "Phone number already exists",
		"invalid_credentials":   "Invalid email or password. Please try again.",
		"all_fields_required":   "All fields are required",
		"logo_required":         "Logo file is required",
		"file_parse_error":      "Failed to parse form data",
		"file_size_error":       "File size exceeds maximum limit of 5MB",
		"file_type_error":       "Invalid file type. Allowed types: %s",
		"server_error":          "Internal server error",
		"create_school_error":   "Failed to create school",
		"create_user_error":     "Failed to create user",
		"token_error":           "Failed to generate tokens",
		"refresh_required":      "Refresh token is required",
		"invalid_token":         "Invalid refresh token",
		"signup_success":        "User registered successfully",
		"school_signup_success": "School registered successfully",
		"login_success":         "Logged in successfully",
		"token_refresh_success": "Token refreshed successfully",
		"account_inactive":      "Your account is inactive. Please contact support.",
		"user_not_found":        "User not found",
		"success":               "Success",
	},
	"ar": {
		"email_exists":          "البريد الإلكتروني موجود مسبقاً",
		"phone_exists":          "رقم الهاتف موجود مسبقاً",
		"invalid_credentials":   "البريد الإلكتروني أو كلمة المرور غير صحيحة.",
		"all_fields_required":   "جميع الحقول مطلوبة",
		"logo_required":         "شعار المدرسة مطلوب",
		"file_parse_error":      "فشل في معالجة البيانات المرسلة",
		"file_size_error":       "حجم الملف يتجاوز الحد الأقصى المسموح به (5 ميجابايت)",
		"file_type_error":       "نوع الملف غير مسموح به. الأنواع المسموح بها: %s",
		"server_error":          "خطأ في الخادم",
		"create_school_error":   "فشل في إنشاء المدرسة",
		"create_user_error":     "فشل في إنشاء المستخدم",
		"token_error":           "فشل في إنشاء رموز المصادقة",
		"refresh_required":      "رمز التحديث مطلوب",
		"invalid_token":         "رمز التحديث غير صالح",
		"signup_success":        "تم تسجيل المستخدم بنجاح",
		"school_signup_success": "تم تسجيل المدرسة بنجاح",
		"login_success":         "تم تسجيل الدخول بنجاح",
		"token_refresh_success": "تم تحديث الرمز بنجاح",
		"account_inactive":      "حسابك غير نشط. يرجى الاتصال بالدعم.",
		"user_not_found":        "المستخدم غير موجود",
		"success":               "تم بنجاح",
	},
}

// GetMessage returns the translated message for the given key and language
func GetMessage(lang, key string) string {
	// Default to English if language not supported
	if lang != "ar" && lang != "en" {
		lang = "en"
	}

	if msg, exists := Messages[lang][key]; exists {
		return msg
	}

	// Return English message if translation not found
	return Messages["en"][key]
}
