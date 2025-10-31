package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// compressWriter wraps http.ResponseWriter and allows transparent compression of data
// sent to the client, setting appropriate HTTP headers. It checks Content-Type before
// deciding whether to compress the response.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
	// Flag indicating if compression has started
	started bool
}

// newCompressWriter creates a new compressWriter instance.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w: w,
		// gzip.Writer is created later if compression is needed
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write compresses data if Content-Type supports compression.
// It calls WriteHeader (if not already called) and sets Content-Encoding.
func (c *compressWriter) Write(p []byte) (int, error) {
	// If WriteHeader hasn't been called, call it with 200 OK
	if !c.started {
		c.WriteHeader(http.StatusOK)
	}

	// If compression started (i.e., zw was created and Content-Encoding was set)
	if c.zw != nil {
		return c.zw.Write(p)
	}

	// Otherwise, write to the original ResponseWriter
	return c.w.Write(p)
}

// WriteHeader checks Content-Type and sets Content-Encoding: gzip if needed.
// Then it calls WriteHeader on the original ResponseWriter.
func (c *compressWriter) WriteHeader(statusCode int) {
	if c.started {
		// Already started, do nothing
		return
	}
	c.started = true

	// Check Content-Type
	contentType := c.Header().Get("Content-Type")
	if isSupportedContentEncoding(contentType) {
		// Create gzip.Writer and set headers
		c.zw = gzip.NewWriter(c.w)
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.Header().Del("Content-Length") // Remove as length changes after compression
	}

	// Always call WriteHeader on the original ResponseWriter
	c.w.WriteHeader(statusCode)
}

// Close closes the gzip.Writer and flushes any remaining data from the buffer.
// Call only if zw is not nil.
func (c *compressWriter) Close() error {
	if c.zw != nil {
		return c.zw.Close()
	}
	return nil
}

// compressReader wraps io.ReadCloser and allows transparent decompression of data
// received from the client.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader creates a new compressReader instance.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// isSupportedContentEncoding checks if we support compression for the given Content-Type.
func isSupportedContentEncoding(contentType string) bool {
	// Extract the main part of the Content-Type, ignoring parameters (e.g., charset)
	ct := strings.Split(contentType, ";")[0]
	ct = strings.TrimSpace(strings.ToLower(ct))

	return ct == "application/json" || ct == "text/html"
}

// GzipMiddleware returns an http.Handler that compresses responses and/or decompresses requests.
// It supports Content-Encoding: gzip for requests and Accept-Encoding: gzip for responses.
// It compresses content only for types application/json and text/html.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Content-Encoding: gzip and decompress request body if needed.
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to read gzipped request body: %v", err), http.StatusBadRequest)
				return
			}
			// Replace request body with the new one
			r.Body = cr
			defer cr.Close()
		}

		// Check if the client supports receiving compressed data in gzip format.
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			// Wrap the original http.ResponseWriter with one that supports compression.
			cw := newCompressWriter(w)
			// Pass the new ResponseWriter to the next handler.
			next.ServeHTTP(cw, r)
			// Ensure all compressed data is sent to the client after the middleware finishes.
			// This is crucial: Close must be called *after* next.ServeHTTP completes.
			cw.Close()
			return
		}

		// Client does not support gzip, pass through as is.
		next.ServeHTTP(w, r)
	})
}
