package ports

import "context"

// MarketChartRepo abstracts the chart-repository HTTP fetch + local
// tar cache used by upgrades. Only upgrading_app.go uses it today,
// but it is a first-class port because:
//
//   - the fetch does external HTTP (not just K8s API), so it must not
//     leak into WorkloadOps / AMStore
//   - it writes to a local directory (appcfg.ChartsPath) — an actual
//     filesystem side effect that must be Runner-executed
//   - swapping the market source (local static-index vs remote
//     registry vs OCI) is a plausible future change; keeping the
//     seam explicit avoids a deep refactor when that lands
//
// Legacy call sites:
//   - upgrading_app.go:214-244   apputils.GetAppConfig (user apps)
//     apputils.GetIndexAndDownloadChart (sys apps)
//     → pkg/utils/app.go:1173-1237
//     → resty HTTP + go-getter tar extract
type MarketChartRepo interface {
	// GetAppConfig fetches the ApplicationConfig for the target app
	// at the requested version from the market. HANDLER-SAFE READ
	// (Runner materialises into domain.World.UpgradeChart — field
	// TBD when writing the Upgrade handler).
	//
	// Returns `any` — concrete type:
	// *pkg/appcfg.ApplicationConfig. Adapter unmarshals the fetched
	// chart into it and returns the pointer. See design note on
	// WorkloadRequest.AppCfg for why the port surface is typed as
	// `any` rather than the concrete type.
	//
	// `token` is the user's kubesphere token, used for private
	// market entries. `marketSource` selects between the static
	// bundled index and remote repositories.
	GetAppConfig(ctx context.Context, req MarketFetchRequest) (any, error)

	// FetchChart downloads the chart tarball for (name, version) to
	// the local charts cache and returns the on-disk path. Written
	// under appcfg.ChartsPath. HANDLER-SAFE only in the sense that
	// it's idempotent (re-downloads over existing paths); the Runner
	// still schedules it as a domain.CallWorkload so its runtime is
	// bounded and its failure feeds WorkloadFailed cleanly.
	FetchChart(ctx context.Context, req MarketFetchRequest) (localPath string, err error)
}

// MarketFetchRequest carries the identifiers the two methods need.
// AdminUsername + ClusterOwner are pre-resolved by Identity so this
// port stays free of user-lookup concerns.
type MarketFetchRequest struct {
	Name          string
	Version       string
	Owner         string
	AdminUsername string
	ClusterOwner  string
	Token         string
	MarketSource  string
	IsSystemApp   bool
}
