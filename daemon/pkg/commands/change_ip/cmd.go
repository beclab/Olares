package changeip

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"bytetrade.io/web3os/terminusd/pkg/cli"
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/utils"
	"github.com/nxadm/tail"
	"k8s.io/klog/v2"
)

type changeIp struct {
	commands.Operation
	isRetryChange bool
	notInstalled  bool
	logFile       string
}

var _ commands.Interface = &changeIp{}

func New() commands.Interface {
	return &changeIp{
		Operation: commands.Operation{
			Name: commands.ChangeIp,
		},
		logFile: commands.COMMAND_BASE_DIR + "/logs/changeip.log",
	}
}

func (i *changeIp) Execute(ctx context.Context, p any) (res any, err error) {
	i.isRetryChange = state.CurrentState.TerminusState == state.IPChangeFailed

	cmd := commands.NewBaseCommand()
	cmd.WithDir_(commands.COMMAND_BASE_DIR).
		WithPid_(commands.CHANGINGIP_PID_FILE).
		AddEnv_("QUIET", "1").
		WithWatchDog_(i.watch)

	if _, err := os.Stat(commands.INSTALL_LOCK); err != nil && os.IsNotExist(err) {
		cmd.AddEnv_("NOT_INSTALLED", "1")
		i.notInstalled = true
	}

	if _, err := os.Stat(commands.PREPARE_LOCK); err != nil && os.IsNotExist(err) {
		cmd.AddEnv_("NOT_PREPARED", "1")
	}

	if _, err := os.Stat(i.logFile); err == nil {
		if err = os.Remove(i.logFile); err != nil {
			klog.Error("remove prev log file error, ", err)
		}
	}

	var cmds []string = []string{
		"change-ip",
		"--version", commands.INSTALLED_VERSION,
		"--base-dir", commands.TERMINUS_BASE_DIR,
	}

	// FIXME: maybe we can always get master node ip from etcd config or anything else.
	masterIp, err := utils.MasterNodeIp(!i.notInstalled)
	if err != nil {
		klog.Error("get master node ip error,", err)
	}

	// backup prev ip, if change failed we can resume with this ip
	err = os.WriteFile(commands.PREV_IP_TO_CHANGE_FILE, []byte(masterIp), 0644)
	if err != nil {
		klog.Error("cannot backup prev ip, ", err)
	}

	state.CurrentState.ChangeTerminusStateTo(state.IPChanging)

	// remove PREV_IP_CHANGE_FAILED tag file if exists
	if _, err = os.Stat(commands.PREV_IP_CHANGE_FAILED); err == nil {
		if err = os.Remove(commands.PREV_IP_CHANGE_FAILED); err != nil {
			klog.Warning("remove ip change failed tag file error, ", err)
		}
	}

	if err = cmd.RunAsync_(ctx, cli.TERMINUS_CLI, cmds...); err != nil {
		i.setFailedState()
		return nil, err
	}

	return nil, nil

}

func (i *changeIp) watch(ctx context.Context) {
	go func() {
		// delay starting to watch
		time.Sleep(10 * time.Second)

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		successState := state.TerminusRunning
		if i.notInstalled {
			successState = state.NotInstalled
		}

		for {
			select {
			case <-ctx.Done():
				klog.Info("change ip command finished")
				_, err := os.Stat(commands.CHANGINGIP_PID_FILE)
				if os.IsNotExist(err) {
					// double check
					if i.tailLog() {
						klog.Info("change ip command finished, change state")
						state.CurrentState.ChangeTerminusStateTo(successState)
						return
					} else {
						klog.Warning("check log file, process not succeed")
						i.setFailedState()
						return
					}
				}
				klog.Warning("change ip killed")
				state.CurrentState.ChangeTerminusStateTo(state.SystemError)
				return
			case <-ticker.C:
				if i.tailLog() {
					klog.Info("watch log succeed")
					state.CurrentState.ChangeTerminusStateTo(successState)
					return
				}
			}
		}
	}()

}

func (i *changeIp) tailLog() (finished bool) {
	tailFile := i.logFile

	info, err := os.Stat(tailFile)
	if err != nil {
		klog.Error(err)
		return false
	}

	filesize := info.Size()
	tailsize := min(filesize, 512)

	t, err := tail.TailFile(tailFile,
		tail.Config{Follow: false, Location: &tail.SeekInfo{Offset: -tailsize, Whence: io.SeekEnd}})
	defer t.Stop()
	if err != nil {
		klog.Error("tail log error, ", err)
		return false
	}

	keyWords := "Olares OS components execute successfully!"
	var line *tail.Line
	for line = range t.Lines {
		if strings.Contains(line.Text, keyWords) {
			return true
		}
	}

	if line != nil {
		// for debug
		klog.Info(line.Text)
	}
	return false
}

func (i *changeIp) setFailedState() {
	if i.isRetryChange {
		klog.Error("retry ip change failed")
		state.CurrentState.ChangeTerminusStateTo(state.SystemError)
	} else {
		// create a ip change failed tag file, make the ip-change command can
		// be resumed if the device get reboot
		err := os.WriteFile(commands.PREV_IP_CHANGE_FAILED, []byte(time.Now().String()), 0644)
		if err != nil {
			klog.Error("write ip change failed tag file error, ", err)
		}

		state.CurrentState.ChangeTerminusStateTo(state.IPChangeFailed)
	}
}
