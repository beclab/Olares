package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beclab/Olares/daemon/cmd/terminusd/version"
	"github.com/beclab/Olares/daemon/internel/apiserver"
	"github.com/beclab/Olares/daemon/internel/ble"
	"github.com/beclab/Olares/daemon/internel/mdns"
	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/internel/watcher/cert"
	intranetwatcher "github.com/beclab/Olares/daemon/internel/watcher/intranet"
	"github.com/beclab/Olares/daemon/internel/watcher/system"
	"github.com/beclab/Olares/daemon/internel/watcher/systemenv"
	"github.com/beclab/Olares/daemon/internel/watcher/upgrade"
	"github.com/beclab/Olares/daemon/internel/watcher/usb"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

func main() {

	state.CurrentState.TerminusdState = state.Initialize

	port := 18088
	var showVersion bool
	var showVendor bool

	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine.BoolVar(&showVersion, "version", false, "show olaresd version")
	pflag.CommandLine.BoolVar(&showVendor, "vendor", false, "show the vendor type of olaresd")

	pflag.Parse()

	if showVersion {
		fmt.Println(version.Version())
		return
	}

	if showVendor {
		fmt.Println(version.VENDOR)
		return
	}

	commands.Init()

	mainCtx, cancel := context.WithCancel(context.Background())

	apis := apiserver.NewServer(mainCtx, port)

	if err := state.CheckCurrentStatus(mainCtx); err != nil {
		klog.Error(err)
	}

	go wait.UntilWithContext(mainCtx, utils.UpdateNetworkTraffic, time.Second)

	state.CurrentState.OlaresdVersion = version.RawVersion()

	bleService, err := ble.NewBleService(mainCtx)
	if err != nil {
		klog.Error(err)
	}

	bleServiceStart := func() {
		if bleService != nil {
			bleService.SetUpdateApListCB(apis.UpdateAps)
			bleService.Start()
		}
	}

	bleServiceStart()

	defer func() {
		if bleService != nil {
			bleService.Stop()
		}
	}()

	s, err := mdns.NewServer(port)
	if err != nil {
		klog.Error(err)
	}

	defer s.Close()

	sunshine := mdns.NewSunShineProxyWithoutStart()
	defer sunshine.Close()

	state.WatchStatus(mainCtx, []watcher.Watcher{
		system.NewSystemWatcher(),
		// usb.NewUsbWatcher(),
		usb.NewUmountWatcher(),
		upgrade.NewUpgradeWatcher(),
		cert.NewCertWatcher(),
		systemenv.NewSystemEnvWatcher(),
		intranetwatcher.NewApplicationWatcher(),
	}, func() {
		if s != nil {
			if err := s.Restart(); err != nil {
				klog.Error(err)
			}
		}

		// try to restart ble service, if ble not enabled when olaresd was started
		if bleService == nil {
			var err error
			bleService, err = ble.NewBleService(mainCtx)
			if err != nil {
				klog.Error(err)
			}

			bleServiceStart()
		}

		// start or close sunshine mdns proxy
		if state.CurrentState.TerminusState == state.TerminusRunning {
			found := false
			if client, err := utils.GetKubeClient(); err == nil {
				if deployments, err := client.AppsV1().Deployments("").List(mainCtx, metav1.ListOptions{}); err == nil {
					for _, d := range deployments.Items {
						if d.Name == "steamheadless" {
							found = true
							if err := sunshine.Restart(); err != nil {
								klog.Error(err)
							}
							break
						}
					}

				}
			}

			if !found {
				sunshine.Close()
			}
		} else {
			// close sunshine mdns proxy, if not started doing nothing
			sunshine.Close()
		}
	})

	// monitor the usb device and mount them automatically
	usb.NewUsbMonitor(mainCtx)

	go func() {
		if err := apis.Start(); err != nil {
			s.Close()
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	state.CurrentState.TerminusdState = state.Running

	<-quit

	cancel()

	if err = apis.Shutdown(); err != nil {
		klog.Error("shutdown error, ", err)
	}
}
