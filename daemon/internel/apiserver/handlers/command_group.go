package handlers

import (
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	changehost "github.com/beclab/Olares/daemon/pkg/commands/change_host"
	collectlogs "github.com/beclab/Olares/daemon/pkg/commands/collect_logs"
	connectwifi "github.com/beclab/Olares/daemon/pkg/commands/connect_wifi"
	"github.com/beclab/Olares/daemon/pkg/commands/install"
	mountnfs "github.com/beclab/Olares/daemon/pkg/commands/mount_nfs"
	mountsmb "github.com/beclab/Olares/daemon/pkg/commands/mount_smb"
	"github.com/beclab/Olares/daemon/pkg/commands/reboot"
	"github.com/beclab/Olares/daemon/pkg/commands/shutdown"
	sshpassword "github.com/beclab/Olares/daemon/pkg/commands/ssh_password"
	umountnfs "github.com/beclab/Olares/daemon/pkg/commands/umount_nfs"
	umountsmb "github.com/beclab/Olares/daemon/pkg/commands/umount_smb"
	umountusb "github.com/beclab/Olares/daemon/pkg/commands/umount_usb"
	"github.com/beclab/Olares/daemon/pkg/commands/uninstall"
	"github.com/beclab/Olares/daemon/pkg/commands/upgrade"
	"k8s.io/klog/v2"
)

func init() {
	s := server.API
	cmd := s.App.Group("command")
	cmd.Post("/install",
		handlers.WaitServerRunning(
			handlers.RunCommand(handlers.PostTerminusInit, install.New)))

	cmd.Post("/uninstall", handlers.RequireSignature(
		handlers.RequireOwner(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostTerminusUninstall, uninstall.New)))))

	cmd.Post("/upgrade", handlers.RequireSignature(
		handlers.RequireOwner(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.RequestOlaresUpgrade, upgrade.NewCreateUpgradeTarget)))))

	cmd.Delete("/upgrade", handlers.RequireSignature(
		handlers.RequireOwner(
			handlers.RunCommand(handlers.CancelOlaresUpgrade, upgrade.NewRemoveUpgradeTarget))))

	cmd.Post("/upgrade/confirm", handlers.RequireSignature(
		handlers.RequireOwner(handlers.ConfirmOlaresUpgrade)))

	cmd.Post("/reboot", handlers.RequireSignature(
		handlers.RequireOwner(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostReboot, reboot.New)))))

	cmd.Post("/shutdown", handlers.RequireSignature(
		handlers.RequireOwner(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostShutdown, shutdown.New)))))

	cmd.Post("/connect-wifi", handlers.RequireSignature(
		handlers.WaitServerRunning(
			handlers.RunCommand(handlers.PostConnectWifi, connectwifi.New))))

	cmd.Post("/change-host", handlers.RequireSignature(
		handlers.WaitServerRunning(
			handlers.RunCommand(handlers.PostChangeHost, changehost.New))))

	cmd.Post("/umount-usb", handlers.RequireMaster(
		handlers.RequireLocal(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostUmountUsb, umountusb.New)))))

	cmd.Post("/umount-usb-incluster", handlers.RequireMaster(
		handlers.RequireLocal(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostUmountUsbInCluster, umountusb.New)))))

	cmd.Post("/collect-logs", handlers.RequireLocal(
		handlers.WaitServerRunning(
			handlers.RunCommand(handlers.PostCollectLogs, collectlogs.New))))

	cmd.Post("/mount-samba", handlers.RequireMaster(
		handlers.RequireLocal(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostMountSambaDriver, mountsmb.New)))))

	cmd.Post("/umount-samba", handlers.RequireMaster(
		handlers.RequireLocal(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostUmountSmb, umountsmb.New)))))

	cmd.Post("/umount-samba-incluster", handlers.RequireMaster(
		handlers.RequireLocal(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostUmountSmbInCluster, umountsmb.New)))))

	cmd.Post("/ssh-password", handlers.RequireSignature(
		handlers.RequireOwner(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostSSHPassword, sshpassword.New)))))

	cmd.Post("/mount-nfs", handlers.RequireMaster(
		handlers.RequireLocal(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostMountNfsDriver, mountnfs.New)))))

	cmd.Post("/umount-nfs", handlers.RequireMaster(
		handlers.RequireLocal(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostUmountNfs, umountnfs.New)))))

	cmd.Post("/umount-nfs-incluster", handlers.RequireMaster(
		handlers.RequireLocal(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostUmountNfsInCluster, umountnfs.New)))))

	cmd.Post("/list-nfs", handlers.RequireLocal(
		handlers.WaitServerRunning(handlers.PostListNfs)))

	cmdv2 := cmd.Group("v2")
	cmdv2.Post("/mount-samba", handlers.RequireMaster(
		handlers.RequireLocal(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostMountSambaDriverV2, mountsmb.New)))))

	klog.V(8).Info("command handlers initialized")
}
