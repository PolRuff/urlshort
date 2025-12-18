package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PolRuff/urlshort/internal/model"
)

// FileRepository implements Repository interface using a JSONL file
// Each URLRecord is stored on a separate line in the file.
type FileRepository struct {
	filePath string
	// map[shortID]URLRecord for fast lookup
	records map[string]model.URLRecord
	mu      sync.RWMutex
}

// NewFileRepository creates a new FileRepository.
// It tries to load existing data from the file if it exists.
func NewFileRepository(filePath string) (*FileRepository, error) {
	repo := &FileRepository{
		filePath: filePath,
		records:  make(map[string]model.URLRecord),
	}

	// Try to load existing data from the file
	err := repo.loadFromFile()
	if err != nil && !os.IsNotExist(err) {
		// If error is not "file not found", return it
		return nil, err
	}
	// If file doesn't exist, it's OK, we start with an empty map and nextUUID=1

	return repo, nil
}

// loadFromFile reads the JSONL file and populates the records map and nextUUID
func (r *FileRepository) loadFromFile() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	r.mu.Lock()
	defer r.mu.Unlock()

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // Skip empty line
		}

		var record model.URLRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return err
		}

		r.records[record.ShortURL] = record
	}

	err = scanner.Err()

	return err
}

// Save stores a URL record and appends it as a new line to the JSONL file.
func (r *FileRepository) Save(ctx context.Context, record model.URLRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store the record in the map
	r.records[record.ShortURL] = record

	// Append the new record as a JSON line to the file
	return r.appendToFile(record)
}

func (r *FileRepository) Delete(ctx context.Context, userID uint32, shortIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// TODO: implement
	return nil
}

// appendToFile writes a single URLRecord as a JSON line to the end of the file
func (r *FileRepository) appendToFile(record model.URLRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(r.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}
	_, err = file.WriteString("\n")
	return err
}

// Get retrieves the original URL by short ID
func (r *FileRepository) Get(ctx context.Context, shortID string) (string, bool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	isDeleted := false // TODO: implement

	record, ok := r.records[shortID]
	if !ok {
		return "", false, isDeleted
	}
	return record.OriginalURL, true, isDeleted
}

func (r *FileRepository) GetByUser(ctx context.Context, userID uint32) ([]model.UserUrls, error) {
	// TODO: implement
	return nil, nil
}

func (r *FileRepository) GetMaxID() (uint64, error) {
	return uint64(len(r.records)), nil
}

func (r *FileRepository) CheckConnection(ctx context.Context) bool {
	dir := filepath.Dir(r.filePath)
	testFile := filepath.Join(dir, ".connection_test.tmp")
	file, err := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return false
	}

	defer func() {
		file.Close()
		os.Remove(testFile)
	}()

	return true
}

func (r *FileRepository) Close() {

}
