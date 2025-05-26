package terminus

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

type InstallVeleroBinary struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *InstallVeleroBinary) Execute(runtime connector.Runtime) error {
	veleroPkg, err := t.Manifest.Get(common.CommandVelero)
	if err != nil {
		return err
	}

	systemInfo := runtime.GetSystemInfo()
	var baseDir = t.BaseDir
	if systemInfo.IsWsl() {
		var wslPackageDir = t.KubeConf.Arg.GetWslUserPath()
		if wslPackageDir != "" {
			baseDir = path.Join(wslPackageDir, cc.DefaultBaseDir)
		}
	}

	path := veleroPkg.FilePath(baseDir)

	var cmd = fmt.Sprintf("rm -rf /tmp/velero* && mkdir /tmp/velero && cp %s /tmp/%s && tar xf /tmp/%s -C /tmp/velero ", path, veleroPkg.Filename, veleroPkg.Filename)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return err
	}
	var binPath string
	filepath.WalkDir("/tmp/velero", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == common.CommandVelero {
			binPath = path
		}
		return nil
	})
	cmd = fmt.Sprintf("install %s /usr/local/bin", binPath)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return err
	}
	return nil
}

type InstallVeleroCRDs struct {
	common.KubeAction
}

func (i *InstallVeleroCRDs) Execute(runtime connector.Runtime) error {
	velero, err := util.GetCommand(common.CommandVelero)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "velero not found")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s install --crds-only --retry 10 --delay 5", velero), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "install velero crds failed")
	}

	return nil
}

type CreateBackupLocation struct {
	common.KubeAction
}

// backup-location
func (c *CreateBackupLocation) Execute(runtime connector.Runtime) error {
	velero, err := util.GetCommand(common.CommandVelero)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "velero not found")
	}

	var ns = "os-system"
	var provider = "terminus"
	var storage = "terminus-cloud"

	var cmd = fmt.Sprintf("%s backup-location get -n %s -l 'name=%s'", velero, ns, storage)
	if res, err := runtime.GetRunner().SudoCmd(cmd, false, true); err == nil && res != "" {
		return nil
	}

	cmd = fmt.Sprintf("%s backup-location create %s --provider %s --namespace %s --prefix '' --bucket %s --labels name=%s",
		velero, storage, provider, ns, storage, storage)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "create backup-location failed")
	}

	// TODO test

	return nil
}

type InstallVeleroPlugin struct {
	common.KubeAction
}

func (i *InstallVeleroPlugin) Execute(runtime connector.Runtime) error {
	velero, err := util.GetCommand(common.CommandVelero)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "velero not found")
	}

	var ns = "os-system"
	var cmd = fmt.Sprintf("%s plugin get -n %s |grep 'velero.io/terminus' |wc -l", velero, ns)
	pluginCounts, _ := runtime.GetRunner().SudoCmd(cmd, false, true)
	if counts := utils.ParseInt(pluginCounts); counts > 0 {
		return nil
	}

	var args string
	var veleroVersion = "v1.11.3"
	var veleroPluginVersion = "v1.0.2"
	if runtime.GetSystemInfo().IsRaspbian() {
		args = " --retry 30 --delay 5"
	}

	cmd = fmt.Sprintf("%s install --no-default-backup-location --namespace %s --image beclab/velero:%s --use-volume-snapshots=false --no-secret --plugins beclab/velero-plugin-for-terminus:%s --velero-pod-cpu-request=10m --velero-pod-cpu-limit=200m --node-agent-pod-cpu-request=10m --node-agent-pod-cpu-limit=200m --wait --wait-minute 30 %s", velero, ns, veleroVersion, veleroPluginVersion, args)

	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "velero install plugin error")
	}

	// cmd = fmt.Sprintf("%s plugin add beclab/velero-plugin-for-terminus:%s -n %s", velero, veleroPluginVersion, ns)
	// if stdout, _ := runtime.GetRunner().SudoCmd(cmd, false, true); stdout != "" && !strings.Contains(stdout, "Duplicate") {
	// 	logger.Debug(stdout)
	// }

	return nil
}

type PatchVelero struct {
	common.KubeAction
}

func (v *PatchVelero) Execute(runtime connector.Runtime) error {
	kubectl, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl not found")
	}

	var ns = "os-system"
	var patch = `[{"op":"replace","path":"/spec/template/spec/volumes","value": [{"name":"plugins","emptyDir":{}},{"name":"scratch","emptyDir":{}},{"name":"terminus-cloud","hostPath":{"path":"/olares/rootfs/k8s-backup", "type":"DirectoryOrCreate"}}]},{"op": "replace", "path": "/spec/template/spec/containers/0/volumeMounts", "value": [{"name":"plugins","mountPath":"/plugins"},{"name":"scratch","mountPath":"/scratch"},{"mountPath":"/data","name":"terminus-cloud"}]},{"op": "replace", "path": "/spec/template/spec/containers/0/securityContext", "value": {"privileged": true, "runAsNonRoot": false, "runAsUser": 0}}]`

	if stdout, _ := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s patch deploy velero -n %s --type='json' -p='%s'", kubectl, ns, patch), false, true); stdout != "" && !strings.Contains(stdout, "patched") {
		logger.Errorf("velero plugin patched error %s", stdout)
	}

	return nil
}

type InstallVeleroModule struct {
	common.KubeModule
	manifest.ManifestModule
}

func (m *InstallVeleroModule) Init() {
	logger.InfoInstallationProgress("Installing backup component ...")
	m.Name = "InstallVelero"

	installVeleroBinary := &task.LocalTask{
		Name: "InstallVeleroBinary",
		Action: &InstallVeleroBinary{
			ManifestAction: manifest.ManifestAction{
				Manifest: m.Manifest,
				BaseDir:  m.BaseDir,
			},
		},
		Retry: 1,
	}

	installVeleroCRDs := &task.LocalTask{
		Name:   "InstallVeleroCRDs",
		Action: new(InstallVeleroCRDs),
		Retry:  1,
	}

	createBackupLocation := &task.LocalTask{
		Name:   "CreateBackupLocation",
		Action: new(CreateBackupLocation),
		Retry:  1,
	}

	installVeleroPlugin := &task.LocalTask{
		Name:   "InstallVeleroPlugin",
		Action: new(InstallVeleroPlugin),
		Retry:  1,
	}

	patchVelero := &task.LocalTask{
		Name:   "PatchVelero",
		Action: new(PatchVelero),
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		installVeleroBinary,
		installVeleroCRDs,
		createBackupLocation,
		installVeleroPlugin,
		patchVelero,
	}
}
