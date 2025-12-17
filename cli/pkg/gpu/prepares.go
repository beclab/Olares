package gpu

import (
	"context"
	"os"
	"strings"

	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/clientset"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GPUEnablePrepare struct {
	common.KubePrepare
}

func (p *GPUEnablePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	systemInfo := runtime.GetSystemInfo()
	if systemInfo.IsWsl() {
		return false, nil
	}

	if systemInfo.IsUbuntu() && systemInfo.IsUbuntuVersionEqual(connector.Ubuntu24) {
		return false, nil
	}
	return p.KubeConf.Arg.GPU.Enable, nil
}

type CudaInstalled struct {
	common.KubePrepare
}

func (p *CudaInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	st, err := utils.GetNvidiaStatus(runtime)
	if err != nil {
		return false, err
	}
	if st == nil || !st.Installed {
		return false, nil
	}

	return true, nil
}

type CudaNotInstalled struct {
	common.KubePrepare
}

func (p *CudaNotInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	st, err := utils.GetNvidiaStatus(runtime)
	if err != nil {
		return false, err
	}
	if st == nil || !st.Installed {
		return true, nil
	}
	return false, nil
}

type CurrentNodeInK8s struct {
	common.KubePrepare
}

func (p *CurrentNodeInK8s) PreCheck(runtime connector.Runtime) (bool, error) {
	client, err := clientset.NewKubeClient()
	if err != nil {
		logger.Debug(errors.Wrap(errors.WithStack(err), "kubeclient create error"))
		return false, nil
	}

	node, err := client.Kubernetes().CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		logger.Debug(errors.Wrap(errors.WithStack(err), "list nodes error"))
		return false, nil
	}

	for _, node := range node.Items {
		if node.Name == runtime.GetSystemInfo().GetHostname() {
			return true, nil
		}
	}

	return false, nil
}

type NvidiaGraphicsCard struct {
	common.KubePrepare
	ExitOnNotFound bool
}

func (p *NvidiaGraphicsCard) PreCheck(runtime connector.Runtime) (found bool, err error) {
	defer func() {
		if !p.ExitOnNotFound {
			return
		}
		if !found {
			logger.Error("ERROR: no graphics card found")
			os.Exit(1)
		}
	}()
	model, _, err := utils.DetectNvidiaModelAndArch(runtime)
	if err != nil {
		logger.Debugf("detect NVIDIA GPU error: %v", err)
	}
	if strings.TrimSpace(model) == "" {
		return false, nil
	}
	logger.Infof("found NVIDIA GPU: %s", model)
	return true, nil
}

type ContainerdInstalled struct {
	common.KubePrepare
}

func (p *ContainerdInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	containerdCheck := precheck.ConflictingContainerdCheck{}
	if err := containerdCheck.Check(runtime); err != nil {
		return true, nil
	}

	logger.Info("containerd is not installed, ignore task")
	return false, nil
}

type GpuDevicePluginInstalled struct {
	common.KubePrepare
}

func (p *GpuDevicePluginInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	client, err := clientset.NewKubeClient()
	if err != nil {
		logger.Debug(errors.Wrap(errors.WithStack(err), "kubeclient create error"))
		return false, nil
	}

	plugins, err := client.Kubernetes().CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=hami-device-plugin"})
	if err != nil {
		logger.Debug(err)
		return false, nil
	}

	return len(plugins.Items) > 0, nil
}
