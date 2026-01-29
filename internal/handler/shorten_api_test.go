package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/PolRuff/urlshort/internal/model"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/PolRuff/urlshort/internal/repository/mock"
	"github.com/golang/mock/gomock"
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
			name:           "request body too large",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           strings.Repeat("x", 5*1024), // 5 KB > 4096
			expectedStatus: http.StatusRequestEntityTooLarge,
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
				_, exists, _ := h.repo.Get(t.Context(), shortID)
				assert.True(t, exists, "short ID %q was not saved", shortID)
			}
		})
	}
}

func TestShortenAPIHandler_ConflictError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	var auditSinks []audit.Sink
	auditManager := audit.NewManager(auditSinks...)

	h := &Handler{
		repo:         mockRepo,
		auditManager: auditManager,
		baseURL:      "http://localhost:8080",
	}

	existingShortID := "existing_id_123"
	originalURL := "https://practicum.yandex.ru/"

	conflictErr := &repository.ConflictError{
		OriginalURL:     originalURL,
		ExistingShortID: existingShortID,
	}

	// Ожидаем, что Save будет вызван с любым URLRecord, и вернёт *ConflictError
	mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", false, false).Times(1)
	mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(conflictErr).Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url": "`+originalURL+`"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ShortenAPIHandler(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp model.ShortenResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/"+existingShortID, resp.Result)
}
