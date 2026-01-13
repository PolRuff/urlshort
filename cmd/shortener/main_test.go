package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/PolRuff/urlshort/internal/handler"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/go-chi/chi/v5"
)

func BenchmarkShorten(b *testing.B) {
	for b.Loop() {
		// Настройка окружения для бенчмарка
		b.StopTimer()

		repo := repository.NewMemoryRepository()
		defer repo.Close()

		var auditSinks []audit.Sink
		auditManager := audit.NewManager(auditSinks...)
		h := handler.New(repo, "http://localhost:8080", "test-key", auditManager)

		r := chi.NewRouter()
		r.Post("/", h.ShortenHandler)
		r.Get("/{id}", h.RedirectHandler)

		ts := httptest.NewServer(r)
		defer ts.Close()

		tsURL := ts.URL + "/"
		originalURL := strings.NewReader("https://practicum.yandex.ru/")

		// Запуск бенчмарка
		b.StartTimer() // возобновляем таймер

		resp, err := http.Post(tsURL, "text/plain", originalURL)
		if err != nil {
			b.Fatalf("Failed to shorten URL: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkRedirect(b *testing.B) {
	// Настройка окружения для бенчмарка
	repo := repository.NewMemoryRepository()
	defer repo.Close()

	var auditSinks []audit.Sink
	auditManager := audit.NewManager(auditSinks...)
	h := handler.New(repo, "http://localhost:8080", "test-key", auditManager)

	r := chi.NewRouter()
	r.Post("/", h.ShortenHandler)
	r.Get("/{id}", h.RedirectHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	tsURL := ts.URL + "/"
	originalURL := strings.NewReader("https://practicum.yandex.ru/")

	resp, err := http.Post(tsURL, "text/plain", originalURL)
	if err != nil {
		b.Fatalf("Failed to shorten URL: %v", err)
	}

	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		b.Fatalf("Failed to read body: %v", err)
	}

	shortenedURL := strings.Replace(string(data), "http://localhost:8080", ts.URL, 1)

	// Запуск бенчмарка
	for b.Loop() {
		resp, err = http.Get(shortenedURL)
		if err != nil {
			b.Fatalf("Failed to redirect: %v", err)
		}
		resp.Body.Close()
	}
}
