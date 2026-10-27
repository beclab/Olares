package apiserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	userspacev1 "github.com/beclab/Olares/framework/app-service/pkg/users/userspace/v1"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"github.com/beclab/Olares/framework/app-service/pkg/webhook/userspaceprep"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// resolveUserspaceRoots derives the absolute host paths of the four
// userspace conventional roots for cfg, gated by the permissions the app
// actually declared.
//
// The pod spec only carries rendered absolute paths, with no trace of the
// helm values they came from, so the roots have to be recomputed here.
// The bfl StatefulSet annotations are the same source the rest of the
// platform reads (see apputils.TryToGetAppdataDirFromDeployment), and the
// path assembly below deliberately mirrors HelmOps.SetValues so the
// strings match what the chart was rendered with.
func (h *Handler) resolveUserspaceRoots(ctx context.Context, cfg *appcfg.ApplicationConfig) (userspaceprep.Roots, error) {
	var roots userspaceprep.Roots

	perms := appinstaller.ParseAppPermission(cfg.Permission)
	var wantAppData, wantAppCache, wantUserData, wantAppCommon bool
	for _, p := range perms {
		switch p {
		case appcfg.AppDataRW:
			wantAppData = true
		case appcfg.AppCacheRW:
			wantAppCache = true
		case appcfg.UserDataRW:
			wantUserData = true
		case appcfg.AppCommonRW:
			wantAppCommon = true
		}
	}

	// appCommon is cluster-wide, so it is derivable without bfl.
	if wantAppCommon {
		rootPath := userspacev1.DefaultRootPath
		if p := os.Getenv(userspacev1.OlaresRootPath); p != "" {
			rootPath = p
		}
		roots.AppCommon = fmt.Sprintf("%s/rootfs/Common", rootPath)
	}

	if !wantAppData && !wantAppCache && !wantUserData {
		return roots, nil
	}

	var sts appsv1.StatefulSet
	key := types.NamespacedName{Name: "bfl", Namespace: utils.UserspaceName(cfg.OwnerName)}
	if err := h.ctrlClient.Get(ctx, key, &sts); err != nil {
		return userspaceprep.Roots{}, fmt.Errorf("get bfl statefulset for owner %s: %w", cfg.OwnerName, err)
	}

	userspaceHostPath := sts.GetAnnotations()[constants.UserSpaceDirKey]
	appCacheHostPath := sts.GetAnnotations()[constants.UserAppDataDirKey]

	if wantAppData && userspaceHostPath != "" {
		roots.AppData = filepath.Join(fmt.Sprintf("%s/Data", userspaceHostPath), cfg.AppName)
	}
	if wantUserData && userspaceHostPath != "" {
		roots.UserData = fmt.Sprintf("%s/Home", userspaceHostPath)
	}
	if wantAppCache && appCacheHostPath != "" {
		roots.AppCache = filepath.Join(appCacheHostPath, cfg.AppName)
	}
	return roots, nil
}

// injectUserspacePrepare prepends the olares-prepare-userspace init
// container to pod, along with any hostPath volume it needs to reach a
// conventional root the chart did not declare itself.
//
// Nothing is injected when the pod mounts no userspace directory, so
// apps without a userspace permission are left exactly as they are.
func (h *Handler) injectUserspacePrepare(ctx context.Context, pod *corev1.Pod, cfg *appcfg.ApplicationConfig) error {
	for _, c := range pod.Spec.InitContainers {
		if c.Name == constants.UserspacePrepareInitContainerName {
			return nil
		}
	}

	roots, err := h.resolveUserspaceRoots(ctx, cfg)
	if err != nil {
		return err
	}
	if roots.IsEmpty() {
		return nil
	}

	plan, ok := userspaceprep.BuildPlan(pod, roots)
	if !ok {
		return nil
	}

	pod.Spec.Volumes = append(pod.Spec.Volumes, plan.ExtraVolumes...)
	// Prepare runs before the chart's own init containers: those often
	// seed config or unpack data into the very directories it owns.
	pod.Spec.InitContainers = append(
		[]corev1.Container{userspaceprep.InitContainer(plan)},
		pod.Spec.InitContainers...,
	)
	klog.Infof("prepare-userspace: injected for app=%s targets=%d", cfg.AppName, len(plan.Targets))
	return nil
}
