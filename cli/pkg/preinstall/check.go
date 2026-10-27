package preinstall

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/manifest"
)

// CheckOptions controls how thoroughly CheckStaticBundle inspects a
// static market bundle directory.
type CheckOptions struct {
	// Full also verifies artifact payload trees under each artifact's
	// source directory against the corresponding manifest entries.
	Full bool

	// InstallationManifest enables the image gate. When set, every image the
	// bundled charts render must appear in it, because LoadImages and
	// PinImages both iterate the manifest and ignore anything absent from it.
	// Leaving it empty keeps the contract-only behaviour. Mutually exclusive
	// with a non-nil OlaresImages.
	InstallationManifest manifest.InstallationManifest

	// OlaresImages is output.containers[].name from a source-tree Olares.yaml.
	// A non-nil slice (including empty) enables the same chart-image gate
	// against that list, so a developer can check before packing
	// installation.manifest. Nil leaves the gate off. Mutually exclusive
	// with InstallationManifest.
	OlaresImages []string

	// ImagesDir is the directory holding the preloaded image payloads,
	// normally <baseDir>/images. When set, an image the manifest lists must
	// also have its tarball on the medium, which is the second half of what
	// prepare needs and a separate failure from a stale image list.
	ImagesDir string

	// GPUModes overrides the .Values.GPU.Type values the charts are rendered
	// under. Empty means the no-override branch plus every known family; see
	// imageRenderModes for why that, and not the bundle's allowedGpuTypes, is
	// the default.
	GPUModes []string
}

// CheckStaticBundle validates a static market bundle directory — the
// directory that directly contains bundle.json (typically
// preinstall/market). It is read-only and does not publish runtime
// state or Hugging Face caches.
//
// By default it checks the V1 JSON contract, chart presence/size/SHA-256,
// and artifact manifests. With opts.Full it also verifies each artifact
// source tree entries declared by its manifest (entry types and digests).
// With opts.InstallationManifest or a non-nil opts.OlaresImages it
// additionally renders every bundled chart and requires each image they
// need to appear in that list.
func CheckStaticBundle(dir string, opts CheckOptions) error {
	dir = filepath.Clean(dir)
	root, err := openDirectoryNoSymlink(dir)
	if err != nil {
		return fmt.Errorf("open static bundle: %w", err)
	}
	defer root.Close()

	bundle, err := decodeBundleRoot(root)
	if err != nil {
		return err
	}
	if err := Validate(bundle); err != nil {
		return err
	}
	if err := preflightCharts(root, bundle); err != nil {
		return err
	}
	var totalChartBytes int64
	for _, app := range bundle.Apps {
		verified, err := verifyChartDigest(root, app, MaxTotalChartBytes-totalChartBytes)
		if err != nil {
			return fmt.Errorf("bundle app %q: %w", app.AppID, err)
		}
		totalChartBytes += verified
	}
	if _, err := hfArtifactTargets(bundleHFArtifacts(bundle)); err != nil {
		return err
	}
	for _, app := range bundle.Apps {
		for _, artifact := range app.Artifacts {
			artifactManifest, err := LoadArtifactManifest(root, artifact)
			if err != nil {
				return fmt.Errorf("bundle app %q artifact manifest: %w", app.AppID, err)
			}
			if !opts.Full {
				continue
			}
			if err := verifyArtifactSource(root, artifact, artifactManifest); err != nil {
				return fmt.Errorf("bundle app %q artifact %q: %w", app.AppID, artifact.Source, err)
			}
		}
	}
	return checkBundleImages(root, bundle, opts)
}

func verifyChartDigest(root *os.Root, app BundleAppV1, totalRemaining int64) (int64, error) {
	spec, err := chartVerifiedCopy(root, app.Chart, app.ChartSHA256, totalRemaining)
	if err != nil {
		return 0, err
	}
	return verifyRegularFile(root, spec)
}

func verifyArtifactSource(staticRoot *os.Root, artifact BundleArtifactV1, manifest ArtifactManifestV1) error {
	sourceRoot, err := openHFArtifactSource(staticRoot, artifact)
	if err != nil {
		return err
	}
	defer sourceRoot.Close()

	for _, entry := range manifest.Entries {
		if err := verifyArtifactEntry(sourceRoot, entry); err != nil {
			return err
		}
	}
	return nil
}

func verifyArtifactEntry(sourceRoot *os.Root, entry ArtifactManifestEntryV1) error {
	_, _, err := inspectHFEntry(sourceRoot, entry)
	if err != nil {
		return err
	}
	if entry.Type != "file" {
		return nil
	}
	_, err = verifyRegularFile(sourceRoot, verifiedCopy{
		Source:  entry.Path,
		Size:    entry.Size,
		MaxSize: entry.Size,
		SHA256:  entry.SHA256,
	})
	return err
}
