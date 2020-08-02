package nonce

import "math/rand"

func GenerateNonce(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz01234567890"

	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}