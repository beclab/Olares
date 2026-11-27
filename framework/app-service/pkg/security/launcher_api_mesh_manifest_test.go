package security

import (
	"reflect"
	"regexp"
	"testing"
)

func TestLauncherAppServiceCapabilitiesFollowTenant(t *testing.T) {
	const template = "../../../bfl/.olares/config/launcher/templates/bfl-tenant-api-mtls.yaml"
	for _, tenant := range []string{"alice", "bob", "admins"} {
		t.Run(tenant, func(t *testing.T) {
			render := func(namespace string, prepare bool) (map[string]map[string]any, error) {
				t.Helper()
				return renderMeshManifestsInNamespace(t, namespace, false, map[string]any{"bfl": map[string]any{
					"username": tenant, "prepareAPIMesh": prepare,
				}}, template)
			}
			ns := "user-space-" + tenant
			name := "launcher-api-" + tenant
			docs, err := render(ns, true)
			if err != nil || len(docs) != 3 {
				t.Fatalf("expected one tenant API capability, got %d: %v", len(docs), err)
			}
			route := docs["HTTPRoute/"+name]["spec"].(map[string]any)
			matches := route["rules"].([]any)[0].(map[string]any)["matches"].([]any)
			checkAppServiceCapability(t, docs, name, "user-system-"+tenant, "bytetrade-sys-ops", matches)
			for _, raw := range matches {
				match := raw.(map[string]any)
				want := appServiceTenantMatch(tenant, match["method"].(string), match["path"].(map[string]any)["type"].(string), match["path"].(map[string]any)["value"].(string))
				if !reflect.DeepEqual(match, want) {
					t.Fatalf("tenant selector missing from %v", match)
				}
			}
			allowed := func(method, path string) bool {
				t.Helper()
				for _, raw := range matches {
					match := raw.(map[string]any)
					if match["method"] != method {
						continue
					}
					p := match["path"].(map[string]any)
					switch v := p["value"].(string); p["type"] {
					case "Exact":
						if v == path {
							return true
						}
					case "RegularExpression":
						if regexp.MustCompile(v).MatchString(path) {
							return true
						}
					default:
						t.Fatalf("unexpected path match: %v", p)
					}
				}
				return false
			}
			for _, check := range []struct {
				method, path string
				want         bool
			}{
				{"GET", "/apps", true}, {"GET", "/control-hub/apps", true},
				{"GET", "/apps/router/env", true}, {"GET", "/apps/router/subject", true},
				{"GET", "/applications/router/api/setup", true}, {"GET", "/users/" + tenant + "/metrics", true},
				{"GET", "/users/" + tenant, tenant != "admins"},
				{"POST", "/appenv/remote-options-proxy", true},
				{"GET", "/users/admins", false}, {"GET", "/all/apps", false},
				{"GET", "/users/foreign/metrics", false}, {"GET", "/apps/router/env/extra", false},
				{"POST", "/users", false}, {"POST", "/apps/router/install", false},
				{"DELETE", "/users/bob", false}, {"PUT", "/systemenvs", false},
			} {
				if got := allowed(check.method, "/app-service/v1"+check.path); got != check.want {
					t.Errorf("%s %s allowed=%v, want %v", check.method, check.path, got, check.want)
				}
			}
			if _, err := render("user-space-foreign", true); err == nil {
				t.Fatal("wrong tenant namespace rendered an authorization")
			}
			if unprepared, err := render(ns, false); err != nil || len(unprepared) != 0 {
				t.Fatalf("unprepared launcher rendered authorizations: %v, %v", unprepared, err)
			}
		})
	}
}
