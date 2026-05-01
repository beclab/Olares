// Package cluster implements the `olares-cli cluster` umbrella command
// tree. Identity (which Olares user) and transport (which cluster) are
// resolved from the currently-selected profile via cmdutil.Factory
// (switch with `olares-cli profile use <name>`), exactly like
// `olares-cli settings` and `olares-cli market`.
//
// The subtree exposes a per-user K8s view of an Olares instance —
// pods / workloads / containers / applications, plus the supporting
// nouns (namespaces, nodes, middleware, jobs). It is the CLI
// counterpart of the ControlHub SPA at
// apps/packages/app/src/apps/controlHub.
//
// Boundary notes:
//
//   - This subtree is NOT cluster maintenance. The existing `olares-cli
//     node` / `gpu` / `os` trees handle host-side install and upgrade
//     via kubeconfig; this tree is purely runtime, profile-based, and
//     visibility-scoped per user.
//
//   - Per-user resource scoping is enforced server-side. CLI verbs
//     here MUST NOT consult the locally cached cluster context to
//     decide whether a call is allowed; the server is the only
//     authority. The cache is strictly a display + error-message
//     convenience. See skills/olares-cluster/SKILL.md for the full
//     rationale.
//
//   - Single-app lifecycle (install / uninstall / upgrade / start /
//     stop) belongs to `olares-cli market`. `cluster application ...`
//     here is the runtime-state view (which workloads are healthy,
//     which pod isn't ready) — never a lifecycle mutator.
package cluster

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/application"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/container"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/cronjob"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/job"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/middleware"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/namespace"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/node"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/workload"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewClusterCommand assembles the `olares-cli cluster` subtree. Verbs
// are added incrementally (Phase 0 ships `cluster context`; Phase 1a
// adds `cluster pod`; Phase 1d (partial) adds `cluster application`);
// the umbrella's --help is the directory of available areas from day
// one even when individual nouns are still pending later phases.
//
// Authentication and transport are inherited from the shared
// cmdutil.Factory (set up in cli/cmd/ctl/root.go) so the
// currently-selected profile flows through unchanged. No per-command
// auth flags, and no per-invocation profile override — switch with
// `olares-cli profile use <name>` ahead of time.
func NewClusterCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Operate the Olares K8s cluster from the active user's profile (per-user view)",
		Long: `Operate the Olares K8s cluster from the active user's profile.

This umbrella exposes a per-user view of the underlying Kubernetes /
KubeSphere cluster — the same surface the ControlHub SPA exposes under
"Cluster" / "Applications" / "Pods". Identity and transport come from
the currently-selected profile (switch with "olares-cli profile use
<name>"), so authentication uses the same access token as "olares-cli
profile login" and the same edge auth chain the Olares web app uses
(Authelia + l4-bfl-proxy).

The base URL is https://control-hub.<terminus>; the same origin fans
out to /capi/* (Olares aggregator), /api/v1/* and /apis/<g>/<v>/*
(K8s native), /kapis/* (KubeSphere paginated), /middleware/* (Olares
middleware controller).

Per-user scoping (which namespaces, workloads, pods you can see) is
enforced server-side by the ControlHub backend. CLI verbs here never
gate based on a locally cached role — the server is the only
authority. Run "cluster context" to see what the server says about
your identity / role / accessible workspaces.

For host-side cluster maintenance (install, upgrade, node operations)
see "olares-cli node" / "gpu" / "os". For app-store-level lifecycle
(install / uninstall / start / stop) see "olares-cli market"; this
tree is the runtime-state view of the resulting K8s objects.
`,
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewContextCommand(f))
	cmd.AddCommand(pod.NewPodCommand(f))
	cmd.AddCommand(container.NewContainerCommand(f))
	cmd.AddCommand(workload.NewWorkloadCommand(f))
	cmd.AddCommand(application.NewApplicationCommand(f))
	cmd.AddCommand(namespace.NewNamespaceCommand(f))
	cmd.AddCommand(node.NewNodeCommand(f))
	cmd.AddCommand(middleware.NewMiddlewareCommand(f))
	cmd.AddCommand(job.NewJobCommand(f))
	cmd.AddCommand(cronjob.NewCronJobCommand(f))

	return cmd
}
