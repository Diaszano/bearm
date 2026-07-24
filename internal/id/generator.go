// Package id creates opaque Bearm operation and item identifiers.
package id

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a 128-bit lowercase hexadecimal identifier.
func New() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
