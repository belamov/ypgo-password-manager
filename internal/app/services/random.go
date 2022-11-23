package services

import (
	"crypto/rand"
)

// Generator generates random bytes and new user id.
type Generator interface {
	GenerateRandomBytes(size int) ([]byte, error)
}

// TrulyRandomGenerator is used for generating truly random values.
type TrulyRandomGenerator struct{}

// GenerateRandomBytes generates size random bytes.
func (g *TrulyRandomGenerator) GenerateRandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}
