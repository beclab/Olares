package oac

import (
	"fmt"

	"github.com/beclab/Olares/framework/oac/internal/resources"
)

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
	m, err := c.LoadManifestFile(oacPath)
	if err != nil {
		return nil, err
	}
	sc := ownerScenario{owner: c.owner, admin: c.admin}
	list, err := c.renderForMode(oacPath, m, sc, mode)
	if err != nil {
		return nil, fmt.Errorf("helm render: %w", err)
	}
	workloadImages := resources.ExtractWorkloadImages(list)
	return resources.MergeImages(workloadImages, m.OptionsImages()), nil
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
