package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/PolRuff/urlshort/internal/model"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// SQLRepository implements db storage for URL pairs
type SQLRepository struct {
	db *sql.DB
}

// SQLRepository creates a new db repository
func NewSQLRepository(databaseDsn string) (*SQLRepository, error) {
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		return nil, err
	}
	return &SQLRepository{
		db: db,
	}, nil
}

// Save stores a URL pair
func (r *SQLRepository) Save(pair model.URLPair) error {
	return nil
}

// Get retrieves the original URL by short ID
func (r *SQLRepository) Get(shortID string) (string, bool) {
	return "", true
}

func (r *SQLRepository) CheckConnection(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err := r.db.PingContext(ctx)

	return err == nil
}

func (r *SQLRepository) Close() {
	r.db.Close()
}
