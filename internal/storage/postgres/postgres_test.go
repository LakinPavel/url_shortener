//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestPostgresStorage_Integration(t *testing.T) {
	dsn := "postgres://user:password@localhost:5433/shortener?sslmode=disable"

	store, err := New(dsn)
	if err != nil {
		t.Fatalf("не удалось подключиться к базе: %v", err)
	}
	defer store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = store.db.ExecContext(ctx, "TRUNCATE TABLE links RESTART IDENTITY")
	if err != nil {
		t.Logf("предупреждение: не удалось очистить таблицу: %v", err)
	}

	t.Run("Save and Get Success", func(t *testing.T) {
		original := "https://google.com"
		short := "google1234"

		gotShort, err := store.Save(ctx, original, short)
		if err != nil {
			t.Errorf("ошибка при сохранении: %v", err)
		}

		gotOriginal, err := store.GetOriginal(ctx, gotShort)
		if err != nil {
			t.Errorf("ошибка при получении: %v", err)
		}

		if gotOriginal != original {
			t.Errorf("ждали %s, получили %s", original, gotOriginal)
		}
	})

	t.Run("Conflict: Same URL returns same ShortID", func(t *testing.T) {
		original := "https://yandex.ru"
		short1 := "yandex1234"
		short2 := "yandex4321"

		s1, _ := store.Save(ctx, original, short1)

		s2, err := store.Save(ctx, original, short2)

		if err != nil {
			t.Errorf("повторное сохранение не должно вызывать ошибку: %v", err)
		}

		if s1 != s2 {
			t.Errorf("логика ON CONFLICT нарушена: ждали %s, получили %s", s1, s2)
		}
	})

	t.Run("Get Non-Existent", func(t *testing.T) {
		_, err := store.GetOriginal(ctx, "non_existent")
		if err == nil {
			t.Error("ожидалась ошибка sql.ErrNoRows, получили nil")
		}
	})
}
