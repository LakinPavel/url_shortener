package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/LakinPavel/url_shortener.git/internal/handler"
	"github.com/LakinPavel/url_shortener.git/internal/storage"
	"github.com/LakinPavel/url_shortener.git/internal/storage/memory"
	"github.com/LakinPavel/url_shortener.git/internal/storage/postgres"
)

func setupLogger(levelStr string) {
	var level slog.Level

	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	slog.SetDefault(logger)
}

func main() {
	logLevel := flag.String("log-level", "info", "log level: debug, info, warn, error")
	flag.Parse()

	setupLogger(*logLevel)

	slog.Info("Starting server", "level", *logLevel)
	storageType := flag.String("storage", "memory", "storage type: 'memory' or 'postgres'")
	dsn := flag.String("dsn", "postgres://user:pass@localhost:5432/shortener?sslmode=disable", "postgres dsn")
	flag.Parse()

	var store storage.Storage
	var err error

	if *storageType == "postgres" {
		store, err = postgres.New(*dsn)
		if err != nil {
			log.Fatalf("Failed to init postgres: %v", err)
		}
		log.Println("Using PostgreSQL storage")
	} else {
		store = memory.New()
		log.Println("Using In-Memory storage")
	}

	h := handler.New(store)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.PostURL(w, r)
		} else if r.Method == http.MethodGet {
			h.GetURL(w, r)
		}
	})

	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
