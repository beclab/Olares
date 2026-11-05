package helmops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
		_, err = apputils.GetIndexAndDownloadChart(ctx, opts)
		return err
	}
	cfg, err := resolveAppCfg(req)
	if err != nil {
		return err
	}
	opts.SelectedGpu = cfg.SelectedGpuType
	appConfig, _, err := apputils.GetAppConfig(ctx, opts)
	if err != nil {
		return err
	}
	appConfig.Ports = cfg.Ports
	appConfig.TailScale = cfg.TailScale
	updated, err := json.Marshal(appConfig)
	if err != nil {
		return err
	}
	cp := req.AM.DeepCopy()
	cp.Spec.Config = string(updated)
	return o.c.Patch(ctx, cp, client.MergeFrom(req.AM))
}
