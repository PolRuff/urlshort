package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// FileSink is a sink that writes audit events to a file
type FileSink struct {
	filePath string
	mu       sync.Mutex // для потокобезопасной записи в файл
}

// NewFileSink creates a new FileSink
func NewFileSink(filePath string) *FileSink {
	return &FileSink{
		filePath: filePath,
	}
}

// Send writes the audit event as a JSON line to the file
func (f *FileSink) Send(event AuditEvent) error {
	// 1. Сериализуем событие вне критической секции
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}
	data = append(data, '\n')

	// 2. Критическая секция: только открытие и запись
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit file for writing: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write audit event to file: %w", err)
	}
	return nil
}
