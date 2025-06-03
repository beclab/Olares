package uninstall

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"bytetrade.io/web3os/terminusd/pkg/cli"
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"bytetrade.io/web3os/terminusd/pkg/commands"
	"github.com/nxadm/tail"
	"k8s.io/klog/v2"
)

type uninstall struct {
	commands.Operation
}

var _ commands.Interface = &uninstall{}

func New() commands.Interface {
	return &uninstall{
		Operation: commands.Operation{
			Name: commands.Uninstall,
		},
	}
}

func (i *uninstall) Execute(ctx context.Context, p any) (res any, err error) {
	if _, err = os.Stat(commands.INSTALLING_PID_FILE); err == nil {
		if err = os.Remove(commands.INSTALLING_PID_FILE); err != nil {
			klog.Warning("remove pid file error, ", err, ", ", commands.INSTALLING_PID_FILE)
		}
	}

	cmd := commands.NewBaseCommand()
	cmd.WithDir_(commands.COMMAND_BASE_DIR).
		WithPid_(commands.UNINSTALLING_PID_FILE).
		AddEnv_("PREPARED", "1").
		WithWatchDog_(i.watch)

	if err = cmd.RunAsync_(ctx, "sudo", cli.TERMINUS_CLI, "uninstall", "--phase", "install", "--base-dir", commands.TERMINUS_BASE_DIR); err != nil {
		return nil, err
	}

	state.CurrentState.ChangeTerminusStateTo(state.Uninstalling)
	state.CurrentState.UninstallingState = state.InProgress
	state.CurrentState.UninstallingProgress = "1%"
	state.CurrentState.UninstallingProgressNum = 1

	return nil, nil

}

func (i *uninstall) watch(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		updateNotInstalled := func() {
			installed, err := state.IsTerminusInstalled()
			if err != nil {
				klog.Error("check install state error, ", err)
				return
			}

			if !installed {
				state.CurrentState.ChangeTerminusStateTo(state.NotInstalled)
				state.CurrentState.UninstallingState = state.Completed
				state.CurrentState.UninstallingProgress = "100%"
				state.CurrentState.UninstallingProgressNum = 100
			} else {
				state.CurrentState.UninstallingState = state.Failed
			}
		}

		for {
			select {
			case <-ctx.Done():
				updateNotInstalled()
				return
			case <-ticker.C:
				// check install.log
				if finished := i.tailLog(); finished {
					klog.Info("install finished")

					updateNotInstalled()
					return
				}

			}
		}
	}()
}

func (i *uninstall) tailLog() (finished bool) {
	tailFile := commands.COMMAND_BASE_DIR + "/logs/uninstall.log"

	info, err := os.Stat(tailFile)
	if err != nil {
		klog.Error(err)
		return false
	}

	filesize := info.Size()
	tailsize := min(filesize, 512)

	t, err := tail.TailFile(tailFile,
		tail.Config{Follow: false, Location: &tail.SeekInfo{Offset: -tailsize, Whence: io.SeekEnd}})
	if err != nil {
		klog.Error("tail log error, ", err)
		return false
	}

	for line := range t.Lines {
		for _, p := range ProgressWords {
			if strings.Contains(line.Text, p.KeyWords) {
				if state.CurrentState.UninstallingProgressNum < p.ProgressNum {
					state.CurrentState.UninstallingProgress = p.Progress
					state.CurrentState.UninstallingProgressNum = p.ProgressNum
					if p.Progress != "100%" {
						state.CurrentState.UninstallingState = state.InProgress
					} else {
						state.CurrentState.UninstallingState = state.Completed
						state.CurrentState.ChangeTerminusStateTo(state.NotInstalled)
						return true
					}
				}
			}
		}
	}

	return false
}

type Progress struct {
	KeyWords    string
	Progress    string
	ProgressNum int
}

var (
	ProgressWords = []Progress{
		{"Uninstalling OS ...", "1%", 1},
		{"remove kubernetes cluster", "2%", 2},
		{"[Module] GetStorage", "5%", 5},
		{"[Module] DeleteClusterModule", "15%", 15},
		{"[Module] ClearOSModule", "35%", 35},
		{"[Module] UninstallAutoRenewCertsModule", "50%", 50},
		{"[Module] KillContainerdProcess", "60%", 60},
		{"[Module] DeleteUserData", "80%", 80},
		{"[Module] DeletePhaseFlag", "95%", 95},
		{"Uninstall Olares execute successfully!!!", "100%", 100},
	}
)
