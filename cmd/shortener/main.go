package main

import (
	"fmt"
	"net/http"

	"github.com/PolRuff/urlshort/internal/config"
	"github.com/PolRuff/urlshort/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	// POST / — shorten a URL and return the short link
	r.Post("/", handler.ShortenHandler)

	// GET /{id} — redirect to the original URL by short ID
	r.Get("/{id}", handler.RedirectHandler)

	fmt.Println("Server is running on http://localhost" + config.ServerPort)
	http.ListenAndServe(config.ServerPort, r)
}
