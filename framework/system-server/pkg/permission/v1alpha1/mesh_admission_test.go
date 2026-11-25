package permission

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/google/cel-go/cel"
)

// Evaluate the shipped CEL expressions, not a Go approximation of admission.
func TestNonceMeshAdmissionRejectsBypass(t *testing.T) {
	data, err := os.ReadFile("../../../../app-gateway/.olares/config/cluster/deploy/http-mesh-admission.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var policy struct {
		Spec struct {
			MatchConditions []struct{ Expression string } `json:"matchConditions"`
			Variables       []struct{ Name, Expression string }
			Validations     []struct{ Expression, Message string }
		}
	}
	if err := yaml.Unmarshal([]byte(strings.SplitN(string(data), "\n---", 2)[0]), &policy); err != nil {
		t.Fatal(err)
	}
	env, err := cel.NewEnv(cel.Variable("object", cel.DynType), cel.Variable("oldObject", cel.DynType), cel.Variable("request", cel.DynType), cel.Variable("variables", cel.DynType))
	if err != nil {
		t.Fatal(err)
	}
	eval := func(expr string, values map[string]any) any {
		t.Helper()
		ast, issues := env.Compile(expr)
		if issues.Err() != nil {
			t.Fatal(issues.Err())
		}
		p, err := env.Program(ast)
		if err != nil {
			t.Fatal(err)
		}
		value, _, err := p.Eval(values)
		if err != nil {
			t.Fatal(err)
		}
		return value.Value()
	}
	fixture := `{"metadata":{"labels":{"app":"systemserver"},"annotations":{"traffic.olares.io/mtls-mode":"Required","linkerd.io/inject":"enabled","config.linkerd.io/default-inbound-policy":"deny","config.linkerd.io/proxy-enable-native-sidecar":"true"}},"spec":{"serviceAccountName":"bytetrade-sys-ops","containers":[{"name":"system-server","env":[{"name":"NONCE_AUTHORIZATION_MODE","value":"linkerd"}]}],"initContainers":[{"name":"linkerd-init"},{"name":"linkerd-proxy","restartPolicy":"Always","startupProbe":{},"env":[{"name":"LINKERD2_PROXY_INBOUND_DEFAULT_POLICY","value":"deny"}]}]}}`
	for _, mode := range []string{"valid", "auth-compat", "bfl-compat", "entry-compat", "no-inject", "no-enrollment", "no-app-label", "skip-inbound", "skip-outbound", "skip-subnet", "host-network", "host-port", "missing-proxy", "proxy-not-deny", "default-not-deny", "missing-native", "nonce-mode-disabled"} {
		t.Run(mode, func(t *testing.T) {
			var object map[string]any
			if err := json.Unmarshal([]byte(fixture), &object); err != nil {
				t.Fatal(err)
			}
			meta := object["metadata"].(map[string]any)
			ann := meta["annotations"].(map[string]any)
			spec := object["spec"].(map[string]any)
			namespace := "user-system-alice"
			switch mode {
			case "auth-compat":
				namespace = "os-framework"
				meta["labels"] = map[string]any{"app": "authelia-backend"}
				ann["config.linkerd.io/default-inbound-policy"] = "cluster-authenticated"
				spec["initContainers"].([]any)[1].(map[string]any)["env"] = []any{map[string]any{"name": "LINKERD2_PROXY_INBOUND_DEFAULT_POLICY", "value": "cluster-authenticated"}}
			case "bfl-compat":
				namespace = "user-space-alice"
				meta["labels"] = map[string]any{"tier": "bfl", "traffic.olares.io/bfl-api-mtls": "required"}
			case "entry-compat":
				namespace = "os-gateway"
				meta["labels"] = map[string]any{"gateway.envoyproxy.io/owning-gateway-name": "app-access-internal"}

			case "no-inject":
				delete(ann, "linkerd.io/inject")
			case "no-enrollment":
				delete(meta, "annotations")
			case "no-app-label":
				delete(meta, "labels")
			case "skip-inbound":
				ann["config.linkerd.io/skip-inbound-ports"] = "80"
			case "skip-outbound":
				ann["config.linkerd.io/skip-outbound-ports"] = "80"
			case "skip-subnet":
				ann["config.linkerd.io/skip-subnets"] = "10.0.0.0/8"
			case "host-network":
				spec["hostNetwork"] = true
			case "host-port":
				spec["containers"].([]any)[0].(map[string]any)["ports"] = []any{map[string]any{"hostPort": 80}}
			case "missing-proxy":
				delete(spec, "initContainers")
			case "proxy-not-deny":
				spec["initContainers"].([]any)[1].(map[string]any)["env"] = []any{map[string]any{"name": "LINKERD2_PROXY_INBOUND_DEFAULT_POLICY", "value": "cluster-authenticated"}}
			case "default-not-deny":
				ann["config.linkerd.io/default-inbound-policy"] = "cluster-authenticated"
			case "missing-native":
				delete(ann, "config.linkerd.io/proxy-enable-native-sidecar")
			case "nonce-mode-disabled":
				spec["containers"].([]any)[0].(map[string]any)["env"] = []any{}
			}
			variables := map[string]any{}
			values := map[string]any{"object": object, "oldObject": nil, "request": map[string]any{"namespace": namespace}, "variables": variables}
			if eval(policy.Spec.MatchConditions[0].Expression, values) != true {
				t.Fatal("system-server escaped admission selection")
			}
			for _, v := range policy.Spec.Variables {
				variables[v.Name] = eval(v.Expression, values)
			}
			allowed := true
			for _, v := range policy.Spec.Validations {
				if eval(v.Expression, values) != true {
					allowed = false
					break
				}
			}
			if allowed != (mode == "valid" || strings.HasSuffix(mode, "-compat")) {
				t.Fatalf("admission boundary failed: %s", mode)
			}
		})
	}
}
