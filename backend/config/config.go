package config

import (
	"os"
)

type Config struct {
	JWTSecret string
}

var appConfig *Config

func init() {
	appConfig = &Config{
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key"), // Default secret key, should be changed in production
	}
}

func GetConfig() *Config {
	return appConfig
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
