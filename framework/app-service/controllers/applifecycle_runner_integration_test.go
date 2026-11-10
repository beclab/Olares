package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/helmops"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/imgpull"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/k8sstore"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/realclock"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

const envtestImageRef = "docker.io/library/alpine:3.20"

var _ = Describe("Application lifecycle Runner integration", func() {
	It("cancels Pending and reaches PendingCanceled", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-pending-cancel",
			State: appv1.Pending,
			Op:    appv1.InstallOp,
		})

		h.pending().Deliver(h.name, domain.TriggerOp{Op: appv1.CancelOp})
		h.reconcile("Pending to PendingCanceling")
		h.waitState(appv1.PendingCanceling, 2*time.Second)

		h.reconcile("dispatch wait-ns")
		h.waitReady(5 * time.Second)

		h.reconcile("wait-ns to PendingCanceled")
		h.waitState(appv1.PendingCanceled, 3*time.Second)
		h.waitNamespaceGone()
	})

	It("cancels an in-progress download and reaches DownloadingCanceled", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-download-cancel",
			State: appv1.Pending,
			Op:    appv1.InstallOp,
		})

		h.reconcile("Pending to Downloading")
		h.waitState(appv1.Downloading, 2*time.Second)

		h.reconcile("start image pull")
		h.waitImageManagerCreated()
		Expect(setImageProgress(ctx, h.name, h.node, envtestImageRef, 40, 100)).To(Succeed())
		h.waitProgress("40.00")
		Expect(h.runner.InFlight()).To(Equal(1), "image pull must still hold the workload claim")

		h.pending().Deliver(h.name, domain.TriggerOp{Op: appv1.CancelOp})
		h.reconcile("Downloading to DownloadingCanceling")
		h.waitState(appv1.DownloadingCanceling, 2*time.Second)

		h.reconcile("dispatch CancelImagePull")
		Eventually(func() error {
			var im appv1.ImageManager
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: h.name}, &im); err != nil {
				return err
			}
			if im.Status.State != appv1.DownloadingCanceled.String() {
				return fmt.Errorf("ImageManager state=%q, want %q; %s",
					im.Status.State, appv1.DownloadingCanceled, h.diag())
			}
			return nil
		}, 3*time.Second, 20*time.Millisecond).Should(Succeed())
		h.waitReady(3 * time.Second)

		h.reconcile("cancel pull callback to namespace wait")
		h.waitReady(5 * time.Second)

		h.reconcile("namespace wait callback to DownloadingCanceled")
		h.waitState(appv1.DownloadingCanceled, 3*time.Second)
		h.waitNamespaceGone()

		Consistently(func() appv1.ApplicationManagerState {
			return getAMState(ctx, h.name)
		}, 1200*time.Millisecond, 50*time.Millisecond).Should(Equal(appv1.DownloadingCanceled),
			"late callback from the preempted pull changed the terminal state")
	})

	It("cancels an in-progress helm install and reaches InstallingCanceled", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-install-cancel",
			State: appv1.Pending,
			Op:    appv1.InstallOp,
			Park:  &helmPark{install: newGate()},
		})

		h.drivePendingThroughHelmInstall()
		h.waitInFlight(1)
		Expect(h.runner.InFlight()).To(Equal(1), "helm install must still hold the workload claim")

		h.pending().Deliver(h.name, domain.TriggerOp{Op: appv1.CancelOp})
		h.reconcile("Installing to InstallingCanceling")
		h.waitState(appv1.InstallingCanceling, 2*time.Second)

		h.reconcile("dispatch Uninstall")
		h.park.install.release()
		h.waitReady(5 * time.Second)

		h.reconcile("uninstall callback to wait-ns")
		h.waitReady(5 * time.Second)

		h.reconcile("wait-ns to InstallingCanceled")
		h.waitState(appv1.InstallingCanceled, 3*time.Second)
		h.waitNamespaceGone()
	})

	It("cancels Initializing mid-launch and reaches Stopped", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:    "runner-init-cancel",
			State:   appv1.Initializing,
			Op:      appv1.InstallOp,
			NeedApp: true,
			Park:    &helmPark{launch: newGate()},
		})
		h.cancelParkedHandoffToStopped()
	})

	It("cancels Resuming mid-scale and reaches Stopped", func(ctx SpecContext) {
		replicas := appcfg.WorkloadReplicas{"web": 1}
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:    "runner-resume-cancel",
			State:   appv1.Resuming,
			Op:      appv1.ResumeOp,
			NeedApp: true,
			Config: &appcfg.ApplicationConfig{
				AppName:          "runner-resume-cancel",
				Namespace:        "runner-resume-cancel-alice",
				WorkloadReplicas: &replicas,
			},
			Park: &helmPark{scale: newGate()},
		})
		h.cancelParkedHandoffToStopped()
	})

	It("cancels ApplyingEnv mid-helm and reaches Stopped", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:    "runner-applyenv-cancel",
			State:   appv1.ApplyingEnv,
			Op:      appv1.ApplyEnvOp,
			NeedApp: true,
			Park:    &helmPark{applyEnv: newGate()},
		})
		h.cancelParkedHandoffToStopped()
	})

	It("cancels an upgrade download mid-pull and restores Running", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:    "runner-upgrade-cancel",
			State:   appv1.Running,
			Op:      appv1.UpgradeOp,
			NeedApp: true,
		})

		h.pending().Deliver(h.name, domain.TriggerOp{Op: appv1.UpgradeOp})
		h.reconcile("Running to UpgradePending")
		h.waitState(appv1.UpgradePending, 2*time.Second)

		h.reconcile("UpgradePending to UpgradeDownloading")
		h.waitState(appv1.UpgradeDownloading, 2*time.Second)

		h.reconcile("upgrade download prepare")
		h.waitReady(3 * time.Second)
		h.reconcile("start upgrade pull")
		h.waitImageManagerCreated()
		Expect(setImageProgress(ctx, h.name, h.node, envtestImageRef, 40, 100)).To(Succeed())
		h.waitProgress("40.00")
		Expect(h.runner.InFlight()).To(Equal(1), "upgrade pull must still hold the workload claim")

		h.pending().Deliver(h.name, domain.TriggerOp{Op: appv1.CancelOp})
		h.reconcile("UpgradeDownloading to UpgradeDownloadingCanceling")
		h.waitState(appv1.UpgradeDownloadingCanceling, 2*time.Second)

		h.reconcile("dispatch CancelImagePull")
		Eventually(func() error {
			var im appv1.ImageManager
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: h.name}, &im); err != nil {
				return err
			}
			if im.Status.State != appv1.DownloadingCanceled.String() {
				return fmt.Errorf("ImageManager state=%q, want %q; %s",
					im.Status.State, appv1.DownloadingCanceled, h.diag())
			}
			return nil
		}, 3*time.Second, 20*time.Millisecond).Should(Succeed())
		h.waitReady(3 * time.Second)

		h.reconcile("cancel pull restores Running")
		h.waitState(appv1.Running, 2*time.Second)
	})

	It("installs Pending through Running against the real store", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-install-running",
			State: appv1.Pending,
			Op:    appv1.InstallOp,
		})
		h.drivePendingThroughRunning()
		h.waitState(appv1.Running, 3*time.Second)

		h.reconcile("healthy Running tick")
		h.waitState(appv1.Running, 2*time.Second)
	})

	It("stops a Running app and reaches Stopped", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:    "runner-running-stop",
			State:   appv1.Running,
			Op:      appv1.StopOp,
			NeedApp: true,
		})
		h.pending().Deliver(h.name, domain.TriggerOp{Op: appv1.StopOp})
		h.reconcile("Running to Stopping")
		h.waitState(appv1.Stopping, 2*time.Second)
		h.reconcile("dispatch Stop")
		h.waitReady(5 * time.Second)
		h.reconcile("Stop to Stopped")
		h.waitState(appv1.Stopped, 3*time.Second)
	})

	It("uninstalls a Running app and reaches Uninstalled", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:    "runner-running-uninst",
			State:   appv1.Running,
			Op:      appv1.UninstallOp,
			NeedApp: true,
		})
		h.pending().Deliver(h.name, domain.TriggerOp{Op: appv1.UninstallOp})
		h.reconcile("Running to Uninstalling")
		h.waitState(appv1.Uninstalling, 2*time.Second)
		h.reconcile("dispatch Uninstall")
		h.waitReady(5 * time.Second)
		h.reconcile("uninstall to wait-ns")
		h.waitReady(5 * time.Second)
		h.reconcile("wait-ns to Uninstalled")
		h.waitState(appv1.Uninstalled, 3*time.Second)
		h.waitNamespaceGone()
	})

	It("resumes Stopped through Initializing to Running", func(ctx SpecContext) {
		replicas := appcfg.WorkloadReplicas{"web": 1}
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:    "runner-stopped-resume",
			State:   appv1.Stopped,
			Op:      appv1.ResumeOp,
			NeedApp: true,
			Config: &appcfg.ApplicationConfig{
				AppName:          "runner-stopped-resume",
				Namespace:        "runner-stopped-resume-alice",
				WorkloadReplicas: &replicas,
			},
		})
		h.pending().Deliver(h.name, domain.TriggerOp{Op: appv1.ResumeOp})
		h.reconcile("Stopped to Resuming")
		h.waitState(appv1.Resuming, 2*time.Second)
		h.reconcile("dispatch Resume")
		h.waitReady(5 * time.Second)
		h.reconcile("Resuming to Initializing")
		h.waitState(appv1.Initializing, 2*time.Second)
		h.reconcile("dispatch WaitForLaunch")
		h.waitReady(5 * time.Second)
		h.reconcile("Initializing to Running")
		h.waitState(appv1.Running, 3*time.Second)
	})

	It("fails an in-progress download and reaches DownloadFailed", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-download-fail",
			State: appv1.Pending,
			Op:    appv1.InstallOp,
		})
		h.reconcile("Pending to Downloading")
		h.waitState(appv1.Downloading, 2*time.Second)
		h.reconcile("start image pull")
		h.waitImageManagerCreated()
		Expect(setImageFailed(ctx, h.name)).To(Succeed())
		h.waitFailed(5 * time.Second)
		h.reconcile("pull failure to DownloadFailed")
		h.waitState(appv1.DownloadFailed, 3*time.Second)
	})

	It("fails helm install and reaches InstallFailed", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-install-fail",
			State: appv1.Pending,
			Op:    appv1.InstallOp,
			Park:  &helmPark{installErr: fmt.Errorf("helm install boom")},
		})
		h.drivePendingThroughHelmInstall()
		h.waitFailed(5 * time.Second)
		h.reconcile("helm failure to InstallingCanceling")
		h.waitState(appv1.InstallingCanceling, 2*time.Second)
		h.reconcile("dispatch uninstall")
		h.waitReady(5 * time.Second)
		h.reconcile("uninstall callback to wait-ns")
		h.waitReady(5 * time.Second)
		h.reconcile("wait-ns to InstallFailed")
		h.waitState(appv1.InstallFailed, 3*time.Second)
		h.waitNamespaceGone()
	})

	It("keeps a fresh Installing state inside its deadline", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-install-within-deadline",
			State: appv1.Installing,
			Op:    appv1.InstallOp,
		})
		h.backdateStatusTime(time.Minute)
		h.reconcile("Installing still inside the deadline")
		Consistently(func() appv1.ApplicationManagerState {
			return getAMState(ctx, h.name)
		}, 400*time.Millisecond, 50*time.Millisecond).Should(Equal(appv1.Installing), h.diag())
	})

	It("times out a parked helm install into InstallFailed", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-install-timeout",
			State: appv1.Pending,
			Op:    appv1.InstallOp,
			Park:  &helmPark{install: newGate()},
		})
		h.drivePendingThroughHelmInstall()
		h.waitInFlight(1)
		h.backdateStatusTime(31 * time.Minute)

		h.reconcile("Installing deadline to InstallingCanceling")
		h.waitState(appv1.InstallingCanceling, 2*time.Second)
		Eventually(func() int { return h.runner.InFlight() }, 2*time.Second, 20*time.Millisecond).Should(Equal(0), h.diag())
		Expect(h.am().Status.Payload[handlers.PayloadInstallingCancelingTerminal]).To(Equal(string(appv1.InstallFailed)))

		h.reconcile("dispatch uninstall")
		h.waitReady(5 * time.Second)
		h.reconcile("uninstall callback to wait-ns")
		h.waitReady(5 * time.Second)
		h.reconcile("wait-ns to InstallFailed")
		h.waitState(appv1.InstallFailed, 3*time.Second)
		h.waitNamespaceGone()
	})

	It("times out a parked launch into Stopped", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:    "runner-init-timeout",
			State:   appv1.Initializing,
			Op:      appv1.InstallOp,
			NeedApp: true,
			Park:    &helmPark{launch: newGate()},
		})
		h.reconcile("dispatch WaitForLaunch")
		h.waitInFlight(1)
		h.backdateStatusTime(2 * time.Hour)

		h.reconcile("Initializing deadline to InitializingCanceling")
		h.waitState(appv1.InitializingCanceling, 2*time.Second)
		h.reconcile("InitializingCanceling to Stopping")
		h.waitState(appv1.Stopping, 2*time.Second)
		h.reconcile("dispatch Stop")
		h.waitReady(5 * time.Second)
		h.reconcile("Stop to Stopped")
		h.waitState(appv1.Stopped, 3*time.Second)
		Eventually(func() int { return h.runner.InFlight() }, 2*time.Second, 20*time.Millisecond).Should(Equal(0), h.diag())
	})

	It("times out Stopping, Uninstalling, and InstallingCanceling", func(ctx SpecContext) {
		stopping := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-stop-timeout",
			State: appv1.Stopping,
			Op:    appv1.StopOp,
		})
		// CallWorkload deadlines: Stopping and Uninstalling are 30m;
		// InstallingCanceling is 10m. Eleven minutes only expires the latter.
		stopping.backdateStatusTime(31 * time.Minute)
		stopping.reconcile("Stopping deadline")
		stopping.waitState(appv1.StopFailed, 2*time.Second)
		Expect(stopping.am().Status.Reason).To(Equal("StopTimedOut"))

		uninstalling := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-uninstall-timeout",
			State: appv1.Uninstalling,
			Op:    appv1.UninstallOp,
		})
		uninstalling.backdateStatusTime(31 * time.Minute)
		uninstalling.reconcile("Uninstalling deadline")
		uninstalling.waitState(appv1.UninstallFailed, 2*time.Second)
		Expect(uninstalling.am().Status.Reason).To(Equal("UninstallTimedOut"))

		canceling := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-install-cancel-timeout",
			State: appv1.InstallingCanceling,
			Op:    appv1.CancelOp,
		})
		Expect(retry.RetryOnConflict(retry.DefaultRetry, func() error {
			var am appv1.ApplicationManager
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: canceling.name}, &am); err != nil {
				return err
			}
			if am.Status.Payload == nil {
				am.Status.Payload = map[string]string{}
			}
			am.Status.Payload[handlers.PayloadInstallingCancelingPhase] = handlers.PhaseCancelWaitNS
			am.Status.Payload[handlers.PayloadInstallingCancelingTerminal] = string(appv1.InstallFailed)
			// ApplicationManager has no status subresource; status is
			// stored on the main object.
			return k8sClient.Update(ctx, &am)
		})).To(Succeed())
		canceling.backdateStatusTime(11 * time.Minute)
		canceling.reconcile("InstallingCanceling deadline")
		canceling.waitState(appv1.InstallingCancelFailed, 2*time.Second)
		Expect(canceling.am().Status.Payload[handlers.PayloadInstallingCancelingTerminal]).To(Equal(string(appv1.InstallFailed)))
	})

	It("keeps ResumingCancelFailed terminal on Tick", func(ctx SpecContext) {
		h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
			Name:  "runner-resume-cf-tick",
			State: appv1.ResumingCancelFailed,
			Op:    appv1.ResumeOp,
		})
		h.reconcile("terminal tick")
		h.waitState(appv1.ResumingCancelFailed, 2*time.Second)
	})

	DescribeTable("routes a terminal op through the real store",
		func(ctx SpecContext, name string, state appv1.ApplicationManagerState, op appv1.OpType, want appv1.ApplicationManagerState) {
			h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
				Name:    "ri-" + name,
				State:   state,
				Op:      op,
				NeedApp: true,
			})
			h.pending().Deliver(h.name, domain.TriggerOp{Op: op})
			h.reconcile("terminal op")
			h.waitState(want, 2*time.Second)
		},
		Entry("uninstalled install", "uninst-in", appv1.Uninstalled, appv1.InstallOp, appv1.Pending),
		Entry("installfailed uninstall", "if-un", appv1.InstallFailed, appv1.UninstallOp, appv1.Uninstalling),
		Entry("downloadfailed install", "df-in", appv1.DownloadFailed, appv1.InstallOp, appv1.Pending),
		Entry("uninstallfailed uninstall", "uf-un", appv1.UninstallFailed, appv1.UninstallOp, appv1.Uninstalling),
		Entry("stopfailed stop", "sf-st", appv1.StopFailed, appv1.StopOp, appv1.Stopping),
		Entry("stopfailed uninstall", "sf-un", appv1.StopFailed, appv1.UninstallOp, appv1.Uninstalling),
		Entry("resumefailed resume", "rf-re", appv1.ResumeFailed, appv1.ResumeOp, appv1.Resuming),
		Entry("resumefailed applyenv", "rf-ae", appv1.ResumeFailed, appv1.ApplyEnvOp, appv1.ApplyingEnv),
		Entry("resumefailed uninstall", "rf-un", appv1.ResumeFailed, appv1.UninstallOp, appv1.Uninstalling),
		Entry("upgradefailed upgrade", "gf-up", appv1.UpgradeFailed, appv1.UpgradeOp, appv1.UpgradePending),
		Entry("upgradefailed uninstall", "gf-un", appv1.UpgradeFailed, appv1.UninstallOp, appv1.Uninstalling),
		Entry("applyenvfailed applyenv", "af-ae", appv1.ApplyEnvFailed, appv1.ApplyEnvOp, appv1.ApplyingEnv),
		Entry("applyenvfailed stop", "af-st", appv1.ApplyEnvFailed, appv1.StopOp, appv1.Stopping),
		Entry("applyenvfailed uninstall", "af-un", appv1.ApplyEnvFailed, appv1.UninstallOp, appv1.Uninstalling),
		Entry("pendingcanceled install", "pc-in", appv1.PendingCanceled, appv1.InstallOp, appv1.Pending),
		Entry("downloadingcanceled install", "dc-in", appv1.DownloadingCanceled, appv1.InstallOp, appv1.Pending),
		Entry("installingcanceled install", "ic-in", appv1.InstallingCanceled, appv1.InstallOp, appv1.Pending),
		Entry("pendingcancelfailed cancel", "pcf-c", appv1.PendingCancelFailed, appv1.CancelOp, appv1.PendingCanceling),
		Entry("downloadingcancelfailed cancel", "dcf-c", appv1.DownloadingCancelFailed, appv1.CancelOp, appv1.DownloadingCanceling),
		Entry("installingcancelfailed cancel", "icf-c", appv1.InstallingCancelFailed, appv1.CancelOp, appv1.InstallingCanceling),
		Entry("installingcancelfailed uninstall", "icf-u", appv1.InstallingCancelFailed, appv1.UninstallOp, appv1.Uninstalling),
		Entry("initializingcancelfailed cancel", "initcf-c", appv1.InitializingCancelFailed, appv1.CancelOp, appv1.InitializingCanceling),
		Entry("initializingcancelfailed uninstall", "initcf-u", appv1.InitializingCancelFailed, appv1.UninstallOp, appv1.Uninstalling),
		Entry("upgradingcancelfailed cancel", "ucf-c", appv1.UpgradingCancelFailed, appv1.CancelOp, appv1.UpgradingCanceling),
		Entry("upgradingcancelfailed uninstall", "ucf-u", appv1.UpgradingCancelFailed, appv1.UninstallOp, appv1.Uninstalling),
		Entry("applyingenvcancelfailed cancel", "acf-c", appv1.ApplyingEnvCancelFailed, appv1.CancelOp, appv1.ApplyingEnvCanceling),
		Entry("applyingenvcancelfailed uninstall", "acf-u", appv1.ApplyingEnvCancelFailed, appv1.UninstallOp, appv1.Uninstalling),
		Entry("resumingcancelfailed cancel", "rcf-c", appv1.ResumingCancelFailed, appv1.CancelOp, appv1.ResumingCanceling),
	)

	for _, tc := range remainingTransitionCases() {
		It("covers "+tc.name, func(ctx SpecContext) {
			h := newRunnerEnvtest(ctx, runnerEnvtestOpts{
				Name:        "gap-" + tc.name,
				State:       tc.state,
				Op:          tc.op,
				Type:        tc.appType,
				Payload:     tc.payload,
				Annotations: tc.annotations,
			})
			if tc.age > 0 {
				h.backdateStatusTime(tc.age)
			}
			if tc.action != nil {
				h.pending().Deliver(h.name, tc.action)
			}
			h.reconcile(tc.name)
			h.waitState(tc.want, 3*time.Second)
		})
	}
})

// transitionCase drives one from→to edge through the real runner.
// A nil action is a Tick. age > 0 backdates StatusTime so the state's
// deadline fires TimeoutFired instead.
type transitionCase struct {
	name        string
	state       appv1.ApplicationManagerState
	op          appv1.OpType
	appType     appv1.Type
	payload     map[string]string
	annotations map[string]string
	age         time.Duration
	action      domain.Action
	want        appv1.ApplicationManagerState
}

func remainingTransitionCases() []transitionCase {
	stopped := map[string]string{handlers.AnnotPreUpgradeState: appv1.Stopped.String()}
	chart := map[string]string{
		api.AppTargetChartVersionKey: "2.0.0",
		api.AppVersionKey:            "1.0.0",
	}
	fromStopped := map[string]string{
		handlers.AnnotPreUpgradeState: appv1.Stopped.String(),
		api.AppTargetChartVersionKey:  "2.0.0",
		api.AppVersionKey:             "1.0.0",
	}
	failed := domain.WorkloadFailed{Reason: "envtest", Message: "envtest failure"}
	unschedulable := domain.WorkloadFailed{Reason: constants.AppUnschedulable, Message: "pod is pending", LandedState: appv1.Stopping}
	ready := domain.WorkloadReady{}
	return []transitionCase{
		{name: "empty-install", op: appv1.InstallOp, action: domain.TriggerOp{Op: appv1.InstallOp}, want: appv1.Pending},
		{name: "mw-running", state: appv1.Installing, appType: appv1.Middleware, payload: map[string]string{handlers.PayloadInstallingPhase: handlers.PhaseInstallingWaitLaunch}, action: ready, want: appv1.Running},
		{name: "inst-startup-cancel", state: appv1.Installing, payload: map[string]string{handlers.PayloadInstallingPhase: handlers.PhaseInstallingWaitStartUp}, action: failed, want: appv1.InstallingCanceling},
		{name: "inst-stopping", state: appv1.Installing, payload: map[string]string{handlers.PayloadInstallingPhase: handlers.PhaseInstallingWaitStartUp}, action: unschedulable, want: appv1.Stopping},
		{name: "inst-failed", state: appv1.Installing, payload: map[string]string{handlers.PayloadInstallingPhase: handlers.PhaseInstallingCleanup}, action: failed, want: appv1.InstallFailed},
		{name: "run-applyenv", state: appv1.Running, op: appv1.ApplyEnvOp, action: domain.TriggerOp{Op: appv1.ApplyEnvOp}, want: appv1.ApplyingEnv},
		{name: "upg-init", state: appv1.Upgrading, payload: map[string]string{handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingWaitStartUp}, annotations: chart, action: ready, want: appv1.Initializing},
		{name: "upg-failed", state: appv1.Upgrading, action: failed, want: appv1.UpgradeFailed},
		{name: "upg-cancel", state: appv1.Upgrading, op: appv1.CancelOp, action: domain.TriggerOp{Op: appv1.CancelOp}, want: appv1.UpgradingCanceling},
		{name: "upg-stopped", state: appv1.Upgrading, payload: map[string]string{handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingHelm}, annotations: fromStopped, action: ready, want: appv1.Stopped},
		{name: "upgpend-cancel", state: appv1.UpgradePending, op: appv1.CancelOp, action: domain.TriggerOp{Op: appv1.CancelOp}, want: appv1.UpgradePendingCanceling},
		{name: "upgdl-upg", state: appv1.UpgradeDownloading, payload: map[string]string{handlers.PayloadUpgradeDownloadingPhase: handlers.PhaseUpgradingPull}, action: ready, want: appv1.Upgrading},
		{name: "upgdl-failed", state: appv1.UpgradeDownloading, action: failed, want: appv1.UpgradeFailed},
		{name: "upgpendc-run", state: appv1.UpgradePendingCanceling, want: appv1.Running},
		{name: "upgpendc-stop", state: appv1.UpgradePendingCanceling, annotations: stopped, want: appv1.Stopped},
		{name: "upgpendc-fail", state: appv1.UpgradePendingCanceling, age: 11 * time.Minute, want: appv1.UpgradePendingCancelFailed},
		{name: "upgdlc-stop", state: appv1.UpgradeDownloadingCanceling, payload: map[string]string{handlers.PayloadUpgradeDownloadingCancelingPhase: handlers.PhaseCancelImagePull}, annotations: stopped, action: ready, want: appv1.Stopped},
		{name: "upgdlc-fail", state: appv1.UpgradeDownloadingCanceling, action: failed, want: appv1.UpgradeDownloadingCancelFailed},
		{name: "aenv-init", state: appv1.ApplyingEnv, payload: map[string]string{handlers.PayloadApplyingEnvPhase: handlers.PhaseApplyingEnvHelm}, action: ready, want: appv1.Initializing},
		{name: "aenv-failed", state: appv1.ApplyingEnv, action: failed, want: appv1.ApplyEnvFailed},
		{name: "aenv-stop", state: appv1.ApplyingEnv, op: appv1.StopOp, action: domain.TriggerOp{Op: appv1.StopOp}, want: appv1.Stopping},
		{name: "aenv-stopped", state: appv1.ApplyingEnv, payload: map[string]string{handlers.PayloadApplyingEnvPhase: handlers.PhaseApplyingEnvHelm}, annotations: stopped, action: ready, want: appv1.Stopped},
		{name: "pendc-fail", state: appv1.PendingCanceling, age: 11 * time.Minute, want: appv1.PendingCancelFailed},
		{name: "dlc-fail", state: appv1.DownloadingCanceling, age: 11 * time.Minute, want: appv1.DownloadingCancelFailed},
		{name: "initc-fail", state: appv1.InitializingCanceling, age: 11 * time.Minute, want: appv1.InitializingCancelFailed},
		{name: "resc-fail", state: appv1.ResumingCanceling, age: 11 * time.Minute, want: appv1.ResumingCancelFailed},
		{name: "upgc-failed", state: appv1.UpgradingCanceling, want: appv1.UpgradeFailed},
		{name: "upgc-fail", state: appv1.UpgradingCanceling, age: 11 * time.Minute, want: appv1.UpgradingCancelFailed},
		{name: "aenvc-fail", state: appv1.ApplyingEnvCanceling, age: 11 * time.Minute, want: appv1.ApplyingEnvCancelFailed},
		{name: "res-failed", state: appv1.Resuming, action: failed, want: appv1.ResumeFailed},
		{name: "res-stop", state: appv1.Resuming, op: appv1.StopOp, action: domain.TriggerOp{Op: appv1.StopOp}, want: appv1.Stopping},
		{name: "stopped-uninst", state: appv1.Stopped, op: appv1.UninstallOp, action: domain.TriggerOp{Op: appv1.UninstallOp}, want: appv1.Uninstalling},
		{name: "stopped-upg", state: appv1.Stopped, op: appv1.UpgradeOp, action: domain.TriggerOp{Op: appv1.UpgradeOp}, want: appv1.UpgradePending},
		{name: "stopped-aenv", state: appv1.Stopped, op: appv1.ApplyEnvOp, action: domain.TriggerOp{Op: appv1.ApplyEnvOp}, want: appv1.ApplyingEnv},
		{name: "instfail-done", state: appv1.InstallFailed, payload: map[string]string{handlers.PayloadInstallFailedCleanup: handlers.PhaseInstallFailedCleanupDone}, want: appv1.Uninstalled},
		{name: "upgpendcf-cancel", state: appv1.UpgradePendingCancelFailed, op: appv1.CancelOp, action: domain.TriggerOp{Op: appv1.CancelOp}, want: appv1.UpgradePendingCanceling},
		{name: "upgdlcf-cancel", state: appv1.UpgradeDownloadingCancelFailed, op: appv1.CancelOp, action: domain.TriggerOp{Op: appv1.CancelOp}, want: appv1.UpgradeDownloadingCanceling},
		// Same edge, the trigger the earlier cases do not use.
		{name: "stop-failed-wl", state: appv1.Stopping, action: failed, want: appv1.StopFailed},
		{name: "uninst-failed-wl", state: appv1.Uninstalling, action: failed, want: appv1.UninstallFailed},
		{name: "dl-timeout", state: appv1.Downloading, age: 25 * time.Hour, want: appv1.DownloadingCanceling},
		{name: "upgdl-timeout", state: appv1.UpgradeDownloading, age: 25 * time.Hour, want: appv1.UpgradeDownloadingCanceling},
		{name: "aenv-timeout", state: appv1.ApplyingEnv, age: 31 * time.Minute, want: appv1.ApplyingEnvCanceling},
		{name: "res-timeout", state: appv1.Resuming, age: 61 * time.Minute, want: appv1.ResumingCanceling},
		{name: "upg-timeout", state: appv1.Upgrading, age: 61 * time.Minute, want: appv1.UpgradingCanceling},
	}
}

type runnerEnvtestOpts struct {
	Name        string
	State       appv1.ApplicationManagerState
	Op          appv1.OpType
	Type        appv1.Type
	NeedApp     bool
	Config      *appcfg.ApplicationConfig
	Payload     map[string]string
	Annotations map[string]string
	Park        *helmPark
}

type runnerEnvtest struct {
	ctx      context.Context
	name     string
	ns       string
	node     string
	park     *helmPark
	runner   *runner.Runner
	recorder *fake.EventRecorder
}

func newRunnerEnvtest(ctx context.Context, opts runnerEnvtestOpts) *runnerEnvtest {
	GinkgoHelper()
	ns := opts.Name + "-alice"
	nodeName := opts.Name + "-node"
	h := &runnerEnvtest{
		ctx:      ctx,
		name:     opts.Name,
		ns:       ns,
		node:     nodeName,
		park:     opts.Park,
		recorder: &fake.EventRecorder{},
	}
	if h.park == nil {
		h.park = &helmPark{}
	}

	Expect(deleteIfPresent(ctx, &appv1.ImageManager{ObjectMeta: metav1.ObjectMeta{Name: opts.Name}})).To(Succeed())
	Expect(deleteIfPresent(ctx, &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: opts.Name}})).To(Succeed())
	Expect(deleteIfPresent(ctx, &appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: opts.Name}})).To(Succeed())
	Expect(deleteIfPresent(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})).To(Succeed())
	Expect(deleteIfPresent(ctx, &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: nodeName}})).To(Succeed())

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
		Spec:       corev1.NodeSpec{Unschedulable: false},
	}
	Expect(k8sClient.Create(ctx, node)).To(Succeed())
	DeferCleanup(func(cleanupCtx SpecContext) {
		_ = deleteIfPresent(cleanupCtx, node)
	})
	Expect(retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var current corev1.Node
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: nodeName}, &current); err != nil {
			return err
		}
		current.Status.Conditions = []corev1.NodeCondition{{
			Type:               corev1.NodeReady,
			Status:             corev1.ConditionTrue,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
		}}
		return k8sClient.Status().Update(ctx, &current)
	})).To(Succeed())

	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
	DeferCleanup(func(cleanupCtx SpecContext) {
		_ = releaseEnvtestNamespace(cleanupCtx, ns)
		_ = deleteIfPresent(cleanupCtx, namespace)
	})

	appCfg := opts.Config
	if appCfg == nil {
		appCfg = &appcfg.ApplicationConfig{AppName: opts.Name, Namespace: ns}
	}
	rawConfig, err := json.Marshal(appCfg)
	Expect(err).NotTo(HaveOccurred())

	appType := opts.Type
	if appType == "" {
		appType = appv1.App
	}
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Annotations: opts.Annotations,
		},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      opts.Name,
			AppOwner:     "alice",
			AppNamespace: ns,
			Source:       "market",
			Type:         appType,
			OpType:       opts.Op,
			Config:       string(rawConfig),
		},
		Status: appv1.ApplicationManagerStatus{
			State:   opts.State,
			OpType:  opts.Op,
			OpID:    "envtest-" + opts.Name,
			Payload: opts.Payload,
		},
	}
	Expect(k8sClient.Create(ctx, am)).To(Succeed())
	DeferCleanup(func(cleanupCtx SpecContext) {
		_ = deleteIfPresent(cleanupCtx, am)
		_ = deleteIfPresent(cleanupCtx, &appv1.ImageManager{ObjectMeta: metav1.ObjectMeta{Name: opts.Name}})
		_ = deleteIfPresent(cleanupCtx, &appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: opts.Name}})
	})

	if opts.NeedApp {
		app := &appv1.Application{
			ObjectMeta: metav1.ObjectMeta{Name: opts.Name},
			Spec: appv1.ApplicationSpec{
				Name:      opts.Name,
				Namespace: ns,
				Owner:     "alice",
			},
		}
		Expect(k8sClient.Create(ctx, app)).To(Succeed())
	}

	table := domain.NewDefaultTable()
	handlers.Register(table)
	store := k8sstore.New(k8sClient, k8sstore.WithShadowMode(false))
	puller := imgpull.New(k8sClient,
		func(context.Context, *appv1.ApplicationManager, *appcfg.ApplicationConfig) ([]appv1.Ref, error) {
			return []appv1.Ref{{
				Name:            envtestImageRef,
				ImagePullPolicy: corev1.PullIfNotPresent,
			}}, nil
		})
	workloads := helmops.New(k8sClient, cfg,
		helmops.WithPollInterval(10*time.Millisecond),
		helmops.WithNamespaceWaitTimeout(2*time.Second),
		helmops.WithHelmOpsFactory(h.park.factory),
	)
	appRunner, err := runner.New(runner.Config{
		Engine:      domain.NewEngine(table),
		AMStore:     store,
		WorkloadOps: workloads,
		Validator:   &fake.InstallValidator{},
		ImagePuller: puller,
		Recorder:    h.recorder,
		Clock:       realclock.New(),
		Requeuer:    &runner.RecordingRequeuer{},
	})
	Expect(err).NotTo(HaveOccurred())
	h.runner = appRunner
	DeferCleanup(func() {
		h.park.releaseAll()
		appRunner.Stop()
	})
	return h
}

func (h *runnerEnvtest) pending() *runner.PendingActionQueue {
	return h.runner.Pending()
}

func (h *runnerEnvtest) reconcile(round string) {
	GinkgoHelper()
	_, err := h.runner.Reconcile(h.ctx, runner.Request{Name: h.name})
	Expect(err).NotTo(HaveOccurred(), "%s: %s", round, h.diag())
}

func (h *runnerEnvtest) waitState(want appv1.ApplicationManagerState, timeout time.Duration) {
	GinkgoHelper()
	Eventually(func() appv1.ApplicationManagerState {
		return getAMState(h.ctx, h.name)
	}, timeout, 20*time.Millisecond).Should(Equal(want), h.diag())
}

func (h *runnerEnvtest) waitReady(timeout time.Duration) {
	GinkgoHelper()
	Eventually(func() bool {
		_ = releaseEnvtestNamespace(h.ctx, h.ns)
		action, ok := h.runner.Pending().Peek(h.name)
		if !ok {
			return false
		}
		_, ready := action.(domain.WorkloadReady)
		return ready
	}, timeout, 20*time.Millisecond).Should(BeTrue(), h.diag())
}

func (h *runnerEnvtest) waitFailed(timeout time.Duration) {
	GinkgoHelper()
	Eventually(func() bool {
		action, ok := h.runner.Pending().Peek(h.name)
		if !ok {
			return false
		}
		_, failed := action.(domain.WorkloadFailed)
		return failed
	}, timeout, 20*time.Millisecond).Should(BeTrue(), h.diag())
}

func (h *runnerEnvtest) waitNamespaceGone() {
	GinkgoHelper()
	Eventually(func() bool {
		_ = releaseEnvtestNamespace(h.ctx, h.ns)
		var current corev1.Namespace
		err := k8sClient.Get(h.ctx, types.NamespacedName{Name: h.ns}, &current)
		return apierrors.IsNotFound(err)
	}, 2*time.Second, 20*time.Millisecond).Should(BeTrue(), h.diag())
}

func (h *runnerEnvtest) waitImageManagerCreated() {
	GinkgoHelper()
	Eventually(func() error {
		var im appv1.ImageManager
		if err := k8sClient.Get(h.ctx, types.NamespacedName{Name: h.name}, &im); err != nil {
			return err
		}
		if len(im.Spec.Nodes) != 1 || im.Spec.Nodes[0] != h.node {
			return fmt.Errorf("ImageManager nodes=%v, want [%s]", im.Spec.Nodes, h.node)
		}
		if len(im.Spec.Refs) != 1 || im.Spec.Refs[0].Name != envtestImageRef {
			return fmt.Errorf("ImageManager refs=%v, want [%s]", im.Spec.Refs, envtestImageRef)
		}
		return nil
	}, 2*time.Second, 20*time.Millisecond).Should(Succeed())
}

func (h *runnerEnvtest) waitProgress(want string) {
	GinkgoHelper()
	Eventually(func() error {
		for _, event := range h.recorder.Events() {
			if event.RawAppName == h.name && event.Progress == want {
				return nil
			}
		}
		return fmt.Errorf("progress %s not observed; %s", want, h.diag())
	}, 3*time.Second, 50*time.Millisecond).Should(Succeed())
}

func (h *runnerEnvtest) waitInFlight(n int) {
	GinkgoHelper()
	Eventually(func() int {
		return h.runner.InFlight()
	}, 2*time.Second, 20*time.Millisecond).Should(Equal(n), h.diag())
}

func (h *runnerEnvtest) ensureApp() {
	GinkgoHelper()
	app := &appv1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: h.name},
		Spec: appv1.ApplicationSpec{
			Name:      h.name,
			Namespace: h.ns,
			Owner:     "alice",
		},
	}
	err := k8sClient.Create(h.ctx, app)
	if apierrors.IsAlreadyExists(err) {
		return
	}
	Expect(err).NotTo(HaveOccurred())
}

func (h *runnerEnvtest) drivePendingThroughRunning() {
	GinkgoHelper()
	h.drivePendingThroughHelmInstall()
	h.waitReady(5 * time.Second)
	h.reconcile("helm to scale-up")
	h.waitReady(5 * time.Second)
	h.reconcile("scale-up to wait-startup")
	h.waitReady(5 * time.Second)
	h.reconcile("wait-startup to Initializing")
	h.waitState(appv1.Initializing, 3*time.Second)
	h.ensureApp()
	h.reconcile("dispatch WaitForLaunch")
	h.waitReady(5 * time.Second)
	h.reconcile("Initializing to Running")
}

func (h *runnerEnvtest) drivePendingThroughHelmInstall() {
	GinkgoHelper()
	h.reconcile("Pending to Downloading")
	h.waitState(appv1.Downloading, 2*time.Second)
	h.reconcile("start image pull")
	h.waitImageManagerCreated()
	Expect(setImageComplete(h.ctx, h.name, h.node, envtestImageRef)).To(Succeed())
	h.waitReady(5 * time.Second)

	h.reconcile("Downloading to Installing")
	h.waitState(appv1.Installing, 2*time.Second)

	h.reconcile("expose-ports")
	h.waitReady(3 * time.Second)
	h.reconcile("validate")
	h.waitReady(3 * time.Second)
	h.reconcile("helm-install")
	Eventually(func() string {
		var am appv1.ApplicationManager
		if err := k8sClient.Get(h.ctx, types.NamespacedName{Name: h.name}, &am); err != nil {
			return ""
		}
		return am.Status.Payload[handlers.PayloadInstallingPhase]
	}, 2*time.Second, 20*time.Millisecond).Should(Equal(handlers.PhaseInstallingHelmInstall), h.diag())
}

func (h *runnerEnvtest) cancelParkedHandoffToStopped() {
	GinkgoHelper()
	h.reconcile("dispatch parked workload")
	h.waitInFlight(1)

	h.pending().Deliver(h.name, domain.TriggerOp{Op: appv1.CancelOp})
	h.reconcile("operating state to *Canceling")
	Eventually(func() bool {
		state := getAMState(h.ctx, h.name)
		return len(state) > 0 && state != appv1.Initializing &&
			state != appv1.Resuming && state != appv1.ApplyingEnv &&
			state != appv1.Running
	}, 2*time.Second, 20*time.Millisecond).Should(BeTrue(), h.diag())

	h.reconcile("Canceling to Stopping")
	h.waitState(appv1.Stopping, 2*time.Second)

	h.reconcile("dispatch Stop")
	h.park.releaseAll()
	h.waitReady(5 * time.Second)
	h.reconcile("Stop to Stopped")
	h.waitState(appv1.Stopped, 3*time.Second)
}

func (h *runnerEnvtest) diag() string {
	return runnerIntegrationDiagnostics(h.ctx, h.runner, h.name, h.ns)
}

func (h *runnerEnvtest) am() appv1.ApplicationManager {
	GinkgoHelper()
	var am appv1.ApplicationManager
	Expect(k8sClient.Get(h.ctx, types.NamespacedName{Name: h.name}, &am)).To(Succeed())
	return am
}

func (h *runnerEnvtest) backdateStatusTime(age time.Duration) {
	GinkgoHelper()
	Expect(retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var am appv1.ApplicationManager
		if err := k8sClient.Get(h.ctx, types.NamespacedName{Name: h.name}, &am); err != nil {
			return err
		}
		stamp := metav1.NewTime(time.Now().Add(-age))
		am.Status.StatusTime = &stamp
		// ApplicationManager has no status subresource; status is
		// stored on the main object.
		return k8sClient.Update(h.ctx, &am)
	})).To(Succeed())
}

type gate struct {
	ch   chan struct{}
	once sync.Once
}

func newGate() *gate { return &gate{ch: make(chan struct{})} }

func (g *gate) wait(ctx context.Context) error {
	if g == nil {
		return nil
	}
	if ctx == nil {
		<-g.ch
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-g.ch:
		return nil
	}
}

func (g *gate) release() {
	if g == nil {
		return
	}
	g.once.Do(func() { close(g.ch) })
}

type helmPark struct {
	install, applyEnv, launch, scale *gate
	scaleN                           atomic.Int32
	installErr                       error
	uninstallErr                     error
}

func (p *helmPark) factory(ctx context.Context, _ *rest.Config, _ *appcfg.ApplicationConfig, _ string, _ appinstaller.Opt) (appinstaller.HelmOpsInterface, error) {
	return &parkedHelmOps{park: p, ctx: ctx}, nil
}

func (p *helmPark) releaseAll() {
	if p == nil {
		return
	}
	p.install.release()
	p.applyEnv.release()
	p.launch.release()
	p.scale.release()
}

type parkedHelmOps struct {
	park *helmPark
	ctx  context.Context
}

func (o *parkedHelmOps) Uninstall() error              { return o.park.uninstallErr }
func (o *parkedHelmOps) Upgrade() error                { return nil }
func (o *parkedHelmOps) RollBack() error               { return nil }
func (o *parkedHelmOps) UninstallAll() error           { return nil }
func (o *parkedHelmOps) WaitForStartUp() (bool, error) { return true, nil }

func (o *parkedHelmOps) Install() error {
	if err := o.park.install.wait(o.ctx); err != nil {
		return err
	}
	return o.park.installErr
}

func (o *parkedHelmOps) ApplyEnv() error {
	if err := o.park.applyEnv.wait(o.ctx); err != nil {
		return err
	}
	return nil
}

func (o *parkedHelmOps) WaitForLaunch() (bool, error) {
	if err := o.park.launch.wait(o.ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (o *parkedHelmOps) Scale(int32) error {
	if o.park.scale != nil && o.park.scaleN.Add(1) == 1 {
		if err := o.park.scale.wait(o.ctx); err != nil {
			return err
		}
	}
	return nil
}

func setImageFailed(ctx context.Context, name string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var im appv1.ImageManager
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, &im); err != nil {
			return err
		}
		now := metav1.Now()
		im.Status.State = "failed"
		im.Status.Message = "image pull failed"
		im.Status.StatusTime = &now
		im.Status.UpdateTime = &now
		return k8sClient.Status().Update(ctx, &im)
	})
}

func setImageComplete(ctx context.Context, name, node, ref string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var im appv1.ImageManager
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, &im); err != nil {
			return err
		}
		now := metav1.Now()
		im.Status.State = "completed"
		im.Status.Message = "completed"
		im.Status.StatusTime = &now
		im.Status.UpdateTime = &now
		im.Status.Conditions = map[string]map[string]map[string]string{
			node: {ref: {"offset": "100", "total": "100"}},
		}
		return k8sClient.Status().Update(ctx, &im)
	})
}

func setImageProgress(
	ctx context.Context,
	name, node, ref string,
	offset, total int,
) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var im appv1.ImageManager
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, &im); err != nil {
			return err
		}
		now := metav1.Now()
		im.Status.State = appv1.Downloading.String()
		im.Status.Message = "downloading"
		im.Status.StatusTime = &now
		im.Status.UpdateTime = &now
		im.Status.Conditions = map[string]map[string]map[string]string{
			node: {
				ref: {
					"offset": fmt.Sprintf("%d", offset),
					"total":  fmt.Sprintf("%d", total),
				},
			},
		}
		return k8sClient.Status().Update(ctx, &im)
	})
}

func getAMState(ctx context.Context, name string) appv1.ApplicationManagerState {
	var am appv1.ApplicationManager
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, &am); err != nil {
		return ""
	}
	return am.Status.State
}

func runnerIntegrationDiagnostics(
	ctx context.Context,
	appRunner *runner.Runner,
	amName, namespace string,
) string {
	var am appv1.ApplicationManager
	amErr := k8sClient.Get(ctx, types.NamespacedName{Name: amName}, &am)
	var im appv1.ImageManager
	imErr := k8sClient.Get(ctx, types.NamespacedName{Name: amName}, &im)
	var ns corev1.Namespace
	nsErr := k8sClient.Get(ctx, types.NamespacedName{Name: namespace}, &ns)
	action, pending := appRunner.Pending().Peek(amName)
	return fmt.Sprintf(
		"AM(state=%q payload=%v err=%v) IM(state=%q conditions=%v err=%v) namespace(err=%v deletion=%v) inFlight=%d pending=%t action=%#v",
		am.Status.State, am.Status.Payload, amErr,
		im.Status.State, im.Status.Conditions, imErr,
		nsErr, ns.DeletionTimestamp,
		appRunner.InFlight(), pending, action,
	)
}

func deleteIfPresent(ctx context.Context, obj client.Object) error {
	err := k8sClient.Delete(ctx, obj)
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// releaseEnvtestNamespace drains the finalizers of a Namespace that is
// already Terminating. envtest runs no namespace controller, so the
// "kubernetes" spec finalizer is never cleared and the object never
// reaches NotFound after Delete. spec.finalizers can only be mutated
// through the finalize subresource, which is exactly what the real
// namespace controller calls once it has purged the contents.
func releaseEnvtestNamespace(ctx context.Context, name string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var ns corev1.Namespace
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, &ns); err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		if ns.DeletionTimestamp == nil {
			return nil
		}
		if len(ns.Finalizers) > 0 {
			ns.Finalizers = nil
			if err := k8sClient.Update(ctx, &ns); err != nil {
				return err
			}
		}
		if len(ns.Spec.Finalizers) == 0 {
			return nil
		}
		ns.Spec.Finalizers = nil
		err := k8sClient.SubResource("finalize").Update(ctx, &ns)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})
}
