package dashboard

import (
	"context"
	"net/http"
)

// ----------------------------------------------------------------------------
// FetchAppsList — myapps_v2 (the SPA's appList store source)
// ----------------------------------------------------------------------------

// RawAppListItem mirrors the subset of `AppListItem`
// (controlPanelCommon/network/network.ts:280) the workload merge consumes.
// We tolerate unknown extra fields — the BFF freely adds new ones, and
// nothing in this struct has tag `json:"-"` to swallow them.
type RawAppListItem struct {
	ID         string                   `json:"id"`
	Name       string                   `json:"name"`
	Title      string                   `json:"title"`
	Icon       string                   `json:"icon"`
	Namespace  string                   `json:"namespace"`
	Deployment string                   `json:"deployment"`
	OwnerKind  string                   `json:"ownerKind"`
	State      string                   `json:"state"`
	Entrances  []map[string]interface{} `json:"entrances"`
}

// FetchAppsList queries `/user-service/api/myapps_v2` and returns the
// SPA's `appsWithNamespace` selector (entries with at least one entrance).
// Empty entrance list ⇒ filtered out, mirroring `appsWithNamespace` in
// stores/AppList.ts.
func FetchAppsList(ctx context.Context, c *Client) ([]RawAppListItem, error) {
	var raw struct {
		Code    int              `json:"code"`
		Message string           `json:"message"`
		Data    []RawAppListItem `json:"data"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, "/user-service/api/myapps_v2", nil, nil, &raw); err != nil {
		return nil, err
	}
	out := make([]RawAppListItem, 0, len(raw.Data))
	for _, it := range raw.Data {
		if len(it.Entrances) == 0 {
			continue
		}
		out = append(out, it)
	}
	return out, nil
}
