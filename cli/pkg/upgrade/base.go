package upgrade

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/preinstall"
	"github.com/beclab/Olares/cli/pkg/storage"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/beclab/Olares/cli/pkg/utils"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type breakingUpgraderBase struct {
	upgraderBase
}

func (u breakingUpgraderBase) AddedBreakingChange() bool {
	return true
}

// upgraderBase is the general-purpose upgrader implementation
// for upgrading across versions without any breaking changes.
// Other implementations of breakingUpgrader,
// targeted for versions with breaking changes,
// should use this as a base for injecting and/or rewriting specific tasks as needed
type upgraderBase struct{}

func (u upgraderBase) AddedBreakingChange() bool {
	return false
}

func (u upgraderBase) NeedRestart() bool {
	return false
}

// ReplaceBinaries and ImportImages are empty in the base: swapping the
// binaries and loading the images are carried out by olaresd before it hands
// a stage to olares-cli, so no olares-cli task implements them today. They
// exist as stages because the flow has those points, and a version that needs
// to do something extra there now has somewhere to put it.
func (u upgraderBase) ReplaceBinaries() []task.Interface { return nil }

func (u upgraderBase) ImportImages() []task.Interface { return nil }

// PreUpgradeNode is what every machine does to itself before the cluster is
// upgraded. /etc/hosts is the machine's own file, and the entries it manages
// are how this node resolves the others.
func (u upgraderBase) PreUpgradeNode() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "Update Hosts",
			Action: new(terminus.UpdateKubeKeyHosts),
			Retry:  5,
		},
	}
}

func (u upgraderBase) PostUpgradeNode() []task.Interface { return nil }

func (u upgraderBase) RebootNodes() []task.Interface { return nil }

func (u upgraderBase) PrepareForUpgrade() []task.Interface {
	var tasks []task.Interface
	tasks = append(tasks, upgradeKSCore()...)
	tasks = append(tasks,
		&task.LocalTask{
			Name:   "PrepareUserInfoForUpgrade",
			Action: new(prepareUserInfoForUpgrade),
			Retry:  5,
		},
	)
	return tasks
}

func (u upgraderBase) ClearAppChartValues() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "ClearAppChartValues",
			Action: new(terminus.ClearAppValues),
		},
	}
}

func (u upgraderBase) ClearBFLChartValues() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "ClearBFLChartValues",
			Action: new(terminus.ClearBFLValues),
		},
	}
}

func (u upgraderBase) UpdateChartsInAppService() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "UpdateChartsInAppService",
			Action: new(terminus.CopyAppServiceHelmFiles),
			Retry:  5,
		},
	}
}

func (u upgraderBase) UpgradeUserComponents() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeUserComponents",
			Action: new(upgradeUserComponents),
			Retry:  5,
			Delay:  15 * time.Second,
		},
	}
}

// UpdateReleaseFile stamps the new version onto each machine. It feeds the
// post-upgrade-node stage: /etc/olares/release and the .prepared/.installed
// markers are per-node facts — they are what that node's olaresd reports about
// itself — so a node the control node stamped on its behalf would claim a
// version it is not running.
func (u upgraderBase) UpdateReleaseFile() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "UpdateReleaseFile",
			Action: new(terminus.WriteReleaseFile),
		},
		&task.LocalTask{
			Name:   "UpdatePreparedMarker",
			Action: new(terminus.PrepareFinished),
		},
		&task.LocalTask{
			Name:   "UpdateInstalledMarker",
			Action: new(terminus.InstallFinished),
		},
	}
}

func (u upgraderBase) UpgradeSystemComponents() []task.Interface {
	// this task updates the version in the CR
	// so put this at last to make the whole pipeline
	// reentrant
	tasks := []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeGPUPlugin",
			Action: new(gpu.InstallPlugin),
		},
	}
	// same Intel GPU stack path as install (labels / drivers / NFD / plugin / xpumd)
	tasks = append(tasks, upgradeIntelGPUPlugin()...)
	tasks = append(tasks,
		&task.LocalTask{
			Name:   "UpgradeSettings",
			Action: new(upgradeSettings),
			Retry:  10,
			Delay:  15 * time.Second,
		},
		&task.LocalTask{
			Name:   "UpgradeSystemEnvs",
			Action: new(terminus.ApplySystemEnv),
			Retry:  5,
			Delay:  15 * time.Second,
		},
		&task.LocalTask{
			Name:   "UpgradeSystemComponents",
			Action: new(upgradeSystemComponents),
			Retry:  10,
			Delay:  15 * time.Second,
		},
		&task.LocalTask{
			Name:   "UpgradeUserEnvs",
			Action: new(terminus.CreateUserEnvConfigMap),
			Retry:  5,
			Delay:  15 * time.Second,
		},
		&task.LocalTask{
			Name:   "UpgradeL4BFLProxy",
			Action: &upgradeL4BFLProxy{Tag: "v0.3.46"},
			Retry:  6,
			Delay:  15 * time.Second,
		},
	)
	return tasks
}

func (u upgraderBase) UpdateOlaresVersion() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "UpdateOlaresVersion",
			Action: new(updateOlaresVersion),
		},
	}
}

func (u upgraderBase) PostUpgrade() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "EnsureSystemComponentsUpAndRunningAgain",
			Action: new(terminus.CheckSystemComponentsReady),
			Delay:  15 * time.Second,
			Retry:  60,
		},
	}
}

type prepareUserInfoForUpgrade struct {
	common.KubeAction
}

func (p *prepareUserInfoForUpgrade) Execute(runtime connector.Runtime) error {
	users, adminUser, err := listUsersToUpgrade()
	if err != nil {
		return err
	}
	p.PipelineCache.Set(common.CacheUpgradeUsers, users)
	p.PipelineCache.Set(common.CacheUpgradeAdminUser, adminUser)

	return nil
}

// listUsersToUpgrade reads the Olares users whose components this upgrade has
// to touch, and names the owner among them.
//
// It is a function rather than only a cached task result because the task that
// records it and the task that consumes it are not guaranteed to share a
// process: on a cluster the upgrade is cut into stages, each its own
// olares-cli run, and the pipeline cache does not cross that boundary. The
// cache is still written — reading it is cheaper and keeps the early failure
// on "no admin user" where it was — but the consumer can do without it.
func listUsersToUpgrade() ([]iamv1alpha2.User, string, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get rest config: %s", err)
	}
	scheme := kruntime.NewScheme()
	err = iamv1alpha2.AddToScheme(scheme)
	if err != nil {
		return nil, "", fmt.Errorf("failed to add user scheme: %s", err)
	}
	userClient, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create client: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	var userList iamv1alpha2.UserList
	err = userClient.List(ctx, &userList)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list users: %s", err)
	}
	var usersToUpgrade []iamv1alpha2.User
	var adminUser string
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create kubernetes client: %s", err)
	}
	for _, user := range userList.Items {
		if user.Status.State == "Failed" {
			logger.Infof("skipping user %s that failed to be created", user.Name)
			continue
		}
		if user.Status.State == "Deleting" || user.DeletionTimestamp != nil {
			logger.Infof("skipping user %s that's being deleted", user.Name)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		_, err := client.CoreV1().Namespaces().Get(ctx, fmt.Sprintf("user-space-%s", user.Name), metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Infof("ignoring non-olares user: %s", user.Name)
				continue
			}
			return nil, "", fmt.Errorf("failed to get user-space-%x: %v", user.Name, err)
		}
		usersToUpgrade = append(usersToUpgrade, user)
		if role, ok := user.Annotations["bytetrade.io/owner-role"]; ok && role == "owner" {
			adminUser = user.Name
		}
	}
	if len(usersToUpgrade) > 0 {
		logger.Infof("found %d users to upgrade", len(usersToUpgrade))
	}
	if adminUser == "" {
		return nil, "", fmt.Errorf("no admin user found")
	}
	return usersToUpgrade, adminUser, nil
}

type upgradeUserComponents struct {
	common.KubeAction
}

// usersToUpgrade prefers what PrepareUserInfoForUpgrade recorded and reads the
// cluster when that task ran in another process. See listUsersToUpgrade.
func (u *upgradeUserComponents) usersToUpgrade() ([]iamv1alpha2.User, string, error) {
	usersCache, usersOK := u.PipelineCache.Get(common.CacheUpgradeUsers)
	adminCache, adminOK := u.PipelineCache.Get(common.CacheUpgradeAdminUser)
	if usersOK && adminOK {
		users, ok := usersCache.([]iamv1alpha2.User)
		admin, adminIsString := adminCache.(string)
		if ok && adminIsString && admin != "" {
			return users, admin, nil
		}
	}
	return listUsersToUpgrade()
}

func (u *upgradeUserComponents) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %s", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %s", err)
	}

	users, adminUser, err := u.usersToUpgrade()
	if err != nil {
		return err
	}

	bflChartPath := path.Join(runtime.GetInstallerDir(), "wizard/config/launcher")

	appsChartDir := path.Join(runtime.GetInstallerDir(), "wizard", "config", "apps")
	appEntries, err := os.ReadDir(appsChartDir)
	if err != nil {
		return fmt.Errorf("failed to list %s: %v", appsChartDir, err)
	}
	var apps []string
	for _, entry := range appEntries {
		if entry.IsDir() {
			apps = append(apps, entry.Name())
		}
	}

	for _, user := range users {
		logger.Infof("upgrading for user: %s", user.Name)
		ns := fmt.Sprintf("user-space-%s", user.Name)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		sts, err := client.AppsV1().StatefulSets(ns).Get(ctx, "bfl", metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get bfl statefulset for user %s: %v", user.Name, err)
		}
		if sts == nil {
			return fmt.Errorf("bfl statefulset for user %s not found", user.Name)
		}
		bflVals := make(map[string]interface{})
		bflInnerVals := make(map[string]interface{})

		// these values are generated by helm during installation
		// and should be retrieved before upgrade
		// for other values, we can reuse them
		for _, key := range []string{"userspace_rand16", "appcache_rand16", "dbdata_rand16",
			"userspace_pv", "userspace_pvc",
			"appcache_pv", "appcache_pvc",
			"dbdata_pv", "dbdata_pvc"} {
			bflInnerVals[key] = sts.Annotations[key]
		}
		bflVals["bfl"] = bflInnerVals

		actionConfig, settings, err := utils.InitConfig(config, ns)
		if err != nil {
			return err
		}
		var bflReleaseName = fmt.Sprintf("launcher-%s", user.Name)
		if err := utils.UpgradeCharts(ctx, actionConfig, settings, bflReleaseName, bflChartPath, "", ns, bflVals, true); err != nil {
			return fmt.Errorf("failed to upgrade launcher: %v", err)
		}

		var wizardNeedUpgrade bool
		if wizardStatus, ok := user.Annotations["bytetrade.io/wizard-status"]; !ok || wizardStatus != "completed" {
			wizardNeedUpgrade = true
		}

		for _, app := range apps {
			if !wizardNeedUpgrade && app == "wizard" {
				logger.Debugf("skipping upgrade wizard as user %s is already activated", user.Name)
				continue
			}
			releaseName := app
			if user.Name != adminUser {
				releaseName = fmt.Sprintf("%s-%s", app, user.Name)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			if err := utils.UpgradeCharts(ctx, actionConfig, settings, releaseName, path.Join(appsChartDir, app), "", ns, nil, true); err != nil {
				return fmt.Errorf("failed to upgrade app %s: %v", app, err)
			}
		}

	}
	return nil
}

type upgradeSettings struct {
	common.KubeAction
}

func (u *upgradeSettings) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %s", err)
	}
	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceDefault)
	if err != nil {
		return err
	}
	ctx, cancelSettings := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancelSettings()
	settingsChartPath := path.Join(runtime.GetInstallerDir(), "wizard", "config", "settings")
	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameSettings, settingsChartPath, "", common.NamespaceDefault, nil, true); err != nil {
		return err
	}
	return nil
}

type upgradeSystemComponents struct {
	common.KubeAction
}

func (u *upgradeSystemComponents) Execute(runtime connector.Runtime) error {
	config, configErr := ctrl.GetConfig()
	if configErr != nil {
		return fmt.Errorf("failed to get rest config: %s", configErr)
	}

	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceOsPlatform)
	if err != nil {
		return err
	}
	ctx, cancelPlatform := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancelPlatform()
	platformChartPath := path.Join(runtime.GetInstallerDir(), "wizard", "config", "os-platform")
	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameOSPlatform, platformChartPath, "", common.NamespaceOsPlatform, nil, true); err != nil {
		return err
	}

	actionConfig, settings, err = utils.InitConfig(config, common.NamespaceOsFramework)
	if err != nil {
		return err
	}
	ctx, cancelFramework := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancelFramework()
	frameworkChartPath := path.Join(runtime.GetInstallerDir(), "wizard", "config", "os-framework")
	vals := map[string]interface{}{
		"ensureApps": preinstall.EnsureAppsPublished(storage.OlaresRootDir),
	}
	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameOSFramework, frameworkChartPath, "", common.NamespaceOsFramework, vals, true); err != nil {
		return err
	}

	return nil
}

type updateOlaresVersion struct {
	common.KubeAction
}

func (u *updateOlaresVersion) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %s", err)
	}
	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceDefault)
	if err != nil {
		return err
	}
	ctx, cancelSettings := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancelSettings()
	settingsChartPath := path.Join(runtime.GetInstallerDir(), "wizard", "config", "settings")

	vals := map[string]interface{}{"version": u.KubeConf.Arg.OlaresVersion}
	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameSettings, settingsChartPath, "", common.NamespaceDefault, vals, true); err != nil {
		return err
	}
	return nil
}
