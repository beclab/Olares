package mwops

import (
	"context"
	"fmt"

	kbopv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/kubeblocks"
	"github.com/beclab/Olares/framework/app-service/pkg/middlewareinstaller"
)

// mongoReleaseSecretName is the Helm v3 release Secret name for the
// legacy pre-KubeBlocks mongodb chart (chart version v1).
// UninstallLegacyMongo Get+Deletes this after the helm uninstall as
// a defensive cleanup — the legacy oldMongodbUninstall used the
// same literal (pkg/appstate/types.go:405).
const mongoReleaseSecretName = "sh.helm.release.v1.mongodb.v1"

// mongoUninstallToleratedErr is the one error string
// oldMongodbUninstall treats as success. Preserved verbatim to keep
// the uninstall path byte-identical to legacy semantics — see
// doc.go "UninstallLegacyMongo error tolerance".
const mongoUninstallToleratedErr = "failed to delete release: mongodb"

// KubeBlocksOps is the subset of pkg/kubeblocks.OperationOptions
// the adapter uses. Extracted so tests can substitute a fake.
type KubeBlocksOps interface {
	Start() error
	Stop() error
}

// KubeBlocksOpsFactory builds a KubeBlocksOps for a given opsType /
// AM / client triple. Production uses defaultKBFactory (below);
// tests inject a fake factory via WithKubeBlocksFactory.
type KubeBlocksOpsFactory func(ctx context.Context, opsType kbopv1alpha1.OpsType, am *appv1.ApplicationManager, c client.Client) KubeBlocksOps

// MongoUninstaller is the seam UninstallLegacyMongo delegates to
// after casting the port-level cfg. Production wires it to a
// closure over kubeConfig calling middlewareinstaller.Uninstall;
// tests pass a stub.
//
// The closure form is used (rather than exposing *rest.Config on
// the adapter) because kubeConfig is exclusively an input to the
// helm client middlewareinstaller.Uninstall builds internally.
// The adapter itself never touches it, so keeping it out of the
// adapter's fields keeps the surface minimal.
type MongoUninstaller func(ctx context.Context, cfg *middlewareinstaller.MiddlewareConfig) error

// Ops is the production ports.MiddlewareOps adapter.
type Ops struct {
	c              client.Client
	newKBOp        KubeBlocksOpsFactory
	mongoUninstall MongoUninstaller
}

// Option customises an Ops at construction time.
type Option func(*Ops)

// WithKubeBlocksFactory overrides the KubeBlocks ops factory.
// Passing nil is a no-op so tests overriding only the mongo path
// don't have to reassert the production factory.
func WithKubeBlocksFactory(f KubeBlocksOpsFactory) Option {
	return func(o *Ops) {
		if f != nil {
			o.newKBOp = f
		}
	}
}

// New constructs an Ops. `c` and `mongoUninstall` are required;
// nil c would NPE in UninstallLegacyMongo (secret Get/Delete) and
// break every Start/Stop call site (KubeBlocks OpsRequest creation
// uses this client); nil mongoUninstall would silently no-op the
// helm uninstall path, leaving the mongodb release installed.
func New(c client.Client, mongoUninstall MongoUninstaller, opts ...Option) *Ops {
	if c == nil {
		panic("mwops.New: client is nil")
	}
	if mongoUninstall == nil {
		panic("mwops.New: mongoUninstall is nil")
	}
	o := &Ops{
		c:              c,
		newKBOp:        defaultKBFactory,
		mongoUninstall: mongoUninstall,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Start implements ports.MiddlewareOps. Dispatches a KubeBlocks
// StartType OpsRequest. Idempotency comes from kubeblocks.NewOperation
// suffixing the OpsRequest name with a UUID (see kubeblocks/operations.go:73)
// — repeat Starts create fresh CRs the KubeBlocks operator
// individually reconciles to no-ops when the target Cluster is
// already Running.
func (o *Ops) Start(ctx context.Context, am *appv1.ApplicationManager) error {
	if am == nil {
		return fmt.Errorf("mwops: Start requires non-nil AM")
	}
	return o.newKBOp(ctx, kbopv1alpha1.StartType, am, o.c).Start()
}

// Stop implements ports.MiddlewareOps. Mirror of Start. Same
// idempotency guarantee (UUID-suffixed OpsRequest name).
func (o *Ops) Stop(ctx context.Context, am *appv1.ApplicationManager) error {
	if am == nil {
		return fmt.Errorf("mwops: Stop requires non-nil AM")
	}
	return o.newKBOp(ctx, kbopv1alpha1.StopType, am, o.c).Stop()
}

// UninstallLegacyMongo implements ports.MiddlewareOps. See doc.go
// "UninstallLegacyMongo error tolerance" for the mongodb-specific
// error skip.
//
// Contract:
//
//   - cfg MUST be a non-nil *middlewareinstaller.MiddlewareConfig
//     (cast on entry; anything else is a programming error).
//   - The helm uninstall runs FIRST. If it returns the tolerated
//     "failed to delete release: mongodb" error, we treat it as
//     success and continue to the Secret cleanup. Any other error
//     is fatal.
//   - The Helm release Secret cleanup Get+Deletes
//     "sh.helm.release.v1.mongodb.v1" in cfg.Namespace. NotFound
//     on Get is success (the Secret is either gone already or was
//     never created); NotFound on Delete is likewise tolerated
//     (racy delete from a parallel uninstall path).
//   - Any other Get/Delete error propagates.
func (o *Ops) UninstallLegacyMongo(ctx context.Context, cfg any) error {
	mc, ok := cfg.(*middlewareinstaller.MiddlewareConfig)
	if !ok || mc == nil {
		return fmt.Errorf("mwops: UninstallLegacyMongo expected non-nil *middlewareinstaller.MiddlewareConfig, got %T", cfg)
	}

	if err := o.mongoUninstall(ctx, mc); err != nil {
		if err.Error() != mongoUninstallToleratedErr {
			return err
		}
		// Tolerated error — fall through to Secret cleanup.
	}

	var secret corev1.Secret
	if err := o.c.Get(ctx, types.NamespacedName{
		Name:      mongoReleaseSecretName,
		Namespace: mc.Namespace,
	}, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := o.c.Delete(ctx, &secret); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// defaultKBFactory is the production KubeBlocksOpsFactory. Wraps
// kubeblocks.NewOperation — returns *OperationOptions which
// structurally implements KubeBlocksOps via its Start/Stop methods.
//
// Kept as a top-level func (rather than an anonymous closure inside
// New) so a signature drift on kubeblocks.NewOperation breaks the
// build here rather than at the first Start/Stop call site.
func defaultKBFactory(ctx context.Context, opsType kbopv1alpha1.OpsType, am *appv1.ApplicationManager, c client.Client) KubeBlocksOps {
	return kubeblocks.NewOperation(ctx, opsType, am, c)
}

// Interface conformance guards. Compile-time check that:
//   - Ops satisfies ports.MiddlewareOps.
//   - *kubeblocks.OperationOptions satisfies KubeBlocksOps (would
//     otherwise fail silently at defaultKBFactory invocation).
var (
	_ ports.MiddlewareOps = (*Ops)(nil)
	_ KubeBlocksOps       = (*kubeblocks.OperationOptions)(nil)
)
