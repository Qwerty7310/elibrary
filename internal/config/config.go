package config

import (
	"log"
	"os"
)

type Config struct {
	HTTPAddr string
}

func Load() *Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	cfg := &Config{
		HTTPAddr: addr,
	}

	log.Println("config loaded:", cfg.HTTPAddr)
	return cfg
}
