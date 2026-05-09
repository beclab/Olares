package clusteropts

import (
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
