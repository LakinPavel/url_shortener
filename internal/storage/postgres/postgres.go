package postgres

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func New(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id SERIAL PRIMARY KEY,
			original_url TEXT UNIQUE NOT NULL,
			short_url VARCHAR(10) UNIQUE NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Save(ctx context.Context, originalURL, shortURL string) (string, error) {
	var finalShortURL string
	query := `
		INSERT INTO links (original_url, short_url) 
		VALUES ($1, $2) 
		ON CONFLICT (original_url) DO UPDATE 
		SET original_url = EXCLUDED.original_url 
		RETURNING short_url;
	`
	err := s.db.QueryRowContext(ctx, query, originalURL, shortURL).Scan(&finalShortURL)
	return finalShortURL, err
}

func (s *PostgresStorage) GetOriginal(ctx context.Context, shortURL string) (string, error) {
	var original string
	err := s.db.QueryRowContext(ctx, "SELECT original_url FROM links WHERE short_url = $1", shortURL).Scan(&original)
	return original, err
}
