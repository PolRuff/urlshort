package repository

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/PolRuff/urlshort/internal/model"
)

// FileRepository implements Repository interface using a JSON file
// It stores URLRecords internally, including their UUIDs for persistence.
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

// loadFromFile reads the JSON file and populates the records map and nextUUID
func (r *FileRepository) loadFromFile() error {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return err
	}

	var records []model.URLRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	maxUUID := 0
	// Populate the map and find the max UUID
	for _, record := range records {
		r.records[record.ShortURL] = record
		// Parse UUID to int to find the maximum
		if uuidInt, err := strconv.Atoi(record.UUID); err == nil {
			if uuidInt > maxUUID {
				maxUUID = uuidInt
			}
		}
	}

	// Set nextUUIDToAssign to maxUUID + 1
	r.nextUUIDToAssign = maxUUID + 1

	return nil
}

// Save stores a URL pair and persists it to the file.
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

	// Persist the entire map to the file
	return r.saveToFile()
}

// saveToFile writes the current state of the records map to the JSON file
func (r *FileRepository) saveToFile() error {
	// Convert map to slice of URLRecord
	var records []model.URLRecord
	for _, record := range r.records {
		// Iterate over map values and append them to the slice
		records = append(records, record)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.filePath, data, 0644)
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
