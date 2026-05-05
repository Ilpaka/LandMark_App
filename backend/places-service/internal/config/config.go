package config

import "os"

type Config struct {
	HTTPAddr             string
	DatabaseURL          string
	InternalAPIKey       string
	ModerationServiceURL string
}

func Load() Config {
	return Config{
		HTTPAddr:             getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		InternalAPIKey:       getenv("INTERNAL_API_KEY", "dev-internal-key-change-me"),
		ModerationServiceURL: getenv("MODERATION_SERVICE_URL", "http://moderation-service:8080"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
