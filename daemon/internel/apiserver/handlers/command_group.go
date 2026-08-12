package handlers

import (
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	changehost "github.com/beclab/Olares/daemon/pkg/commands/change_host"
	collectlogs "github.com/beclab/Olares/daemon/pkg/commands/collect_logs"
	connectwifi "github.com/beclab/Olares/daemon/pkg/commands/connect_wifi"
	disableappoverlaygateway "github.com/beclab/Olares/daemon/pkg/commands/disable_app_overlay_gateway"
	disableoverlaygateway "github.com/beclab/Olares/daemon/pkg/commands/disable_overlay_gateway"
	enableappoverlaygateway "github.com/beclab/Olares/daemon/pkg/commands/enable_app_overlay_gateway"
	enableoverlaygateway "github.com/beclab/Olares/daemon/pkg/commands/enable_overlay_gateway"
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

	// The node-local half of a cluster power operation. Every node serves it,
	// and it acts only on the node it reaches, so it is guarded like the two
	// commands above: the owner's signature, checked here against the
	// operation it names.
	//
	// There is deliberately no access token on this hop. Requiring one would
	// mean the master forwarding the caller's, which hands every node in the
	// cluster a credential good for every route that user can reach — a far
	// wider grant than the one operation being carried out.
	cmd.Post("/power-node", handlers.RequireSignature(
		handlers.RequireOwner(handlers.PostPowerNode)))

	// The same hop for every other cluster operation, guarded identically.
	// It is a second path rather than a wider power-node because an older
	// worker serves power-node and nothing else, so that one's request JSON
	// cannot grow; see clusterop.ClusterOperationPath.
	cmd.Post("/cluster-operation", handlers.RequireSignature(
		handlers.RequireOwner(handlers.PostClusterOperationNode)))

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

	cmd.Post("/collect-logs", handlers.RequireAuthorization(
		handlers.RequireAdmin(
			handlers.RequireMaster(
				handlers.WaitServerRunning(
					handlers.RunCommand(handlers.PostCollectLogs, collectlogs.New))))))

	cmd.Get("/collect-logs/:runID", handlers.RequireAuthorization(
		handlers.RequireMaster(handlers.GetCollectLogsStatus)))

	cmd.Post("/collect-logs-node", handlers.RequireAuthorization(
		handlers.WaitServerRunning(
			handlers.RunCommand(handlers.PostCollectLogsNode, collectlogs.NewNode))))

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

	cmd.Post("/enable-overlay-gateway", handlers.RequireAuthorization(
		handlers.RequireOwner(
			handlers.RequireMaster(
				handlers.RequireLocal(
					handlers.WaitServerRunning(handlers.RunCommand(handlers.EnableOverlayGateway, enableoverlaygateway.New)))))))

	cmd.Post("/disable-overlay-gateway", handlers.RequireAuthorization(
		handlers.RequireOwner(
			handlers.RequireMaster(
				handlers.RequireLocal(
					handlers.WaitServerRunning(handlers.RunCommand(handlers.DisableOverlayGateway, disableoverlaygateway.New)))))))

	cmd.Post("/enable-app-overlay-gateway", handlers.RequireAuthorization(
		handlers.WaitServerRunning(handlers.RunCommand(handlers.EnableAppOverlayGateway, enableappoverlaygateway.New))))

	cmd.Post("/disable-app-overlay-gateway", handlers.RequireAuthorization(
		handlers.WaitServerRunning(handlers.RunCommand(handlers.DisableAppOverlayGateway, disableappoverlaygateway.New))))

	cmdv2 := cmd.Group("v2")
	cmdv2.Post("/mount-samba", handlers.RequireMaster(
		handlers.RequireLocal(
			handlers.WaitServerRunning(
				handlers.RunCommand(handlers.PostMountSambaDriverV2, mountsmb.New)))))

	klog.V(8).Info("command handlers initialized")
}
