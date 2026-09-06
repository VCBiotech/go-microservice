package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// StorageConfig holds storage-related configurations
type StorageConfig struct {
	AWSRegion            string
	AWSAccessKeyID       string
	AWSSecretAccessKey   string
	GCPProjectID         string
	GCPCredentialsFile   string
	DefaultCloud         string // e.g., "aws", "gcp"
	ReplicateToAllClouds bool   // Whether to replicate uploads to all configured clouds
}

// DatabaseConfig holds database-related configurations
type DatabaseConfig struct {
	Username string
	Password string
	Host     string
	Port     uint16
	Name     string
	SslMode  string
}

// AppConfig holds application-wide configurations
type AppConfig struct {
	ServerPort    uint16
	Database      DatabaseConfig
	BucketName    string
	GotenbergURL  string
	StorageConfig StorageConfig
}

func LoadConfig() *AppConfig {
	cfg := AppConfig{
		GotenbergURL: "http://localhost:3001",
		Database: DatabaseConfig{
			Name:     "equilibria_files",
			Port:     5432,
			Username: "postgres",
			Password: "password",
			Host:     "localhost",
			SslMode:  "disable",
		},
		ServerPort: 3000,
		BucketName: "test-file-manager-2025",
		StorageConfig: StorageConfig{
			AWSRegion:            "us-east-1",
			AWSAccessKeyID:       "",
			AWSSecretAccessKey:   "",
			GCPProjectID:         "",
			GCPCredentialsFile:   "",
			DefaultCloud:         "aws",
			ReplicateToAllClouds: false,
		},
	}

	if secrets, exists := os.LookupEnv("SECRETS"); exists {
		// Parse secrets which is a JSON string into a map
		secretsMap := make(map[string]string)
		err := json.Unmarshal([]byte(secrets), &secretsMap)
		if err != nil {
			log.Fatalf("Error parsing secrets: %v", err)
		}

		if dbUrl, exists := secretsMap["DATABASE_URL"]; exists {
			// Remove protocol from URL
			parts := strings.Split(dbUrl, "://")

			// Build regex to extract username, password, host, port, and name
			regex := regexp.MustCompile(`^(.*):(.*)@(.*):(\d+)/(.*)$`)
			matches := regex.FindStringSubmatch(parts[1])

			// Parse port into uint16
			port, err := strconv.ParseUint(matches[4], 10, 16)
			if err != nil {
				log.Fatalf("Error parsing port: %v", err)
			}

			if len(matches) == 6 {
				cfg.Database.Username = matches[1]
				cfg.Database.Password = matches[2]
				cfg.Database.Host = matches[3]
				cfg.Database.Port = uint16(port)
				cfg.Database.Name = matches[5]
			}
		}

		if bucketName, exists := secretsMap["BUCKET_NAME"]; exists {
			cfg.BucketName = bucketName
		}

		if AWSRegion, exists := secretsMap["AWS_REGION"]; exists {
			cfg.StorageConfig.AWSRegion = AWSRegion
		}

		if AWSAccessKeyID, exists := secretsMap["AWS_ACCESS_KEY_ID"]; exists {
			cfg.StorageConfig.AWSAccessKeyID = AWSAccessKeyID
		}

		if AWSSecretAccessKey, exists := secretsMap["AWS_SECRET_ACCESS_KEY"]; exists {
			cfg.StorageConfig.AWSSecretAccessKey = AWSSecretAccessKey
		}

		if GCPProjectID, exists := secretsMap["GCP_PROJECT_ID"]; exists {
			cfg.StorageConfig.GCPProjectID = GCPProjectID
		}
		if GCPCredentialsFile, exists := secretsMap["GCP_CREDENTIALS_FILE"]; exists {
			cfg.StorageConfig.GCPCredentialsFile = GCPCredentialsFile
		}

		if DefaultCloud, exists := secretsMap["DEFAULT_CLOUD"]; exists {
			cfg.StorageConfig.DefaultCloud = DefaultCloud
		}

		if ReplicateToAllClouds, exists := secretsMap["REPLICATE_TO_ALL_CLOUDS"]; exists {
			cfg.StorageConfig.ReplicateToAllClouds = ReplicateToAllClouds == "true"
		}

		if GotenbergURL, exists := secretsMap["GOTENBERG_URL"]; exists {
			cfg.GotenbergURL = GotenbergURL
		}
	}

	if v := os.Getenv("AWS_REGION"); v != "" {
		cfg.StorageConfig.AWSRegion = v
	}
	if v := os.Getenv("AWS_ACCESS_KEY_ID"); v != "" {
		cfg.StorageConfig.AWSAccessKeyID = v
	}
	if v := os.Getenv("AWS_SECRET_ACCESS_KEY"); v != "" {
		cfg.StorageConfig.AWSSecretAccessKey = v
	}
	if v := os.Getenv("BUCKET_NAME"); v != "" {
		cfg.BucketName = v
	}
	if v := os.Getenv("DEFAULT_CLOUD"); v != "" {
		cfg.StorageConfig.DefaultCloud = v
	}
	if v := os.Getenv("GCP_PROJECT_ID"); v != "" {
		cfg.StorageConfig.GCPProjectID = v
	}
	if v := os.Getenv("GCP_CREDENTIALS_FILE"); v != "" {
		cfg.StorageConfig.GCPCredentialsFile = v
	}
	if v := os.Getenv("REPLICATE_TO_ALL_CLOUDS"); v != "" {
		cfg.StorageConfig.ReplicateToAllClouds = v == "true"
	}
	if v := os.Getenv("GOTENBERG_URL"); v != "" {
		cfg.GotenbergURL = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		port, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			log.Fatalf("Error parsing SERVER_PORT: %v", err)
		}
		cfg.ServerPort = uint16(port)
	}

	return &cfg
}

func (c *AppConfig) LoadDbUri() string {
	db := c.Database
	databaseUri := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.Username,
		db.Password,
		db.Host,
		db.Port,
		db.Name,
		db.SslMode,
	)

	return databaseUri
}
