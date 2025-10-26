package logger

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// responseData stores information about the HTTP response
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter wraps http.ResponseWriter to capture status and size
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	// write the response using the original http.ResponseWriter
	size, err := lrw.ResponseWriter.Write(b)
	lrw.responseData.size += size // capture the size
	return size, err
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	// write the status code using the original http.ResponseWriter
	lrw.ResponseWriter.WriteHeader(statusCode)
	lrw.responseData.status = statusCode // capture the status code
}

// Logger is a middleware for logging HTTP requests and responses
func Logger(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Initialize the structure to store response data
		responseData := &responseData{
			status: 0,
			size:   0,
		}

		// Create the wrapper
		lrw := loggingResponseWriter{
			ResponseWriter: w, // embed the original http.ResponseWriter
			responseData:   responseData,
		}

		h.ServeHTTP(&lrw, r) // pass the wrapped ResponseWriter to the next handler

		duration := time.Since(start)

		// Log using zerolog at the Info level
		log.Info().
			Str("uri", r.RequestURI).
			Str("method", r.Method).
			Int("status", responseData.status). // get the captured status code
			Int("size", responseData.size).     // get the captured response size
			Dur("duration", duration).
			Msg("request processed")
	}
	return http.HandlerFunc(logFn)
}
