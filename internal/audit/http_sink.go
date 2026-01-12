package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// HTTPSink is a sink that sends audit events to a remote HTTP server
type HTTPSink struct {
	url string
	cl  *http.Client
}

// NewHTTPSink creates a new HTTPSink with a default HTTP client
func NewHTTPSink(url string) *HTTPSink {
	// Создаём клиент с таймаутом для защиты от зависаний
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	return &HTTPSink{
		url: url,
		cl:  client,
	}
}

// Send posts the audit event as JSON to the remote server
func (h *HTTPSink) Send(event AuditEvent) error {
	// Сериализуем событие в JSON
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Создаём новый HTTP-запрос
	req, err := http.NewRequest(http.MethodPost, h.url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := h.cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
