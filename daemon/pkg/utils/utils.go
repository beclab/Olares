package utils

import (
	"net"
	"regexp"
)

func FilterArray[T any](items []T, fn func(T) bool) []T {
	var filtered []T
	for _, item := range items {
		if fn(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func IsValidIP(s string) bool {
	return net.ParseIP(s) != nil
}

var domainRegexp = regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*)$`)

func IsValidDomain(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	return domainRegexp.MatchString(s)
}
