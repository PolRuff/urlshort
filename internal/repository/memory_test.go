package repository

import (
	"testing"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository_SaveAndGet(t *testing.T) {
	repo := NewMemoryRepository()
	defer func() {
		repo.CheckConnection(t.Context())
		repo.Close()
	}()

	maxID, err := repo.GetMaxID()
	assert.NoError(t, err)
	assert.Equal(t, maxID, uint64(0))

	// Подготовим тестовые данные
	pair := model.URLPair{
		ShortID: "abc123",
		URL:     "https://practicum.yandex.ru/",
	}

	// Проверим, что URL не существует до сохранения
	_, exists := repo.Get(t.Context(), pair.ShortID)
	assert.False(t, exists, "URL should not exist before saving")

	// Сохраним пару
	err = repo.Save(t.Context(), pair)
	assert.NoError(t, err, "Save should not return an error")

	maxID, err = repo.GetMaxID()
	assert.NoError(t, err)
	assert.Equal(t, maxID, uint64(1))

	// Проверим, что URL теперь существует и совпадает
	storedURL, exists := repo.Get(t.Context(), pair.ShortID)
	assert.True(t, exists, "URL should exist after saving")
	assert.Equal(t, pair.URL, storedURL, "Stored URL should match the saved one")
}
