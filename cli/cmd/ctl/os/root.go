package os

import (
	"github.com/spf13/cobra"
)

func NewOSCommands() []*cobra.Command {
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
