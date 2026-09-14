package pathutil

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID returns a 128-bit lowercase hexadecimal identifier.
func NewID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
