package app

import "testing"

func TestFmtAppMgrNameWithExplicitNamespace(t *testing.T) {
	// A non user-space/user-system namespace is used verbatim, so the manager
	// name is "{ns}-{app}" and no API client lookup is needed.
	got, err := FmtAppMgrName("nginx", "alice", "nginx-ns")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "nginx-ns-nginx" {
		t.Errorf("FmtAppMgrName=%q want nginx-ns-nginx", got)
	}
}

func TestFmtAppMgrNameUserSpaceNamespace(t *testing.T) {
	got, err := FmtAppMgrName("nginx", "alice", "user-space-alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "user-space-alice-nginx" {
		t.Errorf("FmtAppMgrName=%q want user-space-alice-nginx", got)
	}
}
