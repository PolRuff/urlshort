package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/PolRuff/urlshort/internal/service"
)

// ShortenAPIHandler handles POST /api/shorten/batch
// Expects JSON: [{"correlation_id":"<строковый идентификатор>","original_url":"<URL для сокращения>"},...]
// Returns JSON: [{"correlation_id":"<строковый идентификатор из объекта запроса>","short_url":"<результирующий сокращённый URL>"},...]  with status 201
func (h *Handler) ShortenBatchAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
		}
		return
	}

	var reqItems []model.BatchShortenRequestItem
	if err := json.Unmarshal(body, &reqItems); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var results []model.BatchShortenResponseItem

	// Проходим по каждому элементу в запросе
	for _, item := range reqItems {
		originalURL := item.OriginalURL
		correlationID := item.CorrelationID

		// Валидируем URL
		if !service.IsValidURL(originalURL) {
			// В зависимости от требований:
			// - можно вернуть 400 для всего запроса
			// - или пропустить этот элемент и добавить ошибку в ответ (требует изменения структуры ответа)
			// Предположим, что валидация обязательна для всех
			http.Error(w, "One or more URLs in the batch are invalid", http.StatusBadRequest)
			return
		}

		// Генерируем shortID (можно вызвать h.generateShortID())
		shortID, err := h.generateShortID()
		if err != nil {
			// Если генерация ID не удалась, возвращаем ошибку для всего запроса
			// В реальности можно было бы обработать ошибки на уровне элемента
			http.Error(w, "Failed to generate short ID", http.StatusInternalServerError)
			return
		}

		// Сохраняем в репозиторий
		err = h.repo.Save(r.Context(), model.URLPair{ShortID: shortID, URL: originalURL})
		if err != nil {
			// Если сохранение не удалось, возвращаем ошибку для всего запроса
			// В реальности можно было бы обработать ошибки на уровне элемента
			http.Error(w, "Failed to save one or more URLs", http.StatusInternalServerError)
			return
		}

		// Формируем результат
		shortenedURL := h.baseURL + "/" + shortID
		results = append(results, model.BatchShortenResponseItem{
			CorrelationID: correlationID,
			ShortURL:      shortenedURL,
		})
	}

	// Устанавливаем заголовки
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Сериализуем и отправляем JSON-ответ
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
