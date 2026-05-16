package events

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

func generate_id() string {
	b := make([]byte, 10) // 10 bytes = 80 bits
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	// Base32 without padding, lowercase
	return strings.ToLower(strings.TrimRight(base32.StdEncoding.EncodeToString(b), "="))
}
