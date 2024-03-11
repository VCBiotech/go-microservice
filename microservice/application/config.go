package application

import (
	"os"
	"strconv"
)

type DatabaseConfig struct {
	Name     string
	Port     uint16
	Username string
	Password string
	Host     string
}

type Config struct {
	RedisAddress string
	ServerPort   uint16
	Database     DatabaseConfig
}

func LoadConfig() Config {
	cfg := Config{
		Database: DatabaseConfig{
			Name:     "test",
			Port:     5432,
			Username: "postgres",
			Password: "password",
			Host:     "localhost",
		},
		RedisAddress: "localhost:6379",
		ServerPort:   3000,
	}

	if databaseName, exists := os.LookupEnv("DB_NAME"); exists {
		cfg.Database.Name = databaseName
	}

	if databaseHost, exists := os.LookupEnv("DB_HOST"); exists {
		cfg.Database.Host = databaseHost
	}

	if redisAddr, exists := os.LookupEnv("REDIS_ADDR"); exists {
		cfg.RedisAddress = redisAddr
	}

	if serverPort, exists := os.LookupEnv("SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	return cfg
}
