package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/LakinPavel/url_shortener.git/internal/storage"
)

type mockStorage struct {
	mu          sync.RWMutex
	shortToOrig map[string]string
	origToShort map[string]string
	err         error
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		shortToOrig: make(map[string]string),
		origToShort: make(map[string]string),
	}
}

func (m *mockStorage) Save(ctx context.Context, originalURL, shortURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Имитация коллизии: если shortURL занят ДРУГИМ оригинальным URL
	if existingOrig, exists := m.shortToOrig[shortURL]; exists && existingOrig != originalURL {
		return "", storage.ErrCollision
	}

	m.shortToOrig[shortURL] = originalURL
	m.origToShort[originalURL] = shortURL
	return shortURL, nil
}

func (m *mockStorage) GetOriginal(ctx context.Context, shortURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	original, exists := m.shortToOrig[shortURL]
	if !exists {
		return "", errors.New("not found")
	}
	return original, nil
}

// Добавляем недостающий метод GetShort
func (m *mockStorage) GetShort(ctx context.Context, originalURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	short, exists := m.origToShort[originalURL]
	if !exists {
		return "", nil // Если не нашли, возвращаем пустую строку (без ошибки)
	}
	return short, nil
}

var _ storage.Storage = (*mockStorage)(nil)

func TestHandler_PostURL(t *testing.T) {
	mockStore := newMockStorage()
	h := New(mockStore)

	t.Run("success", func(t *testing.T) {
		body := bytes.NewBufferString(`{"url": "https://example.com"}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		rec := httptest.NewRecorder()

		h.PostURL(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["short_url"] == "" {
			t.Error("expected short_url in response, got empty string")
		}
	})

	t.Run("invalid_method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.PostURL(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
	})

	t.Run("empty_url", func(t *testing.T) {
		body := bytes.NewBufferString(`{"url": ""}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		rec := httptest.NewRecorder()

		h.PostURL(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}

func TestHandler_GetURL(t *testing.T) {
	mockStore := newMockStorage()
	h := New(mockStore)

	// Правильное заполнение мока перед тестом
	mockStore.shortToOrig["1234567890"] = "https://golang.org"
	mockStore.origToShort["https://golang.org"] = "1234567890"

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/1234567890", nil)
		rec := httptest.NewRecorder()

		h.GetURL(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp map[string]string
		json.NewDecoder(rec.Body).Decode(&resp)
		if resp["original_url"] != "https://golang.org" {
			t.Errorf("expected original_url %s, got %s", "https://golang.org", resp["original_url"])
		}
	})

	t.Run("not_found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/0987654321", nil)
		rec := httptest.NewRecorder()

		h.GetURL(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("invalid_format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/short", nil)
		rec := httptest.NewRecorder()

		h.GetURL(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}

func TestHandler_ConcurrentLoad(t *testing.T) {
	mockStore := newMockStorage()
	h := New(mockStore)

	server := httptest.NewServer(http.HandlerFunc(h.PostURL))
	defer server.Close()

	var wg sync.WaitGroup
	workers := 100

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			body := bytes.NewBufferString(`{"url": "https://highload.com"}`)
			resp, err := http.Post(server.URL, "application/json", body)
			if err != nil {
				t.Errorf("Worker %d failed to send request: %v", workerID, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Worker %d got status: %d", workerID, resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
}
