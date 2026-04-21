package handler

import (
	"encoding/json"
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	shortURL := generator.GenerateShortURL(req.URL)
	finalShortURL, err := h.store.Save(r.Context(), req.URL, shortURL)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	if len(shortURL) != 10 || !isValid(shortURL) {
		http.Error(w, "Invalid short URL", http.StatusBadRequest)
		return
	}

	originalURL, err := h.store.GetOriginal(r.Context(), shortURL)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"original_url": originalURL})
}
