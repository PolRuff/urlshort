package audit_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/stretchr/testify/assert"
)

func TestHTTPSink_Send(t *testing.T) {
	t.Run("should successfully send event to remote server", func(t *testing.T) {
		// Создаём mock-сервер
		receivedEvent := make(chan audit.AuditEvent, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)

			var event audit.AuditEvent
			err = json.Unmarshal(body, &event)
			assert.NoError(t, err)

			// Отправляем полученное событие в канал для проверки
			receivedEvent <- event
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Создаём HTTPSink, указывая на URL mock-сервера
		sink := audit.NewHTTPSink(server.URL)
		eventToSend := audit.AuditEvent{
			Timestamp: 12345,
			Action:    "shorten",
			UserID:    "100",
			URL:       "https://example.com/long-url",
		}

		// Отправляем событие
		err := sink.Send(eventToSend)
		assert.NoError(t, err)

		// Проверяем, что сервер получил правильное событие
		select {
		case received := <-receivedEvent:
			assert.Equal(t, eventToSend, received)
		default:
			t.Fatal("Mock server did not receive the event")
		}
	})

	t.Run("should return error if remote server is unreachable", func(t *testing.T) {
		// Используем несуществующий URL
		sink := audit.NewHTTPSink("http://nonexistent.local:9999")
		event := audit.AuditEvent{}

		err := sink.Send(event)
		assert.Error(t, err, "Ожидалась ошибка при отправке на несуществующий сервер")
	})
}
