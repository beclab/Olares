package appcfg

import (
	"context"
	"encoding/json"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

const (
	ChartsPath = "./charts"
)

func AppChartPath(app string) string {
	return ChartsPath + "/" + app
}

// GetAppInstallationConfig get app installation configuration from app store
func GetAppInstallationConfig(app, owner string) (*ApplicationConfig, error) {
	//chart := AppChartPath(rawAppName)
	appcfg, err := getAppConfigFromAppMgrConfig(app, owner)
	if err != nil {
		return nil, err
	}

	// TODO: app installation namespace
	var namespace string
	if appcfg.Namespace != "" {
		namespace, _ = utils.AppNamespace(app, owner, appcfg.Namespace)
	} else {
		namespace = fmt.Sprintf("%s-%s", app, owner)
	}

	appcfg.Namespace = namespace
	if err := ValidateCallerInClusterManifest(appcfg); err != nil {
		return nil, err
	}
	// Shared apps share one cluster-wide installation across admins.
	// Persist the cluster owner as a stable real-user identity so every
	// consumer (compute allocation, HAMI binding labels, pod labels,
	// kubesphere user APIs, user-scoped namespaces) sees the same value
	// no matter which admin operates the app. See
	// pkg/utils/app::GetAppConfig for the canonical write site at install
	// time. v3+per-user apps fall through to the regular per-user owner.
	if appcfg.IsShared() {
		clusterOwner, err := kubesphere.GetClusterOwner(context.TODO())
		if err != nil {
			return nil, err
		}
		appcfg.OwnerName = clusterOwner
	} else {
		appcfg.OwnerName = owner
	}

	return appcfg, nil
}

// getAppConfigFromAppMgrConfig loads the embedded ApplicationConfig from the
// ApplicationManager that backs the given app.
//
// Shared apps live in a cluster-wide AM named "{app}-shared-{app}" (see
// apputils.V3AppMgrName), while every other app (v1 and v3+per-user) uses
// the per-user "{app}-{owner}-{app}". The shared name is owner-independent
// and only exists for shared apps, so we try it first; if it is absent we
// fall back to the per-user name so the existing behaviour is preserved.
//
// The shared name format is intentionally inlined (rather than calling
// apputils.V3AppMgrName) to avoid an import cycle: pkg/utils/app already
// depends on pkg/appcfg.
func getAppConfigFromAppMgrConfig(appName, owner string) (*ApplicationConfig, error) {
	kclient, err := utils.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()
	sharedName := fmt.Sprintf("%s-shared-%s", appName, appName)
	am, err := kclient.AppV1alpha1().ApplicationManagers().Get(ctx, sharedName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
		legacyName := fmt.Sprintf("%s-%s-%s", appName, owner, appName)
		am, err = kclient.AppV1alpha1().ApplicationManagers().Get(ctx, legacyName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}

	appConfig := ApplicationConfig{}
	if err = json.Unmarshal([]byte(am.Spec.Config), &appConfig); err != nil {
		return nil, err
	}
	return &appConfig, nil
}
