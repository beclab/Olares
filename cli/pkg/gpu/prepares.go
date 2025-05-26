package gpu

import (
	"context"
	"os"

	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/clientset"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
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

type GPUSharePrepare struct {
	common.KubePrepare
}

func (p *GPUSharePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	return p.KubeConf.Arg.GPU.Share || runtime.GetSystemInfo().IsWsl(), nil
}

type CudaInstalled struct {
	common.KubePrepare
	precheck.CudaCheckTask
	FailOnNoInstallation bool
}

func (p *CudaInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	err := p.CudaCheckTask.Execute(runtime)
	if err != nil {
		if err == precheck.ErrCudaInstalled {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

type CudaNotInstalled struct {
	common.KubePrepare
	precheck.CudaCheckTask
}

func (p *CudaNotInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	err := p.CudaCheckTask.Execute(runtime)
	if err != nil {
		if err == precheck.ErrCudaInstalled {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type K8sNodeInstalled struct {
	common.KubePrepare
}

func (p *K8sNodeInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
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

	if len(node.Items) == 0 {
		return false, nil
	}

	return true, nil
}

type NvidiaGraphicsCard struct {
	common.KubePrepare
	ExitOnNotFound bool
}

func (p *NvidiaGraphicsCard) PreCheck(runtime connector.Runtime) (found bool, err error) {
	if runtime.RemoteHost().GetOs() == common.Darwin {
		return false, nil
	}
	defer func() {
		if !p.ExitOnNotFound {
			return
		}
		if !found {
			logger.Error("ERROR: no graphics card found")
			os.Exit(1)
		}
	}()
	output, err := runtime.GetRunner().SudoCmd(
		"lspci | grep -i vga | grep -i nvidia", false, false)
	// an empty grep also results in the exit code to be 1
	// and thus a non-nil err
	if err != nil {
		logger.Debug("try to find nvidia graphics card error ", err)
		logger.Debug("ignore card driver installation")
		return false, nil
	}

	if output != "" {
		logger.Info("found nvidia graphics card: ", output)
	}
	return output != "", nil
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
