package computevalidator

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
)

// ChainForFunc is the seam that maps a port-level ValidationOp to a
// concrete pkg/compute/validation.Validator chain. The default
// production mapping is defaultChainFor (below); tests substitute
// via WithChainFor.
type ChainForFunc func(op ports.ValidationOp) []validation.Validator

// NotifierFunc is the "publish this payload to this subject" seam
// NotifyInsufficientCompute delegates to. The signature matches
// natsrecorder.Recorder.RecordNotification so main.go can pass
// that method directly.
type NotifierFunc func(ctx context.Context, subject string, payload any) error

// Validator is the production ports.InstallValidator adapter.
type Validator struct {
	c        client.Client
	chainFor ChainForFunc
	notify   NotifierFunc
}

// Option customises a Validator at construction time.
type Option func(*Validator)

// WithChainFor overrides the ValidationOp → validator chain
// mapping. Passing nil is a no-op (the default mapping is kept)
// so tests that only need to override the notifier don't have to
// pass a chainFor stub.
func WithChainFor(fn ChainForFunc) Option {
	return func(v *Validator) {
		if fn != nil {
			v.chainFor = fn
		}
	}
}

// New constructs a Validator. Both `c` and `notify` are required —
// nil client would NPE inside the validators (they call c.List /
// c.Get); nil notify would silently drop compute-insufficient
// notifications which is worse than a startup panic.
func New(c client.Client, notify NotifierFunc, opts ...Option) *Validator {
	if c == nil {
		panic("computevalidator.New: client is nil")
	}
	if notify == nil {
		panic("computevalidator.New: notify is nil")
	}
	v := &Validator{
		c:        c,
		notify:   notify,
		chainFor: defaultChainFor,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// ValidateInstall implements ports.InstallValidator. See doc.go
// "Pure-mapping methods" for why ValidatePostHelm delegates here.
func (v *Validator) ValidateInstall(ctx context.Context, in ports.ValidationInput) (ports.ValidationDecision, error) {
	cfg, err := castAppCfg(in.AppCfg)
	if err != nil {
		return ports.ValidationDecision{}, err
	}
	legacyOp, err := toLegacyOp(in.Op)
	if err != nil {
		return ports.ValidationDecision{}, err
	}
	chain := v.chainFor(in.Op)
	if len(chain) == 0 {
		// Explicit short-circuit: an empty chain means "no
		// validators apply to this op" (e.g. upgrade — legacy
		// does not validate). Skipping validation.Run is
		// mandatory here because Run falls back to
		// DefaultValidators() on empty input, which would run
		// the FULL install chain against an upgrade AM and reject
		// it with a spurious pressure error. See doc.go
		// "ValidationOp → legacy Op mapping" for the intent.
		return ports.ValidationDecision{OK: true}, nil
	}
	d, err := validation.Run(ctx, validation.Input{
		Client:    v.c,
		AppConfig: cfg,
		Op:        legacyOp,
		Token:     in.Token,
	}, chain...)
	return toPortDecision(d), err
}

// ValidatePostHelm is the mirror method for the two-phase-install
// post-helm hook. The chain is the same as ValidateInstall — see
// doc.go for the design note.
func (v *Validator) ValidatePostHelm(ctx context.Context, in ports.ValidationInput) (ports.ValidationDecision, error) {
	return v.ValidateInstall(ctx, in)
}

// NotifyInsufficientCompute implements ports.InstallValidator.
// Uses pkg/compute.BuildInsufficientPayload so the subject +
// payload shape is IDENTICAL to legacy
// compute.PublishComputeInsufficientNotification (guaranteed at
// compile time — both call the same helper).
func (v *Validator) NotifyInsufficientCompute(ctx context.Context, appCfg any, reason error) error {
	cfg, err := castAppCfg(appCfg)
	if err != nil {
		return err
	}
	subject, payload := compute.BuildInsufficientPayload(cfg, reason)
	if payload == nil {
		// BuildInsufficientPayload returns nil payload only when
		// cfg is nil, which our cast already rejects — defensive
		// re-check kept for the "compute changes its contract"
		// future.
		return fmt.Errorf("computevalidator: NotifyInsufficientCompute got nil payload for cfg %+v", cfg)
	}
	return v.notify(ctx, subject, payload)
}

// castAppCfg is the strict `any → *appcfg.ApplicationConfig` cast
// every entry point performs. Same shape as the imgpull adapter's
// cast — keeps the failure mode uniform across the ports layer.
func castAppCfg(appCfg any) (*appcfg.ApplicationConfig, error) {
	cfg, ok := appCfg.(*appcfg.ApplicationConfig)
	if !ok || cfg == nil {
		return nil, fmt.Errorf("computevalidator: expected non-nil *appcfg.ApplicationConfig, got %T", appCfg)
	}
	return cfg, nil
}

// toLegacyOp maps the port ValidationOp string to the legacy
// v1alpha1.OpType. Rejecting unknown ops here (rather than passing
// through) means a typo like "instal" fails at the port boundary
// with a clear error instead of running an empty validator chain
// that reports OK (which would silently skip validation).
func toLegacyOp(op ports.ValidationOp) (appv1.OpType, error) {
	switch op {
	case ports.ValidateOpInstall:
		return appv1.InstallOp, nil
	case ports.ValidateOpUpgrade:
		return appv1.UpgradeOp, nil
	case ports.ValidateOpResume:
		return appv1.ResumeOp, nil
	default:
		return "", fmt.Errorf("computevalidator: unknown ValidationOp %q", op)
	}
}

// defaultChainFor is the production ValidationOp → validator chain
// mapping. See doc.go "ValidationOp → legacy Op mapping" for the
// full rationale.
func defaultChainFor(op ports.ValidationOp) []validation.Validator {
	switch op {
	case ports.ValidateOpInstall:
		return validation.InstallRuntimePressureValidators()
	case ports.ValidateOpResume:
		return validation.ResumePressureValidators()
	case ports.ValidateOpUpgrade:
		// Legacy upgrade does not use validation.Run. Returning
		// an empty chain makes Run short-circuit to OK, so a
		// caller who mistakenly asks "validate the upgrade" gets
		// a benign pass rather than a mismatch error (already
		// gated by toLegacyOp for unknown ops).
		return nil
	default:
		return nil
	}
}

// toPortDecision maps the internal validation.Decision to the
// port-level ports.ValidationDecision. String conversion is
// necessary because Decision.Resource / Reason are typed enums
// (constants.ResourceType / ResourceConditionType) inside the
// validation package but the port uses plain strings — the port
// stays free of the constants package's transitive imports.
func toPortDecision(d validation.Decision) ports.ValidationDecision {
	return ports.ValidationDecision{
		OK:        d.OK,
		Resource:  string(d.Resource),
		Reason:    string(d.Reason),
		Message:   d.Message,
		Validator: d.Validator,
	}
}

// Interface conformance guard.
var _ ports.InstallValidator = (*Validator)(nil)
