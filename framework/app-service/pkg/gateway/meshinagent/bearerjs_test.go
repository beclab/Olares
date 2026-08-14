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

func TestBearerJSPickCertBySNICustomDomain(t *testing.T) {
	js := BearerJS()
	for _, want := range []string{
		"CUSTOM_CERTS_DIR",
		"CUSTOM_TLS_HOSTS_FILE",
		CustomCertsMountPath,
		"function customTLSRequired",
		"function pickCert",
		"REJECT_CERT",
		"ssl_server_name",
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("bearer.js missing %q", want)
		}
	}
	idx := strings.Index(js, "function pickCert")
	if idx < 0 {
		t.Fatal("missing pickCert")
	}
	fn := js[idx:]
	end := strings.Index(fn, "\nfunction pickKey")
	if end > 0 {
		fn = fn[:end]
	}
	if !strings.Contains(fn, "return pair.cert") {
		t.Fatal("pickCert must return custom cert path when ready")
	}
	if !strings.Contains(fn, "customTLSRequired(host)") || !strings.Contains(fn, "return REJECT_CERT") {
		t.Fatal("pickCert must fail closed for third-party FQDN without material")
	}
	if !strings.Contains(fn, "REAL_CERT") {
		t.Fatal("pickCert must fall back to viewer REAL_CERT for platform hosts")
	}
}

func TestBearerJSTlsModeCustomBypassesPlaceholder(t *testing.T) {
	js := BearerJS()
	idx := strings.Index(js, "function tlsMode")
	if idx < 0 {
		t.Fatal("missing tlsMode")
	}
	fn := js[idx:]
	end := strings.Index(fn, "\nfunction pickCert")
	if end > 0 {
		fn = fn[:end]
	}
	if !strings.Contains(fn, "customTLSRequired(host)") {
		t.Fatal("tlsMode must gate third-party hosts separately from viewer replica")
	}
	if !strings.Contains(fn, "'reject'") {
		t.Fatal("tlsMode must expose reject when custom material is missing")
	}
}
