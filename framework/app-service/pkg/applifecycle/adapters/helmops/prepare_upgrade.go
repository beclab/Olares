package helmops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
)

// PrepareUpgrade runs the pre-helm half of upgrading_app.go:exec:
// GetDeployedReleaseVersion + version guard, then GetAppConfig or
// GetIndexAndDownloadChart. The fetched config is patched back onto
// AM.Spec.Config so the following image-pull and helm-upgrade phases
// see the new refs.
//
// RepoURL / chart-version annotations empty means the HTTP handler
// already wrote Spec.Config and there is nothing to refresh — the
// call succeeds so handler tests and retries without those
// annotations still proceed.
func (o *Ops) PrepareUpgrade(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	if req.AM == nil {
		return ports.WorkloadResult{Err: fmt.Errorf("helmops: PrepareUpgrade: AM is nil")}
	}
	if err := o.guardUpgradeVersion(ctx, req); err != nil {
		return ports.WorkloadResult{LandedReason: "DowngradeNotAllowed", Err: err}
	}
	if err := o.refreshUpgradeChart(ctx, req); err != nil {
		return ports.WorkloadResult{LandedReason: "UpgradeChartFetchFailed", Err: err}
	}
	return ports.WorkloadResult{}
}

func (o *Ops) guardUpgradeVersion(ctx context.Context, req ports.WorkloadRequest) error {
	version := ""
	if req.AM.Annotations != nil {
		version = req.AM.Annotations[api.AppTargetChartVersionKey]
	}
	if version == "" {
		return nil
	}
	ns := req.Namespace
	if ns == "" {
		ns = req.AM.Spec.AppNamespace
	}
	deployed, _, err := o.GetDeployedVersion(ctx, ns, req.AM.Spec.AppName)
	if err != nil {
		return err
	}
	if !utils.MatchVersion(version, ">= "+deployed) {
		return errors.New("upgrade version should great than deployed version")
	}
	return nil
}

func (o *Ops) refreshUpgradeChart(ctx context.Context, req ports.WorkloadRequest) error {
	if req.AM.Annotations == nil {
		return nil
	}
	repoURL := req.AM.Annotations[api.AppRepoURLKey]
	if repoURL == "" {
		return nil
	}
	version := req.AM.Annotations[api.AppTargetChartVersionKey]
	token := resolveToken(req)
	marketSource := appcfg.GetMarketSource(req.AM)
	kubeConfig, err := resolveKubeConfig(req, o.kubeConfig)
	if err != nil {
		return err
	}
	admin, err := kubesphere.GetAdminUsername(ctx, kubeConfig)
	if err != nil {
		return err
	}
	isAdmin, err := kubesphere.IsAdmin(ctx, kubeConfig, req.AM.Spec.AppOwner)
	if err != nil {
		return err
	}
	if isAdmin {
		admin = req.AM.Spec.AppOwner
	}
	rawAppName := req.AM.Spec.RawAppName
	if rawAppName == "" {
		rawAppName = req.AM.Spec.AppName
	}
	opts := &apputils.ConfigOptions{
		App:          req.AM.Spec.AppName,
		Owner:        req.AM.Spec.AppOwner,
		RawAppName:   rawAppName,
		RepoURL:      repoURL,
		Version:      version,
		Token:        token,
		Admin:        admin,
		MarketSource: marketSource,
		IsAdmin:      isAdmin,
	}
	if userspace.IsSysApp(rawAppName) {
		// System apps are re-downloaded here (their Spec.Config is not
		// refreshed from the chart). Point ChartsName at the fresh
		// download so helm upgrade renders it rather than whatever an
		// older path holds.
		opts.ChartURL = req.AM.Annotations[api.AppChartURLKey]
		opts.ChartDigest = req.AM.Annotations[api.AppChartDigestKey]
		chartPath, err := apputils.GetIndexAndDownloadChart(ctx, opts)
		if err != nil {
			return err
		}
		cfg, cfgErr := resolveAppCfg(req)
		if cfgErr != nil || cfg.ChartsName == chartPath {
			// No stored config to repoint (earlier releases did not
			// need one here either).
			return nil
		}
		cfg.ChartsName = chartPath
		return o.patchSpecConfig(ctx, req, cfg)
	}
	cfg, err := resolveAppCfg(req)
	if err != nil {
		return err
	}
	opts.SelectedGpu = cfg.SelectedGpuType
	// Read the chart the upgrade handler downloaded and recorded as
	// ChartsName. AMs written by earlier releases carry the legacy flat
	// path there, which still resolves. If the directory is gone (cache
	// pruned, node rebuilt), download the target version again instead of
	// failing the upgrade.
	if chartDirExists(cfg.ChartsName) {
		opts.ChartPath = cfg.ChartsName
	} else {
		if opts.Version == "" {
			opts.Version = cfg.Version
		}
		opts.ChartURL = req.AM.Annotations[api.AppChartURLKey]
		opts.ChartDigest = req.AM.Annotations[api.AppChartDigestKey]
		opts.NeedDownloadChart = true
	}
	appConfig, _, err := apputils.GetAppConfig(ctx, opts)
	if err != nil {
		return err
	}
	appConfig.Ports = cfg.Ports
	appConfig.TailScale = cfg.TailScale
	return o.patchSpecConfig(ctx, req, appConfig)
}

func chartDirExists(p string) bool {
	if p == "" {
		return false
	}
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func (o *Ops) patchSpecConfig(ctx context.Context, req ports.WorkloadRequest, cfg *appcfg.ApplicationConfig) error {
	updated, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	cp := req.AM.DeepCopy()
	cp.Spec.Config = string(updated)
	return o.c.Patch(ctx, cp, client.MergeFrom(req.AM))
}
