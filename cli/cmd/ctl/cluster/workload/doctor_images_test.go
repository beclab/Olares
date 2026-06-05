package workload

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/containerdimages"
)

func usageByID(usage []imageUsage, id string) (imageUsage, bool) {
	for _, u := range usage {
		if u.ID == id {
			return u, true
		}
	}
	return imageUsage{}, false
}

func TestComputeImageUsageCountsReferences(t *testing.T) {
	local := []containerdimages.Image{
		{ID: "sha256:used", RepoTags: []string{"docker.io/example/api:v1", "registry.local/example/api:v1"}},
		{ID: "sha256:orphan", RepoTags: []string{"docker.io/example/old:v1"}},
	}
	refs := []workloadImageRef{
		{Workload: "api", Image: "docker.io/example/api:v1"},
		{Workload: "api-canary", Image: "docker.io/example/api:v1"},
	}

	usage := computeImageUsage(local, refs)

	used, ok := usageByID(usage, "sha256:used")
	if !ok {
		t.Fatalf("expected used image in usage: %#v", usage)
	}
	if used.Refs != 2 {
		t.Fatalf("used refs = %d, want 2", used.Refs)
	}
	orphan, ok := usageByID(usage, "sha256:orphan")
	if !ok {
		t.Fatalf("expected orphan image in usage")
	}
	if orphan.Refs != 0 {
		t.Fatalf("orphan refs = %d, want 0", orphan.Refs)
	}
}

func TestComputeImageUsageNormalizesDockerHubShorthand(t *testing.T) {
	local := []containerdimages.Image{
		{ID: "sha256:nginx", RepoTags: []string{"docker.io/library/nginx:latest"}},
	}
	refs := []workloadImageRef{{Image: "nginx"}}

	usage := computeImageUsage(local, refs)

	if usage[0].Refs != 1 {
		t.Fatalf("nginx refs = %d, want 1", usage[0].Refs)
	}
}

func TestComputeImageUsageSkipsPauseImages(t *testing.T) {
	local := []containerdimages.Image{
		{ID: "sha256:pause", RepoTags: []string{"registry.k8s.io/pause:3.10"}},
		{ID: "sha256:orphan", RepoTags: []string{"docker.io/example/old:v1"}},
	}

	usage := computeImageUsage(local, nil)

	if _, ok := usageByID(usage, "sha256:pause"); ok {
		t.Fatalf("pause image must be excluded from usage output")
	}
	if len(usage) != 1 {
		t.Fatalf("usage len = %d, want 1: %#v", len(usage), usage)
	}
}

func TestComputeImageUsageMatchesDigestPinnedReference(t *testing.T) {
	const digest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	local := []containerdimages.Image{
		{
			ID:          digest,
			RepoTags:    []string{"docker.io/example/api:v1"},
			RepoDigests: []string{"docker.io/example/api@" + digest},
		},
		{ID: "sha256:orphan", RepoTags: []string{"docker.io/example/old:v1"}},
	}
	// Workload pins by digest against a mirror host so the tag will not
	// line up — only the digest can prove the image is used.
	refs := []workloadImageRef{{Image: "registry.local/example/api@" + digest}}

	usage := computeImageUsage(local, refs)

	used, ok := usageByID(usage, digest)
	if !ok {
		t.Fatalf("expected digest-pinned image in usage")
	}
	if used.Refs != 1 {
		t.Fatalf("digest-pinned refs = %d, want 1", used.Refs)
	}
}

func TestFilterUnusedKeepsOnlyZeroRefs(t *testing.T) {
	usage := []imageUsage{
		{Image: containerdimages.Image{ID: "sha256:a"}, Refs: 2},
		{Image: containerdimages.Image{ID: "sha256:b"}, Refs: 0},
	}

	got := filterUnused(usage)

	if len(got) != 1 || got[0].ID != "sha256:b" {
		t.Fatalf("filterUnused = %#v, want only sha256:b", got)
	}
}

func TestSummarizeImageUsageCountsReclaimable(t *testing.T) {
	usage := []imageUsage{
		{Image: containerdimages.Image{ID: "sha256:a", Size: 100}, Refs: 2},
		{Image: containerdimages.Image{ID: "sha256:b", Size: 30}, Refs: 0},
		{Image: containerdimages.Image{ID: "sha256:c", Size: 70}, Refs: 0},
	}

	s := summarizeImageUsage(usage)

	if s.Total != 3 {
		t.Fatalf("total = %d, want 3", s.Total)
	}
	if s.Unused != 2 {
		t.Fatalf("unused = %d, want 2", s.Unused)
	}
	if s.ReclaimableBytes != 100 {
		t.Fatalf("reclaimable = %d, want 100", s.ReclaimableBytes)
	}
}

func TestSortBySizeDescBiggestFirst(t *testing.T) {
	usage := []imageUsage{
		{Image: containerdimages.Image{ID: "sha256:small", Size: 10}},
		{Image: containerdimages.Image{ID: "sha256:big", Size: 1000}},
		{Image: containerdimages.Image{ID: "sha256:mid", Size: 100}},
	}

	sortBySizeDesc(usage)

	if usage[0].ID != "sha256:big" || usage[2].ID != "sha256:small" {
		t.Fatalf("size-desc order wrong: %#v", usage)
	}
}

func TestComputeImageUsageSortedDeterministically(t *testing.T) {
	local := []containerdimages.Image{
		{ID: "sha256:c", RepoTags: []string{"docker.io/example/c:v1"}},
		{ID: "sha256:a", RepoTags: []string{"docker.io/example/a:v1"}},
		{ID: "sha256:b", RepoTags: []string{"docker.io/example/b:v1"}},
	}

	usage := computeImageUsage(local, nil)

	want := []string{"docker.io/example/a:v1", "docker.io/example/b:v1", "docker.io/example/c:v1"}
	for i := range want {
		if usage[i].RepoTags[0] != want[i] {
			t.Fatalf("usage order = %v, want %v", usage, want)
		}
	}
}
