package os

import (
	backupssdk "bytetrade.io/web3os/backups-sdk"
	"github.com/spf13/cobra"
)

func NewCmdBackup() *cobra.Command {
	return backupssdk.NewBackupCommands()
}
