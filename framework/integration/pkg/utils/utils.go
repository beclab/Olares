package utils

import "os"

func GetenvOrDefault(env string, d string) string {
	v := os.Getenv(env)
	if v == "" {
		return d
	}

	return v
}

func ListContains[T comparable](items []T, v T) bool {
	for _, item := range items {
		if v == item {
			return true
		}
	}
	return false
}
