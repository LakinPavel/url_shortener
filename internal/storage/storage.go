package storage

import (
	"context"
	"errors"
)

var ErrCollision = errors.New("short url collision")

type Storage interface {
	Save(ctx context.Context, originalURL, shortURL string) (string, error)
	GetOriginal(ctx context.Context, shortURL string) (string, error)
	GetShort(ctx context.Context, originalURL string) (string, error)
}
