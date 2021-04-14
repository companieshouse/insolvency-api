package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateID generates an ID for a resource in Mongo
func GenerateID() string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 2)
	for i := 0; i < 2; i++ {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return string(b) + fmt.Sprintf("%08d", rand.Intn(99999999))
}
