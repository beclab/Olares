package market

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/whoami"
)

// userTotalsResponse mirrors framework/app-service/pkg/apiserver/utils.go's
// ListResult — the shape user-service forwards verbatim from
// /app-service/v1/users on the /api/users/v2 path:
//
//	{ "code": 200, "data": [UserInfo...], "totals": N }
//
// We only need totals; the actual UserInfo array is what
// `settings users list` consumes elsewhere and isn't needed here.
type userTotalsResponse struct {
	Code   int                      `json:"code"`
	Data   []map[string]interface{} `json:"data"`
	Totals int                      `json:"totals"`
}

// fetchUserTotals returns the number of Olares users on this instance,
// mirroring the SPA's `userStore.accounts.length` snapshot that
// csAppUninstall() in apps/.../stores/market/appService.ts uses to decide
// whether to show the "shared server" cascade checkbox.
//
// Endpoint: GET {DesktopURL}/api/users/v2 — user-service's role-filtered
// wrapper around app-service's /app-service/v1/users. Admin / owner sees
// the full instance count; non-admin sees only themselves (server-side
// filter in user-service, see cli/cmd/ctl/settings/users/list.go for the
// fully documented flow). A non-admin result of `totals=1` is therefore
// "single user as far as this caller is concerned", which is the same
// effective signal the SPA's `userStore.accounts.length` produces for the
// same caller — strict parity, no extra admin probe needed.
//
// Lives outside MarketClient because the endpoint isn't under
// /app-store/api/v2; it sits on the base DesktopURL the same way
// `settings users list`'s `prepare()` does. Building a whoami.HTTPClient
// ad-hoc here keeps MarketClient strictly market-scoped — same approach
// used elsewhere when a market verb needs to peek at a non-market endpoint.
func fetchUserTotals(ctx context.Context, opts *MarketOptions) (int, error) {
	if opts == nil || opts.factory == nil {
		return 0, fmt.Errorf("internal error: market options not wired with cmdutil.Factory")
	}
	rp, err := opts.factory.ResolveProfile(ctx)
	if err != nil {
		return 0, err
	}
	hc, err := opts.factory.HTTPClient(ctx)
	if err != nil {
		return 0, err
	}
	client := whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)

	var resp userTotalsResponse
	if err := client.DoJSON(ctx, "GET", "/api/users/v2", nil, &resp); err != nil {
		return 0, fmt.Errorf("GET /api/users/v2: %w", err)
	}
	switch resp.Code {
	case 0, 200:
		// Prefer Totals over len(Data): for admin callers both agree;
		// for non-admin callers user-service still populates Totals
		// on the filtered envelope (the server-side filter rewrites
		// data but keeps a per-caller totals count). Fall back to
		// len(Data) defensively in case an older user-service build
		// drops Totals — len(Data) is the worst-case lower bound and
		// still safely classifies the caller as "single user" when
		// it should.
		if resp.Totals > 0 {
			return resp.Totals, nil
		}
		return len(resp.Data), nil
	default:
		return 0, fmt.Errorf("/api/users/v2 returned code=%d", resp.Code)
	}
}
