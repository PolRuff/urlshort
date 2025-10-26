package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockHandler is a fake handler for testing the middleware
type mockHandler struct {
	status int
	body   string
}

func (h mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(h.status)
	w.Write([]byte(h.body))
}

func TestLogger(t *testing.T) {
	// Create a mock handler that returns 201 and a body
	handler := mockHandler{status: http.StatusCreated, body: "http://localhost:8080/abc123"}

	// Wrap the handler with the Logger middleware
	middleware := Logger(handler)

	// Create a fake request
	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()

	// Call the middleware
	middleware.ServeHTTP(rec, req)

	// Check that the response from the mockHandler passed through the middleware correctly
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "http://localhost:8080/abc123", rec.Body.String())

	// Verify that the middleware does not panic and works correctly.
	// Since we cannot directly check the internal state of loggingResponseWriter,
	// the middleware logic is verified indirectly by checking the correct response.
}
