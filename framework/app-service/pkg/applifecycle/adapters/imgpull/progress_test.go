package imgpull_test

import (
	"context"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/imgpull"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

// resolverStub returns nil, nil — Progress never invokes it, but
// New requires a non-nil resolver.
func resolverStub(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
	return nil, nil
}

// -----------------------------------------------------------------
// Progress: top-level port method
// -----------------------------------------------------------------

func TestProgress_EmptyName(t *testing.T) {
	p := imgpull.New(testutil.NewFakeClient(), resolverStub)
	got, err := p.Progress(context.Background(), "")
	if err != nil {
		t.Fatalf("Progress(\"\"): %v", err)
	}
	if got != (ports.ImageProgress{}) {
		t.Errorf("empty name: got %+v want zero", got)
	}
}

func TestProgress_NotFound(t *testing.T) {
	p := imgpull.New(testutil.NewFakeClient(), resolverStub)
	got, err := p.Progress(context.Background(), "missing")
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
	// Port docs: zero value + nil error means "no IM CR exists".
	if got != (ports.ImageProgress{}) {
		t.Errorf("NotFound: got %+v want zero", got)
	}
}

func TestProgress_Failed(t *testing.T) {
	im := &appv1.ImageManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status: appv1.ImageManagerStatus{
			State:   "failed",
			Message: "pull failed after 6 attempts",
		},
	}
	p := imgpull.New(testutil.NewFakeClient(im), resolverStub)
	got, err := p.Progress(context.Background(), "am")
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
	if !got.Failed {
		t.Errorf("Failed: got %v want true", got.Failed)
	}
	if got.Complete {
		t.Errorf("Complete: got %v want false", got.Complete)
	}
	if got.Message != "pull failed after 6 attempts" {
		t.Errorf("Message: %q", got.Message)
	}
}

func TestProgress_Completed(t *testing.T) {
	im := &appv1.ImageManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status: appv1.ImageManagerStatus{
			State: "completed",
		},
	}
	p := imgpull.New(testutil.NewFakeClient(im), resolverStub)
	got, err := p.Progress(context.Background(), "am")
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
	if !got.Complete {
		t.Errorf("Complete: want true, got %+v", got)
	}
	if got.Percent != 100 {
		t.Errorf("Percent: %v want 100", got.Percent)
	}
	if got.Failed {
		t.Errorf("Failed: got true; completed should not be failed")
	}
}

func TestProgress_InProgress(t *testing.T) {
	im := &appv1.ImageManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Spec: appv1.ImageManagerSpec{
			Nodes: []string{"node-1"},
			Refs: []appv1.Ref{
				{Name: "nginx:1.25"},
			},
		},
		Status: appv1.ImageManagerStatus{
			State: "downloading",
			Conditions: map[string]map[string]map[string]string{
				"node-1": {
					"nginx:1.25": {"total": "100", "offset": "50"},
				},
			},
		},
	}
	p := imgpull.New(testutil.NewFakeClient(im), resolverStub)
	got, err := p.Progress(context.Background(), "am")
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
	if got.Percent != 50 {
		t.Errorf("Percent: %v want 50", got.Percent)
	}
	if got.Complete {
		t.Errorf("Complete should be false")
	}
	if got.Failed {
		t.Errorf("Failed should be false")
	}
	if got.Reason != "downloading" {
		t.Errorf("Reason: %q", got.Reason)
	}
}

// -----------------------------------------------------------------
// computeProgress: pure computation edge cases via Progress()
// -----------------------------------------------------------------

func TestProgress_NoNodes_NoRefs(t *testing.T) {
	im := &appv1.ImageManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status:     appv1.ImageManagerStatus{State: "downloading"},
	}
	p := imgpull.New(testutil.NewFakeClient(im), resolverStub)
	got, _ := p.Progress(context.Background(), "am")
	if got.Percent != 0 {
		t.Errorf("empty spec: got %v want 0", got.Percent)
	}
	if got.Complete || got.Failed {
		t.Errorf("empty spec should not be Complete/Failed: %+v", got)
	}
}

func TestProgress_PartialRefs_IsNotComplete(t *testing.T) {
	// Image-manager reports the small/cached image first. Summing
	// only the refs that have a total makes a 2 MB busybox look
	// like the whole pull is done (100%), so awaitImagePull
	// hands off to Installing while firefox is still at 0 bytes.
	// A missing total on any spec ref means that ref has not
	// started — overall progress must stay below 100.
	im := &appv1.ImageManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: "am",
			Annotations: map[string]string{
				constants.ApplicationImageLabel: `[{"name":"docker.io/beclab/busybox:1.37","size":2213754},{"name":"docker.io/beclab/linuxserver-firefox:1151.0.1","size":1129791514}]`,
			},
		},
		Spec: appv1.ImageManagerSpec{
			Nodes: []string{"olares"},
			Refs: []appv1.Ref{
				{Name: "docker.io/beclab/busybox:1.37"},
				{Name: "docker.io/beclab/linuxserver-firefox:1151.0.1"},
			},
		},
		Status: appv1.ImageManagerStatus{
			State: "downloading",
			Conditions: map[string]map[string]map[string]string{
				"olares": {
					"docker.io/beclab/busybox:1.37": {
						"total": "56782302", "offset": "56782302",
					},
				},
			},
		},
	}
	p := imgpull.New(testutil.NewFakeClient(im), resolverStub)
	got, err := p.Progress(context.Background(), "am")
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
	// Same formula as GetDownloadProgress: firefox total falls
	// back to the annotation size, so 56782302/(56782302+1129791514) ≈ 4.78.
	if got.Percent < 4.7 || got.Percent > 4.8 {
		t.Fatalf("partial refs Percent = %v, want ~4.78", got.Percent)
	}
	if got.Complete {
		t.Fatalf("partial refs: must not be Complete: %+v", got)
	}
}

func TestProgress_EmptyTotalOnNode_ReturnsZero(t *testing.T) {
	// One of two nodes has no total info → the node is treated as
	// unknown → overall progress = 0. This is the documented
	// behavioural delta vs legacy (no maxImageSize fallback).
	im := &appv1.ImageManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Spec: appv1.ImageManagerSpec{
			Nodes: []string{"node-1", "node-2"},
			Refs:  []appv1.Ref{{Name: "nginx:1.25"}},
		},
		Status: appv1.ImageManagerStatus{
			State: "downloading",
			Conditions: map[string]map[string]map[string]string{
				"node-1": {"nginx:1.25": {"total": "100", "offset": "80"}},
				"node-2": {"nginx:1.25": {}}, // no total/offset yet
			},
		},
	}
	p := imgpull.New(testutil.NewFakeClient(im), resolverStub)
	got, _ := p.Progress(context.Background(), "am")
	if got.Percent != 0 {
		t.Errorf("mixed nodes: got %v want 0 (slowest-node dominates)", got.Percent)
	}
}

func TestProgress_MultiNode_MinWins(t *testing.T) {
	im := &appv1.ImageManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Spec: appv1.ImageManagerSpec{
			Nodes: []string{"fast", "slow"},
			Refs:  []appv1.Ref{{Name: "nginx:1.25"}, {Name: "redis:7"}},
		},
		Status: appv1.ImageManagerStatus{
			State: "downloading",
			Conditions: map[string]map[string]map[string]string{
				"fast": {
					"nginx:1.25": {"total": "100", "offset": "100"},
					"redis:7":    {"total": "100", "offset": "100"},
				},
				"slow": {
					"nginx:1.25": {"total": "100", "offset": "30"},
					"redis:7":    {"total": "100", "offset": "10"},
				},
			},
		},
	}
	p := imgpull.New(testutil.NewFakeClient(im), resolverStub)
	got, _ := p.Progress(context.Background(), "am")
	// slow node: (30+10) / (100+100) = 20%; fast node: 100%.
	// Min = 20.
	if got.Percent != 20 {
		t.Errorf("multi-node min: got %v want 20", got.Percent)
	}
}

func TestProgress_ClampsOffsetAboveTotal(t *testing.T) {
	im := &appv1.ImageManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Spec: appv1.ImageManagerSpec{
			Nodes: []string{"node-1"},
			Refs:  []appv1.Ref{{Name: "nginx:1.25"}},
		},
		Status: appv1.ImageManagerStatus{
			State: "downloading",
			Conditions: map[string]map[string]map[string]string{
				"node-1": {"nginx:1.25": {"total": "100", "offset": "9999"}},
			},
		},
	}
	p := imgpull.New(testutil.NewFakeClient(im), resolverStub)
	got, _ := p.Progress(context.Background(), "am")
	if got.Percent != 100 {
		t.Errorf("clamp: got %v want 100", got.Percent)
	}
}

func TestProgress_MalformedNumbers_TreatedAsZero(t *testing.T) {
	im := &appv1.ImageManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Spec: appv1.ImageManagerSpec{
			Nodes: []string{"node-1"},
			Refs:  []appv1.Ref{{Name: "nginx:1.25"}},
		},
		Status: appv1.ImageManagerStatus{
			State: "downloading",
			Conditions: map[string]map[string]map[string]string{
				"node-1": {"nginx:1.25": {"total": "abc", "offset": "xyz"}},
			},
		},
	}
	p := imgpull.New(testutil.NewFakeClient(im), resolverStub)
	got, _ := p.Progress(context.Background(), "am")
	// Malformed total = ignored (no valid total on the node) →
	// treat as unknown → 0.
	if got.Percent != 0 {
		t.Errorf("malformed: got %v want 0", got.Percent)
	}
}
