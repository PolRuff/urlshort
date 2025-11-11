package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PolRuff/urlshort/internal/repository/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPingHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаём мок-объект Repository
	mockRepo := mock.NewMockRepository(ctrl)

	// Создаём Handler с мок-репозиторием
	h := &Handler{repo: mockRepo}

	// Тест 1: CheckConnection возвращает true -> 200 OK
	t.Run("Connection OK", func(t *testing.T) {
		// Ожидаем, что CheckConnection будет вызван и вернёт true
		mockRepo.EXPECT().CheckConnection().Return(true).Times(1)

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		h.PingHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Тест 2: CheckConnection возвращает false -> 500 Internal Server Error
	t.Run("Connection Failed", func(t *testing.T) {
		// Ожидаем, что CheckConnection будет вызван и вернёт false
		mockRepo.EXPECT().CheckConnection().Return(false).Times(1)

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		h.PingHandler(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		// Проверим, что тело ответа содержит ожидаемое сообщение об ошибке
		assert.Contains(t, w.Body.String(), "Failed connection to database")
	})
}
