package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LakinPavel/url_shortener.git/internal/storage"
)

type mockStorage struct {
	data map[string]string
	err  error
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		data: make(map[string]string),
	}
}

func (m *mockStorage) Save(ctx context.Context, originalURL, shortURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.data[shortURL] = originalURL
	return shortURL, nil
}

func (m *mockStorage) GetOriginal(ctx context.Context, shortURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	original, exists := m.data[shortURL]
	if !exists {
		return "", errors.New("not found")
	}
	return original, nil
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

	mockStore.data["1234567890"] = "https://golang.org"

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
