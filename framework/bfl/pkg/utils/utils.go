package utils

import (
	"bytes"
	"encoding/json"
)

func ToJSON(v any) string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		panic(err)
	}
	return buf.String()
}

func PrettyJSON(v any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return buf.String()
}

func ListContains[T comparable](items []T, v T) bool {
	for _, item := range items {
		if v == item {
			return true
		}
	}
	return false
}
