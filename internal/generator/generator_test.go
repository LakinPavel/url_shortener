package generator

import (
	"strings"
	"testing"
)

func TestGenerateShortURL_Length(t *testing.T) {
	urls := []string{
		"https://google.com",
		"https://very-long-url.example.com/path/to/resource?query=123",
		"",
		"short",
	}

	for _, url := range urls {
		short := GenerateShortURL(url, 0)
		if len(short) != 10 {
			t.Errorf("Expected length 10, but got %d for url '%s' (result: %s)", len(short), url, short)
		}
	}
}

func TestGenerateShortURL_Alphabet(t *testing.T) {
	url := "https://yandex.ru"
	short := GenerateShortURL(url, 0)

	for _, char := range short {
		if !strings.ContainsRune(alphabet, char) {
			t.Errorf("Result '%s' contains invalid character: '%c'", short, char)
		}
	}
}

func TestGenerateShortURL_Determinism(t *testing.T) {
	url := "https://github.com"

	short1 := GenerateShortURL(url, 0)
	short2 := GenerateShortURL(url, 0)

	if short1 != short2 {
		t.Errorf("Algorithm is not deterministic. For same URL got different results: '%s' and '%s'", short1, short2)
	}
}

func TestGenerateShortURL_AttemptChangesResult(t *testing.T) {
	url := "https://stackoverflow.com"

	shortAttempt0 := GenerateShortURL(url, 0)
	shortAttempt1 := GenerateShortURL(url, 1)
	shortAttempt2 := GenerateShortURL(url, 2)

	if shortAttempt0 == shortAttempt1 {
		t.Errorf("Attempt parameter did not change the result: both are '%s'", shortAttempt0)
	}
	if shortAttempt1 == shortAttempt2 {
		t.Errorf("Attempt parameter did not change the result: both are '%s'", shortAttempt1)
	}
	if shortAttempt0 == shortAttempt2 {
		t.Errorf("Attempt parameter did not change the result: both are '%s'", shortAttempt0)
	}
}

func TestGenerateShortURL_DifferentURLs(t *testing.T) {
	url1 := "https://example.com/1"
	url2 := "https://example.com/2"

	short1 := GenerateShortURL(url1, 0)
	short2 := GenerateShortURL(url2, 0)

	if short1 == short2 {
		t.Errorf("Different URLs produced the same short URL (collision on first attempt): '%s'", short1)
	}
}
