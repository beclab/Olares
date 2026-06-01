package dashboard

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// BuildRankingEnvelope is the shared workload-grain ranking builder behind
// both `dashboard overview ranking` (default sortBy="cpu") and
// `dashboard applications` (which exposes --sort-by). It mirrors the SPA's
// `formatResult` in Applications2/config.ts: discover the user's app
// inventory via fetchAppsList, fan out to fetchWorkloadsMetrics, then emit
// one Item per workload carrying title / icon / state / pod count + the
// four metric values (cpu / memory / net_in / net_out).
//
// Hoisted to the pkg layer because it is the ONE legitimate horizontal
// share between cmd area subpackages: `applications` MUST NOT import
// `overview/ranking`. The cf parameter replaces the old cmd-side `common`
// global so this function stays cobra-agnostic.
//
// target is the optional --user override (empty = active profile owner).
// sortBy must be one of "cpu" / "memory" / "net_in" / "net_out".
// sortDir must be "asc" or "desc".
func BuildRankingEnvelope(ctx context.Context, c *Client, cf *CommonFlags, target, sortBy, sortDir string, now time.Time) (Envelope, error) {
	apps, userNs, err := LoadAppsForRanking(ctx, c, target)
	if err != nil {
		return Envelope{Kind: KindOverviewRanking}, err
	}
	rows, err := FetchWorkloadsMetrics(ctx, c, cf, WorkloadRequest{
		Apps: apps, UserNamespace: userNs, SortBy: sortBy, Sort: sortDir,
	}, DefaultClusterWindow(), now)
	if err != nil {
		return Envelope{Kind: KindOverviewRanking}, err
	}
	items := make([]Item, 0, len(rows))
	for i, r := range rows {
		title := r.Title
		if title == "" {
			title = r.Name
		}
		raw := map[string]any{
			"rank":       i + 1,
			"app":        r.Name,
			"title":      title,
			"icon":       r.Icon,
			"namespace":  r.Namespace,
			"deployment": r.Deployment,
			"owner_kind": r.OwnerKind,
			"state":      r.State,
			"is_system":  r.IsSystem,
			"pods":       r.PodCount,
			"cpu":        r.CPU,
			"memory":     r.Memory,
			"net_in":     r.NetIn,
			"net_out":    r.NetOut,
		}
		state := r.State
		if state == "" {
			state = "Unknown"
		}
		display := map[string]any{
			"rank":      strconv.Itoa(i + 1),
			"app":       title,
			"namespace": r.Namespace,
			"state":     state,
			"pods":      strconv.Itoa(r.PodCount),
			"cpu":       fmt.Sprintf("%.3f", r.CPU),
			"memory":    format.GetDiskSize(FormatFloat(r.Memory)),
			"net_in":    format.GetThroughput(FormatFloat(r.NetIn)),
			"net_out":   format.GetThroughput(FormatFloat(r.NetOut)),
		}
		items = append(items, Item{Raw: raw, Display: display})
	}
	return Envelope{
		Kind:  KindOverviewRanking,
		Meta:  NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), target),
		Items: items,
	}, nil
}

// LoadAppsForRanking discovers the active user's app inventory the way the
// SPA does — via /user-service/api/myapps_v2 (FetchAppsList). Each app
// entry is then tagged `IsSystem` based on whether it lives in the user's
// `user-space-<username>` namespace, mirroring
// Applications2/IndexPage.vue:330 (`userNamespace = "user-space-${username}"`).
//
// Returns the apps + the user's `user-space-` namespace so the per-pod
// monitoring fetch can target the right ns.
func LoadAppsForRanking(ctx context.Context, c *Client, target string) ([]WorkloadApp, string, error) {
	user, err := ResolveTargetUser(ctx, c, target)
	if err != nil {
		return nil, "", err
	}
	if user.Name == "" {
		return nil, "", fmt.Errorf("LoadAppsForRanking: empty username (server response missing user.username)")
	}
	userNs := fmt.Sprintf("user-space-%s", user.Name)

	raws, err := FetchAppsList(ctx, c)
	if err != nil {
		return nil, "", err
	}
	apps := make([]WorkloadApp, 0, len(raws))
	for _, it := range raws {
		apps = append(apps, WorkloadApp{
			Name:       it.Name,
			Title:      it.Title,
			Icon:       it.Icon,
			Namespace:  it.Namespace,
			Deployment: it.Deployment,
			OwnerKind:  it.OwnerKind,
			State:      it.State,
			IsSystem:   it.Namespace == userNs,
		})
	}
	return apps, userNs, nil
}
