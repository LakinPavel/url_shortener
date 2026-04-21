package generator

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

func GenerateShortURL(originalURL string) string {
	hash := sha256.Sum256([]byte(originalURL))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	encoded = strings.ReplaceAll(encoded, "-", "_")
	return encoded[:10]
}
