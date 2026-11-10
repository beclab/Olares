package fake

import (
	"context"
	"sync"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// MiddlewareOps is an in-memory ports.MiddlewareOps. Default:
// success on every method. Override per method via ...Func hooks.
type MiddlewareOps struct {
	mu    sync.Mutex
	calls []MiddlewareCall

	StartFunc                func(ctx context.Context, am *appv1.ApplicationManager) error
	StopFunc                 func(ctx context.Context, am *appv1.ApplicationManager) error
	UninstallLegacyMongoFunc func(ctx context.Context, cfg any) error
}

// MiddlewareCall is one recorded invocation. Method is the port
// method name; Name is the AM name (or "" for UninstallLegacyMongo
// where the caller passes a raw config).
type MiddlewareCall struct {
	Method string
	Name   string
	Cfg    any
}

// Calls returns a copy of the recorded call log.
func (m *MiddlewareOps) Calls() []MiddlewareCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]MiddlewareCall, len(m.calls))
	copy(out, m.calls)
	return out
}

func (m *MiddlewareOps) recordCall(c MiddlewareCall) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, c)
}

func (m *MiddlewareOps) Start(ctx context.Context, am *appv1.ApplicationManager) error {
	m.recordCall(MiddlewareCall{Method: "Start", Name: am.Name})
	if m.StartFunc != nil {
		return m.StartFunc(ctx, am)
	}
	return nil
}

func (m *MiddlewareOps) Stop(ctx context.Context, am *appv1.ApplicationManager) error {
	m.recordCall(MiddlewareCall{Method: "Stop", Name: am.Name})
	if m.StopFunc != nil {
		return m.StopFunc(ctx, am)
	}
	return nil
}

func (m *MiddlewareOps) UninstallLegacyMongo(ctx context.Context, cfg any) error {
	m.recordCall(MiddlewareCall{Method: "UninstallLegacyMongo", Cfg: cfg})
	if m.UninstallLegacyMongoFunc != nil {
		return m.UninstallLegacyMongoFunc(ctx, cfg)
	}
	return nil
}

// Interface conformance guard.
var _ ports.MiddlewareOps = (*MiddlewareOps)(nil)
