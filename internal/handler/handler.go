package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/LakinPavel/url_shortener.git/internal/generator"
	"github.com/LakinPavel/url_shortener.git/internal/storage"
)

type Handler struct {
	store storage.Storage
}

func New(store storage.Storage) *Handler {
	return &Handler{store: store}
}

func (h *Handler) PostURL(w http.ResponseWriter, r *http.Request) {
	slog.Debug("incoming POST request", "method", r.Method, "path", r.URL.Path)
	if r.Method != http.MethodPost {
		slog.Warn("method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		slog.Warn("invalid request payload", "error", err, "url_empty", req.URL == "")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	shortURL := generator.GenerateShortURL(req.URL)
	finalShortURL, err := h.store.Save(r.Context(), req.URL, shortURL)
	if err != nil {
		slog.Error("failed to save URL to storage", "error", err, "original_url", req.URL)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("URL shortened successfully", "short", finalShortURL, "original", req.URL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"short_url": finalShortURL})
}

func isValid(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	slog.Debug("incoming GET request", "shortURL", shortURL)

	if r.Method != http.MethodGet {
		slog.Warn("method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(shortURL) != 10 || !isValid(shortURL) {
		slog.Warn("invalid short URL format", "shortURL", shortURL)
		http.Error(w, "Invalid short URL format", http.StatusBadRequest)
		return
	}

	originalURL, err := h.store.GetOriginal(r.Context(), shortURL)
	if err != nil {
		slog.Warn("URL not found in storage", "shortURL", shortURL)
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	slog.Info("redirecting to original URL", "short", shortURL, "original", originalURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"original_url": originalURL})
}
