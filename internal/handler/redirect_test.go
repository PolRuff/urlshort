package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func newRequestWithID(method, id string) *http.Request {
	req := httptest.NewRequest(method, "/"+id, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		id               string
		setup            func(*Handler)
		expectedCode     int
		expectedLocation string
	}{
		{
			name:   "valid redirect",
			method: http.MethodGet,
			id:     "test123",
			setup: func(h *Handler) {
				h.repo.Save(struct{ ShortID, URL string }{"test123", "https://practicum.yandex.ru/"})
			},
			expectedCode:     http.StatusTemporaryRedirect,
			expectedLocation: "https://practicum.yandex.ru/",
		},
		{
			name:         "ID not found",
			method:       http.MethodGet,
			id:           "missing",
			setup:        func(h *Handler) {},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "empty ID",
			method:       http.MethodGet,
			id:           "",
			setup:        func(h *Handler) {},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()
			if tt.setup != nil {
				tt.setup(h)
			}

			req := newRequestWithID(tt.method, tt.id)
			w := httptest.NewRecorder()

			h.RedirectHandler(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedLocation != "" {
				assert.Equal(t, tt.expectedLocation, w.Header().Get("Location"))
			}
		})
	}
}
