package audit_test

import (
	"os"
	"testing"

	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/stretchr/testify/assert"
)

func TestFileSink_Send(t *testing.T) {
	t.Run("should return error when writing to inaccessible file", func(t *testing.T) {
		f := audit.NewFileSink("/proc/self/fd/9999") // Путь, куда нет прав на запись
		event := audit.AuditEvent{
			Timestamp: 12345,
			Action:    "test",
			UserID:    "1",
			URL:       "http://example.com",
		}
		err := f.Send(event)
		assert.Error(t, err, "Ожидалась ошибка при записи в недоступный файл")
	})

	t.Run("should successfully write event to file", func(t *testing.T) {
		// Создаём временный файл для теста
		tempFile := t.TempDir() + "/audit_test.log"
		f := audit.NewFileSink(tempFile)
		event := audit.AuditEvent{
			Timestamp: 12345,
			Action:    "test",
			UserID:    "1",
			URL:       "http://example.com",
		}

		err := f.Send(event)
		assert.NoError(t, err, "Не ожидалось ошибки при записи в временный файл")

		// Читаем файл и проверяем содержимое
		data, err := os.ReadFile(tempFile)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"ts":12345`)
		assert.Contains(t, string(data), `"action":"test"`)
		assert.Contains(t, string(data), `"user_id":"1"`)
		assert.Contains(t, string(data), `"url":"http://example.com"`)
		assert.Contains(t, string(data), "\n") // Проверка новой строки
	})
}
