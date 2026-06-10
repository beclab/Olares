package terminus

import (
	"testing"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"
)

func TestBuildAppGatewayHelmValuesDefaults(t *testing.T) {
	vals := buildAppGatewayHelmValues("app-gateway", agwconfig.Defaults{})

	if vals["namespace"] != "app-gateway" {
		t.Errorf("namespace = %v, want app-gateway", vals["namespace"])
	}
	if vals["namespaceCreate"] != false {
		t.Errorf("namespaceCreate = %v, want false", vals["namespaceCreate"])
	}

	gw, ok := vals["gateway"].(map[string]interface{})
	if !ok {
		t.Fatal("gateway block missing")
	}
	if gw["name"] != "app-gateway" || gw["gatewayClassName"] != "olares-app-gateway" {
		t.Errorf("gateway defaults wrong: %v", gw)
	}

	// Lite invariant: no service-mesh keys are ever surfaced into Helm values.
	for _, k := range []string{"mesh", "linkerd", "extAuthz", "demo"} {
		if _, present := vals[k]; present {
			t.Errorf("unexpected mesh-related key %q present in lite values", k)
		}
	}
}
