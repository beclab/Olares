package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// PreviousImagesAnnotation records what a container's image and pull
// policy were before `set-image` overrode them, so `dev revert` can put
// them back exactly.
//
// It is ONE annotation holding a JSON object keyed by container name
// rather than one annotation per container. A per-container key
// ("previous-image.<container>") would blow the 63-character limit on
// the annotation name segment for long container names, and a partial
// write would leave a half-reverted workload. Container names are
// unique across containers and initContainers within a pod spec (K8s
// validation enforces it), so the container name alone is a safe key.
const PreviousImagesAnnotation = "dev.olares.io/previous-images"

// PreviousContainerState is one entry in the PreviousImagesAnnotation
// map. ImagePullPolicy is empty when the container did not set one
// explicitly — restoring that means dropping the field, which is what
// the revert patch does by sending null.
type PreviousContainerState struct {
	Image           string `json:"image"`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}

// DefaultDevPullPolicy is what set-image applies unless told otherwise.
// A side-loaded image (imported straight into the node's containerd by
// `dev push`) exists in no registry, so leaving the policy at Always
// guarantees ErrImagePull on the next kubelet sync.
const DefaultDevPullPolicy = "IfNotPresent"

// SetImageResult is the `-o json` shape. Mirrors scaleResult's
// philosophy: a stable synthesized summary rather than the (verbose)
// post-patch object.
type SetImageResult struct {
	Operation     string `json:"operation"`
	Kind          string `json:"kind"`
	Namespace     string `json:"namespace"`
	Name          string `json:"name"`
	Container     string `json:"container"`
	ContainerType string `json:"containerType"`
	PreviousImage string `json:"previousImage"`
	Image         string `json:"image"`
	PullPolicy    string `json:"pullPolicy,omitempty"`
}

// NewSetImageCommand: `olares-cli cluster workload set-image
// <ns/name | name> --kind X --image REF [-c CONTAINER] [-n NS]
// [--pull-policy P] [--watch] [--yes]`.
//
// Unlike `scale`, this cannot use application/merge-patch+json:
// merge-patch replaces whole arrays, so patching one container's image
// would delete every other container in the pod template. Strategic
// merge patch understands the containers list's patchMergeKey ("name")
// and merges by it, which is what the SPA and kubectl both rely on.
func NewSetImageCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace  string
		kindRaw    string
		container  string
		image      string
		pullPolicy string
		watch      bool
		interval   time.Duration
		timeout    time.Duration
		assumeYes  bool
	)
	cmd := &cobra.Command{
		Use:   "set-image <ns/name | name>",
		Short: "point a container at a different image (records the old one for revert)",
		Long: `Replace the image of one container in a Deployment / StatefulSet /
DaemonSet pod template.

PATCHes ` + "`/apis/apps/v1/namespaces/<ns>/<kind>/<name>`" + ` with
` + "`application/strategic-merge-patch+json`" + `. Strategic merge is
required here (not the merge-patch ` + "`scale`" + ` uses): the containers
list has patchMergeKey "name", so the server merges the entry by name
instead of replacing the whole array and dropping sibling containers.

--container may be omitted when the pod template has exactly one
container; otherwise it is required. initContainers are addressed by
the same flag — the verb finds the container by name across both lists.

The replaced image and pull policy are stashed in the
` + "`" + PreviousImagesAnnotation + "`" + ` annotation so
` + "`olares-cli dev revert`" + ` can restore them exactly.

--pull-policy defaults to ` + DefaultDevPullPolicy + ` because the main
consumer is a locally built image side-loaded into the node's
containerd, which no registry can serve. Pass --pull-policy="" to leave
the existing policy untouched.

Examples:
  olares-cli cluster workload set-image os-framework/app-service \
      --kind sts --image beclab/app-service:dev -w
  olares-cli cluster workload set-image myapp -n user-space-alice \
      --kind deploy -c web --image ghcr.io/me/web:pr-42
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			plural, err := NormalizeKind(kindRaw)
			if err != nil {
				return err
			}
			if plural == KindAll {
				return fmt.Errorf("--kind must be one of: deployment, statefulset, daemonset (not %q)", kindRaw)
			}
			if strings.TrimSpace(image) == "" {
				return fmt.Errorf("--image is required")
			}
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			// Same contract as scale: watch-only knobs are dead
			// without --watch, so reject rather than silently ignore.
			if c.Flags().Changed("interval") && !watch {
				return fmt.Errorf("--interval requires --watch")
			}
			if c.Flags().Changed("timeout") && !watch {
				return fmt.Errorf("--timeout requires --watch")
			}
			_, err = RunSetImage(c.Context(), o, SetImageParams{
				Namespace:  ns,
				Name:       name,
				KindPlural: plural,
				Container:  container,
				Image:      image,
				PullPolicy: pullPolicy,
				Watch:      watch,
				Interval:   interval,
				Timeout:    timeout,
				AssumeYes:  assumeYes,
			})
			return err
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().StringVar(&kindRaw, "kind", "", "workload kind: deployment | statefulset | daemonset (REQUIRED)")
	cmd.Flags().StringVarP(&container, "container", "c", "", "container name (optional when the pod template has exactly one container)")
	cmd.Flags().StringVar(&image, "image", "", "new image reference (REQUIRED)")
	cmd.Flags().StringVar(&pullPolicy, "pull-policy", DefaultDevPullPolicy, `imagePullPolicy to set alongside the image; "" leaves it untouched`)
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "after patching, wait for the rollout to converge")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "polling interval for --watch")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "give up after this duration when --watch is set; 0 = no timeout")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	o.AddDetailOutputFlags(cmd)
	return cmd
}

// SetImageParams is the input to RunSetImage. Exported because
// pkg/devdeploy drives the same code path — `dev deploy` must not
// re-implement the patch, so that the verb behaves identically whether
// it is invoked directly or as one step of the dev loop.
type SetImageParams struct {
	Namespace  string
	Name       string
	KindPlural string
	Container  string
	Image      string
	PullPolicy string
	Watch      bool
	Interval   time.Duration
	Timeout    time.Duration
	AssumeYes  bool
	// SkipConfirm bypasses the prompt entirely (not just pre-answering
	// it). `dev deploy` confirms once for the whole resolved target
	// set, so re-prompting per workload would be noise.
	SkipConfirm bool
}

// RunSetImage performs the GET → validate → PATCH → (optional) wait
// sequence and returns what it did. The returned result is populated
// even when Watch is set and the wait later fails, so callers can
// report the patch that did land.
func RunSetImage(ctx context.Context, o *clusteropts.ClusterOptions, p SetImageParams) (SetImageResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var zero SetImageResult
	client, err := o.Prepare()
	if err != nil {
		return zero, err
	}

	path := buildGetPath(p.Namespace, p.KindPlural, p.Name)
	var w Workload
	if err := clusterclient.GetK8sObject(ctx, client, path, &w); err != nil {
		return zero, fmt.Errorf("get %s %s/%s: %w", SingularKind(p.KindPlural), p.Namespace, p.Name, err)
	}

	current, containerType, err := resolveTargetContainer(w, p.Container)
	if err != nil {
		return zero, fmt.Errorf("%s %s/%s: %w", SingularKind(p.KindPlural), p.Namespace, p.Name, err)
	}

	if current.Image == p.Image && (p.PullPolicy == "" || current.ImagePullPolicy == p.PullPolicy) {
		// No-op. Report it rather than issuing a PATCH that would bump
		// the generation and trigger a pointless rollout.
		result := SetImageResult{
			Operation: "set-image (no-op)", Kind: SingularKind(p.KindPlural),
			Namespace: p.Namespace, Name: p.Name,
			Container: current.Name, ContainerType: containerType,
			PreviousImage: current.Image, Image: p.Image, PullPolicy: p.PullPolicy,
		}
		if o.IsJSON() {
			return result, o.PrintJSON(result)
		}
		if !o.Quiet {
			fmt.Fprintf(os.Stdout, "%s %s/%s container %q already runs %s — nothing to do\n",
				SingularKind(p.KindPlural), p.Namespace, p.Name, current.Name, p.Image)
		}
		return result, nil
	}

	if !p.SkipConfirm {
		if err := clusteropts.ConfirmDestructive(os.Stderr, os.Stdin,
			fmt.Sprintf("Repoint %s %s/%s container %q from %s to %s? Pods will be recreated",
				SingularKind(p.KindPlural), p.Namespace, p.Name, current.Name, current.Image, p.Image),
			p.AssumeYes); err != nil {
			return zero, err
		}
	}

	body, err := buildSetImagePatch(w, current, containerType, p.Image, p.PullPolicy)
	if err != nil {
		return zero, err
	}
	if err := clusterclient.Patch[Workload](ctx, client, path,
		"application/strategic-merge-patch+json", body, nil); err != nil {
		return zero, fmt.Errorf("set image on %s %s/%s container %q: %w",
			SingularKind(p.KindPlural), p.Namespace, p.Name, current.Name, err)
	}

	result := SetImageResult{
		Operation: "set-image", Kind: SingularKind(p.KindPlural),
		Namespace: p.Namespace, Name: p.Name,
		Container: current.Name, ContainerType: containerType,
		PreviousImage: current.Image, Image: p.Image, PullPolicy: p.PullPolicy,
	}

	if !p.Watch {
		if o.IsJSON() {
			return result, o.PrintJSON(result)
		}
		if !o.Quiet {
			fmt.Fprintf(os.Stdout, "%s %s/%s container %q: %s -> %s\n",
				SingularKind(p.KindPlural), p.Namespace, p.Name, current.Name, current.Image, p.Image)
		}
		return result, nil
	}

	if !o.Quiet && !o.IsJSON() {
		fmt.Fprintf(os.Stdout, "%s %s/%s container %q: %s -> %s; waiting for rollout to converge\n",
			SingularKind(p.KindPlural), p.Namespace, p.Name, current.Name, current.Image, p.Image)
	}
	return result, runRolloutStatus(ctx, o, p.Namespace, p.Name, p.KindPlural, true, p.Interval, p.Timeout)
}

// resolveTargetContainer finds the container to patch. With an explicit
// name it searches containers then initContainers (names are unique
// across both). Without one it defaults to the sole container, and
// refuses to guess when there is more than one.
func resolveTargetContainer(w Workload, name string) (WorkloadContainer, string, error) {
	if w.Spec.Template == nil {
		return WorkloadContainer{}, "", fmt.Errorf("has no spec.template — not a workload with a pod template")
	}
	containers := w.Spec.Template.Spec.Containers
	initContainers := w.Spec.Template.Spec.InitContainers

	if name == "" {
		switch len(containers) {
		case 0:
			return WorkloadContainer{}, "", fmt.Errorf("pod template has no containers")
		case 1:
			return containers[0], "container", nil
		default:
			return WorkloadContainer{}, "", fmt.Errorf(
				"pod template has %d containers (%s) — pass -c/--container to pick one",
				len(containers), strings.Join(containerNames(containers), ", "))
		}
	}
	for _, c := range containers {
		if c.Name == name {
			return c, "container", nil
		}
	}
	for _, c := range initContainers {
		if c.Name == name {
			return c, "initContainer", nil
		}
	}
	known := append(containerNames(containers), containerNames(initContainers)...)
	sort.Strings(known)
	return WorkloadContainer{}, "", fmt.Errorf("no container named %q (have: %s)",
		name, strings.Join(known, ", "))
}

func containerNames(cs []WorkloadContainer) []string {
	out := make([]string, 0, len(cs))
	for _, c := range cs {
		out = append(out, c.Name)
	}
	return out
}

// buildSetImagePatch assembles the strategic-merge body: the container
// entry keyed by name, plus the updated previous-images annotation.
//
// The annotation is merged with whatever is already there so that
// overriding a second container does not erase the first container's
// recorded original. An existing entry is NOT overwritten — the first
// override is the one that recorded the true pre-dev state, and
// overwriting it on a second `set-image` would make revert restore a
// dev image instead of the original.
func buildSetImagePatch(w Workload, current WorkloadContainer, containerType, image, pullPolicy string) (map[string]interface{}, error) {
	previous, err := DecodePreviousImages(w.Metadata.Annotations)
	if err != nil {
		return nil, err
	}
	if _, already := previous[current.Name]; !already {
		previous[current.Name] = PreviousContainerState{
			Image:           current.Image,
			ImagePullPolicy: current.ImagePullPolicy,
		}
	}
	encoded, err := json.Marshal(previous)
	if err != nil {
		return nil, fmt.Errorf("encode %s annotation: %w", PreviousImagesAnnotation, err)
	}

	entry := map[string]interface{}{
		"name":  current.Name,
		"image": image,
	}
	if pullPolicy != "" {
		entry["imagePullPolicy"] = pullPolicy
	}
	listKey := "containers"
	if containerType == "initContainer" {
		listKey = "initContainers"
	}

	return map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				PreviousImagesAnnotation: string(encoded),
			},
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					listKey: []interface{}{entry},
				},
			},
		},
	}, nil
}

// DecodePreviousImages reads the revert annotation off a workload's
// metadata. A missing annotation yields an empty (non-nil) map; a
// malformed one is an error rather than a silent reset, because
// silently dropping it would strand the workload on a dev image with
// no recorded way back.
func DecodePreviousImages(annotations map[string]string) (map[string]PreviousContainerState, error) {
	out := map[string]PreviousContainerState{}
	raw, ok := annotations[PreviousImagesAnnotation]
	if !ok || strings.TrimSpace(raw) == "" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("annotation %s is malformed (%w); fix or remove it before retrying",
			PreviousImagesAnnotation, err)
	}
	return out, nil
}
