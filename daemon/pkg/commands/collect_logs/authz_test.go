package collectlogs

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func fakeClient(owner string, owned ...string) kubernetes.Interface {
	objs := make([]runtime.Object, 0, len(owned))
	for _, ns := range owned {
		objs = append(objs, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ns,
				Labels: map[string]string{security.NamespaceOwnerLabel: owner},
			},
		})
	}
	return fake.NewSimpleClientset(objs...)
}

func TestAuthorizeOwnerWildcard(t *testing.T) {
	p := &Param{
		CallerUsername: "alice",
		CallerRole:     utils.Owner,
		Systemd:        SystemdGroup{Components: []string{Wildcard}, Since: "1h", MaxLines: 100},
		Host:           HostGroup{Dmesg: true, Network: true},
		Cluster:        ClusterGroup{Info: true},
		Namespaces:     NamespacesGroup{Names: []string{Wildcard}},
	}
	rs, err := authorize(context.Background(), fakeClient("alice"), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rs.collectSystemd || rs.systemdComponents != nil {
		t.Errorf("systemd: want all (nil components), got %+v collect=%v", rs.systemdComponents, rs.collectSystemd)
	}
	if rs.since != "1h" || rs.maxLines != 100 {
		t.Errorf("since/maxLines not propagated: %q %d", rs.since, rs.maxLines)
	}
	if !rs.collectDmesg || !rs.collectNetwork || !rs.collectClusterInfo {
		t.Errorf("host/cluster scopes not granted: %+v", rs)
	}
	if !rs.collectPods || !rs.allNamespaces {
		t.Errorf("owner wildcard namespaces should map to allNamespaces: %+v", rs)
	}
}

func TestAuthorizeOwnerConcrete(t *testing.T) {
	p := &Param{
		CallerUsername: "alice",
		CallerRole:     utils.Owner,
		Systemd:        SystemdGroup{Components: []string{"k3s", "containerd"}},
		Namespaces:     NamespacesGroup{Names: []string{"a", "b"}},
	}
	rs, err := authorize(context.Background(), fakeClient("alice"), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(rs.systemdComponents, []string{"k3s", "containerd"}) {
		t.Errorf("systemd components: got %+v", rs.systemdComponents)
	}
	if rs.allNamespaces {
		t.Errorf("concrete namespaces should not be allNamespaces")
	}
	if !reflect.DeepEqual(rs.podNamespaces, []string{"a", "b"}) {
		t.Errorf("podNamespaces: got %+v", rs.podNamespaces)
	}
}

func TestAuthorizeNormalDeniedPrivilegedScopes(t *testing.T) {
	cases := map[string]*Param{
		"systemd": {CallerUsername: "bob", CallerRole: utils.Normal, Systemd: SystemdGroup{Components: []string{"k3s"}}},
		"dmesg":   {CallerUsername: "bob", CallerRole: utils.Normal, Host: HostGroup{Dmesg: true}},
		"network": {CallerUsername: "bob", CallerRole: utils.Normal, Host: HostGroup{Network: true}},
		"cluster": {CallerUsername: "bob", CallerRole: utils.Normal, Cluster: ClusterGroup{Info: true}},
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := authorize(context.Background(), fakeClient("bob"), p)
			var denied *ScopeDeniedError
			if !errors.As(err, &denied) {
				t.Fatalf("want ScopeDeniedError, got %v", err)
			}
		})
	}
}

func TestAuthorizeNormalNamespaceWildcardOwned(t *testing.T) {
	p := &Param{
		CallerUsername: "bob",
		CallerRole:     utils.Normal,
		Namespaces:     NamespacesGroup{Names: []string{Wildcard}},
	}
	rs, err := authorize(context.Background(), fakeClient("bob", "user-space-bob", "bob-app"), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rs.allNamespaces {
		t.Errorf("normal user must never get allNamespaces")
	}
	got := append([]string(nil), rs.podNamespaces...)
	sort.Strings(got)
	want := []string{"bob-app", "user-space-bob"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("podNamespaces: got %+v want %+v", got, want)
	}
}

func TestAuthorizeNormalNamespaceUnownedDenied(t *testing.T) {
	p := &Param{
		CallerUsername: "bob",
		CallerRole:     utils.Normal,
		Namespaces:     NamespacesGroup{Names: []string{"bob-app", "alice-secret"}},
	}
	_, err := authorize(context.Background(), fakeClient("bob", "bob-app"), p)
	var denied *ScopeDeniedError
	if !errors.As(err, &denied) {
		t.Fatalf("want ScopeDeniedError for unowned namespace, got %v", err)
	}
}

func TestAuthorizeNormalWildcardOwnsNothing(t *testing.T) {
	// Privilege-escalation guard: a normal user who owns nothing must not fall
	// through to collecting every namespace.
	p := &Param{
		CallerUsername: "bob",
		CallerRole:     utils.Normal,
		Namespaces:     NamespacesGroup{Names: []string{Wildcard}},
	}
	_, err := authorize(context.Background(), fakeClient("bob"), p)
	if !errors.Is(err, ErrNothingRequested) {
		t.Fatalf("want ErrNothingRequested, got %v", err)
	}
}

func TestAuthorizeEmptyRequest(t *testing.T) {
	p := &Param{CallerUsername: "alice", CallerRole: utils.Owner}
	_, err := authorize(context.Background(), fakeClient("alice"), p)
	if !errors.Is(err, ErrNothingRequested) {
		t.Fatalf("want ErrNothingRequested, got %v", err)
	}
}

func TestAuthorizeMissingIdentity(t *testing.T) {
	p := &Param{CallerRole: utils.Owner, Cluster: ClusterGroup{Info: true}}
	_, err := authorize(context.Background(), fakeClient(""), p)
	if err == nil {
		t.Fatalf("want error for missing caller identity")
	}
}
