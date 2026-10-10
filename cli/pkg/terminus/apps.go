package terminus

import (
	"context"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/logger"

	"github.com/beclab/Olares/cli/pkg/clientset"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrepareUserChartsModule makes personal charts available to app-service without
// deploying any user's services during machine installation.
type PrepareUserChartsModule struct {
	common.KubeModule
}

func (m *PrepareUserChartsModule) Init() {
	logger.InfoInstallationProgress("Preparing charts for user creation ...")
	m.Name = "PrepareUserCharts"
	m.Tasks = []task.Interface{
		&task.LocalTask{Name: "CopyAppServiceHelmFiles", Action: new(CopyAppServiceHelmFiles), Retry: 5},
	}
}

type ClearAppValues struct {
	common.KubeAction
}

func (c *ClearAppValues) Execute(runtime connector.Runtime) error {
	// clear apps values.yaml
	_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("cat /dev/null > %s/wizard/config/apps/values.yaml", runtime.GetInstallerDir()), false, false)

	return nil
}

type CopyAppServiceHelmFiles struct {
	common.KubeAction
}

func (c *CopyAppServiceHelmFiles) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	appServiceName, err := getAppServiceName(client, runtime)
	if err != nil {
		return err
	}

	kubeclt, _ := util.GetCommand(common.CommandKubectl)
	for _, app := range []string{"launcher", "apps"} {
		var cmd = fmt.Sprintf("%s cp %s/wizard/config/%s os-framework/%s:/userapps -c app-service", kubeclt, runtime.GetInstallerDir(), app, appServiceName)
		if _, err = runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "copy files failed")
		}
	}

	return nil
}

func getAppServiceName(client clientset.Client, runtime connector.Runtime) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pods, err := client.Kubernetes().CoreV1().Pods(common.NamespaceOsFramework).List(ctx, metav1.ListOptions{LabelSelector: "tier=app-service"})
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get app-service failed")
	}

	if len(pods.Items) == 0 {
		return "", errors.New("app-service not found")
	}

	return pods.Items[0].Name, nil
}
