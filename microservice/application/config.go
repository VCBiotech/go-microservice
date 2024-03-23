package application

import (
	"fmt"
	"os"
	"strconv"
)

type DatabaseConfig struct {
	Name     string
	Port     uint16
	Username string
	Password string
	Host     string
	SslMode  string
}

type Config struct {
	RedisAddress string
	ServerPort   uint16
	Database     DatabaseConfig
}

func LoadConfig() Config {
	cfg := Config{
		Database: DatabaseConfig{
			Name:     "core",
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

	return cfg
}

func (c *Config) LoadDbUri() string {
	db := c.Database
	// Generate db uri
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
