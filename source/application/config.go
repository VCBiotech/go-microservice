package application

import (
	"file-manager/config"
)

// Re-export types for backward compatibility
type AppConfig = config.AppConfig
type StorageConfig = config.StorageConfig
type DatabaseConfig = config.DatabaseConfig

// LoadConfig is a convenience function that calls the config package
func LoadConfig() *AppConfig {
	return config.LoadConfig()
}
