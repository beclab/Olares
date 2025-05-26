package os

import (
	"os/exec"

	"github.com/spf13/cobra"
)

func NewOSCommands() []*cobra.Command {
	_ = exec.Command("/bin/bash", "-c", "ulimit -u 65535").Run()
	_ = exec.Command("/bin/bash", "-c", "ulimit -n 65535").Run()

	return []*cobra.Command{
		NewCmdPrecheck(),
		NewCmdRootDownload(),
		NewCmdPrepare(),
		NewCmdInstallOs(),
		NewCmdUninstallOs(),
		NewCmdChangeIP(),
		NewCmdRelease(),
		NewCmdPrintInfo(),
		NewCmdBackup(),
		NewCmdLogs(),
		NewCmdStart(),
		NewCmdStop(),
		NewCmdUpgradeOs(),
	}
}
