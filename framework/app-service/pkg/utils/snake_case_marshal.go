package utils

import (
	"bytes"
	"encoding/json"
	"regexp"
)

var keyMatch = regexp.MustCompile(`\"(\w+)\":`)
var wordSplit = regexp.MustCompile(`(\w)([A-Z0-9])`)

// SnakeCaseMarshaller is a struct that contains a field `Value`.
// This struct is intended to implement custom JSON marshalling by converting
// the field names in `Value` from camelCase or PascalCase to snake_case which
// is often used in JSON keys.
type SnakeCaseMarshaller struct {
	Value interface{}
}

// MarshalJSON implements the marshalling to JSON for the SnakeCaseMarshaller.
// This method overrides the standard marshalling behavior to convert
// the field names of the 'Value' to snake_case.
func (sc SnakeCaseMarshaller) MarshalJSON() ([]byte, error) {
	s, err := json.Marshal(sc.Value)
	converted := keyMatch.ReplaceAllFunc(
		s,
		func(match []byte) []byte {
			return bytes.ToLower(wordSplit.ReplaceAll(
				match,
				[]byte(`${1}_${2}`),
			))
		},
	)
	return converted, err
}
