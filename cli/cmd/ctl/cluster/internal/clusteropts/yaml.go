package clusteropts

import (
	"bytes"
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

// JSONToYAML converts the K8s native API's JSON object response into
// the canonical YAML rendering expected by `cluster <noun> yaml`.
// Uses sigs.k8s.io/yaml so JSON tag conventions and field ordering
// match what `kubectl get -o yaml` produces.
func JSONToYAML(body []byte) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}
	return yaml.Marshal(v)
}

// IndentJSON re-indents a server response so a verb can forward every field
// it returned while matching PrintJSON's two-space layout. Use it instead of
// PrintJSON whenever the verb's typed struct is narrower than the response:
// marshalling the struct would silently drop whatever it does not model.
func IndentJSON(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, body, "", "  "); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}
	if buf.Len() > 0 && !bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}
