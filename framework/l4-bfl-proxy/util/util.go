package util

import (
	"bytes"
	"k8s.io/apimachinery/pkg/util/json"
)

func PrettyJSON(v any) string {
	var bf bytes.Buffer

	enc := json.NewEncoder(&bf)
	enc.SetIndent("", "  ")
	enc.Encode(v) // noqa

	return bf.String()
}
