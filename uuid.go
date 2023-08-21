package qstash

import (
	"crypto/rand"
	"io"
	"math/big"
)

type uuid struct {
}


// NewV4 is a 16 byte universally unique identifier
// generated for each message published with this package by default
func (*uuid) NewV4() (string, error) {
	// Generate a random uuid
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid[:])
	if err != nil {
		return "", err
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	// Base62 encode the uuid
	var i big.Int
	i.SetBytes(uuid)
	return i.Text(62), nil
}	