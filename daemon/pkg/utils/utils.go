// Package utils provides common utility functions for the daemon.
package utils

import (
	"net"
	"regexp"

	_ "github.com/prometheus/node_exporter/collector"
)

// FilterArray returns a new slice containing only the items from the input
// slice for which the provided function returns true.
func FilterArray[T any](items []T, fn func(T) bool) []T {
	var filtered []T
	for _, item := range items {
		if fn(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// IsValidIP reports whether s is a valid IPv4 or IPv6 address.
func IsValidIP(s string) bool {
	return net.ParseIP(s) != nil
}

var domainRegexp = regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*)$`)

// IsValidDomain reports whether s is a syntactically valid domain name.
func IsValidDomain(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	return domainRegexp.MatchString(s)
}
