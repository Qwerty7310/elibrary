package http

import (
	"elibrary/internal/config"
	"elibrary/internal/http/handler"
	"elibrary/internal/repository/postgres"
	"elibrary/internal/service"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-chi/cors"

	httpMiddleware "elibrary/internal/http/middleware"
)

const (
	prefix = 200
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	allowCredentials := true
	for _, origin := range cfg.CORSAllowedOrigins {
		if origin == "*" {
			allowCredentials = false
			break
		}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: allowCredentials,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	userRepo := postgres.NewUserRepository(db)

	jwtManager := &service.JWTManager{
		Secret: []byte(cfg.JWTSecret),
		TTL:    24 * time.Hour,
	}

	authService := service.NewAuthService(userRepo, jwtManager)
	authHandler := &handler.AuthHandler{
		Service: authService,
	}

	bookRepo := postgres.NewBookRepository(db)
	sequenceRepo := postgres.NewSequenceRepository(db)

	barcodeService := service.NewBarcodeService(sequenceRepo, prefix)
	bookService := service.NewBookService(bookRepo, barcodeService)

	bookHandler := &handler.BookHandler{
		Service:        bookService,
		BarcodeService: barcodeService,
	}

	scanHandler := &handler.ScanHandler{
		BookService: bookService,
	}

	r.Get("/health", handler.Health)
	r.Get("/scan/{value}", scanHandler.Scan)

	r.Post("/auth/login", authHandler.Login)

	r.Route("/books", func(r chi.Router) {
		r.Use(httpMiddleware.Auth(jwtManager))

		r.Post("/", bookHandler.Create)
		r.Get("/search", bookHandler.Search)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", bookHandler.Get)
			r.Get("/barcode", bookHandler.Barcode)
		})
	})

	return r
}
