package config

import (
	"log"
	"os"
)

type Config struct {
	HTTPAddr string
	DBURL    string
}

func Load() *Config {
	cfg := &Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		DBURL:    os.Getenv("DB_URL"),
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
