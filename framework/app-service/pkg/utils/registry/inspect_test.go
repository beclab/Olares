package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/remotes"
	digest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type inspectTestResolver struct {
	root    ocispec.Descriptor
	blobs   map[digest.Digest][]byte
	fetched []digest.Digest
}

func (r *inspectTestResolver) Resolve(_ context.Context, ref string) (string, ocispec.Descriptor, error) {
	return ref, r.root, nil
}

func (r *inspectTestResolver) Fetcher(context.Context, string) (remotes.Fetcher, error) {
	return remotes.FetcherFunc(func(_ context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
		r.fetched = append(r.fetched, desc.Digest)
		data, ok := r.blobs[desc.Digest]
		if !ok {
			return nil, fmt.Errorf("unexpected fetch of %s", desc.Digest)
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}), nil
}

func (*inspectTestResolver) Pusher(context.Context, string) (remotes.Pusher, error) {
	return nil, fmt.Errorf("inspect must not push content")
}

func TestInspectPlatformSelection(t *testing.T) {
	amd64 := &ocispec.Platform{OS: "linux", Architecture: "amd64"}
	amd64v2 := &ocispec.Platform{OS: "linux", Architecture: "amd64", Variant: "v2"}
	amd64v3 := &ocispec.Platform{OS: "linux", Architecture: "amd64", Variant: "v3"}
	i386 := &ocispec.Platform{OS: "linux", Architecture: "386"}
	arm64 := &ocispec.Platform{OS: "linux", Architecture: "arm64"}
	armv7 := &ocispec.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}
	unknown := &ocispec.Platform{OS: "unknown", Architecture: "unknown"}
	unused := func(label, mediaType string, platform *ocispec.Platform) ocispec.Descriptor {
		return ocispec.Descriptor{MediaType: mediaType, Digest: digest.FromString(label), Size: 1, Platform: platform}
	}

	for _, tc := range []struct {
		name     string
		platform *ocispec.Platform
		selected *ocispec.Platform
		before   []ocispec.Descriptor
		after    []ocispec.Descriptor
		nested   bool
		wantErr  bool
	}{
		{name: "amd64 preferred over 386", platform: amd64, selected: amd64,
			before: []ocispec.Descriptor{unused("386", ocispec.MediaTypeImageManifest, i386)}},
		{name: "arm64 preferred over arm v7", platform: arm64, selected: arm64,
			before: []ocispec.Descriptor{unused("arm v7", ocispec.MediaTypeImageManifest, armv7)}},
		{name: "native CPU variant preferred", platform: amd64v3, selected: amd64v3,
			before: []ocispec.Descriptor{unused("amd64 v2", ocispec.MediaTypeImageManifest, amd64v2)}},
		{name: "known platform preferred over nil", platform: amd64, selected: amd64,
			before: []ocispec.Descriptor{unused("nil platform manifest", ocispec.MediaTypeImageManifest, nil)}},
		{name: "unsupported metadata with nil platform skipped", platform: amd64, selected: amd64,
			before: []ocispec.Descriptor{unused("SBOM", "application/spdx+json", nil)}},
		{name: "unsupported metadata with matching platform skipped", platform: amd64, selected: amd64,
			before: []ocispec.Descriptor{unused("SBOM", "application/spdx+json", amd64)}},
		{name: "unknown attestation platform skipped", platform: amd64, selected: amd64,
			before: []ocispec.Descriptor{unused("attestation", ocispec.MediaTypeImageManifest, unknown)}},
		{name: "equal platform preference preserves index order", platform: amd64, selected: amd64,
			after: []ocispec.Descriptor{unused("second amd64", ocispec.MediaTypeImageManifest, amd64)}},
		{name: "compatible architecture fallback retained", platform: amd64, selected: i386},
		{name: "nil platform fallback retained", platform: amd64},
		{name: "nested index retained", platform: amd64, selected: amd64, nested: true},
		{name: "no matching platform", platform: amd64, selected: unknown, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolver := &inspectTestResolver{blobs: make(map[digest.Digest][]byte)}
			store := func(mediaType string, value interface{}) ocispec.Descriptor {
				data, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				desc := ocispec.Descriptor{MediaType: mediaType, Digest: digest.FromBytes(data), Size: int64(len(data))}
				resolver.blobs[desc.Digest] = data
				return desc
			}
			configPlatform := tc.selected
			if configPlatform == nil {
				configPlatform = tc.platform
			}
			config := store(ocispec.MediaTypeImageConfig, ocispec.Image{Platform: *configPlatform})
			layers := []ocispec.Descriptor{{MediaType: ocispec.MediaTypeImageLayerGzip, Digest: digest.FromString("layer"), Size: 123}}
			manifest := store(ocispec.MediaTypeImageManifest, ocispec.Manifest{
				Versioned: specs.Versioned{SchemaVersion: 2}, MediaType: ocispec.MediaTypeImageManifest, Config: config, Layers: layers,
			})
			manifest.Platform = tc.selected
			selected := manifest
			if tc.nested {
				selected = store(ocispec.MediaTypeImageIndex, ocispec.Index{
					Versioned: specs.Versioned{SchemaVersion: 2}, MediaType: ocispec.MediaTypeImageIndex, Manifests: []ocispec.Descriptor{manifest},
				})
			}
			manifests := append(append(append([]ocispec.Descriptor(nil), tc.before...), selected), tc.after...)
			resolver.root = store(ocispec.MediaTypeImageIndex, ocispec.Index{
				Versioned: specs.Versioned{SchemaVersion: 2}, MediaType: ocispec.MediaTypeImageIndex, Manifests: manifests,
			})

			image, gotLayers, err := inspectWithPlatform(context.Background(), "example.com/app:v1", resolver, platforms.Only(*tc.platform))
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "no manifest for the current platform") {
					t.Fatalf("expected no platform match, got %v", err)
				}
				if len(resolver.fetched) != 1 {
					t.Fatalf("fetched content despite no platform match: %v", resolver.fetched)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(image.Platform, *configPlatform) || !reflect.DeepEqual(gotLayers, layers) {
				t.Fatalf("wrong image metadata: platform=%+v layers=%+v", image.Platform, gotLayers)
			}
			wantFetched := []digest.Digest{resolver.root.Digest}
			if tc.nested {
				wantFetched = append(wantFetched, selected.Digest)
			}
			wantFetched = append(wantFetched, manifest.Digest, config.Digest)
			if !reflect.DeepEqual(resolver.fetched, wantFetched) {
				t.Fatalf("fetched %v, want only selected metadata %v", resolver.fetched, wantFetched)
			}
		})
	}
}

func TestInspectRejectsIndexWithoutImageDescriptors(t *testing.T) {
	data := []byte(`{"schemaVersion":2,"manifests":[{"mediaType":"application/spdx+json","digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","size":1}]}`)
	resolver := &inspectTestResolver{
		root:  ocispec.Descriptor{MediaType: images.MediaTypeDockerSchema2ManifestList, Digest: digest.FromBytes(data), Size: int64(len(data))},
		blobs: map[digest.Digest][]byte{digest.FromBytes(data): data},
	}
	_, _, err := inspectWithResolver(context.Background(), "example.com/app:v1", resolver)
	if err == nil || !strings.Contains(err.Error(), "no manifest for the current platform") {
		t.Fatalf("expected no image descriptors, got %v", err)
	}
	if len(resolver.fetched) != 1 {
		t.Fatalf("fetched unsupported content: %v", resolver.fetched)
	}
}
