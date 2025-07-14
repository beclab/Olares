package handlers

import (
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"k8s.io/klog/v2"
)

func init() {
	s := server.API
	system := s.App.Group("system")
	system.Get("/status", handlers.RequireSignature(handlers.GetTerminusState))
	system.Get("/ifs", handlers.RequireSignature(handlers.GetNetIfs))
	system.Get("/hosts-file", handlers.RequireSignature(handlers.GetHostsfile))
	system.Post("/hosts-file", handlers.RequireSignature(handlers.PostHostsfile))
	system.Get("/mounted-usb", handlers.RequireSignature(handlers.GetMountedUsb))
	system.Get("/mounted-hdd", handlers.RequireSignature(handlers.GetMountedHdd))
	system.Get("/mounted-smb", handlers.RequireSignature(handlers.GetMountedSmb))
	system.Get("/mounted-path", handlers.RequireSignature(handlers.GetMountedPath))
	system.Get("/mounted-usb-incluster", handlers.RequireSignature(handlers.GetMountedUsbInCluster))
	system.Get("/mounted-hdd-incluster", handlers.RequireSignature(handlers.GetMountedHddInCluster))
	system.Get("/mounted-smb-incluster", handlers.RequireSignature(handlers.GetMountedSmbInCluster))
	system.Get("/mounted-path-incluster", handlers.RequireSignature(handlers.GetMountedPathInCluster))

	klog.Info("system handlers initialized")
}
