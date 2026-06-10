package gateway_test

import (
	"context"
	"reflect"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
)

// TC-T1-2-01: characterization for gateway.NamespaceOptedIntoGateway, byte-equal
// with the former (*CallerReconciler).namespaceOptedIntoGateway behavior
// (annotation hit + non-empty clusterAppRef -> true; everything else -> false).

func newOptInApp(name, ns, inClusterValue, clusterAppRef string) *appv1alpha1.Application {
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      name,
			Namespace: ns,
			Settings:  map[string]string{},
		},
	}
	if inClusterValue != "" {
		app.Annotations = map[string]string{
			gateway.AnnotationInCluster: inClusterValue,
		}
	}
	if clusterAppRef != "" {
		app.Spec.Settings["clusterAppRef"] = clusterAppRef
	}
	return app
}

func newFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("clientgo scheme: %v", err)
	}
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("app scheme: %v", err)
	}
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		Build()
}

func TestNamespaceOptedIntoGateway_TCT12_01(t *testing.T) {
	tests := []struct {
		name string
		ns   string
		objs []client.Object
		want bool
	}{
		{
			name: "hit annotation + non-empty clusterAppRef => true",
			ns:   "litellm-alice",
			objs: []client.Object{
				newOptInApp("litellm", "litellm-alice", gateway.InClusterGateway, "ollama"),
			},
			want: true,
		},
		{
			name: "hit annotation + empty clusterAppRef => false",
			ns:   "litellm-alice",
			objs: []client.Object{
				newOptInApp("litellm", "litellm-alice", gateway.InClusterGateway, ""),
			},
			want: false,
		},
		{
			name: "missing annotation + non-empty clusterAppRef => false",
			ns:   "litellm-alice",
			objs: []client.Object{
				newOptInApp("litellm", "litellm-alice", "", "ollama"),
			},
			want: false,
		},
		{
			name: "annotation direct (not gateway) => false",
			ns:   "litellm-alice",
			objs: []client.Object{
				newOptInApp("litellm", "litellm-alice", gateway.InClusterDirect, "ollama"),
			},
			want: false,
		},
		{
			name: "annotation case-insensitive (EqualFold) => true",
			ns:   "litellm-alice",
			objs: []client.Object{
				newOptInApp("litellm", "litellm-alice", "Gateway", "ollama"),
			},
			want: true,
		},
		{
			name: "annotation trimmed (whitespace tolerant) => true",
			ns:   "litellm-alice",
			objs: []client.Object{
				newOptInApp("litellm", "litellm-alice", "  gateway  ", "ollama"),
			},
			want: true,
		},
		{
			name: "different namespace => false",
			ns:   "litellm-alice",
			objs: []client.Object{
				newOptInApp("litellm", "litellm-bob", gateway.InClusterGateway, "ollama"),
			},
			want: false,
		},
		{
			name: "empty Application list => false",
			ns:   "litellm-alice",
			objs: nil,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newFakeClient(t, tc.objs...)
			got, err := gateway.NamespaceOptedIntoGateway(context.Background(), c, tc.ns)
			if err != nil {
				t.Fatalf("NamespaceOptedIntoGateway err: %v", err)
			}
			if got != tc.want {
				t.Fatalf("NamespaceOptedIntoGateway(%q) = %v, want %v", tc.ns, got, tc.want)
			}
		})
	}
}

// TC-T1-2-02: characterization for the three resolve helpers
// (BuildClusterAppOwnerIndex / ResolveClusterAppOwner / SplitClusterAppRefs).

func newClusterApp(name, owner string, clusterScoped string) appv1alpha1.Application {
	return appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: name + "-" + owner},
		Spec: appv1alpha1.ApplicationSpec{
			Name:  name,
			Owner: owner,
			Settings: map[string]string{
				"clusterScoped": clusterScoped,
			},
		},
	}
}

// newV3SharedApp builds a shared v3 app (shared marker + shared entrances),
// which qualifies as a shared server without settings.clusterScoped.
func newV3SharedApp(name, owner string) appv1alpha1.Application {
	return appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-" + owner,
			Labels: map[string]string{
				constants.AppApiVersionLabel: constants.AppVersionV3,
				constants.AppSharedLabel:     constants.AppSharedTrue,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:            name,
			Owner:           owner,
			SharedEntrances: []appv1alpha1.Entrance{{Name: "api", Host: "svc"}},
		},
	}
}

func TestBuildClusterAppOwnerIndex_TCT12_02(t *testing.T) {
	tests := []struct {
		name string
		apps []appv1alpha1.Application
		want gateway.ClusterAppOwnerIndex
	}{
		{
			name: "empty input => empty index",
			apps: nil,
			want: gateway.ClusterAppOwnerIndex{},
		},
		{
			name: "single clusterScoped app indexed",
			apps: []appv1alpha1.Application{
				newClusterApp("ollama", "alice", "true"),
			},
			want: gateway.ClusterAppOwnerIndex{"ollama": "alice"},
		},
		{
			name: "non-clusterScoped app skipped",
			apps: []appv1alpha1.Application{
				newClusterApp("ollama", "alice", ""),
			},
			want: gateway.ClusterAppOwnerIndex{},
		},
		{
			name: "v3 shared app indexed without clusterScoped",
			apps: []appv1alpha1.Application{
				newV3SharedApp("ollamav3", "alice"),
			},
			want: gateway.ClusterAppOwnerIndex{"ollamav3": "alice"},
		},
		{
			name: "v3 app without shared entrances skipped",
			apps: []appv1alpha1.Application{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "novel-bob",
						Labels: map[string]string{constants.AppApiVersionLabel: constants.AppVersionV3},
					},
					Spec: appv1alpha1.ApplicationSpec{Name: "novel", Owner: "bob"},
				},
			},
			want: gateway.ClusterAppOwnerIndex{},
		},
		{
			name: "empty name skipped",
			apps: []appv1alpha1.Application{
				newClusterApp("", "alice", "true"),
			},
			want: gateway.ClusterAppOwnerIndex{},
		},
		{
			name: "empty owner skipped",
			apps: []appv1alpha1.Application{
				newClusterApp("ollama", "", "true"),
			},
			want: gateway.ClusterAppOwnerIndex{},
		},
		{
			name: "two distinct names indexed independently",
			apps: []appv1alpha1.Application{
				newClusterApp("ollama", "alice", "true"),
				newClusterApp("redis", "bob", "true"),
			},
			want: gateway.ClusterAppOwnerIndex{
				"ollama": "alice",
				"redis":  "bob",
			},
		},
		{
			name: "same name two owners => warn-only comma-joined",
			apps: []appv1alpha1.Application{
				newClusterApp("dup", "bob", "true"),
				newClusterApp("dup", "charlie", "true"),
			},
			want: gateway.ClusterAppOwnerIndex{"dup": "bob,charlie"},
		},
		{
			name: "same name same owner duplicated => no change",
			apps: []appv1alpha1.Application{
				newClusterApp("dup", "bob", "true"),
				newClusterApp("dup", "bob", "true"),
			},
			want: gateway.ClusterAppOwnerIndex{"dup": "bob"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := gateway.BuildClusterAppOwnerIndex(tc.apps)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("BuildClusterAppOwnerIndex: got %v want %v", got, tc.want)
			}
		})
	}
}

func TestResolveClusterAppOwner_TCT12_02(t *testing.T) {
	idx := gateway.ClusterAppOwnerIndex{
		"ollama": "alice",
		"dup":    "bob,charlie",
	}
	tests := []struct {
		name   string
		idx    gateway.ClusterAppOwnerIndex
		appRef string
		want   string
	}{
		{"nil idx => empty", nil, "ollama", ""},
		{"empty idx => empty", gateway.ClusterAppOwnerIndex{}, "ollama", ""},
		{"hit => owner", idx, "ollama", "alice"},
		{"miss => empty", idx, "ghost", ""},
		{"whitespace appRef trimmed", idx, "  ollama  ", "alice"},
		{"comma-joined multi-owner pass-through", idx, "dup", "bob,charlie"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := gateway.ResolveClusterAppOwner(tc.idx, tc.appRef)
			if got != tc.want {
				t.Fatalf("ResolveClusterAppOwner(%q) = %q want %q", tc.appRef, got, tc.want)
			}
		})
	}
}

func TestSplitClusterAppRefs_TCT12_02(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{"empty => nil", "", nil},
		{"whitespace-only => nil", "   ", nil},
		{"single ref", "ollama", []string{"ollama"}},
		{"multi refs order preserved (no sort)", "wise,calendar,ollama", []string{"wise", "calendar", "ollama"}},
		{"trims whitespace per entry", " wise , calendar ", []string{"wise", "calendar"}},
		{"skips empty entries", "wise,,calendar,", []string{"wise", "calendar"}},
		{"duplicates preserved (no dedupe)", "wise,wise", []string{"wise", "wise"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := gateway.SplitClusterAppRefs(tc.raw)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("SplitClusterAppRefs(%q) = %v want %v", tc.raw, got, tc.want)
			}
		})
	}
}

// TC-T1-2-02 extension: test seam hooks fire when set (cross-package wire-up
// regression that backstops TC-402d in routecontrol).
func TestExportedTestHooksFire_TCT12_02(t *testing.T) {
	buildCalls := 0
	resolveCalls := 0
	gateway.TestBuildClusterAppOwnerIndexHook = func() { buildCalls++ }
	gateway.TestResolveClusterAppOwnerHook = func() { resolveCalls++ }
	defer func() {
		gateway.TestBuildClusterAppOwnerIndexHook = nil
		gateway.TestResolveClusterAppOwnerHook = nil
	}()

	idx := gateway.BuildClusterAppOwnerIndex([]appv1alpha1.Application{
		newClusterApp("ollama", "alice", "true"),
	})
	_ = gateway.ResolveClusterAppOwner(idx, "ollama")
	_ = gateway.ResolveClusterAppOwner(idx, "ghost")

	if buildCalls != 1 {
		t.Fatalf("TestBuildClusterAppOwnerIndexHook fired %d times, want 1", buildCalls)
	}
	if resolveCalls != 2 {
		t.Fatalf("TestResolveClusterAppOwnerHook fired %d times, want 2", resolveCalls)
	}
}
