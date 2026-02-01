package http

import (
	"elibrary/internal/config"
	"elibrary/internal/http/auth"
	"elibrary/internal/http/handler"
	"elibrary/internal/repository/postgres"
	"elibrary/internal/service"
	"elibrary/internal/storage/local"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-chi/cors"

	httpMiddleware "elibrary/internal/http/middleware"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// ---------- Static images ----------
	if cfg.ImagesURL != "" {
		r.Handle(cfg.ImagesURL+"/*",
			http.StripPrefix(cfg.ImagesURL, http.FileServer(http.Dir(cfg.ImagesPath))))
	}

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
	locationRepo := postgres.NewLocationRepository(db)
	sequenceRepo := postgres.NewSequenceRepository(db)

	imageStorage := local.NewImageStorage(cfg.ImagesPath, cfg.ImagesURL)

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
	locationService := service.NewLocationService(locationRepo, barcodeService)
	userService := service.NewUserService(userRepo)
	imageService := service.NewImageService(imageStorage)

	// ---------- Handlers ----------
	authHandler := handler.NewAuthHandler(authService)
	bookPublicHandler := handler.NewBookPublicHandler(bookService)
	bookInternalHandler := handler.NewBookInternalHandler(bookService)
	bookAdminHandler := handler.NewBookAdminHandler(bookService)

	authorHandler := handler.NewAuthorHandler(authorService)
	workHandler := handler.NewWorkHandler(workService)
	publisherHandler := handler.NewPublisherHandler(publisherService)
	locationHandler := handler.NewLocationHandler(locationService)
	userHandler := handler.NewUserHandler(userService)
	imageHandler := handler.NewImageHandler(imageService)

	// ---------- Public routes ----------
	r.Get("/health", handler.Health)
	r.Post("/auth/login", authHandler.Login)

	// ---------- Protected routes ----------
	r.Route("/", func(r chi.Router) {
		r.Use(httpMiddleware.Auth(jwtManager, userRepo))

		// ---------- books ----------
		r.Route("/books", func(r chi.Router) {
			r.Route("/public", func(r chi.Router) {
				r.Get("/", bookPublicHandler.List)
				r.Get("/{id}", bookPublicHandler.GetByID)
			})

			r.Route("/internal", func(r chi.Router) {
				r.Get("/", bookInternalHandler.List)
				r.Get("/{id}", bookInternalHandler.GetByID)
				//r.Get("/{id}/barcode", booksInternalHandler.Barcode)
			})
		})

		// ---------- works ----------
		r.Route("/works", func(r chi.Router) {
			r.Get("/{id}", workHandler.GetByID)
		})

		// ---------- authors ----------
		r.Route("/authors", func(r chi.Router) {
			r.Get("/{id}", authorHandler.GetByID)
		})

		// ---------- publishers ----------
		r.Route("/publishers", func(r chi.Router) {
			r.Get("/{id}", publisherHandler.GetByID)
		})

		// ---------- location ----------
		r.Route("/locations", func(r chi.Router) {
			r.Get("/type/{type}", locationHandler.GetByType)
			r.Get("/{id}", locationHandler.GetByID)
			r.Get("/child/{id}/{type}", locationHandler.GetByParentID)
		})

		// ---------- admin ----------
		r.Route("/admin", func(r chi.Router) {
			r.Use(httpMiddleware.RequireRole(auth.RoleAdmin))

			r.Post("/{entity}/{id}/image", imageHandler.Upload)

			r.Route("/users", func(r chi.Router) {
				r.Get("/{id}", userHandler.GetByID)
				r.Post("/", userHandler.Create)
				r.Put("/{id}", userHandler.Update)
				r.Delete("/{id}", userHandler.Delete)
			})

			r.Route("/books", func(r chi.Router) {
				r.Post("/", bookAdminHandler.Create)
				r.Put("/{id}", bookAdminHandler.Update)
			})

			r.Route("/works", func(r chi.Router) {
				r.Post("/", workHandler.Create)
				r.Put("/{id}", workHandler.Update)
				r.Delete("/{id}", workHandler.Delete)
			})

			r.Route("/authors", func(r chi.Router) {
				r.Post("/", authorHandler.Create)
				r.Put("/{id}", authorHandler.Update)
				r.Delete("/{id}", authorHandler.Delete)
			})

			r.Route("/publishers", func(r chi.Router) {
				r.Post("/", publisherHandler.Create)
				r.Put("/{id}", publisherHandler.Update)
				r.Delete("/{id}", publisherHandler.Delete)
			})

			r.Route("/locations", func(r chi.Router) {
				r.Post("/", locationHandler.Create)
				r.Put("/{id}", locationHandler.Update)
				r.Delete("/{id}", locationHandler.Delete)
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
