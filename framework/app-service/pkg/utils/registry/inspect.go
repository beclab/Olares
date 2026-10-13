package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Inspect fetches metadata through the same resolver used to pre-pull images.
// Only the selected platform's manifest and config are downloaded, not layers.
func Inspect(ctx context.Context, ref string) (*ocispec.Image, []ocispec.Descriptor, error) {
	return inspectWithResolver(ctx, ref, NewResolver(ctx))
}

func inspectWithResolver(ctx context.Context, ref string, resolver remotes.Resolver) (*ocispec.Image, []ocispec.Descriptor, error) {
	return inspectWithPlatform(ctx, ref, resolver, platforms.Default())
}

func inspectWithPlatform(ctx context.Context, ref string, resolver remotes.Resolver, matcher platforms.MatchComparer) (*ocispec.Image, []ocispec.Descriptor, error) {
	name, desc, err := resolver.Resolve(ctx, ref)
	if err != nil {
		return nil, nil, err
	}
	fetcher, err := resolver.Fetcher(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	for depth := 0; depth < 8; depth++ {
		data, err := fetchMetadata(ctx, fetcher, desc)
		if err != nil {
			return nil, nil, err
		}
		if images.IsIndexType(desc.MediaType) {
			var index ocispec.Index
			if err := json.Unmarshal(data, &index); err != nil {
				return nil, nil, err
			}
			children := images.HandlerFunc(func(context.Context, ocispec.Descriptor) ([]ocispec.Descriptor, error) {
				var manifests []ocispec.Descriptor
				for _, candidate := range index.Manifests {
					if images.IsManifestType(candidate.MediaType) || images.IsIndexType(candidate.MediaType) {
						manifests = append(manifests, candidate)
					}
				}
				return manifests, nil
			})
			// Match the puller's platform preference, including compatible fallbacks
			// and placing descriptors without platform information last.
			children = images.LimitManifests(images.FilterPlatforms(children, matcher), matcher, 1)
			manifests, err := children(ctx, desc)
			if err != nil {
				return nil, nil, fmt.Errorf("no manifest for the current platform: %w", err)
			}
			desc = manifests[0]
			continue
		}
		if !images.IsManifestType(desc.MediaType) {
			return nil, nil, fmt.Errorf("unsupported manifest media type %s", desc.MediaType)
		}
		var manifest ocispec.Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, nil, err
		}
		data, err = fetchMetadata(ctx, fetcher, manifest.Config)
		if err != nil {
			return nil, nil, err
		}
		var config ocispec.Image
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, nil, err
		}
		return &config, manifest.Layers, nil
	}
	return nil, nil, fmt.Errorf("image index nesting exceeds limit")
}

func fetchMetadata(ctx context.Context, fetcher remotes.Fetcher, desc ocispec.Descriptor) ([]byte, error) {
	const limit = 16 << 20
	reader, err := fetcher.Fetch(ctx, desc)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("image metadata exceeds limit")
	}
	if desc.Digest != "" && desc.Digest.Algorithm().Available() && desc.Digest.Algorithm().FromBytes(data) != desc.Digest {
		return nil, fmt.Errorf("image metadata digest mismatch")
	}
	return data, nil
}
