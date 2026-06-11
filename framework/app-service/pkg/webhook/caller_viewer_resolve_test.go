package webhook

import (
	"errors"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// callerApp builds a caller-mode opt-in Application (Annotations[in-cluster]=
// gateway + Settings.clusterAppRef=ref). Spec.Namespace = ns (caller NS), and
// Spec.Name == metadata.name for readability.
func callerApp(name, ns, ref string) v1alpha1.Application {
	return v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				gateway.AnnotationInCluster: gateway.InClusterGateway,
			},
		},
		Spec: v1alpha1.ApplicationSpec{
			Name:      name,
			Namespace: ns,
			Settings:  map[string]string{"clusterAppRef": ref},
		},
	}
}

// clusterApp builds a cluster-scoped Shared Application that contributes
// <name>->owner into BuildClusterAppOwnerIndex.
func clusterApp(name, owner string) v1alpha1.Application {
	return v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.ApplicationSpec{
			Name:     name,
			Owner:    owner,
			Settings: map[string]string{"clusterScoped": "true"},
		},
	}
}

// TC-T1-5-01: single ref -> ViewerSet has the cluster-app owner (lowercased
// per allowset contract); PrimaryViewer keeps the raw owner-index value so
// downstream callers can join multi-owner strings without losing original
// case; renderCallerAllowset round-trips ViewerSet unchanged.
func TestResolveCallerViewers_SingleRef(t *testing.T) {
	apps := []v1alpha1.Application{
		callerApp("litellm", "litellm-userA", "ollama"),
		clusterApp("ollama", "userA"),
	}

	got, err := resolveCallerViewersFromSnapshot("litellm-userA", apps)
	require.NoError(t, err)
	require.Equal(t, []string{"usera"}, got.ViewerSet)
	require.Equal(t, "ollama", got.PrimaryRef)
	require.Equal(t, "userA", got.PrimaryViewer)
	require.Equal(t, []string{"usera"}, renderCallerAllowset(got.ViewerSet))
}

// TC-T1-5-02: multi-ref MVP -> primary = sorted-first ref; viewerSet aggregates
// every ref's owner; multi_ref_unsupported metric increments exactly once.
func TestResolveCallerViewers_MultiRefPrimaryAndMetric(t *testing.T) {
	before := testutil.ToFloat64(d2InjectSkippedTotal.WithLabelValues(d2SkipReasonMultiRefUnsupported))

	apps := []v1alpha1.Application{
		callerApp("litellm", "litellm-userA", "ollama,redis"),
		clusterApp("ollama", "userA"),
		clusterApp("redis", "userB"),
	}
	got, err := resolveCallerViewersFromSnapshot("litellm-userA", apps)
	require.NoError(t, err)

	require.Equal(t, "ollama", got.PrimaryRef)
	require.Equal(t, "userA", got.PrimaryViewer)
	require.Equal(t, []string{"usera", "userb"}, got.ViewerSet)
	require.Equal(t, []string{"usera", "userb"}, renderCallerAllowset(got.ViewerSet))

	after := testutil.ToFloat64(d2InjectSkippedTotal.WithLabelValues(d2SkipReasonMultiRefUnsupported))
	require.Equal(t, before+1, after)
}

// TC-T1-5-03: unresolved -> ErrD2CallerViewerUnresolved sentinel; ClassifyD2SkipReason
// maps it to the caller_viewer_unresolved metric label.
func TestResolveCallerViewers_Unresolved(t *testing.T) {
	apps := []v1alpha1.Application{
		callerApp("litellm", "litellm-userA", "ghost"),
	}

	got, err := resolveCallerViewersFromSnapshot("litellm-userA", apps)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrD2CallerViewerUnresolved))
	require.Empty(t, got.ViewerSet)
	require.Empty(t, got.PrimaryRef)
	require.Empty(t, got.PrimaryViewer)
	require.Equal(t, d2SkipReasonCallerViewerUnresolved, ClassifyD2SkipReason(err))
}

// TC-T1-5-04: caller-mode path must never invoke the server-mode resolver
// deriveViewerFromPodNS (no shared code path; same-source resolution must come
// from gateway.BuildClusterAppOwnerIndex / ResolveClusterAppOwner).
func TestResolveCallerViewers_DoesNotCallServerModeViewerDerive(t *testing.T) {
	calls := 0
	testDeriveViewerFromPodNSHook = func() { calls++ }
	t.Cleanup(func() { testDeriveViewerFromPodNSHook = nil })

	apps := []v1alpha1.Application{
		callerApp("litellm", "litellm-userA", "ollama"),
		clusterApp("ollama", "userA"),
	}
	_, err := resolveCallerViewersFromSnapshot("litellm-userA", apps)
	require.NoError(t, err)
	require.Equal(t, 0, calls, "caller-mode resolver must not call deriveViewerFromPodNS")
}

// Empty / missing clusterAppRef in the caller NS yields the same sentinel as
// an unresolved ref so the fail-open path reports a single reason family.
func TestResolveCallerViewers_NoOptInAppInNamespace(t *testing.T) {
	apps := []v1alpha1.Application{
		// in NS but no in-cluster annotation
		{
			ObjectMeta: metav1.ObjectMeta{Name: "x"},
			Spec: v1alpha1.ApplicationSpec{
				Name:      "x",
				Namespace: "litellm-userA",
				Settings:  map[string]string{"clusterAppRef": "ollama"},
			},
		},
		// in NS with annotation but empty ref
		callerApp("y", "litellm-userA", ""),
		// matching opt-in but a different NS
		callerApp("z", "other-ns", "ollama"),
		clusterApp("ollama", "userA"),
	}
	_, err := resolveCallerViewersFromSnapshot("litellm-userA", apps)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrD2CallerViewerUnresolved))
}

// Multiple opt-in apps in the same caller NS merge their refs; surrounding
// whitespace on refs is stripped by gateway.SplitClusterAppRefs; comma-joined
// owner values (per BuildClusterAppOwnerIndex multi-owner contract) expand to
// multiple ViewerSet entries; owner case is normalised by the resolver before
// dedupe so case-only collisions collapse to a single viewer.
func TestResolveCallerViewers_MergeAndNormalise(t *testing.T) {
	apps := []v1alpha1.Application{
		callerApp("a", "litellm-userA", " ollama "),
		callerApp("b", "litellm-userA", "redis"),
		// two cluster-scoped apps with the same Name -> comma-joined owner
		clusterApp("ollama", "USERA"),
		clusterApp("ollama", "userC"),
		clusterApp("redis", "userB"),
	}

	got, err := resolveCallerViewersFromSnapshot("litellm-userA", apps)
	require.NoError(t, err)
	require.Equal(t, []string{"usera", "userb", "userc"}, got.ViewerSet)
	require.Equal(t, "ollama", got.PrimaryRef)
	// PrimaryViewer is the raw owner-index value for the primary ref:
	// BuildClusterAppOwnerIndex joins multi-owners with a comma (original case).
	require.Equal(t, "USERA,userC", got.PrimaryViewer)
}

// Empty pod namespace short-circuits to the sentinel (defensive: caller should
// never reach the patch path with an empty pod.Namespace, but the resolver
// stays safe).
func TestResolveCallerViewers_EmptyPodNamespace(t *testing.T) {
	_, err := resolveCallerViewersFromSnapshot("", []v1alpha1.Application{
		callerApp("a", "litellm-userA", "ollama"),
		clusterApp("ollama", "userA"),
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrD2CallerViewerUnresolved))
}

// renderCallerAllowset is a defensive sanity pass: it tolerates an
// unsorted / mixed-case / duplicate-laden input from any future caller and
// produces the WI-N1 §2.3 literal-allowset shape (lowercased + sorted + uniq).
func TestRenderCallerAllowset_Defensive(t *testing.T) {
	require.Empty(t, renderCallerAllowset(nil))
	require.Empty(t, renderCallerAllowset([]string{"", "  "}))
	require.Equal(t, []string{"alice", "bob"}, renderCallerAllowset([]string{"Bob", " alice ", "ALICE"}))
}
