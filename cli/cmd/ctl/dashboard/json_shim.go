package dashboard

import "encoding/json"

// jsonMarshal / jsonUnmarshal are tiny shims kept in their own file so the
// rest of the package doesn't have to import "encoding/json" everywhere.
// The names sit in helpers.go's import space; tests can swap them at link
// time via build tags if a non-stdlib JSON encoder is ever desired.
func jsonMarshal(v interface{}) ([]byte, error)  { return json.Marshal(v) }
func jsonUnmarshal(b []byte, v interface{}) error { return json.Unmarshal(b, v) }
