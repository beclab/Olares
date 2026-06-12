package appcfg

import "testing"

func TestGetAppIDUserApp(t *testing.T) {
	// With no SYS_APPS configured, "nginx" is a user app: the ID is the first
	// 8 hex chars of its md5 and is stable across calls.
	id := AppName("nginx").GetAppID()
	if len(id) != 8 {
		t.Fatalf("user app id length=%d want 8 (%q)", len(id), id)
	}
	if id != AppName("nginx").GetAppID() {
		t.Error("GetAppID is not deterministic")
	}
	if AppName("nginx").GetAppID() == AppName("redis").GetAppID() {
		t.Error("different app names produced the same id")
	}
}

func TestGetAppIDSystemApp(t *testing.T) {
	// System apps use their raw name as the ID.
	t.Setenv("SYS_APPS", "market,settings")
	if got := AppName("market").GetAppID(); got != "market" {
		t.Errorf("system app id=%q want market", got)
	}
}
