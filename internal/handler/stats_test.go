package handler

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	mock_repo "github.com/PolRuff/urlshort/internal/repository/mock"
	mock_user "github.com/PolRuff/urlshort/internal/service/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestStatsHandler_NoTrustedSubnet(t *testing.T) {
	h := &Handler{
		trustedNet: nil, // Нет доверенной подсети
	}

	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)

	w := httptest.NewRecorder()
	h.StatsHandler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStatsHandler_MissingXRealIPHeader(t *testing.T) {
	_, trustedNet, _ := net.ParseCIDR("192.168.1.0/24")

	h := &Handler{
		trustedNet: trustedNet,
	}

	// Заголовок отсутствует
	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)

	w := httptest.NewRecorder()
	h.StatsHandler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStatsHandler_InvalidIPFormat(t *testing.T) {
	_, trustedNet, _ := net.ParseCIDR("192.168.1.0/24")

	h := &Handler{
		trustedNet: trustedNet,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "invalid-ip-address")

	w := httptest.NewRecorder()
	h.StatsHandler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStatsHandler_IPNotInTrustedSubnet(t *testing.T) {
	_, trustedNet, _ := net.ParseCIDR("192.168.1.0/24")

	h := &Handler{
		trustedNet: trustedNet,
	}

	// IP 10.0.0.1 не входит в 192.168.1.0/24
	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")

	w := httptest.NewRecorder()
	h.StatsHandler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStatsHandler_IPInTrustedSubnet_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repo.NewMockRepository(ctrl)
	mockUser := mock_user.NewMockUserService(ctrl)

	_, trustedNet, _ := net.ParseCIDR("192.168.1.0/24")

	h := &Handler{
		repo:        mockRepo,
		userService: mockUser,
		trustedNet:  trustedNet,
	}

	// IP 192.168.1.100 входит в 192.168.1.0/24
	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")

	mockRepo.EXPECT().CountURLs(req.Context()).Return(100, nil).Times(1)
	mockUser.EXPECT().CountUsers(req.Context()).Return(200, nil).Times(1)

	w := httptest.NewRecorder()
	h.StatsHandler(w, req)

	// Проверяем статус
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем заголовки
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Проверяем тело ответа
	assert.JSONEq(t, w.Body.String(), `{"urls": 100,"users": 200}`)
}

func TestStatsHandler_CountURLsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repo.NewMockRepository(ctrl)

	_, trustedNet, _ := net.ParseCIDR("192.168.1.0/24")

	h := &Handler{
		repo:       mockRepo,
		trustedNet: trustedNet,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")

	mockRepo.EXPECT().CountURLs(req.Context()).Return(0, assert.AnError).Times(1)

	w := httptest.NewRecorder()
	h.StatsHandler(w, req)

	// Ожидаем 500 при ошибке подсчёта
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStatsHandler_CountUsersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repo.NewMockRepository(ctrl)
	mockUser := mock_user.NewMockUserService(ctrl)

	_, trustedNet, _ := net.ParseCIDR("192.168.1.0/24")

	h := &Handler{
		repo:        mockRepo,
		userService: mockUser,
		trustedNet:  trustedNet,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")

	mockRepo.EXPECT().CountURLs(req.Context()).Return(100, nil).Times(1)
	mockUser.EXPECT().CountUsers(req.Context()).Return(0, assert.AnError).Times(1)

	w := httptest.NewRecorder()
	h.StatsHandler(w, req)

	// Ожидаем 500 при ошибке подсчёта
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
