package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/PolRuff/urlshort/internal/config"
	"github.com/PolRuff/urlshort/internal/handler"
	"github.com/PolRuff/urlshort/internal/logger"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.MustLoad(os.Args[1:])

	repo := repository.NewMemoryRepository()
	h := handler.New(repo, cfg.BaseURL)

	r := chi.NewRouter()

	r.Use(logger.Logger)

	r.Post("/", h.ShortenHandler)
	r.Get("/{id}", h.RedirectHandler)

	fmt.Printf("Server is running on http://%s\n", cfg.ServerAddr)
	fmt.Printf("Base URL for short links: %s\n", cfg.BaseURL)

	http.ListenAndServe(cfg.ServerAddr, r)
}
