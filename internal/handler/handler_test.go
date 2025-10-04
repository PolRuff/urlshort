package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetStorage() {
	shortURLs = make(map[string]string)
}

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		contentType  string
		body         string
		expectedCode int
	}{
		{
			name:         "valid URL",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "https://practicum.yandex.ru/",
			expectedCode: http.StatusCreated,
		},
		{
			name:         "invalid Content-Type",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         "https://example.com",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid URL",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "not-a-url",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty body",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "GET instead of POST",
			method:       http.MethodGet,
			contentType:  "text/plain",
			body:         "https://example.com",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStorage()

			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			ShortenHandler(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusCreated {
				assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
				body := w.Body.String()
				assert.True(t, strings.HasPrefix(body, "http://localhost:8080/"),
					"expected short URL to start with http://localhost:8080/, got %q", body)

				shortID := strings.TrimPrefix(body, "http://localhost:8080/")
				_, exists := shortURLs[shortID]
				assert.True(t, exists, "short ID %q was not saved to storage", shortID)
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		setup            func()
		expectedCode     int
		expectedLocation string
	}{
		{
			name:   "valid redirect",
			method: http.MethodGet,
			path:   "/test123",
			setup: func() {
				shortURLs["test123"] = "https://practicum.yandex.ru/"
			},
			expectedCode:     http.StatusTemporaryRedirect,
			expectedLocation: "https://practicum.yandex.ru/",
		},
		{
			name:         "ID not found",
			method:       http.MethodGet,
			path:         "/missing",
			setup:        func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty ID (root path)",
			method:       http.MethodGet,
			path:         "/",
			setup:        func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "POST instead of GET",
			method:       http.MethodPost,
			path:         "/test123",
			setup:        func() {},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStorage()
			if tt.setup != nil {
				tt.setup()
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			RedirectHandler(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			assert.Equal(t, tt.expectedLocation, w.Header().Get("Location"))
		})
	}
}
