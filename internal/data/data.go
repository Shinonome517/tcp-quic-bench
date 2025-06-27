package data

import (
	"crypto/rand"
	"fmt"
)

// Generate a byte slice with 1GB of random data.
func Generate() ([]byte, error) {
	const size = 1 << 30 // 1GB
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		return nil, fmt.Errorf("failed to generate random data: %w", err)
	}
	return data, nil
}
