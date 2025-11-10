package repository

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
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
	// nextUUIDToAssign is used for generating new UUIDs sequentially
	nextUUIDToAssign int
	mu               sync.RWMutex
}

// NewFileRepository creates a new FileRepository.
// It tries to load existing data from the file if it exists.
func NewFileRepository(filePath string) (*FileRepository, error) {
	repo := &FileRepository{
		filePath:         filePath,
		records:          make(map[string]model.URLRecord),
		nextUUIDToAssign: 1, // начинаем с 1
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
	maxUUID := 0

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
		// Parse UUID to int to find the maximum
		if uuidInt, err := strconv.Atoi(record.UUID); err == nil {
			if uuidInt > maxUUID {
				maxUUID = uuidInt
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Set nextUUIDToAssign to maxUUID + 1
	r.nextUUIDToAssign = maxUUID + 1

	return nil
}

// Save stores a URL pair and appends it as a new line to the JSONL file.
// It generates a new sequential UUID for the record.
func (r *FileRepository) Save(pair model.URLPair) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a new URLRecord with a generated UUID
	record := model.URLRecord{
		UUID:        strconv.Itoa(r.nextUUIDToAssign),
		ShortURL:    pair.ShortID,
		OriginalURL: pair.URL,
	}

	// Store the record in the map
	r.records[pair.ShortID] = record
	// Increment the next UUID counter
	r.nextUUIDToAssign++

	// Append the new record as a JSON line to the file
	return r.appendToFile(record)
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
func (r *FileRepository) Get(shortID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, ok := r.records[shortID]
	if !ok {
		return "", false
	}
	return record.OriginalURL, true
}
