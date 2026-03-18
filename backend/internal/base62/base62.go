package base62

import (
	"strings"
)

// We use an explicit 62-character alphabet. 
// Shuffling this string creates a "custom" alphabet, making your URLs harder to guess.
const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const base = uint64(len(alphabet))

// Encode converts a base-10 integer ID to a Base-62 string
func Encode(id uint64) string {
	if id == 0 {
		return string(alphabet[0])
	}

	var result []byte
	for id > 0 {
		remainder := id % base
		result = append(result, alphabet[remainder])
		id = id / base
	}

	// The division process builds the string backwards, so we must reverse it
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// Decode converts a Base-62 string back to a base-10 integer ID
func Decode(shortURL string) uint64 {
	var id uint64
	for _, char := range shortURL {
		id = id*base + uint64(strings.IndexRune(alphabet, char))
	}
	return id
}