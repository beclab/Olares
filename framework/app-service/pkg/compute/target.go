package compute

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func resolveComputeTarget(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig, includeSharedServer bool) (*appcfg.ApplicationConfig, bool, error) {
	if appConfig == nil {
		return nil, false, fmt.Errorf("app config is nil")
	}
	if !appConfig.IsV2() || !appConfig.HasClusterSharedCharts() {
		return appConfig, true, nil
	}
	owner, found, err := sharedServerOwner(ctx, c, appConfig)
	if err != nil {
		return nil, false, err
	}
	if !found || owner == "" || owner == appConfig.OwnerName {
		return appConfig, true, nil
	}
	if !includeSharedServer {
		return appConfig, false, nil
	}
	target, err := loadAppConfigForOwner(ctx, c, appConfig.AppName, owner)
	if err != nil {
		return nil, false, err
	}
	return target, true, nil
}

func DeleteAllocationsForComputeTarget(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig, includeSharedServer bool) error {
	if appConfig == nil {
		return nil
	}
	// For v2 cluster-shared apps the only allocation row lives at
	// (appName, sharedServerOwner). When the caller does not intend to
	// touch the shared server (suspend / uninstall the client only),
	// leave that row and its HAMI bindings alone.
	if appConfig.IsV2() && appConfig.HasClusterSharedCharts() && !includeSharedServer {
		return nil
	}
	// resolveComputeTarget redirects to the actual server owner's config
	// when stop/uninstall-all is triggered by someone who is not the
	// original installer of the shared server; in every other reachable
	// case it returns appConfig unchanged.
	target, _, err := resolveComputeTarget(ctx, c, appConfig, includeSharedServer)
	if err != nil {
		return err
	}
	return DeleteAllocationsForApp(ctx, c, target.AppName, target.OwnerName)
}

func ManagesSharedServer(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig) (bool, error) {
	target, manage, err := resolveComputeTarget(ctx, c, appConfig, false)
	if err != nil {
		return false, err
	}
	return manage && target != nil && target.IsV2() && target.HasClusterSharedCharts(), nil
}

func ShouldIncludeSharedServerForResume(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig, isAdmin bool) (bool, error) {
	if !isAdmin {
		return false, nil
	}
	if appConfig == nil || !appConfig.IsV2() || !appConfig.HasClusterSharedCharts() {
		return false, nil
	}
	suspended, found, err := sharedServerSuspended(ctx, c, appConfig)
	if err != nil {
		return false, err
	}
	if found {
		return suspended, nil
	}
	return false, nil
}

func sharedServerSuspended(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig) (bool, bool, error) {
	for _, chart := range appConfig.SubCharts {
		if !chart.Shared {
			continue
		}
		namespace := appcfg.ChartNamespace(&chart, appConfig.OwnerName)
		var deployments appsv1.DeploymentList
		if err := c.List(ctx, &deployments, client.InNamespace(namespace)); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return false, false, err
		}
		found := false
		for i := range deployments.Items {
			found = true
			if !replicasSuspended(deployments.Items[i].Spec.Replicas) {
				return false, true, nil
			}
		}
		var statefulSets appsv1.StatefulSetList
		if err := c.List(ctx, &statefulSets, client.InNamespace(namespace)); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return false, false, err
		}
		for i := range statefulSets.Items {
			found = true
			if !replicasSuspended(statefulSets.Items[i].Spec.Replicas) {
				return false, true, nil
			}
		}
		if found {
			return true, true, nil
		}
	}
	return false, false, nil
}

func replicasSuspended(replicas *int32) bool {
	return replicas != nil && *replicas == 0
}

func sharedServerOwner(ctx context.Context, c client.Client, appConfig *appcfg.ApplicationConfig) (string, bool, error) {
	for _, chart := range appConfig.SubCharts {
		if !chart.Shared {
			continue
		}
		var ns corev1.Namespace
		err := c.Get(ctx, client.ObjectKey{Name: appcfg.ChartNamespace(&chart, appConfig.OwnerName)}, &ns)
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return "", false, err
		}
		if ns.Labels[constants.ApplicationNameLabel] != "" && ns.Labels[constants.ApplicationNameLabel] != appConfig.AppName {
			continue
		}
		return ns.Labels[constants.ApplicationInstallUserLabel], true, nil
	}
	return "", false, nil
}

func loadAppConfigForOwner(ctx context.Context, c client.Client, appName, owner string) (*appcfg.ApplicationConfig, error) {
	manager, err := loadAppManagerForOwner(ctx, c, appName, owner)
	if err != nil {
		return nil, err
	}
	var cfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(manager.Spec.Config), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func loadAppManagerForOwner(ctx context.Context, c client.Client, appName, owner string) (*appv1alpha1.ApplicationManager, error) {
	var managers appv1alpha1.ApplicationManagerList
	if err := c.List(ctx, &managers); err != nil {
		return nil, err
	}
	for i := range managers.Items {
		manager := &managers.Items[i]
		if manager.Spec.AppName != appName || manager.Spec.AppOwner != owner {
			continue
		}
		if manager.Spec.Type != appv1alpha1.App && manager.Spec.Type != appv1alpha1.Middleware {
			continue
		}
		return manager, nil
	}
	return nil, fmt.Errorf("compute target app %s owned by %s not found", appName, owner)
}
