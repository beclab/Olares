package pod

import (
	"context"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/picker"
	"github.com/beclab/Olares/cli/pkg/clusterctx"
)

// resolveExecIdentity fetches the active profile's ControlHub identity
// (username / globalrole / systemNamespaces) straight from /capi/app/detail —
// the same endpoint the SPA's AppDetail store hits. We deliberately fetch
// fresh (cfg=nil ⇒ in-memory only, no config.json write) rather than reading
// the cached clusterContext: the exec gate applies the SPA's hasPermission
// rule to *server-authoritative* identity, so it never trusts (or drifts with)
// the local cache. A fetch failure is returned as-is — it carries the
// canonical "profile login" CTA, and since the exec WebSocket rides the same
// ControlHub ingress, a detail-endpoint failure means the dial would fail too.
func resolveExecIdentity(ctx context.Context, o *clusteropts.ClusterOptions) (clusterctx.Info, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return clusterctx.Info{}, err
	}
	rp, err := o.Factory().ResolveProfile(ctx)
	if err != nil {
		return clusterctx.Info{}, err
	}
	res, err := clusterctx.FetchAndCache(ctx, client, nil, rp.OlaresID, time.Now)
	if err != nil {
		return clusterctx.Info{}, err
	}
	return res.Info, nil
}

// gateExecPermission enforces the SPA's per-namespace exec rule client-side so
// `cluster {pod,container} exec` matches ControlHub: the main account
// (platform-admin) cannot open a shell inside a sub-account's container
// (e.g. user-space-alice), and non-admins are confined to their own
// namespaces. Viewing pods/containers/images/logs is unaffected — this gates
// exec only, exactly like the SPA hides just the Terminal button.
func gateExecPermission(ctx context.Context, o *clusteropts.ClusterOptions, namespace string) error {
	info, err := resolveExecIdentity(ctx, o)
	if err != nil {
		return err
	}
	if info.Username == "" {
		return fmt.Errorf(
			"cannot verify exec permission: the ControlHub identity for the active profile is empty; " +
				"run `olares-cli cluster context --refresh` and try again")
	}
	if !info.CanExec(namespace) {
		return execPermissionError(info, namespace)
	}
	return nil
}

// execPermissionError renders the "not allowed to exec here" message,
// tailored to whether the caller is an admin (whose scope also spans system /
// *-shared / os-protected namespaces) or a regular user. Only admins are told
// they can still view the namespace — a regular user usually can't see another
// user's namespace at all.
func execPermissionError(info clusterctx.Info, namespace string) error {
	who := fmt.Sprintf("%s (%s)", info.Username, clusterctx.FriendlyGlobalRole(info.GlobalRole))
	if info.GlobalRole == clusterctx.GlobalRoleAdmin {
		return fmt.Errorf(
			"permission denied: %s may not exec into namespace %q. "+
				"You can exec only into your own namespaces, system namespaces, any *-shared namespace, or os-protected; "+
				"you can still list/view its pods, containers, images and logs",
			who, namespace)
	}
	return fmt.Errorf(
		"permission denied: %s may not exec into namespace %q. "+
			"You can exec only into your own namespaces",
		who, namespace)
}

// filterExecEntries drops picker entries whose namespace the identity may not
// exec into, so the interactive picker only offers containers the user can
// actually open — mirroring the SPA, which never renders a Terminal button for
// off-limits namespaces.
func filterExecEntries(entries []picker.Entry, info clusterctx.Info) []picker.Entry {
	if info.Username == "" {
		return entries
	}
	out := entries[:0]
	for _, e := range entries {
		if info.CanExec(e.Namespace) {
			out = append(out, e)
		}
	}
	return out
}
