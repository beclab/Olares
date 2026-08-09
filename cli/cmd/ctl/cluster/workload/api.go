package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// This file is the workload package's Factory-based facade.
//
// clusteropts lives under cmd/ctl/cluster/internal/, so Go's internal
// rule makes it unreachable from cmd/ctl/dev. Rather than hoisting
// clusteropts out of internal/ — which would widen a deliberately
// narrow surface for every future caller — the dev tree talks to
// workloads through these functions, which take a *cmdutil.Factory and
// build the options bag themselves. Everything they expose is a type
// this package owns.

// OutputMode is the caller-settable slice of ClusterOptions that
// matters to non-cobra callers. Table rendering is the default; JSON
// and Quiet mirror the -o json / -q flags.
type OutputMode struct {
	JSON  bool
	Quiet bool
}

func (m OutputMode) options(f *cmdutil.Factory) *clusteropts.ClusterOptions {
	o := clusteropts.NewClusterOptions(f)
	if m.JSON {
		o.Output = "json"
	}
	o.Quiet = m.Quiet
	return o
}

// SetImage is the Factory-based entry point to the set-image code path.
// `dev deploy` uses it so that repointing a workload is byte-for-byte
// the same operation whether it was reached through
// `cluster workload set-image` or through the dev loop.
func SetImage(ctx context.Context, f *cmdutil.Factory, m OutputMode, p SetImageParams) (SetImageResult, error) {
	return RunSetImage(ctx, m.options(f), p)
}

// FindImageRefsFor is FindImageRefs with the options bag built here.
func FindImageRefsFor(ctx context.Context, f *cmdutil.Factory, m OutputMode, namespace, image string) ([]ImageRef, error) {
	return FindImageRefs(ctx, m.options(f), namespace, image)
}

// DevOverride is one workload carrying a `set-image` override, as
// reported by `dev status`.
type DevOverride struct {
	Namespace  string `json:"namespace"`
	Kind       string `json:"kind"`
	KindPlural string `json:"-"`
	Name       string `json:"name"`
	Container  string `json:"container"`
	// Current is the image the workload runs now (the dev build).
	Current string `json:"current"`
	// Previous is what it ran before the override, i.e. what `dev
	// revert` will restore.
	Previous string `json:"previous"`
	// Missing is true when the annotation names a container that is no
	// longer in the pod template — a leftover from a chart upgrade that
	// renamed or dropped it. Revert skips these and says so.
	Missing bool `json:"missing,omitempty"`
}

// ListDevOverrides scans every workload visible to the profile for the
// previous-images annotation.
//
// This is the answer to "what on this cluster is currently running a
// dev build?" — the first thing to check when a workload misbehaves
// after a dev session, and the reason a side-loaded image that a node
// prune has since collected is diagnosable at all.
func ListDevOverrides(ctx context.Context, f *cmdutil.Factory, m OutputMode, namespace string) ([]DevOverride, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	o := m.options(f)
	p := clusteropts.NewPaginationOptions()
	p.All = true

	collected, _, err := fetchWorkloads(ctx, o, p, namespace, KindAll, "", true)
	if err != nil {
		return nil, err
	}

	var out []DevOverride
	for _, result := range collected {
		for _, w := range result.Items {
			previous, err := DecodePreviousImages(w.Metadata.Annotations)
			if err != nil {
				// One malformed annotation must not blind the whole
				// scan — report it in place and keep going.
				out = append(out, DevOverride{
					Namespace: w.Metadata.Namespace,
					Kind:      nonEmpty(w.Kind, SingularKind(result.Kind)),
					Name:      w.Metadata.Name,
					Previous:  fmt.Sprintf("<malformed: %v>", err),
					Missing:   true,
				})
				continue
			}
			if len(previous) == 0 {
				continue
			}
			live := map[string]WorkloadContainer{}
			if w.Spec.Template != nil {
				for _, c := range w.Spec.Template.Spec.Containers {
					live[c.Name] = c
				}
				for _, c := range w.Spec.Template.Spec.InitContainers {
					live[c.Name] = c
				}
			}
			for name, state := range previous {
				entry := DevOverride{
					Namespace:  w.Metadata.Namespace,
					Kind:       nonEmpty(w.Kind, SingularKind(result.Kind)),
					KindPlural: result.Kind,
					Name:       w.Metadata.Name,
					Container:  name,
					Previous:   state.Image,
				}
				if c, ok := live[name]; ok {
					entry.Current = c.Image
				} else {
					entry.Missing = true
				}
				out = append(out, entry)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Container < b.Container
	})
	return out, nil
}

// RevertResult reports one restored container.
type RevertResult struct {
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Container string `json:"container"`
	From      string `json:"from"`
	To        string `json:"to"`
	Skipped   string `json:"skipped,omitempty"`
}

// RevertParams selects what to restore.
type RevertParams struct {
	Namespace  string
	Name       string
	KindPlural string
	// Container restricts the revert to one container; empty reverts
	// every container recorded in the annotation.
	Container string
	Watch     bool
	Interval  time.Duration
	Timeout   time.Duration
}

// RevertImages restores the images recorded in the previous-images
// annotation and clears the entries it consumed.
//
// Restoring the pull policy matters as much as the image: set-image
// forces IfNotPresent so a side-loaded build is usable, and leaving
// that behind would silently pin the workload to whatever happens to
// be cached on the node, defeating the next legitimate chart upgrade.
// An entry whose original policy was unset is restored by sending
// null, which strategic merge patch treats as "delete this field" so
// the API server reapplies its default.
func RevertImages(ctx context.Context, f *cmdutil.Factory, m OutputMode, p RevertParams) ([]RevertResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	o := m.options(f)
	client, err := o.Prepare()
	if err != nil {
		return nil, err
	}

	path := buildGetPath(p.Namespace, p.KindPlural, p.Name)
	var w Workload
	if err := clusterclient.GetK8sObject(ctx, client, path, &w); err != nil {
		return nil, fmt.Errorf("get %s %s/%s: %w", SingularKind(p.KindPlural), p.Namespace, p.Name, err)
	}
	previous, err := DecodePreviousImages(w.Metadata.Annotations)
	if err != nil {
		return nil, err
	}
	if len(previous) == 0 {
		return nil, fmt.Errorf("%s %s/%s carries no %s annotation — nothing to revert",
			SingularKind(p.KindPlural), p.Namespace, p.Name, PreviousImagesAnnotation)
	}
	if p.Container != "" {
		state, ok := previous[p.Container]
		if !ok {
			return nil, fmt.Errorf("%s %s/%s has no recorded original for container %q (recorded: %s)",
				SingularKind(p.KindPlural), p.Namespace, p.Name, p.Container,
				strings.Join(sortedKeys(previous), ", "))
		}
		previous = map[string]PreviousContainerState{p.Container: state}
	}

	live := map[string]string{} // container name -> list key
	if w.Spec.Template != nil {
		for _, c := range w.Spec.Template.Spec.Containers {
			live[c.Name] = "containers"
		}
		for _, c := range w.Spec.Template.Spec.InitContainers {
			live[c.Name] = "initContainers"
		}
	}

	entries := map[string][]interface{}{}
	var results []RevertResult
	remaining, err := DecodePreviousImages(w.Metadata.Annotations)
	if err != nil {
		return nil, err
	}

	for _, name := range sortedKeys(previous) {
		state := previous[name]
		listKey, present := live[name]
		if !present {
			// The container is gone (renamed or removed by a chart
			// upgrade). Patching it back would ADD a container to the
			// pod template, which is worse than leaving it alone.
			results = append(results, RevertResult{
				Namespace: p.Namespace, Kind: SingularKind(p.KindPlural), Name: p.Name,
				Container: name, To: state.Image,
				Skipped: "container is no longer in the pod template",
			})
			delete(remaining, name)
			continue
		}
		entry := map[string]interface{}{
			"name":  name,
			"image": state.Image,
		}
		if state.ImagePullPolicy != "" {
			entry["imagePullPolicy"] = state.ImagePullPolicy
		} else {
			entry["imagePullPolicy"] = nil
		}
		entries[listKey] = append(entries[listKey], entry)
		results = append(results, RevertResult{
			Namespace: p.Namespace, Kind: SingularKind(p.KindPlural), Name: p.Name,
			Container: name, From: currentImage(w, name), To: state.Image,
		})
		delete(remaining, name)
	}

	if len(entries) == 0 {
		return results, nil
	}

	spec := map[string]interface{}{}
	for listKey, list := range entries {
		spec[listKey] = list
	}
	// An emptied map is written as a deleted annotation (null) rather
	// than "{}" so a fully reverted workload looks untouched.
	var annotationValue interface{}
	if len(remaining) == 0 {
		annotationValue = nil
	} else {
		encoded, merr := json.Marshal(remaining)
		if merr != nil {
			return nil, fmt.Errorf("encode %s annotation: %w", PreviousImagesAnnotation, merr)
		}
		annotationValue = string(encoded)
	}

	body := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				PreviousImagesAnnotation: annotationValue,
			},
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": spec,
			},
		},
	}
	if err := clusterclient.Patch[Workload](ctx, client, path,
		"application/strategic-merge-patch+json", body, nil); err != nil {
		return nil, fmt.Errorf("revert %s %s/%s: %w", SingularKind(p.KindPlural), p.Namespace, p.Name, err)
	}

	if p.Watch {
		if err := runRolloutStatus(ctx, o, p.Namespace, p.Name, p.KindPlural, true, p.Interval, p.Timeout); err != nil {
			return results, err
		}
	}
	return results, nil
}

func currentImage(w Workload, container string) string {
	if w.Spec.Template == nil {
		return ""
	}
	for _, c := range w.Spec.Template.Spec.Containers {
		if c.Name == container {
			return c.Image
		}
	}
	for _, c := range w.Spec.Template.Spec.InitContainers {
		if c.Name == container {
			return c.Image
		}
	}
	return ""
}

func sortedKeys(m map[string]PreviousContainerState) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
