package generator

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
const base = uint64(len(alphabet))

func GenerateShortURL(originalURL string, attempt int) string {
	input := originalURL

	if attempt > 0 {
		input = fmt.Sprintf("%s[%d]", originalURL, attempt)
	}

	hash := sha256.Sum256([]byte(input))

	num := binary.BigEndian.Uint64(hash[:8])

	var result []byte
	for i := 0; i < 10; i++ {
		idx := num % base
		result = append(result, alphabet[idx])
		num /= base
	}

	return string(result)
}
