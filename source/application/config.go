package application

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type DatabaseConfig struct {
	Username string
	Password string
	Host     string
	Port     uint16
	Name     string
	SslMode  string
}

type Config struct {
	ServerPort uint16
	Database   DatabaseConfig
}

func LoadConfig() Config {

	cfg := Config{
		Database: DatabaseConfig{
			Name:     "file_manager",
			Port:     5432,
			Username: "postgres",
			Password: "password",
			Host:     "localhost",
			SslMode:  "disable",
		},
		ServerPort: 3000,
	}

	if databaseName, exists := os.LookupEnv("DB_NAME"); exists {
		cfg.Database.Name = databaseName
	}

	if databaseHost, exists := os.LookupEnv("DB_HOST"); exists {
		cfg.Database.Host = databaseHost
	}

	if serverPort, exists := os.LookupEnv("SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	if sslMode, exists := os.LookupEnv("SSL_MODE"); exists {
		cfg.Database.SslMode = sslMode
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
	}

	return cfg
}

func (c *Config) LoadDbUri() string {
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
