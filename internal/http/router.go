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

	// ---------- CORS ----------
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

	// ---------- Base middleware ----------
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// ----------  Repositories ----------
	userRepo := postgres.NewUserRepository(db)
	bookRepo := postgres.NewBookRepository(db)
	authorRepo := postgres.NewAuthorRepository(db)
	workRepo := postgres.NewWorkRepository(db)
	bookWorksRepo := postgres.NewBookWorksRepository(db)
	workAuthorsRepo := postgres.NewWorkAuthorsRepository(db)
	publisherRepo := postgres.NewPublisherRepository(db)
	sequenceRepo := postgres.NewSequenceRepository(db)

	// ---------- Services ----------
	jwtManager := &service.JWTManager{
		Secret: []byte(cfg.JWTSecret),
		TTL:    24 * time.Hour,
	}

	authService := service.NewAuthService(userRepo, jwtManager)
	barcodeService := service.NewBarcodeService(sequenceRepo)

	bookService := service.NewBookService(bookRepo, bookWorksRepo, workRepo, workAuthorsRepo, barcodeService)
	authorService := service.NewAuthorService(authorRepo)
	workService := service.NewWorkService(workRepo)
	publisherService := service.NewPublisherService(publisherRepo)

	// ---------- Handlers ----------
	authHandler := &handler.AuthHandler{Service: authService}
	booksPublicHandler := &handler.BooksPublicHandler{Service: bookService}
	booksInternalHandler := &handler.BooksInternalHandler{Service: bookService}
	booksAdminHandler := &handler.BooksAdminHandler{Service: bookService}

	authorHandler := handler.NewAuthorHandler(authorService)
	workHandler := handler.NewWorkHandler(workService)
	publisherHandler := handler.NewPublisherHandler(publisherService)

	// ---------- Public routes ----------
	r.Get("/health", handler.Health)
	r.Post("/auth/login", authHandler.Login)

	// ---------- Protected routes ----------
	r.Route("/", func(r chi.Router) {
		r.Use(httpMiddleware.Auth(jwtManager))

		// ---------- Books ----------
		r.Route("/books", func(r chi.Router) {
			r.Route("/public", func(r chi.Router) {
				r.Get("/", booksPublicHandler.List)
				r.Get("/{id}", booksPublicHandler.GetByID)
			})

			r.Route("/internal", func(r chi.Router) {
				r.Get("/", booksInternalHandler.List)
				r.Get("/{id}", booksInternalHandler.GetByID)
				//r.Get("/{id}/barcode", booksInternalHandler.Barcode)
			})
		})

		// ---------- admin ----------
		r.Route("/admin", func(r chi.Router) {
			r.Route("/books", func(r chi.Router) {
				r.Post("/", booksAdminHandler.Create)
				r.Put("/{id}", booksAdminHandler.Update)
				r.Delete("/{id}", booksAdminHandler.Delete)
			})
		})

		// ---------- reference ----------
		r.Route("/reference", func(r chi.Router) {
			r.Get("/authors", authorHandler.GetAll)
			r.Get("/works", workHandler.GetAll)
			r.Get("/publishers", publisherHandler.GetAll)
		})
	})

	return r
}
