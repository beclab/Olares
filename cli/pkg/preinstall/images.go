package preinstall

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/distribution/reference"

	pkgchart "github.com/beclab/Olares/cli/pkg/chart"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/utils"
	oac "github.com/beclab/Olares/framework/oac"
)

// chartArchiveName is where a bundled chart lands inside the scratch directory
// before it is unpacked. The bundle's own path is not reused: it may name
// subdirectories, and the scratch tree only ever holds one chart.
const chartArchiveName = "chart.tgz"

// imageTarSuffixes are the archive extensions LoadImages accepts for a preloaded
// image payload. The gate has to agree with it exactly, or it will pass a medium
// prepare then refuses.
var imageTarSuffixes = []string{".tar.gz", ".tgz", ".tar"}

// ImageGapKind separates the two ways a bundled chart's image fails to be
// preloaded. They are reported apart because they are fixed apart: one needs a
// line in images.mf, the other needs the payload to actually ship.
type ImageGapKind string

const (
	// ImageGapUnlisted means installation.manifest never names the image.
	// LoadImages iterates the manifest, so nothing imports it and PinImages
	// never pins it; at install time the cluster falls back to pulling from a
	// registry, which an offline device cannot do.
	ImageGapUnlisted ImageGapKind = "missing from installation.manifest"
	// ImageGapUnlistedOlaresYAML means the source-tree Olares.yaml never names
	// the image in output.containers, so packing would omit it from
	// installation.manifest and the same registry fallback would happen.
	ImageGapUnlistedOlaresYAML ImageGapKind = "missing from olares.yaml"
	// ImageGapNoPayload means the manifest names the image but the medium
	// carries no tarball for it. Prepare aborts in LoadImages with
	// "image %s not found in %s".
	ImageGapNoPayload ImageGapKind = "no image payload on the medium"
)

// ImageGap is one image a bundled chart needs that the installation medium will
// not preload.
type ImageGap struct {
	AppID   string
	AppName string
	// Image is spelled the way installation.manifest and the CDN object names
	// spell it: the familiar form, without the docker.io/ prefix. Payload
	// filenames are md5 of this exact string, so a line added to images.mf in
	// any other spelling resolves to a tarball that was never uploaded.
	Image string
	Kind  ImageGapKind
}

func (g ImageGap) String() string {
	return fmt.Sprintf("app %s (%s): %s: %s", g.AppName, g.AppID, g.Image, g.Kind)
}

// imageGapError renders every gap in one error. Reporting the first one only
// would hide the shape of the problem: a stale image list usually misses one
// image per app, and the build needs to see all of them in a single run.
func imageGapError(gaps []ImageGap) error {
	if len(gaps) == 0 {
		return nil
	}
	lines := make([]string, 0, len(gaps)+1)
	lines = append(lines, fmt.Sprintf("%d preinstall image(s) will not be preloaded:", len(gaps)))
	for _, gap := range gaps {
		lines = append(lines, "  "+gap.String())
	}
	return fmt.Errorf("%s", strings.Join(lines, "\n"))
}

// checkBundleImages reports every image the bundled charts need that the listed
// source will not cover. Neither an empty InstallationManifest nor a nil
// OlaresImages enables the check, which is what keeps CheckStaticBundle's
// contract-only mode working unchanged. A non-nil OlaresImages, even empty,
// does enable it: the caller asked for the source-tree gate.
func checkBundleImages(root *os.Root, bundle BundleV1, opts CheckOptions) error {
	listed, unlistedKind, imagesDir, enabled, err := listedFromOptions(opts)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	var (
		gaps  []ImageGap
		total int64
	)
	for _, app := range bundle.Apps {
		images, consumed, err := chartImages(root, app, opts.GPUModes, MaxTotalChartBytes-total)
		if err != nil {
			return fmt.Errorf("bundle app %q: %w", app.AppID, err)
		}
		total += consumed
		for _, image := range images {
			gap, ok := imageGapFor(app, image, listed, unlistedKind, imagesDir)
			if ok {
				gaps = append(gaps, gap)
			}
		}
	}
	return imageGapError(gaps)
}

func listedFromOptions(opts CheckOptions) (map[string]string, ImageGapKind, string, bool, error) {
	hasManifest := len(opts.InstallationManifest) > 0
	hasYAML := opts.OlaresImages != nil
	if hasManifest && hasYAML {
		return nil, "", "", false, fmt.Errorf("installation manifest and olares.yaml image lists are mutually exclusive")
	}
	if hasManifest {
		return listedImages(opts.InstallationManifest), ImageGapUnlisted, opts.ImagesDir, true, nil
	}
	if hasYAML {
		return listedFromRefs(opts.OlaresImages), ImageGapUnlistedOlaresYAML, "", true, nil
	}
	return nil, "", "", false, nil
}

// imageGapFor answers, for one image of one app, whether the listed source will
// cover it. The image arrives normalized; `listed` maps normalized refs back
// to the source's own spelling, which is both what names a packed payload file
// and what a fix has to be written as.
func imageGapFor(app BundleAppV1, image string, listed map[string]string, unlistedKind ImageGapKind, imagesDir string) (ImageGap, bool) {
	raw, ok := listed[image]
	if !ok {
		return ImageGap{
			AppID:   app.AppID,
			AppName: app.AppName,
			Image:   familiarImageRef(image),
			Kind:    unlistedKind,
		}, true
	}
	if imagesDir == "" || imagePayloadExists(imagesDir, raw) {
		return ImageGap{}, false
	}
	return ImageGap{
		AppID:   app.AppID,
		AppName: app.AppName,
		Image:   raw,
		Kind:    ImageGapNoPayload,
	}, true
}

// listedImages indexes installation.manifest by normalized ref so a chart's
// docker.io/beclab/x:1 matches the manifest's beclab/x:1. The value keeps the
// manifest's original spelling because that is what md5s into the payload
// filename.
func listedImages(installation manifest.InstallationManifest) map[string]string {
	_, refs := installation.GetImageList()
	return listedFromRefs(refs)
}

// listedFromRefs indexes image references by normalized form so a chart's
// docker.io/beclab/x:1 matches a list's beclab/x:1. The value keeps the
// original spelling so a reported gap can be pasted back into images.mf or
// Olares.yaml as-is.
func listedFromRefs(refs []string) map[string]string {
	listed := make(map[string]string, len(refs))
	for _, ref := range refs {
		normalized, err := normalizeImageRef(ref)
		if err != nil {
			continue
		}
		listed[normalized] = ref
	}
	return listed
}

// imagePayloadExists mirrors how LoadImages finds a preloaded image: a file in
// the images directory whose name starts with the md5 of the manifest's ref and
// carries an archive suffix.
func imagePayloadExists(imagesDir, rawRef string) bool {
	prefix := utils.MD5(rawRef)
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		lower := strings.ToLower(entry.Name())
		for _, suffix := range imageTarSuffixes {
			if strings.HasSuffix(lower, suffix) {
				return true
			}
		}
	}
	return false
}

// chartImages renders an app's bundled chart under every requested GPU mode and
// returns the normalized images it needs, plus the chart bytes it consumed
// against the bundle's total budget.
//
// Rendering rather than scanning templates is not optional: a chart may name its
// image through .Values, and a textual scan would return the template string
// instead of the image that will actually be pulled.
func chartImages(root *os.Root, app BundleAppV1, modes []string, totalRemaining int64) ([]string, int64, error) {
	dir, consumed, cleanup, err := extractBundledChart(root, app, totalRemaining)
	if err != nil {
		return nil, 0, err
	}
	defer cleanup()

	refs, err := oac.ListImagesFromOACForModes(dir, imageRenderModes(modes))
	if err != nil {
		return nil, 0, fmt.Errorf("render chart %q: %w", app.Chart, err)
	}
	seen := make(map[string]struct{}, len(refs))
	images := make([]string, 0, len(refs))
	for _, ref := range refs {
		normalized, err := normalizeImageRef(ref)
		if err != nil {
			return nil, 0, fmt.Errorf("chart %q: %w", app.Chart, err)
		}
		if err := rejectFloatingTag(normalized); err != nil {
			return nil, 0, fmt.Errorf("chart %q: %w", app.Chart, err)
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		images = append(images, normalized)
	}
	sort.Strings(images)
	return images, consumed, nil
}

// imageRenderModes decides which GPU families the chart is rendered under.
//
// The default is every family plus the no-override branch, and deliberately not
// the bundle's allowedGpuTypes: that field gates which env and GPU selections an
// operator may make, while the branch a chart actually takes comes from
// .Values.GPU.Type, which app-service fills in from the hardware it finds. One
// image ships to many machines, so the medium has to satisfy all of them.
func imageRenderModes(modes []string) []string {
	if len(modes) == 0 {
		return []string{"", oac.AllModes}
	}
	return modes
}

// extractBundledChart unpacks one bundled chart into a scratch directory and
// returns the chart root. The archive is copied out through the same verified
// path publish uses, so the size ceilings and the digest in bundle.json are
// enforced before anything is unpacked.
func extractBundledChart(root *os.Root, app BundleAppV1, totalRemaining int64) (string, int64, func(), error) {
	noop := func() {}
	spec, err := chartVerifiedCopy(root, app.Chart, app.ChartSHA256, totalRemaining)
	if err != nil {
		return "", 0, noop, err
	}
	spec.Target = chartArchiveName

	scratch, err := os.MkdirTemp("", "olares-preinstall-chart-*")
	if err != nil {
		return "", 0, noop, fmt.Errorf("create chart scratch directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(scratch) }

	scratchRoot, err := os.OpenRoot(scratch)
	if err != nil {
		cleanup()
		return "", 0, noop, fmt.Errorf("open chart scratch directory: %w", err)
	}
	copied, err := copyVerifiedRegularFile(root, scratchRoot, spec)
	closeErr := scratchRoot.Close()
	if err != nil {
		cleanup()
		return "", 0, noop, err
	}
	if closeErr != nil {
		cleanup()
		return "", 0, noop, fmt.Errorf("close chart scratch directory: %w", closeErr)
	}

	dir, unpackCleanup, err := pkgchart.ResolveDir(filepath.Join(scratch, chartArchiveName))
	if err != nil {
		cleanup()
		return "", 0, noop, fmt.Errorf("unpack chart %q: %w", app.Chart, err)
	}
	return dir, copied, func() {
		unpackCleanup()
		cleanup()
	}, nil
}

// normalizeImageRef puts a reference in the canonical fully-qualified form so
// that beclab/x:1 and docker.io/beclab/x:1 compare equal. installation.manifest
// uses the short spelling and charts generally use the long one, so without this
// every image would look absent.
func normalizeImageRef(ref string) (string, error) {
	parsed, err := reference.ParseNormalizedNamed(strings.TrimSpace(ref))
	if err != nil {
		return "", fmt.Errorf("parse image reference %q: %w", ref, err)
	}
	return parsed.String(), nil
}

// familiarImageRef spells a normalized reference the way installation.manifest
// and the CDN object names do, so a reported gap can be pasted into images.mf
// as-is.
func familiarImageRef(normalized string) string {
	parsed, err := reference.ParseNormalizedNamed(normalized)
	if err != nil {
		return normalized
	}
	return reference.FamiliarString(parsed)
}

// rejectFloatingTag refuses images that do not pin what they resolve to. A
// preloaded :latest is a lie: the tag the medium froze and the tag the registry
// serves later are different images, and app-service pulls rather than reusing
// the local copy.
func rejectFloatingTag(normalized string) error {
	parsed, err := reference.ParseNormalizedNamed(normalized)
	if err != nil {
		return fmt.Errorf("parse image reference %q: %w", normalized, err)
	}
	if _, ok := parsed.(reference.Canonical); ok {
		return nil
	}
	tagged, ok := parsed.(reference.Tagged)
	if !ok {
		return fmt.Errorf("image %q has no tag", familiarImageRef(normalized))
	}
	if tagged.Tag() == "latest" {
		return fmt.Errorf("image %q uses the latest tag", familiarImageRef(normalized))
	}
	return nil
}
