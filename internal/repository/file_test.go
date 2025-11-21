package repository

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
		defer func() {
			exists := repo.CheckConnection(t.Context())
			assert.True(t, exists)
			repo.Close()
		}()
		assert.NotNil(t, repo)

		// Check that the repository is empty
		_, exists := repo.Get(t.Context(), "nonexistent")
		assert.False(t, exists)
	})

	// Prepare initial data for the next test by manually writing JSONL
	initialData := []model.URLRecord{
		{UUID: "1", ShortURL: "abc123", OriginalURL: "http://example.com"},
		{UUID: "2", ShortURL: "def456", OriginalURL: "http://practicum.yandex.ru"},
	}

	// Manually create the file with JSONL format
	file, err := os.Create(tempFilePath)
	require.NoError(t, err)
	for _, record := range initialData {
		data, err := json.Marshal(record)
		require.NoError(t, err)
		_, err = file.Write(data)
		require.NoError(t, err)
		_, err = file.WriteString("\n") // Add newline after each record
		require.NoError(t, err)
	}
	file.Close()

	// Test 2: Creating a repository with an existing JSONL file -> data is loaded
	t.Run("NewFileRepository loads data from existing JSONL file", func(t *testing.T) {
		// Create a new repository; it should load data from the JSONL file
		repo, err := NewFileRepository(tempFilePath)
		require.NoError(t, err)
		defer func() {
			exists := repo.CheckConnection(t.Context())
			assert.True(t, exists)
			repo.Close()
		}()
		assert.NotNil(t, repo)

		// Check that the data is loaded
		url, exists := repo.Get(t.Context(), "abc123")
		assert.True(t, exists)
		assert.Equal(t, "http://example.com", url)

		url, exists = repo.Get(t.Context(), "def456")
		assert.True(t, exists)
		assert.Equal(t, "http://practicum.yandex.ru", url)

		// Check that a non-existent ID is not returned
		_, exists = repo.Get(t.Context(), "nonexistent")
		assert.False(t, exists)

		// Check that nextUUID is set correctly (max + 1)
		// This is harder to check directly without exposing internal state,
		// but we can verify by saving a new item and checking its UUID.
		// Let's add this check to the next test.
	})

	// Test 3: Save adds a new record as a new line and updates the file
	t.Run("Save adds new record as a new line and updates file", func(t *testing.T) {
		// Use the same file as in the previous test.
		// The repo instance from the previous test is gone.
		// We create a new one to ensure it reloads the file's state before our save.
		// But actually, the file already contains 2 records (UUID 1, 2).
		// After loading, nextUUIDToAssign should be 3.
		// Our save should create a record with UUID 3.

		// First, reload the repo to get the latest state from the file
		repo, err := NewFileRepository(tempFilePath)
		require.NoError(t, err)
		defer func() {
			exists := repo.CheckConnection(t.Context())
			assert.True(t, exists)
			repo.Close()
		}()

		// Count lines before save to verify new line is added
		fileBefore, err := os.Open(tempFilePath)
		require.NoError(t, err)
		scannerBefore := bufio.NewScanner(fileBefore)
		linesBefore := 0
		for scannerBefore.Scan() {
			linesBefore++
		}
		fileBefore.Close()

		// Save a new pair
		newPair := model.URLPair{
			ShortID: "ghi789",
			URL:     "http://newsite.com",
		}
		err = repo.Save(t.Context(), newPair)
		require.NoError(t, err)

		// Check that it is available in the repository
		url, exists := repo.Get(t.Context(), "ghi789")
		assert.True(t, exists)
		assert.Equal(t, "http://newsite.com", url)

		// Check that the file has one more line
		fileAfter, err := os.Open(tempFilePath)
		require.NoError(t, err)
		scannerAfter := bufio.NewScanner(fileAfter)
		linesAfter := 0
		for scannerAfter.Scan() {
			linesAfter++
		}
		fileAfter.Close()

		assert.Equal(t, linesBefore+1, linesAfter)

		// Read the last line of the file to check the new record
		fileForReading, err := os.Open(tempFilePath)
		require.NoError(t, err)
		defer fileForReading.Close()

		scanner := bufio.NewScanner(fileForReading)
		var lastLine string
		for scanner.Scan() {
			lastLine = scanner.Text()
		}
		require.NoError(t, scanner.Err())

		var lastRecordFromFile model.URLRecord
		err = json.Unmarshal([]byte(lastLine), &lastRecordFromFile)
		require.NoError(t, err)

		// Verify the last record matches what we saved and has the correct UUID
		assert.Equal(t, "3", lastRecordFromFile.UUID) // Because max was 2, next should be 3
		assert.Equal(t, "ghi789", lastRecordFromFile.ShortURL)
		assert.Equal(t, "http://newsite.com", lastRecordFromFile.OriginalURL)

	})

	// Test 4: Save multiple records sequentially, check file integrity
	t.Run("Save multiple records sequentially", func(t *testing.T) {
		// Use the same file. It currently has 3 records (UUIDs 1, 2, 3).
		// After loading, nextUUIDToAssign should be 4.
		repo, err := NewFileRepository(tempFilePath)
		require.NoError(t, err)
		defer func() {
			exists := repo.CheckConnection(t.Context())
			assert.True(t, exists)
			repo.Close()
		}()

		// Save two more pairs
		pair1 := model.URLPair{ShortID: "jkl012", URL: "http://site1.com"}
		pair2 := model.URLPair{ShortID: "mno345", URL: "http://site2.com"}

		err = repo.Save(t.Context(), pair1)
		require.NoError(t, err)
		err = repo.Save(t.Context(), pair2)
		require.NoError(t, err)

		// Check that they are available in the current repo instance
		url, exists := repo.Get(t.Context(), "jkl012")
		assert.True(t, exists)
		assert.Equal(t, "http://site1.com", url)

		url, exists = repo.Get(t.Context(), "mno345")
		assert.True(t, exists)
		assert.Equal(t, "http://site2.com", url)

		// Reload the repo to check persistence
		reloadRepo, err := NewFileRepository(tempFilePath)
		require.NoError(t, err)
		defer func() {
			exists := repo.CheckConnection(t.Context())
			assert.True(t, exists)
			repo.Close()
		}()

		// Check that the newly saved items are also present after reload
		url, exists = reloadRepo.Get(t.Context(), "jkl012")
		assert.True(t, exists)
		assert.Equal(t, "http://site1.com", url)

		url, exists = reloadRepo.Get(t.Context(), "mno345")
		assert.True(t, exists)
		assert.Equal(t, "http://site2.com", url)

		// Check that all previously saved items are still there
		url, exists = reloadRepo.Get(t.Context(), "abc123")
		assert.True(t, exists)
		assert.Equal(t, "http://example.com", url)

		url, exists = reloadRepo.Get(t.Context(), "def456")
		assert.True(t, exists)
		assert.Equal(t, "http://practicum.yandex.ru", url)

		url, exists = reloadRepo.Get(t.Context(), "ghi789")
		assert.True(t, exists)
		assert.Equal(t, "http://newsite.com", url)

		// Check that nextUUID is correctly tracked internally by saving another and verifying its UUID
		pair3 := model.URLPair{ShortID: "xyz999", URL: "http://site3.com"}
		err = reloadRepo.Save(t.Context(), pair3)
		require.NoError(t, err)

		// The new record should have UUID 6 (previous max was 3, then 4, then 5 were assigned during saves, next is 6)
		// Or, based on the *loaded* state: max was 3, so nextUUIDToAssign starts at 4.
		// Save(pair1) -> assigns 4 (jkl012)
		// Save(pair2) -> assigns 5 (mno345)
		// Save(pair3) -> assigns 6 (xyz999)
		// When we reloaded the repo, it read the file again, saw up to UUID 5, so nextUUIDToAssign became 6.
		// Save(pair3) should assign 6.
		// Let's re-evaluate the expected UUID based on the *state* of the reloaded repo.
		// The reloaded repo sees 5 records (1,2,3,4,5). So nextUUIDToAssign is 6.
		// Save(pair3) will assign 6.
		// We can verify this by checking the file again or by internal state if exposed.
		// Since we don't expose internal state easily, we rely on the file format.
		// After this save, the last line should be the record for xyz999 with UUID 6.
		// Let's re-check the file content.
		fileForCheck, err := os.Open(tempFilePath)
		require.NoError(t, err)
		defer fileForCheck.Close()

		scanner := bufio.NewScanner(fileForCheck)
		var lines []string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) != "" { // Skip empty lines if any
				lines = append(lines, line)
			}
		}
		require.NoError(t, scanner.Err())

		// The last line should be the record for xyz999 with UUID "6"
		require.NotEmpty(t, lines)
		lastLineContent := lines[len(lines)-1]
		var finalRecord model.URLRecord
		err = json.Unmarshal([]byte(lastLineContent), &finalRecord)
		require.NoError(t, err)

		assert.Equal(t, "xyz999", finalRecord.ShortURL)
		assert.Equal(t, "http://site3.com", finalRecord.OriginalURL)
		// The UUID depends on how many records were assigned during the reloadRepo Saves.
		// If reloadRepo started with nextUUID=6 (based on file scan), then xyz999 gets UUID=6.
		assert.Equal(t, "6", finalRecord.UUID)

	})

	// Test 5: Get returns false for a non-existent ID
	t.Run("Get returns false for non-existent ID", func(t *testing.T) {
		// Use the file with data
		repo, err := NewFileRepository(tempFilePath)
		require.NoError(t, err)
		defer func() {
			exists := repo.CheckConnection(t.Context())
			assert.True(t, exists)
			repo.Close()
		}()

		_, exists := repo.Get(t.Context(), "nonexistent")
		assert.False(t, exists)
	})

	// Test 6: Get returns false for check connection
	t.Run("Get returns false for check connection", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "non_existent_dir")
		nonExistentPath := filepath.Join(nonExistentDir, "non_existent_file.json")
		repo, err := NewFileRepository(nonExistentPath)
		require.NoError(t, err)
		defer repo.Close()

		exists := repo.CheckConnection(t.Context())
		assert.False(t, exists)
	})
}
