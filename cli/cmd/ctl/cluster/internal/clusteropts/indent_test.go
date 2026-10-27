package clusteropts

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestIndentJSONPreservesEveryField is the reason IndentJSON exists: verbs
// whose typed struct is narrower than the response must not reach for
// PrintJSON, which would drop whatever the struct does not model. Here the
// container carries fields `cluster workload get`'s Workload does not.
func TestIndentJSONPreservesEveryField(t *testing.T) {
	body := []byte(`{"kind":"Deployment","spec":{"template":{"spec":{"containers":[` +
		`{"name":"app","image":"busybox:1.36","args":["-c","sleep"],` +
		`"resources":{"limits":{"cpu":"1"}},"ports":[{"containerPort":8080}]}]}}}}`)

	out, err := IndentJSON(body)
	if err != nil {
		t.Fatalf("IndentJSON: %v", err)
	}
	for _, want := range []string{"args", "resources", "containerPort", "busybox:1.36"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output dropped %q:\n%s", want, out)
		}
	}
	if !strings.HasSuffix(string(out), "\n") {
		t.Error("output must end with a newline")
	}

	// Indentation must not change the decoded value.
	var before, after interface{}
	if err := json.Unmarshal(body, &before); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &after); err != nil {
		t.Fatalf("indented output is not valid JSON: %v", err)
	}
	if string(mustMarshal(t, before)) != string(mustMarshal(t, after)) {
		t.Error("IndentJSON changed the decoded value")
	}
}

func TestIndentJSONRejectsNonJSON(t *testing.T) {
	if _, err := IndentJSON([]byte("not json")); err == nil {
		t.Error("expected an error for a non-JSON body")
	}
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
