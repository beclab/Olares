package systemcomponents

import "testing"

func TestDefaultReadinessDoesNotWaitForPersonalServices(t *testing.T) {
	bfl := false
	network := map[string]bool{"headscale": false, "tailscale": false}
	for _, c := range Default() {
		if c.Name == "bfl" && c.Namespace == NamespaceOsFrontend && c.Kind == Deployment && c.Presence == Required {
			bfl = true
		}
		if c.isPerUser() {
			t.Fatalf("requires a user before setup: %+v", c)
		}
		if c.Name == "wizard" {
			t.Fatalf("waits for user activation: %+v", c)
		}
		if _, ok := network[c.Name]; ok {
			if c.Namespace != NamespaceOsNetwork || c.Kind != Deployment || c.Presence != Required {
				t.Fatalf("network service must be a required global deployment: %+v", c)
			}
			network[c.Name] = true
		}
		if c.Namespace == "user-space" || c.Namespace == "user-system" {
			t.Fatalf("base readiness assumes a shared user component: %+v", c)
		}
	}
	if !bfl {
		t.Fatal("missing shared BFL")
	}
	for name, found := range network {
		if !found {
			t.Errorf("missing global network service %s", name)
		}
	}
}
