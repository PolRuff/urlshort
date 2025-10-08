package main

import (
	"fmt"
	"net/http"

	"github.com/PolRuff/urlshort/internal/config"
	"github.com/PolRuff/urlshort/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Load configuration from command-line flags
	cfg := config.MustLoad()

	// Configure the base URL used to generate short links
	handler.SetBaseURL(cfg.BaseURL)

	r := chi.NewRouter()

	// POST / — shorten a URL and return the short link
	r.Post("/", handler.ShortenHandler)

	// GET /{id} — redirect to the original URL by short ID
	r.Get("/{id}", handler.RedirectHandler)

	fmt.Printf("Server is running on http://%s\n", cfg.ServerAddr)
	fmt.Printf("Base URL for short links: %s\n", cfg.BaseURL)

	// Start HTTP server
	http.ListenAndServe(cfg.ServerAddr, r)
}
