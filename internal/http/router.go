package http

import (
	"elibrary/internal/http/handler"
	"elibrary/internal/repository/postgres"
	"elibrary/internal/service"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	bookRepo := postgres.NewBookRepository(db)
	bookService := service.NewBookService(bookRepo)
	bookHandler := &handler.BookHandler{Service: bookService}

	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/books", bookHandler.Create)
	mux.HandleFunc("/books/", bookHandler.Get)

	return mux
}
