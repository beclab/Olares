package workload

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
)

// batchKinds are the extra image-referencing kinds the unused-image
// scan always folds into the "used" set. Without them a one-shot Job
// or scheduled CronJob image would look orphaned (no Deployment /
// StatefulSet / DaemonSet references it) and get reported as unused.
var batchKinds = []string{"jobs", "cronjobs"}

// batchWorkload is the minimal typed view of a Job or CronJob list
// item. Jobs carry the pod template at spec.template; CronJobs nest it
// under spec.jobTemplate.spec.template.
type batchWorkload struct {
	Kind     string            `json:"kind,omitempty"`
	Metadata WorkloadMetadata  `json:"metadata"`
	Spec     batchWorkloadSpec `json:"spec"`
}

type batchWorkloadSpec struct {
	Template    *WorkloadTemplate `json:"template,omitempty"`    // Job
	JobTemplate *batchJobTemplate `json:"jobTemplate,omitempty"` // CronJob
}

type batchJobTemplate struct {
	Spec struct {
		Template *WorkloadTemplate `json:"template,omitempty"`
	} `json:"spec,omitempty"`
}

func (b batchWorkload) podTemplate() *WorkloadTemplate {
	if b.Spec.Template != nil {
		return b.Spec.Template
	}
	if b.Spec.JobTemplate != nil {
		return b.Spec.JobTemplate.Spec.Template
	}
	return nil
}

// fetchBatchImageRefs lists the given batch kinds (a subset of
// batchKinds) and returns the image references plus a per-kind scan
// summary, reusing the same pagination contract as the primary
// workload scan.
func fetchBatchImageRefs(
	ctx context.Context,
	o *clusteropts.ClusterOptions,
	p *clusteropts.PaginationOptions,
	namespace, labelSelector string,
	kinds []string,
) ([]workloadImageRef, []workloadImageScan, error) {
	if len(kinds) == 0 {
		return nil, nil, nil
	}
	client, err := o.Prepare()
	if err != nil {
		return nil, nil, err
	}
	var (
		refs []workloadImageRef
		scan []workloadImageScan
	)
	for _, kind := range kinds {
		items, total, err := clusteropts.FetchAllKubeSphere[batchWorkload](ctx, client, p, func(page int) string {
			return buildListPath(kind, namespace, labelSelector, p, page)
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list %s: %w", kind, err)
		}
		refs = append(refs, collectBatchImageRefs(kind, items)...)
		scan = append(scan, workloadImageScan{
			Kind:          kind,
			ReturnedItems: len(items),
			TotalItems:    total,
			HasMore:       !p.All && len(items) < total,
		})
	}
	return refs, scan, nil
}

func collectBatchImageRefs(kindPlural string, items []batchWorkload) []workloadImageRef {
	var refs []workloadImageRef
	for _, b := range items {
		tmpl := b.podTemplate()
		if tmpl == nil {
			continue
		}
		base := workloadImageRef{
			Namespace: b.Metadata.Namespace,
			Kind:      nonEmpty(b.Kind, SingularKind(kindPlural)),
			Workload:  b.Metadata.Name,
		}
		for _, c := range tmpl.Spec.InitContainers {
			if c.Image == "" {
				continue
			}
			ref := base
			ref.ContainerType = "initContainer"
			ref.Container = c.Name
			ref.Image = c.Image
			refs = append(refs, ref)
		}
		for _, c := range tmpl.Spec.Containers {
			if c.Image == "" {
				continue
			}
			ref := base
			ref.ContainerType = "container"
			ref.Container = c.Name
			ref.Image = c.Image
			refs = append(refs, ref)
		}
	}
	return refs
}
