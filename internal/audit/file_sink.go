package audit

import (
	"encoding/json"
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
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Сериализуем событие в JSON
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Добавляем символ новой строки и записываем в файл
	data = append(data, '\n')
	_, err = file.Write(data)
	return err
}
