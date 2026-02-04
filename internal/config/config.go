package config

import (
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
		JWTSecret: getEnv("JWT_SECRET", "dev-secret"),

		CORSAllowedOrigins: parseCSVEnv(
			"CORS_ALLOWED_ORIGINS",
			[]string{"http://localhost:5173", "http://localhost:3000"},
		),

		ImagesPath: getEnv("IMAGES_PATH", "./data/images"),
		ImagesURL:  getEnv("IMAGES_URL", "/static/images"),

		RabbitURL:   getEnv("RABBIT_URL", "amqp://printer_user:strong_pass@rabbitmq:5672/"),
		RabbitQueue: getEnv("RABBIT_QUEUE", "print_queue"),
	}

	log.Println("config loaded:", cfg.HTTPAddr)
	return cfg
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
