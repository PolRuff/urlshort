package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// HTTPSink is a sink that sends audit events to a remote HTTP server
type HTTPSink struct {
	url string
	cl  *retryablehttp.Client
}

// NewHTTPSink creates a new HTTPSink with a retryable HTTP client
func NewHTTPSink(url string) *HTTPSink {
	// Создаём retryable клиент
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3                          // Максимум 3 попытки
	retryClient.RetryWaitMin = 100 * time.Millisecond // Мин. задержка
	retryClient.RetryWaitMax = 1 * time.Second        // Макс. задержка

	return &HTTPSink{
		url: url,
		cl:  retryClient,
	}
}

// Send posts the audit event as JSON to the remote server with retries
func (h *HTTPSink) Send(event AuditEvent) error {
	// Сериализуем событие в JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Используем retryablehttp.Post
	resp, err := h.cl.Post(h.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send audit event after retries: %w", err)
	}
	resp.Body.Close()

	return nil
}
