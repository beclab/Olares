package plugins

import (
	"fmt"
	"path"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

type CopyEmbedFiles struct {
	common.KubeAction
}

func (t *CopyEmbedFiles) Execute(runtime connector.Runtime) error {
	var dst = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir)
	return utils.CopyEmbed(assets, ".", dst)
}

type CheckNodeState struct {
	common.KubeAction
}

func (t *CheckNodeState) Execute(runtime connector.Runtime) error {
	var kubectlpath, _ = t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if kubectlpath == "" {
		kubectlpath = path.Join(common.BinDir, common.CommandKubectl)
	}
	var cmd = fmt.Sprintf("%s get node --no-headers", kubectlpath)
	stdout, err := runtime.GetRunner().SudoCmd(cmd, false, false)

	if err != nil || stdout == "" {
		return fmt.Errorf("Node Pending")
	}

	var nodeInfo = strings.Fields(stdout)
	if len(nodeInfo) != 5 {
		logger.Errorf("node info invalid: %s", stdout)
		return fmt.Errorf("Node Pending")
	}

	var state = nodeInfo[1]
	var version = nodeInfo[4]

	if state != "Ready" {
		return fmt.Errorf("Node Pending")
	}
	t.PipelineCache.Set(common.CacheKubeletVersion, version)

	return nil
}

type InitNamespace struct {
	common.KubeAction
}

func (t *InitNamespace) Execute(runtime connector.Runtime) error {
	var kubectlpath, _ = t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if kubectlpath == "" {
		kubectlpath = path.Join(common.BinDir, common.CommandKubectl)
	}

	for _, ns := range []string{common.NamespaceKubesphereControlsSystem, common.NamespaceKubesphereMonitoringFederated} {
		if stdout, err := runtime.GetRunner().Cmd(fmt.Sprintf("%s create ns %s", kubectlpath, ns), false, true); err != nil {
			if !strings.Contains(stdout, "already exists") {
				logger.Errorf("create ns %s failed: %v", ns, err)
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("create namespace %s failed: %v", ns, err))
			}
		}
	}
	// _, err := runtime.GetRunner().SudoCmd(
	// 	fmt.Sprintf(`cat <<EOF | /usr/local/bin/kubectl apply -f -
	// apiVersion: v1
	// kind: Namespace
	// metadata:
	//   name: %s
	// ---
	// apiVersion: v1
	// kind: Namespace
	// metadata:
	//   name: %s
	// EOF
	// `, common.NamespaceKubesphereControlsSystem, common.NamespaceKubesphereMonitoringFederated), false, true)
	// if err != nil {
	// 	return errors.Wrap(errors.WithStack(err), fmt.Sprintf("create namespace: %s and %s",
	// 		common.NamespaceKubesphereControlsSystem, common.NamespaceKubesphereMonitoringFederated))
	// }

	var allNs = []string{
		common.NamespaceDefault,
		common.NamespaceKubeNodeLease,
		common.NamespaceKubePublic,
		common.NamespaceKubeSystem,
		common.NamespaceKubekeySystem,
		common.NamespaceKubesphereControlsSystem,
		common.NamespaceKubesphereMonitoringFederated,
		common.NamespaceKubesphereMonitoringSystem,
		common.NamespaceKubesphereSystem,
	}

	for _, ns := range allNs {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s label ns %s kubesphere.io/workspace=system-workspace --overwrite", kubectlpath, ns), false, true); err != nil {
			logger.Errorf("label ns %s kubesphere.io/workspace=system-workspace failed: %v", ns, err)
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("label namespace %s kubesphere.io/workspace=system-workspace failed: %v", ns, err))
		}

		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s label ns %s kubesphere.io/namespace=%s --overwrite", kubectlpath, ns, ns), false, true); err != nil {
			logger.Errorf("label ns %s kubesphere.io/namespace=%s failed: %v", ns, ns, err)
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("label namespace %s kubesphere.io/namespace=%s failed: %v", ns, ns, err))
		}
	}

	return nil
}
