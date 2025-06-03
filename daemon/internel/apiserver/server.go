package apiserver

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/terminusd/internel/ble"
	changehost "bytetrade.io/web3os/terminusd/pkg/commands/change_host"
	collectlogs "bytetrade.io/web3os/terminusd/pkg/commands/collect_logs"
	connectwifi "bytetrade.io/web3os/terminusd/pkg/commands/connect_wifi"
	"bytetrade.io/web3os/terminusd/pkg/commands/install"
	mountsmb "bytetrade.io/web3os/terminusd/pkg/commands/mount_smb"
	"bytetrade.io/web3os/terminusd/pkg/commands/reboot"
	"bytetrade.io/web3os/terminusd/pkg/commands/shutdown"
	umountsmb "bytetrade.io/web3os/terminusd/pkg/commands/umount_smb"
	umountusb "bytetrade.io/web3os/terminusd/pkg/commands/umount_usb"
	"bytetrade.io/web3os/terminusd/pkg/commands/uninstall"
	"bytetrade.io/web3os/terminusd/pkg/commands/upgrade"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"k8s.io/klog/v2"
)

type server struct {
	handlers *handlers
	port     int
	app      *fiber.App
}

func NewServer(ctx context.Context, port int) (*server, error) {
	return &server{handlers: &handlers{mainCtx: ctx}, port: port}, nil
}

func (s *server) Start() error {
	app := fiber.New()
	s.app = app

	app.Use(cors.New())
	app.Use(logger.New())

	cmd := app.Group("command")
	cmd.Post("/install", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostTerminusInit, install.New))))

	cmd.Post("/uninstall", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostTerminusUninstall, uninstall.New))))

	cmd.Post("/upgrade", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.RequestOlaresUpgrade, upgrade.NewCreateTarget))))

	cmd.Delete("/upgrade", s.handlers.RequireSignature(
		s.handlers.RunCommand(s.handlers.CancelOlaresUpgrade, upgrade.NewRemoveTarget)))

	cmd.Post("/reboot", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostReboot, reboot.New))))

	cmd.Post("/shutdown", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostShutdown, shutdown.New))))

	cmd.Post("/connect-wifi", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostConnectWifi, connectwifi.New))))

	cmd.Post("/change-host", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostChangeHost, changehost.New))))

	cmd.Post("/umount-usb", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostUmountUsb, umountusb.New))))

	cmd.Post("/umount-usb-incluster", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostUmountUsbInCluster, umountusb.New))))

	cmd.Post("/collect-logs", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostCollectLogs, collectlogs.New))))

	cmd.Post("/mount-samba", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostMountSambaDriver, mountsmb.New))))

	cmd.Post("/umount-samba", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostUmountSmb, umountsmb.New))))

	cmd.Post("/umount-samba-incluster", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostUmountSmbInCluster, umountsmb.New))))

	cmdv2 := cmd.Group("v2")
	cmdv2.Post("/mount-samba", s.handlers.RequireSignature(
		s.handlers.WaitServerRunning(
			s.handlers.RunCommand(s.handlers.PostMountSambaDriverV2, mountsmb.New))))

	system := app.Group("system")
	system.Get("/status", s.handlers.RequireSignature(s.handlers.GetTerminusState))
	system.Get("/ifs", s.handlers.RequireSignature(s.handlers.GetNetIfs))
	system.Get("/hosts-file", s.handlers.RequireSignature(s.handlers.GetHostsfile))
	system.Post("/hosts-file", s.handlers.RequireSignature(s.handlers.PostHostsfile))
	system.Get("/mounted-usb", s.handlers.RequireSignature(s.handlers.GetMountedUsb))
	system.Get("/mounted-hdd", s.handlers.RequireSignature(s.handlers.GetMountedHdd))
	system.Get("/mounted-smb", s.handlers.RequireSignature(s.handlers.GetMountedSmb))
	system.Get("/mounted-path", s.handlers.RequireSignature(s.handlers.GetMountedPath))
	system.Get("/mounted-usb-incluster", s.handlers.RequireSignature(s.handlers.GetMountedUsbInCluster))
	system.Get("/mounted-hdd-incluster", s.handlers.RequireSignature(s.handlers.GetMountedHddInCluster))
	system.Get("/mounted-smb-incluster", s.handlers.RequireSignature(s.handlers.GetMountedSmbInCluster))
	system.Get("/mounted-path-incluster", s.handlers.RequireSignature(s.handlers.GetMountedPathInCluster))

	containerd := app.Group("containerd")
	containerd.Get("/registries", s.handlers.RequireSignature(s.handlers.ListRegistries))

	registry := containerd.Group("registry")
	mirrors := registry.Group("mirrors")

	mirrors.Get("/", s.handlers.RequireSignature(s.handlers.GetRegistryMirrors))
	mirrors.Get("/:registry", s.handlers.RequireSignature(s.handlers.GetRegistryMirror))
	mirrors.Put("/:registry", s.handlers.RequireSignature(s.handlers.UpdateRegistryMirror))
	mirrors.Delete("/:registry", s.handlers.RequireSignature(s.handlers.DeleteRegistryMirror))

	image := containerd.Group("images")

	image.Get("/", s.handlers.RequireSignature(s.handlers.ListImages))
	image.Delete("/:image", s.handlers.RequireSignature(s.handlers.DeleteImage))
	image.Post("/prune", s.handlers.RequireSignature(s.handlers.PruneImages))

	return app.Listen(fmt.Sprintf(":%d", s.port))
}

func (s *server) Shutdown() error {
	klog.Info("shutdown api server")
	if s.app == nil {
		return nil
	}
	return s.app.Shutdown()
}

func (s *server) UpdateAps(aplist []ble.AccessPoint) {
	s.handlers.apList = aplist
}
