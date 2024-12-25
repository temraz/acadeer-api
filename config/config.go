package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
	JWTConfig  JWTConfig
	R2Config   R2Config
}

type JWTConfig struct {
	Secret               string
	AccessTokenDuration  string
	RefreshTokenDuration string
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBPort:     os.Getenv("DB_PORT"),
		ServerPort: os.Getenv("SERVER_PORT"),
		JWTConfig: JWTConfig{
			Secret:               os.Getenv("JWT_SECRET"),
			AccessTokenDuration:  os.Getenv("ACCESS_TOKEN_DURATION"),
			RefreshTokenDuration: os.Getenv("REFRESH_TOKEN_DURATION"),
		},
		R2Config: R2Config{
			AccountID:       os.Getenv("R2_ACCOUNT_ID"),
			AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
			AccessKeySecret: os.Getenv("R2_ACCESS_KEY_SECRET"),
			BucketName:      os.Getenv("R2_BUCKET_NAME"),
		},
	}

	// Validate required R2 settings
	if config.R2Config.BucketName == "" {
		return nil, fmt.Errorf("R2_BUCKET_NAME is required")
	}
	if config.R2Config.AccountID == "" {
		return nil, fmt.Errorf("R2_ACCOUNT_ID is required")
	}
	if config.R2Config.AccessKeyID == "" {
		return nil, fmt.Errorf("R2_ACCESS_KEY_ID is required")
	}
	if config.R2Config.AccessKeySecret == "" {
		return nil, fmt.Errorf("R2_ACCESS_KEY_SECRET is required")
	}

	return config, nil
}
