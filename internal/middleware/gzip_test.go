package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockHandler is a simple handler for testing middleware.
// It returns the request body as is, with the Content-Type passed as an argument.
type mockHandler struct {
	contentType string
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", h.contentType)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		log.Printf("Failed to write response body: %v", err)
	}
}

func TestGzipMiddleware(t *testing.T) {
	// Test 1: Request with Content-Encoding: gzip -> body should be decompressed
	t.Run("Decompresses request body with gzip", func(t *testing.T) {
		originalBody := "This is a test body for decompression."

		// Compress the body
		var compressedBuffer bytes.Buffer
		gz := gzip.NewWriter(&compressedBuffer)
		_, err := gz.Write([]byte(originalBody))
		assert.NoError(t, err)
		err = gz.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/", &compressedBuffer)
		req.Header.Set("Content-Encoding", "gzip")
		// Set Content-Type so the mockHandler knows what to return
		req.Header.Set("Content-Type", "text/plain") // unsupported type for response compression

		w := httptest.NewRecorder()

		// Test the middleware
		handler := GzipMiddleware(&mockHandler{contentType: "text/plain"})
		handler.ServeHTTP(w, req)

		// Check that the response contains the original (decompressed) body
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, originalBody, w.Body.String())
		// Response should not be compressed as Content-Type is not supported
		assert.NotEqual(t, "gzip", w.Header().Get("Content-Encoding"))
	})

	// Test 2: Request without gzip -> body remains unchanged
	t.Run("Does not decompress request body without gzip", func(t *testing.T) {
		originalBody := "This is a plain text body."

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalBody))
		// Set Content-Type so the mockHandler knows what to return
		req.Header.Set("Content-Type", "text/plain") // unsupported type for response compression

		w := httptest.NewRecorder()

		// Test the middleware
		handler := GzipMiddleware(&mockHandler{contentType: "text/plain"})
		handler.ServeHTTP(w, req)

		// Check that the response contains the original body
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, originalBody, w.Body.String())
		// Response should not be compressed as Content-Type is not supported
		assert.NotEqual(t, "gzip", w.Header().Get("Content-Encoding"))
	})

	// Test 3: Response with Accept-Encoding: gzip and Content-Type: application/json -> should be compressed
	t.Run("Compresses response body for JSON if client accepts gzip", func(t *testing.T) {
		originalBody := `{"message": "test json response"}`

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalBody))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json") // supported type for response compression

		w := httptest.NewRecorder()

		// Test the middleware
		handler := GzipMiddleware(&mockHandler{contentType: "application/json"})
		handler.ServeHTTP(w, req)

		// Check that the response is compressed
		assert.Equal(t, http.StatusOK, w.Code)
		// Check that Content-Encoding is set
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

		// Decompress the response body and check its content
		gr, err := gzip.NewReader(w.Body)
		assert.NoError(t, err)
		defer gr.Close()

		uncompressedBody, err := io.ReadAll(gr)
		assert.NoError(t, err)

		assert.Equal(t, originalBody, string(uncompressedBody))
	})

	// Test 4: Response with Accept-Encoding: gzip and Content-Type: text/html -> should be compressed
	t.Run("Compresses response body for HTML if client accepts gzip", func(t *testing.T) {
		originalBody := "<html><body><h1>Test HTML</h1></body></html>"

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalBody))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "text/html") // supported type for response compression

		w := httptest.NewRecorder()

		// Test the middleware
		handler := GzipMiddleware(&mockHandler{contentType: "text/html"})
		handler.ServeHTTP(w, req)

		// Check that the response is compressed
		assert.Equal(t, http.StatusOK, w.Code)
		// Check that Content-Encoding is set
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

		// Decompress the response body and check its content
		gr, err := gzip.NewReader(w.Body)
		assert.NoError(t, err)
		defer gr.Close()

		uncompressedBody, err := io.ReadAll(gr)
		assert.NoError(t, err)

		assert.Equal(t, originalBody, string(uncompressedBody))
	})

	// Test 5: Response with Accept-Encoding: gzip and Content-Type: text/plain -> should NOT be compressed
	t.Run("Does not compress response body for unsupported Content-Type", func(t *testing.T) {
		originalBody := "This is a plain text response."

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalBody))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "text/plain") // NOT a supported type for response compression

		w := httptest.NewRecorder()

		// Test the middleware
		handler := GzipMiddleware(&mockHandler{contentType: "text/plain"})
		handler.ServeHTTP(w, req)

		// Check that the response is NOT compressed
		assert.Equal(t, http.StatusOK, w.Code)
		// Check that Content-Encoding is NOT set
		assert.NotEqual(t, "gzip", w.Header().Get("Content-Encoding"))
		// Check that the body is not compressed and matches
		assert.Equal(t, originalBody, w.Body.String())
	})

	// Test 6: Response without Accept-Encoding: gzip -> should NOT be compressed
	t.Run("Does not compress response body if client does not accept gzip", func(t *testing.T) {
		originalBody := `{"message": "test json response without gzip"}`

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalBody))
		// Do NOT set Accept-Encoding
		req.Header.Set("Content-Type", "application/json") // supported type, but client doesn't ask for gzip

		w := httptest.NewRecorder()

		// Test the middleware
		handler := GzipMiddleware(&mockHandler{contentType: "application/json"})
		handler.ServeHTTP(w, req)

		// Check that the response is NOT compressed
		assert.Equal(t, http.StatusOK, w.Code)
		// Check that Content-Encoding is NOT set
		assert.NotEqual(t, "gzip", w.Header().Get("Content-Encoding"))
		// Check that the body matches
		assert.Equal(t, originalBody, w.Body.String())
	})

	// Test 7: Invalid gzipped request body -> returns error
	t.Run("Returns error for invalid gzipped request body", func(t *testing.T) {
		// Pass obviously invalid compressed data
		invalidGzippedData := []byte("this is not gzipped data")

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(invalidGzippedData))
		req.Header.Set("Content-Encoding", "gzip")

		w := httptest.NewRecorder()

		// Test the middleware
		handler := GzipMiddleware(&mockHandler{contentType: "text/plain"})
		handler.ServeHTTP(w, req)

		// Check that an error is returned (e.g., 400)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		// Check that the response body contains a mention of the error
		assert.Contains(t, w.Body.String(), "Failed to read gzipped request body")
	})
}
