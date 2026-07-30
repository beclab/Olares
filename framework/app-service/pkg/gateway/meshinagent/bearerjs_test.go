package meshinagent

import (
	"strings"
	"testing"
)

func TestBearerJSNoSNILogsError(t *testing.T) {
	js := BearerJS()
	if !strings.Contains(js, `s.error('mesh-in: no SNI on hijacked 443`) {
		t.Fatal("decideOffload must log empty SNI via s.error")
	}
	idx := strings.Index(js, "function decideOffload")
	if idx < 0 {
		t.Fatal("missing decideOffload")
	}
	fn := js[idx:]
	end := strings.Index(fn, "\nexport default")
	if end > 0 {
		fn = fn[:end]
	}
	if !strings.Contains(fn, "if (!host)") {
		t.Fatal("empty host branch required")
	}
	if !strings.Contains(fn, "return passthrough(host)") {
		t.Fatal("empty host must still passthrough to invalid.local")
	}
}

func TestBearerJSDecideOffloadKeepsPlatformPath(t *testing.T) {
	js := BearerJS()
	for _, want := range []string{
		"isPlatformHost(host)",
		"reloadTLSHosts()",
		"return TERMINATE",
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("decideOffload missing %q", want)
		}
	}
}
