package tools

import (
	"time"

	"math/rand"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomString(strlen int) string {
	b := make([]rune, strlen)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] // #nosec
	}

	return string(b)
}
