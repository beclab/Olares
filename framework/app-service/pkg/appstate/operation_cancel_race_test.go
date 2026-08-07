package appstate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"k8s.io/client-go/rest"
)

// The three tests below pin the cancel guards added to InstallingApp,
// UpgradingApp and ApplyingEnvApp, mirroring InitializingApp: when the
// operation context is canceled (a *Canceling app's Cleanup, or a force
// uninstall) the in-flight goroutine must NOT publish its own terminal state,
// because the initiator already owns the transition and *Canceling -> *Failed
// is not a declared edge in StateTransitions.
//
// Each test blocks the operation goroutine on a seam, cancels the operation via
// Cleanup (what cancelOperation does in production), releases the seam with an
// error and then asserts that the ApplicationManager was left in the operating
// state for the initiator to move on.

// blockingKubeConfig returns a KubeConfig seam that reports it was reached, then
// blocks until release is closed and finally fails. Every operation goroutine
// calls it first, so it is a reliable place to freeze one mid-flight.
func blockingKubeConfig(started chan<- struct{}, release <-chan struct{}) func() (*rest.Config, error) {
	return func() (*rest.Config, error) {
		close(started)
		<-release
		return nil, errors.New("operation aborted")
	}
}

func TestUpgrade_CanceledExecSkipsUpgradeFailed(t *testing.T) {
	isolateKubeconfig(t)

	am := buildAM("upgrade-race", appv1alpha1.App, appv1alpha1.Upgrading, appv1alpha1.UpgradeOp,
		configJSON(t, "upgrade-race", false))
	c := newFakeClient(t, am)
	deps, _ := newTestDeps(c)

	started, release := make(chan struct{}), make(chan struct{})
	deps.KubeConfig = blockingKubeConfig(started, release)

	p := &UpgradingApp{
		baseOperationApp: &baseOperationApp{
			ttl:             time.Hour,
			baseStatefulApp: &baseStatefulApp{manager: am, client: c, deps: deps},
		},
		downloadTTL: time.Hour,
		imageClient: &fakeImageManager{},
	}

	in, err := p.Exec(context.TODO())
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}

	<-started
	in.Cleanup(context.TODO())
	close(release)

	// Finally blocks until the goroutine either publishes a transition or closes
	// finallyCh on its way out, so it doubles as a barrier: no sleep needed.
	p.Finally()

	if got := getAM(t, c, am.Name).Status.State; got != appv1alpha1.Upgrading {
		t.Fatalf("state = %q, want %q left for the cancel handler", got, appv1alpha1.Upgrading)
	}
}

func TestApplyEnv_CanceledExecSkipsApplyEnvFailed(t *testing.T) {
	isolateKubeconfig(t)

	am := buildAM("applyenv-race", appv1alpha1.App, appv1alpha1.ApplyingEnv, appv1alpha1.ApplyEnvOp,
		configJSON(t, "applyenv-race", false))
	c := newFakeClient(t, am)
	deps, _ := newTestDeps(c)

	started, release := make(chan struct{}), make(chan struct{})
	deps.KubeConfig = blockingKubeConfig(started, release)

	a := &ApplyingEnvApp{
		baseOperationApp: &baseOperationApp{
			ttl:             time.Hour,
			baseStatefulApp: &baseStatefulApp{manager: am, client: c, deps: deps},
		},
	}

	in, err := a.Exec(context.TODO())
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}

	<-started
	in.Cleanup(context.TODO())
	close(release)

	// No finallyCh barrier here, so give the goroutine room to reach its guard
	// before running the hook it would have installed without the fix.
	time.Sleep(200 * time.Millisecond)
	in.Finally()

	if got := getAM(t, c, am.Name).Status.State; got != appv1alpha1.ApplyingEnv {
		t.Fatalf("state = %q, want %q left for the cancel handler", got, appv1alpha1.ApplyingEnv)
	}
}

func TestInstall_CanceledExecSkipsInstallFailed(t *testing.T) {
	isolateKubeconfig(t)

	am := buildAM("install-race", appv1alpha1.App, appv1alpha1.Installing, appv1alpha1.InstallOp,
		configJSON(t, "install-race", false))
	c := newFakeClient(t, am)
	deps, tf := newTestDeps(c)

	// Freeze inside validation: it is the first step of the install goroutine
	// and its failure funnels straight into toInstallFailed.
	started, release := make(chan struct{}), make(chan struct{})
	tf.validation = func(ctx context.Context, in validation.Input) (validation.Decision, error) {
		close(started)
		<-release
		return validation.Decision{}, errors.New("validation aborted")
	}

	p := &InstallingApp{
		baseOperationApp: &baseOperationApp{
			ttl:             time.Hour,
			baseStatefulApp: &baseStatefulApp{manager: am, client: c, deps: deps},
		},
	}

	in, err := p.Exec(context.TODO())
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}

	<-started
	in.Cleanup(context.TODO())
	close(release)

	time.Sleep(200 * time.Millisecond)
	in.Finally()

	if got := getAM(t, c, am.Name).Status.State; got != appv1alpha1.Installing {
		t.Fatalf("state = %q, want %q left for the cancel handler", got, appv1alpha1.Installing)
	}
}
