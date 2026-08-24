package preinstall

import (
	"fmt"
	"os"
	"path/filepath"
)

// CheckOptions controls how thoroughly CheckStaticBundle inspects a
// static market bundle directory.
type CheckOptions struct {
	// Full also verifies artifact payload trees under each artifact's
	// source directory against the corresponding manifest entries.
	Full bool
}

// CheckStaticBundle validates a static market bundle directory — the
// directory that directly contains bundle.json (typically
// preinstall/market). It is read-only and does not publish runtime
// state or Hugging Face caches.
//
// By default it checks the V1 JSON contract, chart presence/size/SHA-256,
// and artifact manifests. With opts.Full it also verifies each artifact
// source tree entries declared by its manifest (entry types and digests).
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
			manifest, err := LoadArtifactManifest(root, artifact)
			if err != nil {
				return fmt.Errorf("bundle app %q artifact manifest: %w", app.AppID, err)
			}
			if !opts.Full {
				continue
			}
			if err := verifyArtifactSource(root, artifact, manifest); err != nil {
				return fmt.Errorf("bundle app %q artifact %q: %w", app.AppID, artifact.Source, err)
			}
		}
	}
	return nil
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
