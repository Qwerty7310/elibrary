package http

import (
	"elibrary/internal/http/handler"
	"elibrary/internal/repository/postgres"
	"elibrary/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	bookRepo := postgres.NewBookRepository(db)
	bookService := service.NewBookService(bookRepo)
	barcodeService := service.NewBarcodeService()

	bookHandler := &handler.BookHandler{
		Service:        bookService,
		BarcodeService: barcodeService,
	}

	scanHandler := &handler.ScanHandler{
		BookService: bookService,
	}

	r.Get("/health", handler.Health)
	r.Get("/scan/{value}", scanHandler.Scan)

	r.Route("/books", func(r chi.Router) {
		r.Post("/", bookHandler.Create)
		r.Get("/search", bookHandler.Search)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", bookHandler.Get)
			r.Get("/barcode", bookHandler.Barcode)
		})
	})

	return r
}
