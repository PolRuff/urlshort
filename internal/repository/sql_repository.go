package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/PolRuff/urlshort/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
)

// ConflictError represents a conflict when trying to save a URL that already exists
type ConflictError struct {
	OriginalURL     string
	ExistingShortID string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("URL %q already exists with short ID %q", e.OriginalURL, e.ExistingShortID)
}

// Is reports whether target is a ConflictError
func (e *ConflictError) Is(target error) bool {
	_, ok := target.(*ConflictError)
	return ok
}

// SQLRepository implements db storage for URL records
type SQLRepository struct {
	db *sql.DB
}

// SQLRepository creates a new db repository
func NewSQLRepository(databaseDsn string) (*SQLRepository, error) {
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		return nil, err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		db.Close()
		return nil, err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		db.Close()
		return nil, err
	}
	return &SQLRepository{
		db: db,
	}, nil
}

// Save stores a URL record
func (r *SQLRepository) Save(ctx context.Context, record model.URLRecord) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO shortened_urls (short_url, original_url) VALUES ($1, $2)", record.ShortURL, record.OriginalURL)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				var existingShortID string
				err = r.db.QueryRowContext(ctx, "SELECT short_url FROM shortened_urls WHERE original_url = $1", record.OriginalURL).Scan(&existingShortID)
				if err != nil {
					return fmt.Errorf("failed to retrieve existing short URL after conflict: %w", err)
				}
				return &ConflictError{
					OriginalURL:     record.OriginalURL,
					ExistingShortID: existingShortID,
				}
			}
		}
		return fmt.Errorf("failed to save URL record: %w", err)
	}

	return err
}

// Get retrieves the original URL by short ID
func (r *SQLRepository) Get(ctx context.Context, shortID string) (string, bool) {
	row := r.db.QueryRowContext(ctx, "SELECT original_url FROM shortened_urls WHERE short_url = $1", shortID)

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
