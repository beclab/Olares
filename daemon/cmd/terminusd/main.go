package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"bytetrade.io/web3os/terminusd/cmd/terminusd/version"
	"bytetrade.io/web3os/terminusd/internel/apiserver"
	"bytetrade.io/web3os/terminusd/internel/ble"
	"bytetrade.io/web3os/terminusd/internel/mdns"
	"bytetrade.io/web3os/terminusd/internel/watcher"
	"bytetrade.io/web3os/terminusd/internel/watcher/system"
	"bytetrade.io/web3os/terminusd/internel/watcher/upgrade"
	"bytetrade.io/web3os/terminusd/internel/watcher/usb"
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/utils"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func main() {

	state.CurrentState.TerminusdState = state.Initialize

	port := 18088
	var showVersion bool

	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine.BoolVar(&showVersion, "version", false, "show olaresd version")

	pflag.Parse()

	if showVersion {
		fmt.Println(version.Version())
		return
	}

	commands.Init()

	mainCtx, cancel := context.WithCancel(context.Background())

	apis, err := apiserver.NewServer(mainCtx, port)
	if err != nil {
		panic(err)
	}

	if err := state.CheckCurrentStatus(mainCtx); err != nil {
		klog.Error(err)
	}

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
