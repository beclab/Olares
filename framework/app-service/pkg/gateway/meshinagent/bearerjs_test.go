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

// TestBearerJSDecideOffloadUsesTLSHostsOnly pins DEFECT-SH-MESHIN-TLS-01: HTTPS
// offload must not gate on a hard-coded platform suffix (e.g. .olares.com), so
// zones like .olares.cn terminate when listed in tls-hosts.
func TestBearerJSDecideOffloadUsesTLSHostsOnly(t *testing.T) {
	js := BearerJS()
	for _, ban := range []string{
		"isPlatformHost",
		"PLATFORM_SUFFIX",
	} {
		if strings.Contains(js, ban) {
			t.Fatalf("bearer.js must not contain %q after tls-hosts-only offload", ban)
		}
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
	for _, want := range []string{
		"reloadTLSHosts()",
		"return TERMINATE",
		"return passthrough(host)",
	} {
		if !strings.Contains(fn, want) {
			t.Fatalf("decideOffload missing %q", want)
		}
	}
}
