package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/PolRuff/urlshort/internal/repository/mock"
	"github.com/golang/mock/gomock"
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
			name:         "request body too large",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         strings.Repeat("x", 5*1024), // 5 KB > 4096
			expectedCode: http.StatusRequestEntityTooLarge,
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
				_, exists, _ := h.repo.Get(t.Context(), shortID)
				assert.True(t, exists, "short ID %q was not saved", shortID)
			}
		})
	}
}

func TestShortenHandlerConflictError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)

	h := &Handler{
		repo:    mockRepo,
		baseURL: "http://localhost:8080",
	}

	existingShortID := "existing_id_123"
	originalURL := "https://practicum.yandex.ru/"

	conflictErr := &repository.ConflictError{
		OriginalURL:     originalURL,
		ExistingShortID: existingShortID,
	}

	// Ожидаем, что Save будет вызван с любым URLRecord, и вернёт ConflictError
	mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", false, false).Times(1)
	mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(conflictErr).Times(1)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	h.ShortenHandler(w, req)

	// Проверяем, что возвращён статус 409 Conflict
	assert.Equal(t, http.StatusConflict, w.Code)
	// Проверяем, что Content-Type - text/plain
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	// Проверяем, что тело ответа содержит существующий URL
	expectedBody := "http://localhost:8080/" + existingShortID
	assert.Equal(t, expectedBody, w.Body.String())
}
