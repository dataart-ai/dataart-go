package randomutil

import (
	"math/rand"
	"time"
)

var (
	chars = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func String(length int) string {
	rand.Seed(time.Now().UnixNano())

	pick := make([]rune, length)
	for i := range pick {
		pick[i] = chars[rand.Intn(len(chars))]
	}

	return string(pick)
}
