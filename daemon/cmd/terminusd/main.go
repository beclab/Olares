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
	"github.com/beclab/Olares/daemon/internel/apiserver/handlers"
	"github.com/beclab/Olares/daemon/internel/ble"
	"github.com/beclab/Olares/daemon/internel/mdns"
	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/internel/watcher/cert"
	intranetwatcher "github.com/beclab/Olares/daemon/internel/watcher/intranet"
	"github.com/beclab/Olares/daemon/internel/watcher/lpvpndns"
	mountwatcher "github.com/beclab/Olares/daemon/internel/watcher/mount"
	"github.com/beclab/Olares/daemon/internel/watcher/system"
	"github.com/beclab/Olares/daemon/internel/watcher/systemenv"
	"github.com/beclab/Olares/daemon/internel/watcher/upgrade"
	"github.com/beclab/Olares/daemon/internel/watcher/usb"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	upgradecmd "github.com/beclab/Olares/daemon/pkg/commands/upgrade"
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

	// Clear leftover overlay op locks from a previous crash so status does not
	// report a permanent activating/deactivating state. Do not reconcile app
	// settings or cni-dhcp here.
	handlers.ClearOverlayGatewayOpLocks()

	// Cluster power operations are recorded on disk before they are carried
	// out. A daemon that cannot record them serves the routes and refuses,
	// rather than powering off a cluster it will not be able to report on.
	//
	// This is also the only place anything is given the ability to act on
	// this machine: the node-local half of an operation runs through the
	// modules built into this daemon, and they are handed to the routes here
	// rather than at package level, so no test binary holds them.
	//
	// Both halves are given the same set, which InitClusterOperations then
	// closes: the master decides an operation exists and dispatches its node
	// half against these modules, and a node carries that half out against
	// the same ones.
	handlers.InstallNodeOperations(clusterop.DefaultRegistry())

	// The upgrade's node-local half is assembled here for the same reason the
	// power one is a field of Deps: it acts on this machine, so the only place
	// it should be built is the one binary that is allowed to. A test never
	// reaches this file, and so never holds a runner that can execute
	// olares-cli against the host it is running on.
	//
	// Every node builds one, control node or not. A compute node serves the
	// node-local upgrade routes and never builds a manager at all, and the
	// control node reaches its own stages through the same runner it reaches
	// a compute node's through.
	deps := clusterop.NewDeps()
	stageRunner, err := clusterop.NewLocalUpgradeStageRunner(
		commands.CLUSTER_OPERATIONS_DIR, runUpgradeStage)
	if err != nil {
		// Logged, not Fatal. Hardware, mDNS, the API and status reporting do
		// not need this directory writable, and a full disk — the usual
		// reason settleInterrupted cannot record — is exactly when the rest
		// of the daemon has to stay up. The upgrade routes refuse while the
		// runner is missing; see InstallUpgradeStageRunner.
		klog.Error("upgrades are unavailable, ", err)
	} else {
		handlers.InstallUpgradeStageRunner(stageRunner)
		deps.Upgrade = clusterop.NewUpgradeDeps(stageRunner)
	}

	if err := handlers.InitClusterOperations(commands.CLUSTER_OPERATIONS_DIR, deps); err != nil {
		// Logged, not Fatal: the same directory being unwritable used to
		// disable cluster operations and leave everything else running. That
		// is still the right trade. A missing orchestrator is refused when
		// an upgrade is asked for — see upgradeOrchestrator — rather than
		// by taking the whole daemon down before anyone can ask.
		klog.Error("cluster operations are unavailable, ", err)
	}

	mainCtx, cancel := context.WithCancel(context.Background())

	// Set up the shared informers' lifecycle context. The factory itself starts
	// lazily once the cluster is reachable; readers fall back to live Lists
	// until the cache has synced.
	utils.InitInformers(mainCtx)

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

	sunshine := mdns.NewSunShineProxyWithoutStart(mainCtx)
	defer sunshine.Close()

	state.WatchStatus(mainCtx, []watcher.Watcher{
		system.NewSystemWatcher(),
		system.NewBridgeConnectionWatcher(),
		// usb.NewUsbWatcher(),
		usb.NewUmountWatcher(),
		upgrade.NewUpgradeWatcher(),
		cert.NewCertWatcher(),
		systemenv.NewSystemEnvWatcher(),
		intranetwatcher.NewApplicationWatcher(),
		mountwatcher.NewMountWatcher(),
		lpvpndns.NewWatcher(),
	}, func() {
		if s != nil {
			startMDNS := true
			if client, err := utils.GetKubeClient(); err == nil {
				if _, _, nodeRole, err := utils.GetThisNodeName(mainCtx, client); err == nil && nodeRole != "master" {
					startMDNS = false
				}
			}

			if startMDNS {
				if err := s.Restart(); err != nil {
					klog.Error(err)
				}
			} else {
				s.Close()
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
				// Use a name field selector so the apiserver returns only the
				// steamheadless deployment instead of every deployment in the
				// cluster, which the status loop otherwise listed and
				// deserialized on every tick.
				if deployments, err := client.AppsV1().Deployments("").List(mainCtx, metav1.ListOptions{
					FieldSelector: "metadata.name=steamheadless",
				}); err == nil {
					if len(deployments.Items) > 0 {
						// check if the overlay gateway is enabled and the steamheadless
						// is running with the overlay gateway enabled, if not restart the sunshine mdns proxy
						var overlaygatewayEnabled bool

						c, err := utils.FindBridgeConnection(mainCtx)
						if err != nil {
							klog.Error("find bridge connection error, ", err)
						} else {
							if c != nil && c.Active {
								enabled, err := utils.GetApplicationSettings(mainCtx, "steamheadless", "enableOverlayGateway")
								if err != nil {
									klog.Error("get application settings error, ", err)
								} else {
									if enabled == "true" {
										overlaygatewayEnabled = true
									}
								}
							}
						}

						if !overlaygatewayEnabled {
							found = true
							if err := sunshine.Restart(); err != nil {
								klog.Error(err)
							}
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

// runUpgradeStage is this node's whole ability to be told to upgrade itself.
//
// Both kinds of work arrive through the same route and are recorded the same
// way, so they are dispatched here by stage name rather than by two endpoints
// with two sets of state. The node prepare stage is the one that cannot go
// through olares-cli's plan: it is what installs the olares-cli that has the
// plan.
func runUpgradeStage(ctx context.Context, req clusterop.UpgradeStageRequest) error {
	if req.Stage == clusterop.StageNodePrepare {
		return upgradecmd.PrepareNode(ctx, req.Version, req.CliURL, req.WizardURL)
	}
	return clusterop.RunUpgradeStage(ctx, req)
}
