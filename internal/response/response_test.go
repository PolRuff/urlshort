package response

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	message := "Something went wrong"
	statusCode := http.StatusBadRequest

	WriteError(w, message, statusCode)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, statusCode, w.Code)
	assert.JSONEq(t, fmt.Sprintf(`{"error":"%s"}`, message), w.Body.String())
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{
		"key": "value",
	}
	statusCode := http.StatusOK

	WriteJSON(w, data, statusCode)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, statusCode, w.Code)
	assert.JSONEq(t, `{"key":"value"}`, w.Body.String())
}
