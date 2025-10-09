package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			h := newTestHandler()

			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			h.ShortenHandler(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusCreated {
				assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
				body := w.Body.String()
				assert.True(t, strings.HasPrefix(body, "http://localhost:8080/"))

				shortID := strings.TrimPrefix(body, "http://localhost:8080/")
				_, exists := h.repo.Get(shortID)
				assert.True(t, exists, "short ID %q was not saved", shortID)
			}
		})
	}
}
