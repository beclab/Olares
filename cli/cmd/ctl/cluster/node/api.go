package node

import (
	"context"
	"fmt"
	"net/url"
	"sort"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// Architectures returns the distinct CPU architectures reported by the
// cluster's nodes (status.nodeInfo.architecture — "amd64", "arm64", …).
//
// Exported as a Factory-based facade for the same reason as
// workload/api.go: clusteropts is internal to the cluster tree, so
// cmd/ctl/dev cannot build an options bag itself. `dev push` uses this
// to refuse an image built for the wrong architecture *before* paying
// for the transfer — a wrong-arch image imports cleanly and then fails
// at runtime as an opaque CrashLoopBackOff ("exec format error" buried
// in the container log), which is a genuinely expensive thing to debug.
func Architectures(ctx context.Context, f *cmdutil.Factory) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	o := clusteropts.NewClusterOptions(f)
	o.Quiet = true
	client, err := o.Prepare()
	if err != nil {
		return nil, err
	}
	p := clusteropts.NewPaginationOptions()
	p.All = true

	items, _, err := clusteropts.FetchAllKubeSphere[Node](ctx, client, p, func(page int) string {
		q := url.Values{}
		p.AppendQueryForPage(q, page)
		path := "/kapis/resources.kubesphere.io/v1alpha3/nodes"
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
		return path
	})
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}

	seen := map[string]struct{}{}
	var out []string
	for _, n := range items {
		arch := n.Status.NodeInfo.Architecture
		if arch == "" {
			continue
		}
		if _, dup := seen[arch]; dup {
			continue
		}
		seen[arch] = struct{}{}
		out = append(out, arch)
	}
	sort.Strings(out)
	return out, nil
}
