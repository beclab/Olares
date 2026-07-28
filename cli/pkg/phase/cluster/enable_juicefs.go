package cluster

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/k3s"
	"github.com/beclab/Olares/cli/pkg/kubernetes"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/storage"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/beclab/Olares/cli/pkg/utils"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	k8sclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// EnableJuiceFS builds the pipeline that converts a single-node Olares master
// from a local-filesystem rootfs to a JuiceFS-backed shared rootfs, in place,
// without uninstalling. Once it completes, the master satisfies the JuiceFS
// precondition required by `olares-cli node add`, so worker nodes can join.
//
// The caller must already have checked that JuiceFS is not yet enabled (the
// "already migrated" case short-circuits earlier, before this pipeline is even
// built). This pipeline therefore always performs the migration.
//
// The migration requires Olares to be stopped beforehand (it fails fast with a
// hint to run `olares-cli stop` otherwise): with the kubernetes/container layer
// down nothing writes to the rootfs, so a single full rsync migrates all the
// data in one pass. The bundled object store (MinIO) and metadata engine
// (Redis) are (re)started as part of the migration so they can serve the
// JuiceFS mount during the sync. Olares is brought back up at the end so the
// rootfs type and fsnotify charts can be updated.
func EnableJuiceFS(runtime *common.KubeRuntime, manifestMap manifest.InstallationManifest) *pipeline.Pipeline {
	manifestModule := manifest.ManifestModule{
		Manifest: manifestMap,
		BaseDir:  runtime.GetBaseDir(),
	}

	// the bundled MinIO is only installed when using the managed-minio
	// backend; for external object storage (s3/oss/cos/external minio) we
	// validate the provided credentials instead and JuiceFS is formatted
	// against that remote bucket.
	useManagedMinIO := runtime.Arg.Storage == nil || runtime.Arg.Storage.StorageType == common.ManagedMinIO

	// When the juicefs service is already enabled, the data has already been
	// migrated and the rootfs swapped, so a resumed run must skip every
	// migration step (prechecks, the object store / metadata engine install,
	// and the rootfs migration itself) and only re-run the idempotent
	// finalization modules below (StartOlares, the rootfs-type flip and the
	// fsnotify re-render). This is also why the prechecks must be skipped:
	// they require Olares to be stopped, whereas finalization needs it running.
	alreadyEnabled := storage.IsJuiceFSEnabled()

	modules := []module.Module{
		&MigratePrecheckModule{Skip: alreadyEnabled},
		&storage.ValidateModule{Skip: useManagedMinIO || alreadyEnabled},
		&storage.InstallMinioModule{
			ManifestModule: manifestModule,
			Skip:           !useManagedMinIO || alreadyEnabled,
		},
		&storage.InstallRedisModule{ManifestModule: manifestModule, Skip: alreadyEnabled},
		&MigrateRootFSToJuiceFSModule{
			ManifestModule: manifestModule,
			Skip:           alreadyEnabled,
		},
		// Always run (idempotent): regenerate the container-runtime service unit
		// so it gains the JuiceFS pre-check now that juicefs.service exists. This
		// lives in its own module rather than inside the migration module so that
		// a run interrupted between enabling the juicefs service and this step is
		// still repaired on resume (when the migration module is skipped).
		&RegenerateRuntimeServiceModule{},
		&terminus.StartOlaresModule{},
		&UpdateRootFSTypeModule{},
		&ReRenderFsnotifyChartsModule{},
	}

	return &pipeline.Pipeline{
		Name:    "Enable JuiceFS on the master node and migrate rootfs",
		Modules: modules,
		Runtime: runtime,
	}
}

// MigratePrecheckModule validates that the node can be migrated and that there
// is enough disk space to do so.
type MigratePrecheckModule struct {
	common.KubeModule
	Skip bool
}

func (m *MigratePrecheckModule) IsSkip() bool {
	return m.Skip
}

func (m *MigratePrecheckModule) Init() {
	m.Name = "MigratePrecheck"
	m.Tasks = []task.Interface{
		// Olares must be stopped before we touch the rootfs; bail out early
		// (with a hint to run `olares-cli stop`) if it is still running.
		&task.LocalTask{
			Name:   "CheckOlaresStopped",
			Action: new(storage.CheckOlaresStopped),
		},
		&task.LocalTask{
			Name:   "CheckMigrationPrecheck",
			Action: new(storage.CheckMigrationPrecheck),
		},
	}
}

// MigrateRootFSToJuiceFSModule installs the JuiceFS client, formats the
// filesystem, syncs the local rootfs into it in a single full pass (Olares is
// already stopped, so the rootfs is quiescent), swaps the rootfs directory for
// the JuiceFS mount, and enables the juicefs service. The container-runtime
// service unit is regenerated separately (RegenerateRuntimeServiceModule) so
// that step survives resumption even when this module is skipped.
type MigrateRootFSToJuiceFSModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool
}

func (m *MigrateRootFSToJuiceFSModule) IsSkip() bool {
	return m.Skip
}

func (m *MigrateRootFSToJuiceFSModule) Init() {
	m.Name = "MigrateRootFSToJuiceFS"

	// 1. install JuiceFS client + format the filesystem (no mount on rootfs yet)
	getRedisConfig := &task.RemoteTask{
		Name:   "GetRedisConfig",
		Hosts:  m.Runtime.GetHostsByRole(common.Master),
		Action: new(storage.GetOrSetRedisConfig),
		Retry:  1,
	}
	// the osnode-init daemonset subPath-mounts /usr/local/bin/juicefs, so on a
	// node that never had JuiceFS the path already exists as an empty directory;
	// remove it so CheckJuiceFsExists doesn't mistake it for an installed binary.
	cleanupStaleBinary := &task.LocalTask{
		Name:   "CleanupStaleJuiceFsBinaryPath",
		Action: new(storage.CleanupStaleJuiceFsBinaryPath),
	}
	installJuiceFs := &task.LocalTask{
		Name:    "InstallJuiceFs",
		Prepare: &storage.CheckJuiceFsExists{},
		Action: &storage.InstallJuiceFs{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Retry: 1,
	}
	formatJuiceFs := &task.LocalTask{
		Name:   "FormatJuiceFs",
		Action: new(storage.ConfigJuiceFsMetaDB),
		Retry:  1,
	}

	// 2. mount JuiceFS on a temp mount point and do a single full sync. Olares
	// is already stopped (enforced by the precheck), so nothing writes to the
	// rootfs and one pass with --delete migrates everything.
	mountForMigration := &task.LocalTask{
		Name:   "MountJuiceFSForMigration",
		Action: new(storage.MountJuiceFSForMigration),
		Retry:  1,
	}
	syncRootFS := &task.LocalTask{
		Name:   "SyncRootFSData",
		Action: &storage.SyncRootFSData{Delete: true},
		Retry:  1,
	}

	// 3. unmount temp, swap rootfs, mount JuiceFS on rootfs, verify
	unmountMigration := &task.LocalTask{
		Name:   "UnmountJuiceFSMigration",
		Action: new(storage.UnmountJuiceFSMigration),
		Retry:  3,
	}
	swapRootFS := &task.LocalTask{
		Name:   "BackupAndSwapRootFS",
		Action: new(storage.BackupAndSwapRootFS),
		Retry:  1,
	}
	enableJuiceFs := &task.LocalTask{
		Name:   "EnableJuiceFsService",
		Action: new(storage.EnableJuiceFsService),
		Retry:  1,
	}
	checkJuiceFs := &task.LocalTask{
		Name:   "CheckJuiceFsState",
		Action: new(storage.CheckJuiceFsState),
		Retry:  5,
		Delay:  5 * time.Second,
	}
	m.Tasks = []task.Interface{
		getRedisConfig,
		cleanupStaleBinary,
		installJuiceFs,
		formatJuiceFs,
		mountForMigration,
		syncRootFS,
		unmountMigration,
		swapRootFS,
		enableJuiceFs,
		checkJuiceFs,
	}
}

// RegenerateRuntimeServiceModule regenerates the container-runtime service unit
// so it gains the JuiceFS pre-check (After=juicefs.service + ExecStartPre=juicefs
// summary) now that juicefs.service exists.
//
// It is a standalone, always-run module (not part of the migration module) and
// every task in it is idempotent, so a run that was interrupted between enabling
// the juicefs service and this regeneration is still repaired on resume, even
// though the migration module itself is skipped once JuiceFS is enabled.
type RegenerateRuntimeServiceModule struct {
	common.KubeModule
}

func (m *RegenerateRuntimeServiceModule) Init() {
	m.Name = "RegenerateRuntimeService"

	// Only the container-runtime *service* unit needs regenerating: now that
	// juicefs.service exists, the template adds the JuiceFS pre-check
	// (After=juicefs.service + ExecStartPre=juicefs summary). The service *env*
	// file (e.g. k3s.service.env, which holds K3S_TOKEN) is unchanged by this
	// migration, so we must NOT regenerate it - doing so would require the
	// cluster status from the pipeline cache and risks clobbering the token.
	if m.KubeConf.Arg.Kubetype == common.K8s {
		m.Tasks = []task.Interface{
			&task.LocalTask{
				Name:   "RegenerateKubeletService",
				Action: new(kubernetes.GenerateKubeletService),
			},
			&task.LocalTask{
				Name:   "ReloadSystemdUnits",
				Action: &terminus.SystemctlCommand{DaemonReloadPreExec: true},
			},
		}
		return
	}
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "RegenerateK3sService",
			Action: new(k3s.GenerateK3sService),
		},
		// EnableK3sService does `systemctl daemon-reload && systemctl enable --now k3s`,
		// which picks up the regenerated unit (with the JuiceFS pre-check).
		&task.LocalTask{
			Name:   "EnableK3sService",
			Action: new(k3s.EnableK3sService),
		},
	}
}

// UpdateRootFSTypeModule flips OLARES_SYSTEM_ROOTFS_TYPE to "jfs". It runs after
// Olares is back up, so it retries to give the kube-apiserver time to come back.
type UpdateRootFSTypeModule struct {
	common.KubeModule
}

func (m *UpdateRootFSTypeModule) Init() {
	m.Name = "UpdateRootFSType"
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "UpdateRootFSTypeSystemEnv",
			Action: new(storage.UpdateRootFSTypeSystemEnv),
			Retry:  30,
			Delay:  10 * time.Second,
		},
	}
}

// ReRenderFsnotifyChartsModule re-renders the charts whose templates branch on
// fs_type so the JuiceFS-specific fsnotify components are deployed. Flipping the
// SystemEnv alone is not enough: already-installed releases were rendered with
// fs_type=fs and only re-render on a helm upgrade.
type ReRenderFsnotifyChartsModule struct {
	common.KubeModule
}

func (m *ReRenderFsnotifyChartsModule) Init() {
	m.Name = "ReRenderFsnotifyCharts"
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "ReRenderFsnotifyCharts",
			Action: new(ReRenderFsnotifyCharts),
			Retry:  3,
			Delay:  15 * time.Second,
		},
	}
}

// ReRenderFsnotifyCharts re-upgrades, with fs_type=jfs and --reuse-values:
//   - the os-platform release, which carries the cluster-scoped fsnotify proxy
//     and daemon (gated on fs_type==jfs);
//   - each user's per-user fsnotify release (fsnotify for the admin/owner,
//     fsnotify-<user> for others).
//
// Only the fs_type-gated templates change; everything else renders identically
// thanks to --reuse-values, so this is effectively a no-op for the rest of the
// release. fsnotify is the only consumer of fs_type that actually changes
// behavior (the file-watch bridge needed because inotify does not work over the
// JuiceFS FUSE mount).
type ReRenderFsnotifyCharts struct {
	common.KubeAction
}

func (t *ReRenderFsnotifyCharts) Execute(runtime connector.Runtime) error {
	if !storage.IsJuiceFSEnabled() {
		logger.Info("JuiceFS is not enabled, skipping fsnotify chart re-render")
		return nil
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %w", err)
	}

	jfsVals := map[string]interface{}{"fs_type": "jfs"}

	// 1. cluster-scoped fsnotify lives in the os-platform release
	platformActionConfig, platformSettings, err := utils.InitConfig(config, common.NamespaceOsPlatform)
	if err != nil {
		return err
	}
	platformChartPath := path.Join(runtime.GetInstallerDir(), "wizard", "config", "os-platform")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	logger.Infof("re-rendering %s release with fs_type=jfs to deploy the cluster fsnotify components", common.ChartNameOSPlatform)
	if err := utils.UpgradeCharts(ctx, platformActionConfig, platformSettings, common.ChartNameOSPlatform, platformChartPath, "", common.NamespaceOsPlatform, jfsVals, true); err != nil {
		return fmt.Errorf("failed to re-render %s release: %w", common.ChartNameOSPlatform, err)
	}

	// 2. per-user fsnotify charts
	users, adminUser, err := listOlaresUsers(config)
	if err != nil {
		// the cluster-level component is the critical one; don't fail the whole
		// migration if user enumeration has a transient problem
		logger.Warnf("failed to list users for per-user fsnotify re-render: %v", err)
		return nil
	}
	fsnotifyChartPath := path.Join(runtime.GetInstallerDir(), "wizard", "config", "apps", "fsnotify")
	for _, user := range users {
		ns := fmt.Sprintf("user-space-%s", user)
		releaseName := "fsnotify"
		if user != adminUser {
			releaseName = fmt.Sprintf("fsnotify-%s", user)
		}
		userActionConfig, userSettings, err := utils.InitConfig(config, ns)
		if err != nil {
			logger.Warnf("failed to init helm config for user %s: %v", user, err)
			continue
		}
		// Explicitly pin bfl.username to this user. The per-user fsnotify chart
		// derives its target namespace (user-system-<username>) from this value;
		// relying on --reuse-values alone proved unreliable and could render a
		// user's resources into another user's namespace.
		userVals := map[string]interface{}{
			"fs_type": "jfs",
			"bfl":     map[string]interface{}{"username": user},
		}
		uctx, ucancel := context.WithTimeout(context.Background(), 3*time.Minute)
		logger.Infof("re-rendering fsnotify release %q in %s with fs_type=jfs", releaseName, ns)
		if err := utils.UpgradeCharts(uctx, userActionConfig, userSettings, releaseName, fsnotifyChartPath, "", ns, userVals, true); err != nil {
			logger.Warnf("failed to re-render fsnotify for user %s: %v", user, err)
		}
		ucancel()
	}

	// The cluster-scoped fsnotify re-render (the critical step) succeeded, so
	// the migration is now fully finalized. Record it so subsequent
	// enable-juicefs runs short-circuit instead of resuming finalization.
	if err := storage.MarkMigrationFinalized(); err != nil {
		return fmt.Errorf("failed to record the rootfs migration as finalized: %w", err)
	}

	return nil
}

// listOlaresUsers returns the names of all active Olares users (those with a
// user-space namespace) and the name of the admin/owner user.
func listOlaresUsers(config *rest.Config) (users []string, adminUser string, err error) {
	scheme := kruntime.NewScheme()
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		return nil, "", fmt.Errorf("failed to add user scheme: %w", err)
	}
	userClient, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user client: %w", err)
	}
	k8sClient, err := k8sclientset.NewForConfig(config)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	var userList iamv1alpha2.UserList
	if err := userClient.List(ctx, &userList); err != nil {
		return nil, "", fmt.Errorf("failed to list users: %w", err)
	}

	for _, user := range userList.Items {
		if user.Status.State == "Failed" || user.Status.State == "Deleting" || user.DeletionTimestamp != nil {
			continue
		}
		nsCtx, nsCancel := context.WithTimeout(context.Background(), 1*time.Minute)
		_, nsErr := k8sClient.CoreV1().Namespaces().Get(nsCtx, fmt.Sprintf("user-space-%s", user.Name), metav1.GetOptions{})
		nsCancel()
		if nsErr != nil {
			if apierrors.IsNotFound(nsErr) {
				continue
			}
			return nil, "", fmt.Errorf("failed to get user-space namespace for %s: %w", user.Name, nsErr)
		}
		users = append(users, user.Name)
		if role, ok := user.Annotations["bytetrade.io/owner-role"]; ok && role == "owner" {
			adminUser = user.Name
		}
	}
	return users, adminUser, nil
}
