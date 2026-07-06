package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	CorsAllowOrigins []string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
	}

	corsAllowOrigins := strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	if len(corsAllowOrigins) == 1 && corsAllowOrigins[0] == "" {
		corsAllowOrigins = []string{}
	}
	for i := range corsAllowOrigins {
		corsAllowOrigins[i] = strings.TrimSpace(corsAllowOrigins[i])
	}

	return Config{
		Port:             getEnvOr("PORT", "8080"),
		CorsAllowOrigins: corsAllowOrigins,
	}
}

func getEnvOr(key string, defaultValue string) string {
	result := os.Getenv(key)
	if result == "" {
		return defaultValue
	}
	return result
}
