package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRepository(t *testing.T) {
	// Create a temporary file for the test
	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "test_storage.json")

	// Ensure the file does not exist before the test (although TempDir should be empty)
	os.Remove(tempFilePath)

	// Test 1: Creating a new repository with a non-existent file -> empty repository
	t.Run("NewFileRepository with non-existent file", func(t *testing.T) {
		repo, err := NewFileRepository(tempFilePath)
		require.NoError(t, err)
		assert.NotNil(t, repo)

		// Check that the repository is empty
		_, exists := repo.Get("nonexistent")
		assert.False(t, exists)
	})

	// Prepare initial data for the next test
	initialData := []model.URLRecord{
		{UUID: "1", ShortURL: "abc123", OriginalURL: "http://example.com"},
		{UUID: "2", ShortURL: "def456", OriginalURL: "http://practicum.yandex.ru"},
	}

	initialDataBytes, err := json.MarshalIndent(initialData, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(tempFilePath, initialDataBytes, 0644)
	require.NoError(t, err)

	// Test 2: Creating a repository with an existing file -> data is loaded
	t.Run("NewFileRepository loads data from existing file", func(t *testing.T) {
		// Create a new repository; it should load data from the file
		repo, err := NewFileRepository(tempFilePath)
		require.NoError(t, err)
		assert.NotNil(t, repo)

		// Check that the data is loaded
		url, exists := repo.Get("abc123")
		assert.True(t, exists)
		assert.Equal(t, "http://example.com", url)

		url, exists = repo.Get("def456")
		assert.True(t, exists)
		assert.Equal(t, "http://practicum.yandex.ru", url)

		// Check that a non-existent ID is not returned
		_, exists = repo.Get("nonexistent")
		assert.False(t, exists)

		// Check that nextUUID is set correctly (max + 1)
		// This is harder to check directly, but can be inferred via Save
	})

	// Test 3: Save adds a new record and updates the file
	t.Run("Save adds new record and updates file", func(t *testing.T) {
		// Use the same file as in the previous test
		repo, err := NewFileRepository(tempFilePath)
		require.NoError(t, err)

		// Save a new pair
		newPair := model.URLPair{
			ShortID: "ghi789",
			URL:     "http://newsite.com",
		}
		err = repo.Save(newPair)
		require.NoError(t, err)

		// Check that it is available in the repository
		url, exists := repo.Get("ghi789")
		assert.True(t, exists)
		assert.Equal(t, "http://newsite.com", url)

		// Check that the file is updated
		fileData, err := os.ReadFile(tempFilePath)
		require.NoError(t, err)

		var recordsFromFile []model.URLRecord
		err = json.Unmarshal(fileData, &recordsFromFile)
		require.NoError(t, err)

		expectedRecords := []model.URLRecord{
			{UUID: "1", ShortURL: "abc123", OriginalURL: "http://example.com"},
			{UUID: "2", ShortURL: "def456", OriginalURL: "http://practicum.yandex.ru"},
			// The new item should have UUID 3, as the max was 2, so nextUUID = 3
			{UUID: "3", ShortURL: "ghi789", OriginalURL: "http://newsite.com"},
		}

		// Check that the file contains three records
		assert.Len(t, recordsFromFile, 3)
		// Check that the records match (order may differ, but content should be the same)
		assert.ElementsMatch(t, expectedRecords, recordsFromFile)
	})

	// Test 4: Saving multiple records sequentially
	t.Run("Save multiple records sequentially", func(t *testing.T) {
		// Use the same file as in the previous test
		repo, err := NewFileRepository(tempFilePath)
		require.NoError(t, err)

		// Save two more pairs
		pair1 := model.URLPair{ShortID: "jkl012", URL: "http://site1.com"}
		pair2 := model.URLPair{ShortID: "mno345", URL: "http://site2.com"}

		err = repo.Save(pair1)
		require.NoError(t, err)
		err = repo.Save(pair2)
		require.NoError(t, err)

		// Check that they are available
		url, exists := repo.Get("jkl012")
		assert.True(t, exists)
		assert.Equal(t, "http://site1.com", url)

		url, exists = repo.Get("mno345")
		assert.True(t, exists)
		assert.Equal(t, "http://site2.com", url)

		// Check that the file is updated
		fileData, err := os.ReadFile(tempFilePath)
		require.NoError(t, err)

		var recordsFromFile []model.URLRecord
		err = json.Unmarshal(fileData, &recordsFromFile)
		require.NoError(t, err)

		expectedRecords := []model.URLRecord{
			{UUID: "1", ShortURL: "abc123", OriginalURL: "http://example.com"},
			{UUID: "2", ShortURL: "def456", OriginalURL: "http://practicum.yandex.ru"},
			{UUID: "3", ShortURL: "ghi789", OriginalURL: "http://newsite.com"},
			{UUID: "4", ShortURL: "jkl012", OriginalURL: "http://site1.com"}, // nextUUID was 3, now 4
			{UUID: "5", ShortURL: "mno345", OriginalURL: "http://site2.com"}, // nextUUID was 4, now 5
		}

		assert.Len(t, recordsFromFile, 5)
		assert.ElementsMatch(t, expectedRecords, recordsFromFile)
	})

	// Test 5: Loading from a file with existing records with a high UUID
	t.Run("Loads and continues UUID sequence from high number", func(t *testing.T) {
		// Create a temporary file for this test
		tempFilePathHighUUID := filepath.Join(tempDir, "high_uuid_test.json")
		defer os.Remove(tempFilePathHighUUID)

		// Prepare data with a high UUID
		highUUIDData := []model.URLRecord{
			{UUID: "1", ShortURL: "abc123", OriginalURL: "http://example.com"},
			{UUID: "999", ShortURL: "def456", OriginalURL: "http://practicum.yandex.ru"},
		}
		highUUIDDataBytes, err := json.MarshalIndent(highUUIDData, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(tempFilePathHighUUID, highUUIDDataBytes, 0644)
		require.NoError(t, err)

		// Load the repository
		repo, err := NewFileRepository(tempFilePathHighUUID)
		require.NoError(t, err)

		// Save a new record
		newPair := model.URLPair{ShortID: "new123", URL: "http://new.com"}
		err = repo.Save(newPair)
		require.NoError(t, err)

		// Check that the new UUID is 1000
		fileData, err := os.ReadFile(tempFilePathHighUUID)
		require.NoError(t, err)

		var recordsFromFile []model.URLRecord
		err = json.Unmarshal(fileData, &recordsFromFile)
		require.NoError(t, err)

		var foundNewRecord bool
		for _, record := range recordsFromFile {
			if record.ShortURL == "new123" {
				assert.Equal(t, "1000", record.UUID)
				foundNewRecord = true
				break
			}
		}
		assert.True(t, foundNewRecord, "New record with UUID 1000 was not found in the file")
	})

	// Test 6: Error when loading from a file with invalid JSON
	t.Run("Fails to load from file with invalid JSON", func(t *testing.T) {
		// Create a temporary file for this test
		tempFilePathInvalid := filepath.Join(tempDir, "invalid_test.json")
		defer os.Remove(tempFilePathInvalid)

		// Write deliberately invalid JSON
		invalidData := []byte(`[{"uuid":"1","short_url":"abc123","original_url":"http://example.com"},`)
		err := os.WriteFile(tempFilePathInvalid, invalidData, 0644)
		require.NoError(t, err)

		// Try to create a repository -> expect an error
		repo, err := NewFileRepository(tempFilePathInvalid)
		// Expect an error when parsing the invalid JSON
		assert.Error(t, err)
		assert.Nil(t, repo)
	})

	// Test 7: Get returns false for a non-existent ID
	t.Run("Get returns false for non-existent ID", func(t *testing.T) {
		// Use a file with data
		repo, err := NewFileRepository(tempFilePath)
		require.NoError(t, err)

		_, exists := repo.Get("nonexistent")
		assert.False(t, exists)
	})
}
