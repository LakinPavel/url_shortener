package postgres

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"log/slog"
)

type PostgresStorage struct {
	db *sql.DB
}

func New(dsn string) (*PostgresStorage, error) {
	slog.Info("connecting to postgres database")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("failed to open database connection", "error", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		slog.Error("failed to ping database", "error", err)
		return nil, err
	}

	slog.Debug("checking/creating table 'links'")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id SERIAL PRIMARY KEY,
			original_url TEXT UNIQUE NOT NULL,
			short_url VARCHAR(10) UNIQUE NOT NULL
		);
	`)
	if err != nil {
		slog.Error("failed to initialize table", "error", err)
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Save(ctx context.Context, originalURL, shortURL string) (string, error) {
	var finalShortURL string
	slog.Debug("executing Save URL query", "original_url", originalURL)
	query := `
		INSERT INTO links (original_url, short_url) 
		VALUES ($1, $2) 
		ON CONFLICT (original_url) DO UPDATE 
		SET original_url = EXCLUDED.original_url 
		RETURNING short_url;
	`
	err := s.db.QueryRowContext(ctx, query, originalURL, shortURL).Scan(&finalShortURL)

	if err != nil {
		slog.Error("failed to save url", "error", err, "url", originalURL)
		return "", err
	}

	slog.Debug("url saved/retrieved", "short", finalShortURL)

	return finalShortURL, err
}

func (s *PostgresStorage) GetOriginal(ctx context.Context, shortURL string) (string, error) {
	var original string
	slog.Debug("executing GetOriginal query", "short_url", shortURL)
	err := s.db.QueryRowContext(ctx, "SELECT original_url FROM links WHERE short_url = $1", shortURL).Scan(&original)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("short url not found in database", "short_url", shortURL)
		} else {
			slog.Error("database query error during GetOriginal", "error", err, "short_url", shortURL)
		}
		return "", err
	}
	return original, err
}
