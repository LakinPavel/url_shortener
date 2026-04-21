package storage

import "context"

type Storage interface {
	Save(ctx context.Context, originalURL, shortURL string) (string, error)
	GetOriginal(ctx context.Context, shortURL string) (string, error)
}
