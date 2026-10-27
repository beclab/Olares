package pod

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/picker"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
)

// maxPickerPages bounds the interactive picker's pod-list drain so a
// pathologically large cluster can't hang the CLI. 20 pages * the default
// limit of 100 ≈ 2000 pods; past that we show what we have and hint at -n.
const maxPickerPages = 20

// pickerFetch is the result of draining the pod list for the picker: the
// flattened+filtered container entries plus the pod-level counts used to
// describe a truncated scan. capped is true when the drain hit maxPickerPages
// before exhausting the list; fetchedPods is how many pods were actually
// fetched and totalPods is the server-reported total (so the header can say
// "first X of Y pods" rather than misreporting the container-row count).
type pickerFetch struct {
	entries     []picker.Entry
	capped      bool
	fetchedPods int
	totalPods   int
}

// buildPickerEntries fetches the pods visible to the active profile (scoped to
// namespace when non-empty) and flattens them into one picker.Entry per
// container. The KubeSphere pods list envelope already carries spec.containers
// and status, so a single paginated drain yields everything — no per-pod Get.
func buildPickerEntries(ctx context.Context, o *clusteropts.ClusterOptions, namespace string) (pickerFetch, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return pickerFetch{}, err
	}
	p := clusteropts.NewPaginationOptions() // limit 100, page 1

	var pods []Pod
	capped := false
	totalPods := 0
	for page := 1; page <= maxPickerPages; page++ {
		path := buildListPath(namespace, "", nil, p, page)
		resp, err := clusterclient.GetKubeSphereList[Pod](ctx, client, path)
		if err != nil {
			return pickerFetch{}, fmt.Errorf("list pods: %w", err)
		}
		pods = append(pods, resp.Items...)
		totalPods = resp.TotalItems
		if resp.TotalItems == 0 || len(pods) >= resp.TotalItems || len(resp.Items) < p.Limit {
			break
		}
		if page == maxPickerPages {
			capped = true
		}
	}

	entries := podsToEntries(pods, namespace)

	// Only offer containers the active profile may actually exec into, so the
	// picker matches the SPA (which never shows a Terminal button for off-limits
	// namespaces) and a selection never dead-ends on a permission error. If the
	// identity can't be resolved we leave the list intact — RunExec re-checks
	// and surfaces the reason on the chosen target.
	if info, ierr := resolveExecIdentity(ctx, o); ierr == nil {
		entries = filterExecEntries(entries, info)
	}

	picker.Sort(entries)
	return pickerFetch{
		entries:     entries,
		capped:      capped,
		fetchedPods: len(pods),
		totalPods:   totalPods,
	}, nil
}

// podsToEntries flattens pods into one picker.Entry per container. nsFallback
// supplies the namespace when the list envelope omits it (namespace-scoped
// requests may leave metadata.namespace blank). Pure — unit-tested directly.
func podsToEntries(pods []Pod, nsFallback string) []picker.Entry {
	entries := make([]picker.Entry, 0, len(pods))
	for _, pd := range pods {
		ns := pd.Metadata.Namespace
		if ns == "" {
			ns = nsFallback
		}
		running := pd.Status.Phase == "Running"
		status := pd.statusReason()
		if status == "" {
			status = "-"
		}
		for _, c := range pd.Spec.Containers {
			entries = append(entries, picker.Entry{
				Namespace: ns,
				Pod:       pd.Metadata.Name,
				Container: c.Name,
				Running:   running,
				Status:    status,
			})
		}
	}
	return entries
}

// PickInteractiveTarget fetches selectable containers and runs the interactive
// picker, returning the chosen namespace/pod/container. canceled is true when
// the user aborted (Esc/Ctrl-C) — callers should treat that as a clean no-op
// exit, not an error. Only meaningful on a real terminal; the caller gates this
// behind -it (which already requires a TTY).
func PickInteractiveTarget(ctx context.Context, o *clusteropts.ClusterOptions, namespace string) (ns, podName, container string, canceled bool, err error) {
	// A slow/large pod-list drain would otherwise leave the user staring at a
	// blank line. Animate a spinner until the fetch returns. (No live count:
	// the list arrives one page at a time, so there is no sub-page progress to
	// report — the animation alone signals work.) No-op on non-TTY stderr.
	msg := "Loading containers\u2026"
	if namespace != "" {
		msg = fmt.Sprintf("Loading containers in %s\u2026", namespace)
	}
	sp := picker.StartSpinner(func() string { return msg })
	fetch, err := buildPickerEntries(ctx, o, namespace)
	sp.Stop()
	if err != nil {
		return "", "", "", false, err
	}
	entries := fetch.entries
	if len(entries) == 0 {
		if namespace != "" {
			return "", "", "", false, fmt.Errorf("no containers visible in namespace %q for the active profile", namespace)
		}
		return "", "", "", false, fmt.Errorf("no containers visible to the active profile")
	}

	header := "Select a container to exec into"
	if namespace != "" {
		header = fmt.Sprintf("Select a container in %s to exec into", namespace)
	}
	// When the drain was truncated, describe it in POD terms — the cap is a
	// pod-page limit, so reporting the flattened container-row count (len
	// entries) would understate what was skipped and read as "list complete".
	if fetch.capped {
		if fetch.totalPods > fetch.fetchedPods {
			header += fmt.Sprintf("  (scanned first %d of %d pods — narrow with -n)", fetch.fetchedPods, fetch.totalPods)
		} else {
			header += fmt.Sprintf("  (scanned first %d pods — narrow with -n)", fetch.fetchedPods)
		}
	}

	sel, perr := picker.Pick(entries, header)
	if perr != nil {
		if perr == picker.ErrCanceled {
			return "", "", "", true, nil
		}
		return "", "", "", false, perr
	}
	return sel.Namespace, sel.Pod, sel.Container, false, nil
}
