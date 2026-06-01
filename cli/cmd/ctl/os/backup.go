package os

import (
	"github.com/spf13/cobra"
	backupssdk "olares.com/backups-sdk"
)

func NewCmdBackup() *cobra.Command {
	return backupssdk.NewBackupCommands()
}
