package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestShortenAPIHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		expectedStatus int
		expectedJSON   bool
	}{
		{
			name:           "valid JSON request",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url": "https://practicum.yandex.ru"}`,
			expectedStatus: http.StatusCreated,
			expectedJSON:   true,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":}`,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   false,
		},
		{
			name:           "invalid URL in JSON",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url": "not-a-url"}`,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   false,
		},
		{
			name:           "non-JSON Content-Type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           `{"url": "https://example.com"}`,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()

			req := httptest.NewRequest(tt.method, "/api/shorten", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			h.ShortenAPIHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedJSON {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				var resp model.ShortenResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.True(t, strings.HasPrefix(resp.Result, "http://localhost:8080/"))

				shortID := strings.TrimPrefix(resp.Result, "http://localhost:8080/")
				_, exists := h.repo.Get(shortID)
				assert.True(t, exists, "short ID %q was not saved", shortID)
			}
		})
	}
}
