package main

import (
	"fmt"
	"log"
	"os"
	"subteacher/backend/config"
	"subteacher/backend/internal/handlers/auth"
	"subteacher/backend/internal/handlers/booking"
	"subteacher/backend/internal/handlers/lookup"
	"subteacher/backend/internal/handlers/notification"
	"subteacher/backend/internal/handlers/user"
	"subteacher/backend/internal/middleware"
	"subteacher/backend/internal/repository"
	"subteacher/backend/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := &config.Config{
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "3306"),
		DBUser:     getEnvOrDefault("DB_USER", "root"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", ""),
		DBName:     getEnvOrDefault("DB_NAME", "subteacher"),
		ServerPort: getEnvOrDefault("SERVER_PORT", "8080"),
		JWTConfig: config.JWTConfig{
			Secret:               getEnvOrDefault("JWT_SECRET", "your_jwt_secret"),
			AccessTokenDuration:  getEnvOrDefault("ACCESS_TOKEN_DURATION", "24h"),
			RefreshTokenDuration: getEnvOrDefault("REFRESH_TOKEN_DURATION", "720h"),
		},
		R2Config: config.R2Config{
			AccountID:       getEnvOrDefault("R2_ACCOUNT_ID", ""),
			AccessKeyID:     getEnvOrDefault("R2_ACCESS_KEY_ID", ""),
			AccessKeySecret: getEnvOrDefault("R2_ACCESS_KEY_SECRET", ""),
			BucketName:      getEnvOrDefault("R2_BUCKET_NAME", ""),
		},
	}

	// Setup database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize R2 client
	r2Client, err := storage.NewR2Client(
		cfg.R2Config.AccountID,
		cfg.R2Config.AccessKeyID,
		cfg.R2Config.AccessKeySecret,
		cfg.R2Config.BucketName,
	)
	if err != nil {
		log.Fatalf("Failed to initialize R2 client: %v", err)
	}

	// Initialize handlers
	authHandler := auth.NewHandler(db, cfg)
	userHandler := user.NewHandler(db, cfg)
	lookupHandler := lookup.NewHandler(db, cfg)
	educationHandler := user.NewEducationHandler(repository.NewEducationRepository(db), r2Client)
	bookingHandler := booking.NewHandler(db, r2Client)
	notificationHandler := notification.NewHandler(db)

	// Setup router
	router := gin.Default()

	// Add middlewares
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LanguageMiddleware())
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes
		authHandler.SetupRoutes(api.Group("/auth"))

		// User routes
		users := api.Group("/users")
		user.RegisterRoutes(users, userHandler, educationHandler, cfg)

		// Lookup routes
		lookupHandler.SetupRoutes(api.Group("/lookups"))

		// Booking routes
		booking.RegisterRoutes(api.Group("/bookings"), bookingHandler, cfg)

		// Notification routes
		notification.RegisterRoutes(api.Group("/notifications"), notificationHandler, cfg)
	}

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
