package config

import (
	"reflect"
	"testing"
)

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_GET_ENV", "value")
	if got := getEnv("TEST_GET_ENV", "default"); got != "value" {
		t.Fatalf("getEnv() = %q, want %q", got, "value")
	}

	t.Setenv("TEST_GET_ENV_EMPTY", "")
	if got := getEnv("TEST_GET_ENV_EMPTY", "default"); got != "default" {
		t.Fatalf("getEnv() with empty value = %q, want %q", got, "default")
	}
}

func TestParseCSVEnv(t *testing.T) {
	t.Setenv("TEST_CSV", "  https://a.example , , https://b.example  ")
	got := parseCSVEnv("TEST_CSV", []string{"fallback"})
	want := []string{"https://a.example", "https://b.example"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseCSVEnv() = %#v, want %#v", got, want)
	}

	t.Setenv("TEST_CSV_EMPTY", " ,  , ")
	got = parseCSVEnv("TEST_CSV_EMPTY", []string{"fallback"})
	want = []string{"fallback"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseCSVEnv() for empty list = %#v, want %#v", got, want)
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("DB_URL", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("IMAGES_PATH", "")
	t.Setenv("IMAGES_URL", "")
	t.Setenv("RABBIT_URL", "")
	t.Setenv("RABBIT_QUEUE", "")

	cfg := Load()

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":8080")
	}
	if cfg.DBURL != "" {
		t.Fatalf("DBURL = %q, want empty string", cfg.DBURL)
	}
	if cfg.JWTSecret != "" {
		t.Fatalf("JWTSecret = %q, want empty string", cfg.JWTSecret)
	}
	wantOrigins := []string{"http://localhost:5173", "http://localhost:3000"}
	if !reflect.DeepEqual(cfg.CORSAllowedOrigins, wantOrigins) {
		t.Fatalf("CORSAllowedOrigins = %#v, want %#v", cfg.CORSAllowedOrigins, wantOrigins)
	}
	if cfg.ImagesPath != "./data/images" {
		t.Fatalf("ImagesPath = %q, want %q", cfg.ImagesPath, "./data/images")
	}
	if cfg.ImagesURL != "/static/images" {
		t.Fatalf("ImagesURL = %q, want %q", cfg.ImagesURL, "/static/images")
	}
	if cfg.RabbitQueue != "print_queue" {
		t.Fatalf("RabbitQueue = %q, want %q", cfg.RabbitQueue, "print_queue")
	}
	if cfg.RabbitURL != "" {
		t.Fatalf("RabbitURL = %q, want empty string", cfg.RabbitURL)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DB_URL", "postgres://db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example, https://b.example")
	t.Setenv("IMAGES_PATH", "/srv/images")
	t.Setenv("IMAGES_URL", "https://cdn.example.com/img")
	t.Setenv("RABBIT_URL", "amqp://guest:guest@localhost:5672/")
	t.Setenv("RABBIT_QUEUE", "jobs")

	cfg := Load()

	if cfg.HTTPAddr != ":9090" || cfg.DBURL != "postgres://db" || cfg.JWTSecret != "secret" {
		t.Fatalf("Load() did not apply scalar overrides: %+v", cfg)
	}
	wantOrigins := []string{"https://a.example", "https://b.example"}
	if !reflect.DeepEqual(cfg.CORSAllowedOrigins, wantOrigins) {
		t.Fatalf("CORSAllowedOrigins = %#v, want %#v", cfg.CORSAllowedOrigins, wantOrigins)
	}
	if cfg.ImagesPath != "/srv/images" || cfg.ImagesURL != "https://cdn.example.com/img" {
		t.Fatalf("image settings mismatch: %+v", cfg)
	}
	if cfg.RabbitURL != "amqp://guest:guest@localhost:5672/" || cfg.RabbitQueue != "jobs" {
		t.Fatalf("rabbit settings mismatch: %+v", cfg)
	}
}

func TestValidate(t *testing.T) {
	t.Run("missing required values", func(t *testing.T) {
		cfg := &Config{}
		if err := cfg.Validate(); err == nil {
			t.Fatal("Validate() error = nil, want error")
		}
	})

	t.Run("all required values provided", func(t *testing.T) {
		cfg := &Config{
			DBURL:     "postgres://db",
			JWTSecret: "secret",
			RabbitURL: "amqp://guest:guest@localhost:5672/",
		}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})
}
