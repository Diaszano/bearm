// Package trash contains platform-independent trash helpers.
package trash

import "strings"

const uppercaseHex = "0123456789ABCDEF"

// EncodePath percent-encodes a UTF-8 path for a FreeDesktop .trashinfo file.
func EncodePath(path string) string {
	var builder strings.Builder
	builder.Grow(len(path))

	for index := 0; index < len(path); index++ {
		value := path[index]
		if isSafePathByte(value) {
			builder.WriteByte(value)
			continue
		}

		builder.WriteByte('%')
		builder.WriteByte(uppercaseHex[value>>4])
		builder.WriteByte(uppercaseHex[value&0x0f])
	}

	return builder.String()
}

func isSafePathByte(value byte) bool {
	switch {
	case value >= 'a' && value <= 'z':
		return true
	case value >= 'A' && value <= 'Z':
		return true
	case value >= '0' && value <= '9':
		return true
	}

	switch value {
	case '-', '_', '.', '~', '/':
		return true
	default:
		return false
	}
}
