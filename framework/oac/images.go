package oac

import (
	"fmt"
	"strings"

	"github.com/beclab/Olares/framework/oac/internal/resources"
)

// AllModes is the literal keyword that ListImagesForModes / the related
// top-level shortcuts expand into AllImageRenderModes. Matched
// case-insensitively so "ALL", "All", "all" all mean the same thing.
const AllModes = "all"

// AllImageRenderModes is the ordered list of .Values.GPU.Type values that
// "all" expands into when ListImagesForModes (or its top-level shortcut)
// is invoked. Each mode triggers a separate helm render under the
// matching GPU-Type override; the union of workload images across every
// render is deduped and returned alongside options.images.
//
// The list intentionally mirrors the resource modes the Olares app store
// advertises, not the broader resource_mode enum that OlaresManifest
// validation accepts (validResourceModes). They can diverge over time:
// a mode may still be a valid manifest value while no longer being part
// of the default image-extraction set, and vice versa. Callers that
// want to drive image extraction off the manifest's own
// spec.resources[] can build their own slice from cfg.Spec.Resources
// instead of passing "all".
var AllImageRenderModes = []string{
	"cpu",
	"apple-m",
	"nvidia",
	"nvidia-gb10",
	"moore-soc",
	"amd",
	"intel",
}

// ListImages returns the sorted, deduplicated set of container images used by
// oacPath. The set is the union of:
//
//  1. Images discovered by walking the Deployment/StatefulSet/DaemonSet
//     workloads produced by a helm dry-run (primary containers only).
//  2. Images listed under options.images in OlaresManifest.yaml, which is
//     how apps declare extra images that are pulled outside the chart (e.g.
//     images referenced at runtime or by client-side tooling).
//
// ListImages is the no-mode shortcut for ListImagesForMode -- the chart is
// rendered without any .Values.GPU.Type override, which surfaces the images
// of the chart's default (non-GPU) branch only.
func (c *OAC) ListImages(oacPath string) ([]string, error) {
	return c.ListImagesForMode(oacPath, "")
}

// ListImagesForMode is the mode-aware variant of ListImages: it renders the
// chart with .Values.GPU.Type set to mode so chart templates that branch
// per GPU family (e.g. {{ if eq .Values.GPU.Type "nvidia" }}) emit the
// matching workload set. The returned list is still the union of those
// rendered workload images and options.images, sorted and deduplicated.
//
// Passing an empty mode is identical to calling ListImages: no GPU.Type
// override is injected and the default branch of the chart renders.
func (c *OAC) ListImagesForMode(oacPath, mode string) ([]string, error) {
	return c.ListImagesForModes(oacPath, []string{mode})
}

// ListImagesForModes returns the union of container images across each
// .Values.GPU.Type mode in modes. The chart is helm-rendered once per
// expanded mode, the resulting Deployment/StatefulSet workload images
// are collected, then unioned with the manifest's options.images and
// returned as a sorted, deduplicated slice.
//
// Mode semantics:
//
//   - A nil / empty modes slice is treated as a single render with no
//     GPU.Type override, identical to ListImages.
//   - An empty string entry renders the chart's default branch (no
//     override), same as ListImages.
//   - Any element equal to AllModes ("all", case-insensitive) expands
//     in-place into AllImageRenderModes. Duplicates introduced by mixing
//     "all" with explicit modes are collapsed, so each mode renders at
//     most once per call.
//   - Other entries are passed straight through as the
//     .Values.GPU.Type value for that mode's render.
//
// Errors from any single render fail the whole call and identify the
// offending mode in the wrapping message.
func (c *OAC) ListImagesForModes(oacPath string, modes []string) ([]string, error) {
	expanded := expandImageRenderModes(modes)
	m, err := c.LoadManifestFile(oacPath)
	if err != nil {
		return nil, err
	}
	sc := ownerScenario{owner: c.owner, admin: c.admin}
	var workloadUnion []string
	for _, mode := range expanded {
		list, err := c.renderForMode(oacPath, m, sc, mode)
		if err != nil {
			return nil, fmt.Errorf("helm render (mode=%q): %w", mode, err)
		}
		workloadUnion = append(workloadUnion, resources.ExtractWorkloadImages(list)...)
	}
	return resources.MergeImages(workloadUnion, m.OptionsImages()), nil
}

// expandImageRenderModes normalises a caller-supplied list of modes into
// the actual sequence of .Values.GPU.Type values to render under. It
// expands AllModes, drops duplicates while preserving insertion order
// (so the first occurrence wins), and collapses an empty input to a
// single no-override render.
func expandImageRenderModes(modes []string) []string {
	if len(modes) == 0 {
		return []string{""}
	}
	seen := make(map[string]struct{}, len(modes))
	out := make([]string, 0, len(modes))
	add := func(mode string) {
		if _, ok := seen[mode]; ok {
			return
		}
		seen[mode] = struct{}{}
		out = append(out, mode)
	}
	for _, mode := range modes {
		if strings.EqualFold(mode, AllModes) {
			for _, m := range AllImageRenderModes {
				add(m)
			}
			continue
		}
		add(mode)
	}
	if len(out) == 0 {
		// Caller passed only "all" but AllImageRenderModes was set to an
		// empty slice (e.g. cleared by an embedder). Fall back to the
		// no-override render so we still produce a meaningful result.
		return []string{""}
	}
	return out
}

// ListImagesFromOAC is the Checker-less shortcut for (*Checker).ListImages.
func ListImagesFromOAC(oacPath string, opts ...Option) ([]string, error) {
	return New(opts...).ListImages(oacPath)
}

// ListImagesFromOACForMode is the Checker-less shortcut for
// (*Checker).ListImagesForMode.
func ListImagesFromOACForMode(oacPath, mode string, opts ...Option) ([]string, error) {
	return New(opts...).ListImagesForMode(oacPath, mode)
}

// ListImagesFromOACForModes is the Checker-less shortcut for
// (*Checker).ListImagesForModes. Use it when you want to extract images
// across several GPU-Type modes (or the "all" keyword) without holding
// a Checker yourself.
func ListImagesFromOACForModes(oacPath string, modes []string, opts ...Option) ([]string, error) {
	return New(opts...).ListImagesForModes(oacPath, modes)
}
