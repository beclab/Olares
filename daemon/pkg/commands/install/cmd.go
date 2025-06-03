package install

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"bytetrade.io/web3os/terminusd/pkg/cli"
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/tools"
	"github.com/nxadm/tail"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

type install struct {
	commands.Operation
	username string
	password string
}

var _ commands.Interface = &install{}

func New() commands.Interface {
	return &install{
		Operation: commands.Operation{
			Name: commands.Install,
		},
	}
}

func (i *install) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		return nil, errors.New("invalid param")
	}

	klog.Info("install olares with user: ", param.Username)
	i.username = param.Username
	i.password = param.Password

	if i.password == "" {
		// create a random password
		i.password = tools.RandomString(6)
	}

	// start installing olares async
	cmd := commands.NewBaseCommand()
	cmd.WithDir_(commands.COMMAND_BASE_DIR).
		WithPid_(commands.INSTALLING_PID_FILE).
		AddEnv_("TERMINUS_OS_USERNAME", param.Username).
		AddEnv_("TERMINUS_OS_PASSWORD", param.Password).
		AddEnv_("TERMINUS_OS_EMAIL", param.Email).
		AddEnv_("TERMINUS_OS_DOMAINNAME", param.Domain).
		AddEnv_("VERSION", commands.INSTALLED_VERSION).
		AddEnv_("TERMINUS_BOX", "1").
		WithWatchDog_(i.watch)

	params := []string{
		"install",
		"--version", commands.INSTALLED_VERSION,
		"--base-dir", commands.TERMINUS_BASE_DIR,
		"--kube", commands.KUBE_TYPE,
	}
	if err = cmd.RunAsync_(ctx, cli.TERMINUS_CLI, params...); err != nil {
		return nil, err
	}

	state.CurrentState.InstallingState = state.InProgress
	state.CurrentState.InstallingProgress = "1%"
	state.CurrentState.InstallingProgressNum = 1
	return nil, nil
}

func (i *install) watch(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				if state.CurrentState.InstallingProgress != "100%" {
					// double check
					if finished := i.tailLog(); finished {
						klog.Info("install finished")
						state.CurrentState.InstallFinishedTime = ptr.To[time.Time](time.Now())
						return
					} else {
						state.CurrentState.InstallingState = state.Failed
						state.CurrentState.ChangeTerminusStateTo(state.InstallFailed)
					}
				}
				return
			case <-ticker.C:
				// check install.log
				if finished := i.tailLog(); finished {
					klog.Info("install finished")
					state.CurrentState.InstallFinishedTime = ptr.To[time.Time](time.Now())
					return
				}

			}
		}
	}()
}

func (i *install) tailLog() (finished bool) {
	info, err := os.Stat(commands.LOG_FILE)
	if err != nil {
		klog.Error(err)
		return false
	}

	filesize := info.Size()
	tailsize := min(filesize, 4096)

	t, err := tail.TailFile(commands.LOG_FILE,
		tail.Config{Follow: false, Location: &tail.SeekInfo{Offset: -tailsize, Whence: io.SeekEnd}})
	if err != nil {
		klog.Error("tail log error, ", err)
		return false
	}

	updated := false
	for line := range t.Lines {
		for _, p := range ProgressWords {
			if strings.Contains(line.Text, p.KeyWords) {
				if state.CurrentState.InstallingProgressNum < p.ProgressNum {
					state.CurrentState.InstallingProgress = p.Progress
					state.CurrentState.InstallingProgressNum = p.ProgressNum
					updated = true
					if p.Progress != "100%" {
						state.CurrentState.InstallingState = state.InProgress
					} else {
						state.CurrentState.InstallingState = state.Completed
						return true
					}
				}
			}
		}
	}

	// smooth progress
	if !updated {
		next := state.CurrentState.InstallingProgressNum
		for _, p := range ProgressWords {
			if p.ProgressNum > state.CurrentState.InstallingProgressNum {
				next = p.ProgressNum
				break
			}
		}

		if next > state.CurrentState.InstallingProgressNum+1 {
			state.CurrentState.InstallingProgressNum += 1
			state.CurrentState.InstallingProgress = fmt.Sprintf("%d%%", state.CurrentState.InstallingProgressNum)
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
		// {"Start to Install Olares ...", "1%", 1},
		// {"Precheck and Installing dependencies ...", "2%", 2},
		// {"Installing Olares ...", "2%", 2},
		// {"Setup your first user ...", "2%", 2},
		// {"parse user info from env or stdin", "2%", 2},
		// {"generate app values", "2%", 2},
		// {"installing k8s and kubesphere", "3%", 3},
		// {"Generating \"ca\" certificate and key", "3%", 3},
		// {"PatchKsCoreStatus success", "6%", 6},
		{"InstallGPUShared", "9%", 9},
		{"k8s and kubesphere installation is complete", "10%", 10},
		{"Installing account ...", "15%", 15},
		{"Installing settings ...", "20%", 20},
		{"Installing appservice ...", "25%", 25},
		{"waiting for appservice", "30%", 30},
		{"Installing launcher ...", "35%", 35},
		{"LocalHost: CheckLauncherStatus success", "40%", 40},
		{"Installing built-in apps ...", "45%", 45},
		{"Performing the final configuration ...", "65%", 65},
		{"Installing backup component ...", "70%", 70},
		{"Waiting for Vault ...", "80%", 80},
		{"Starting Olares ...", "90%", 90},
		{"Installation wizard is complete", "95%", 95},
		{"All done", "100%", 100},
	}
)
