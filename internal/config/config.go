package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr           string
	DBURL              string
	CORSAllowedOrigins []string

	JWTSecret string

	ImagesPath string
	ImagesURL  string

	RabbitURL   string
	RabbitQueue string
}

func Load() *Config {
	cfg := &Config{
		HTTPAddr:  getEnv("HTTP_ADDR", ":8080"),
		DBURL:     os.Getenv("DB_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),

		CORSAllowedOrigins: parseCSVEnv(
			"CORS_ALLOWED_ORIGINS",
			[]string{"http://localhost:5173", "http://localhost:3000"},
		),

		ImagesPath: getEnv("IMAGES_PATH", "./data/images"),
		ImagesURL:  getEnv("IMAGES_URL", "/static/images"),

		RabbitURL:   os.Getenv("RABBIT_URL"),
		RabbitQueue: getEnv("RABBIT_QUEUE", "print_queue"),
	}

	log.Println("config loaded:", cfg.HTTPAddr)
	return cfg
}

func (c *Config) Validate() error {
	var missing []string

	if strings.TrimSpace(c.DBURL) == "" {
		missing = append(missing, "DB_URL")
	}
	if strings.TrimSpace(c.JWTSecret) == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if strings.TrimSpace(c.RabbitURL) == "" {
		missing = append(missing, "RABBIT_URL")
	}

	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("%w: %s", errors.New("missing required environment variables"), strings.Join(missing, ", "))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseCSVEnv(key string, def []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}
