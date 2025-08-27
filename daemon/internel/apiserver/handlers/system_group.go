package handlers

import (
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"k8s.io/klog/v2"
)

func init() {
	s := server.API
	system := s.App.Group("system")
	system.Get("/status", handlers.RequireLocal(handlers.GetTerminusState))
	system.Get("/ifs", handlers.RequireLocal(handlers.GetNetIfs))
	system.Get("/hosts-file", handlers.RequireLocal(handlers.GetHostsfile))
	system.Post("/hosts-file", handlers.RequireLocal(handlers.PostHostsfile))
	system.Get("/mounted-usb", handlers.RequireLocal(handlers.GetMountedUsb))
	system.Get("/mounted-hdd", handlers.RequireLocal(handlers.GetMountedHdd))
	system.Get("/mounted-smb", handlers.RequireLocal(handlers.GetMountedSmb))
	system.Get("/mounted-path", handlers.RequireLocal(handlers.GetMountedPath))
	system.Get("/mounted-usb-incluster", handlers.RequireLocal(handlers.GetMountedUsbInCluster))
	system.Get("/mounted-hdd-incluster", handlers.RequireLocal(handlers.GetMountedHddInCluster))
	system.Get("/mounted-smb-incluster", handlers.RequireLocal(handlers.GetMountedSmbInCluster))
	system.Get("/mounted-path-incluster", handlers.RequireLocal(handlers.GetMountedPathInCluster))
	system.Get("/1.0/name/:olaresName", handlers.RequireLocal(handlers.ResolveOlaresName))
	system.Post("/checkjws", handlers.RequireLocal(handlers.CheckJWS))

	klog.V(8).Info("system handlers initialized")
}
