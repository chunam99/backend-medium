package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// AWSConfig holds AWS configuration values
type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	Endpoint        string
	UsePathStyle    bool
}

// LoadEnv loads environment variables from the .env file
func LoadEnv() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("error loading .env file")
	}
	return nil
}

// GetAWSConfig loads AWS configurations from environment variables
func GetAWSConfig() (*AWSConfig, error) {
	usePathStyle := os.Getenv("AWS_USE_PATH_STYLE_ENDPOINT") == "true"

	cfg := &AWSConfig{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region:          os.Getenv("AWS_DEFAULT_REGION"),
		Bucket:          os.Getenv("AWS_BUCKET"),
		Endpoint:        os.Getenv("AWS_ENDPOINT"),
		UsePathStyle:    usePathStyle,
	}

	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.Region == "" || cfg.Bucket == "" || cfg.Endpoint == "" {
		return nil, fmt.Errorf("missing required AWS environment variables")
	}

	return cfg, nil
}
