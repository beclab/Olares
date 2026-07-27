package images

import (
	"math"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func imWith(nodes []string, refs []string, conditions map[string]map[string]map[string]string) *appv1alpha1.ImageManager {
	refList := make([]appv1alpha1.Ref, 0, len(refs))
	for _, r := range refs {
		refList = append(refList, appv1alpha1.Ref{Name: r})
	}
	return &appv1alpha1.ImageManager{
		Spec: appv1alpha1.ImageManagerSpec{
			Nodes: nodes,
			Refs:  refList,
		},
		Status: appv1alpha1.ImageManagerStatus{
			Conditions: conditions,
		},
	}
}

func TestAggregateDownloadProgress_SingleNodeHalfway(t *testing.T) {
	im := imWith(
		[]string{"node1"},
		[]string{"img-a"},
		map[string]map[string]map[string]string{
			"node1": {"img-a": {"offset": "50", "total": "100"}},
		},
	)
	if got := aggregateDownloadProgress(im, nil); got != 50 {
		t.Fatalf("aggregateDownloadProgress = %v, want 50", got)
	}
}

func TestAggregateDownloadProgress_SlowestNodeWins(t *testing.T) {
	im := imWith(
		[]string{"fast", "slow"},
		[]string{"img-a"},
		map[string]map[string]map[string]string{
			"fast": {"img-a": {"offset": "90", "total": "100"}},
			"slow": {"img-a": {"offset": "20", "total": "100"}},
		},
	)
	if got := aggregateDownloadProgress(im, nil); got != 20 {
		t.Fatalf("aggregateDownloadProgress = %v, want 20 (slowest node)", got)
	}
}

func TestAggregateDownloadProgress_MultipleRefsSummedPerNode(t *testing.T) {
	im := imWith(
		[]string{"node1"},
		[]string{"img-a", "img-b"},
		map[string]map[string]map[string]string{
			"node1": {
				"img-a": {"offset": "30", "total": "100"},
				"img-b": {"offset": "70", "total": "100"},
			},
		},
	)
	// (30+70) / (100+100) = 0.5 -> 50
	if got := aggregateDownloadProgress(im, nil); got != 50 {
		t.Fatalf("aggregateDownloadProgress = %v, want 50", got)
	}
}

func TestAggregateDownloadProgress_MissingTotalUsesImageSizeEstimate(t *testing.T) {
	im := imWith(
		[]string{"node1"},
		[]string{"img-a"},
		map[string]map[string]map[string]string{
			"node1": {"img-a": {"offset": "50"}}, // no "total"
		},
	)
	// total falls back to the known image size (100), offset 50 -> 50%.
	if got := aggregateDownloadProgress(im, []Image{{Name: "img-a", Size: 100}}); got != 50 {
		t.Fatalf("aggregateDownloadProgress = %v, want 50 (size estimate)", got)
	}
}

func TestAggregateDownloadProgress_NoNodesReturnsZero(t *testing.T) {
	im := imWith(nil, []string{"img-a"}, map[string]map[string]map[string]string{})
	if got := aggregateDownloadProgress(im, nil); got != 0 {
		t.Fatalf("aggregateDownloadProgress = %v, want 0 for empty node set", got)
	}
	// Guard against the pre-refactor bug where an empty node map surfaced
	// math.MaxFloat64 instead of 0.
	if aggregateDownloadProgress(im, nil) == math.MaxFloat64 {
		t.Fatal("aggregateDownloadProgress returned MaxFloat64 sentinel")
	}
}
