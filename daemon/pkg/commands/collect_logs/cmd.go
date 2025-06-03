package collectlogs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/utils"
	"k8s.io/klog/v2"
)

type collectLogs struct {
	commands.Operation
	*commands.BaseCommand
}

var _ commands.Interface = &collectLogs{}

func New() commands.Interface {
	return &collectLogs{
		Operation: commands.Operation{
			Name: commands.CollectLogs,
		},
		BaseCommand: commands.NewBaseCommand(),
	}
}

func (i *collectLogs) Execute(ctx context.Context, p any) (res any, err error) {
	state.CurrentState.CollectingLogsState = state.InProgress
	state.CurrentState.CollectingLogsError = ""

	go func() {
		var (
			errStr string
			err    error
		)
		defer func() {
			if err != nil {
				klog.Error(errStr)
				state.CurrentState.CollectingLogsState = state.Failed
				state.CurrentState.CollectingLogsError = errStr
			}
		}()

		kubeClient, err := utils.GetKubeClient()
		if err != nil {
			errStr = fmt.Sprintf("create k8s / k3s client error, %v", err)
			return
		}

		dynamicClient, err := utils.GetDynamicClient()
		if err != nil {
			errStr = fmt.Sprintf("create k8s / k3s client error, %v", err)
			return
		}

		adminUser, err := utils.GetAdminUser(ctx, dynamicClient)
		if err != nil {
			errStr = fmt.Sprintf("get admin user error, %v", err)
			return
		}

		hostPath, err := utils.GetUserspacePvcHostPath(ctx, adminUser.GetName(), kubeClient)
		if err != nil {
			errStr = fmt.Sprintf("get admin user host path error, %v", err)
			return
		}

		exportPodLogsDir := filepath.Join(hostPath, commands.EXPORT_POD_LOGS_DIR)

		err = os.MkdirAll(exportPodLogsDir, 0755)
		if err != nil {
			errStr = fmt.Sprintf("mkdir error, %v", err)
			return
		}

		err = os.Chown(exportPodLogsDir, 1000, 1000)
		if err != nil {
			errStr = fmt.Sprintf("change logs dir owner error, %v", err)
			return
		}

		var cmds []string = []string{
			"logs",
			"--output-dir", exportPodLogsDir,
			"--ignore-kube-errors", "true",
		}

		_, err = i.BaseCommand.Run_(ctx, "olares-cli", cmds...)
		if err != nil {
			state.CurrentState.CollectingLogsState = state.Failed
			state.CurrentState.CollectingLogsError = err.Error()
			errStr = fmt.Sprintf("collect logs error, %v", err)
			return
		}

		// logTgz := fmt.Sprintf("%s/olares-logs-*.tar.gz", exportPodLogsDir)
		// cmds = []string{
		// 	"1000:1000",
		// 	logTgz,
		// }

		// _, err = i.BaseCommand.Run_(ctx, "chown", cmds...)
		// if err != nil {
		// 	errStr = fmt.Sprintf("change logs file owner error, %v", err)
		// 	return
		// }

		state.CurrentState.CollectingLogsState = state.Completed
		state.CurrentState.CollectingLogsError = ""
		klog.Info("collect log completed")
	}()

	return nil, nil
}
