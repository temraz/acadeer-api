package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UploadDir    = "uploads/schools/logos"
	MaxFileSize  = 5 << 20 // 5MB
	AllowedTypes = ".jpg,.jpeg,.png"
)

func SaveUploadedFile(c *gin.Context, file *multipart.FileHeader) (string, error) {
	// Check file size
	if file.Size > MaxFileSize {
		return "", fmt.Errorf("file size exceeds maximum limit of 5MB")
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !strings.Contains(AllowedTypes, ext) {
		return "", fmt.Errorf("invalid file type. Allowed types: %s", AllowedTypes)
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s-%s%s",
		time.Now().Format("20060102"),
		uuid.New().String(),
		ext,
	)
	filepath := filepath.Join(UploadDir, filename)

	// Save the file using Gin's SaveUploadedFile
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return filepath, nil
}
