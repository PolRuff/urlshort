package main

import (
	"net/http"
	"os"

	"github.com/PolRuff/urlshort/internal/config"
	"github.com/PolRuff/urlshort/internal/handler"
	"github.com/PolRuff/urlshort/internal/middleware"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.MustLoad(os.Args[1:])

	var repo repository.Repository
	var err error

	if cfg.DatabaseDsn != "" {
		repo, err = repository.NewSQLRepository(cfg.DatabaseDsn)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed top open sql repository")
		}
	} else if cfg.FileStoragePath != "" {
		// If a file path is provided, create a FileRepository
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize file repository")
		}
	} else {
		// Otherwise, create an in-memory repository
		repo = repository.NewMemoryRepository()
	}

	defer repo.Close()

	h := handler.New(repo, cfg.BaseURL, cfg.SecretKey)

	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.Logger)

	r.Post("/", h.ShortenHandler)
	r.Post("/api/shorten", h.ShortenAPIHandler)
	r.Post("/api/shorten/batch", h.ShortenBatchAPIHandler)
	r.Delete("/api/user/urls", h.DeleteUserUrlsHandler)
	r.Get("/api/user/urls", h.UserUrlsHandler)
	r.Get("/{id}", h.RedirectHandler)
	r.Get("/ping", h.PingHandler)

	log.Debug().Msgf("Server is running on http://%s", cfg.ServerAddr)
	log.Debug().Msgf("Base URL for short links: %s", cfg.BaseURL)

	log.Fatal().Err(http.ListenAndServe(cfg.ServerAddr, r))
}
