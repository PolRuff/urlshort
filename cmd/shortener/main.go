package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PolRuff/urlshort/internal/config"
	"github.com/PolRuff/urlshort/internal/handler"
	"github.com/PolRuff/urlshort/internal/logger"
	"github.com/PolRuff/urlshort/internal/middleware"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.MustLoad(os.Args[1:])

	var repo repository.Repository
	var err error

	if cfg.FileStoragePath != "" {
		// If a file path is provided, create a FileRepository
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			log.Fatalf("Failed to initialize file repository: %v", err)
		}
	} else {
		// Otherwise, create an in-memory repository
		repo = repository.NewMemoryRepository()
	}

	h := handler.New(repo, cfg.BaseURL)

	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(logger.Logger)

	r.Post("/", h.ShortenHandler)
	r.Post("/api/shorten", h.ShortenAPIHandler)
	r.Get("/{id}", h.RedirectHandler)

	fmt.Printf("Server is running on http://%s\n", cfg.ServerAddr)
	fmt.Printf("Base URL for short links: %s\n", cfg.BaseURL)

	http.ListenAndServe(cfg.ServerAddr, r)
}
