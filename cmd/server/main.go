package main

import (
	"context"
	"elibrary/internal/config"
	httpTransport "elibrary/internal/http"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	db, err := pgxpool.New(context.Background(), cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := httpTransport.NewRouter(db)

	log.Println("starting server on", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, router))
}
