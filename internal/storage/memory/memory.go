package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/LakinPavel/url_shortener.git/internal/storage"
)

type MemoryStorage struct {
	mu          sync.RWMutex
	shortToOrig map[string]string
	origToShort map[string]string
}

func New() *MemoryStorage {
	return &MemoryStorage{
		shortToOrig: make(map[string]string),
		origToShort: make(map[string]string),
	}
}

func (s *MemoryStorage) Save(ctx context.Context, originalURL, shortURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingShort, exists := s.origToShort[originalURL]; exists {
		return existingShort, nil
	}

	if _, exists := s.shortToOrig[shortURL]; exists {
		return "", storage.ErrCollision
	}

	s.shortToOrig[shortURL] = originalURL
	s.origToShort[originalURL] = shortURL
	return shortURL, nil
}

func (s *MemoryStorage) GetOriginal(ctx context.Context, shortURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orig, exists := s.shortToOrig[shortURL]
	if !exists {
		return "", errors.New("url not found")
	}
	return orig, nil
}

func (s *MemoryStorage) GetShort(ctx context.Context, originalURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	short, exists := s.origToShort[originalURL]
	if !exists {
		return "", nil
	}
	return short, nil
}
