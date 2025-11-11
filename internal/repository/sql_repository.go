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
	_, err := r.db.ExecContext(context.Background(), "INSERT INTO shortened_urls (short_url, original_url) VALUES ($1, $2)", pair.ShortID, pair.URL)

	return err
}

// Get retrieves the original URL by short ID
func (r *SQLRepository) Get(shortID string) (string, bool) {
	row := r.db.QueryRowContext(context.Background(), "SELECT original_url FROM shortened_urls WHERE short_url = $1", shortID)

	var originalURL sql.NullString
	err := row.Scan(&originalURL)

	if err != nil || !originalURL.Valid {
		return "", false
	}

	return originalURL.String, true
}

func (r *SQLRepository) GetMaxID() (uint64, error) {
	row := r.db.QueryRowContext(context.Background(), "SELECT MAX(id) FROM shortened_urls")

	var maxID sql.NullInt64
	err := row.Scan(&maxID)

	if err != nil {
		return 0, err
	}

	if maxID.Valid {
		return uint64(maxID.Int64), nil
	}

	return 0, err
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
