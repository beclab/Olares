package pod

import (
	"testing"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/picker"
	"github.com/beclab/Olares/cli/pkg/clusterctx"
)

func TestFilterExecEntries_AdminDropsSubAccounts(t *testing.T) {
	entries := []picker.Entry{
		{Namespace: "user-space-admin", Pod: "a", Container: "c"},
		{Namespace: "user-space-alice", Pod: "b", Container: "c"},
		{Namespace: "kube-system", Pod: "k", Container: "c"},
		{Namespace: "app-shared", Pod: "s", Container: "c"},
	}
	info := clusterctx.Info{
		Username:         "admin",
		GlobalRole:       clusterctx.GlobalRoleAdmin,
		SystemNamespaces: []string{"kube-system"},
	}
	got := filterExecEntries(entries, info)
	if len(got) != 3 {
		t.Fatalf("want 3 exec-able entries (own + system + shared), got %d", len(got))
	}
	for _, e := range got {
		if e.Namespace == "user-space-alice" {
			t.Fatalf("admin must not be offered the sub-account namespace %q", e.Namespace)
		}
	}
}

func TestFilterExecEntries_NonAdminOwnOnly(t *testing.T) {
	entries := []picker.Entry{
		{Namespace: "user-space-alice", Pod: "a", Container: "c"},
		{Namespace: "user-space-admin", Pod: "b", Container: "c"},
		{Namespace: "kube-system", Pod: "k", Container: "c"},
	}
	info := clusterctx.Info{Username: "alice", GlobalRole: "platform-regular"}
	got := filterExecEntries(entries, info)
	if len(got) != 1 || got[0].Namespace != "user-space-alice" {
		t.Fatalf("non-admin should only keep their own namespace, got %+v", got)
	}
}

func TestFilterExecEntries_EmptyIdentityKeepsAll(t *testing.T) {
	entries := []picker.Entry{
		{Namespace: "user-space-alice", Pod: "a", Container: "c"},
	}
	got := filterExecEntries(entries, clusterctx.Info{})
	if len(got) != 1 {
		t.Fatalf("unresolved identity must leave entries intact for RunExec to gate, got %d", len(got))
	}
}
