package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	OSRMUrl     string
}

func Load() Config {
	return Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		OSRMUrl:     getenv("OSRM_URL", "http://router.project-osrm.org"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
