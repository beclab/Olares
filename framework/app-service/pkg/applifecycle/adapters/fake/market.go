package fake

import (
	"context"
	"sync"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// MarketChartRepo is an in-memory ports.MarketChartRepo. Programmed
// via two maps keyed by request Name:
//
//   - Configs[name] = pre-loaded config any (e.g.
//     *appcfg.ApplicationConfig) returned by GetAppConfig
//   - Charts[name+"@"+version] = pre-loaded local path string
//     returned by FetchChart
//
// Missing entries default to (nil, nil) for GetAppConfig and
// ("", nil) for FetchChart, matching the "no chart cached" case.
type MarketChartRepo struct {
	mu sync.Mutex

	Configs map[string]any
	Charts  map[string]string

	GetAppConfigFunc func(ctx context.Context, req ports.MarketFetchRequest) (any, error)
	FetchChartFunc   func(ctx context.Context, req ports.MarketFetchRequest) (string, error)

	calls []MarketCall
}

// MarketCall is one recorded invocation. Method is the port method
// name; Name / Version populated per request.
type MarketCall struct {
	Method  string
	Name    string
	Version string
}

// Calls returns a copy of the recorded call log.
func (m *MarketChartRepo) Calls() []MarketCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]MarketCall, len(m.calls))
	copy(out, m.calls)
	return out
}

func (m *MarketChartRepo) recordCall(c MarketCall) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, c)
}

func (m *MarketChartRepo) GetAppConfig(ctx context.Context, req ports.MarketFetchRequest) (any, error) {
	m.recordCall(MarketCall{Method: "GetAppConfig", Name: req.Name, Version: req.Version})
	if m.GetAppConfigFunc != nil {
		return m.GetAppConfigFunc(ctx, req)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Configs[req.Name], nil
}

func (m *MarketChartRepo) FetchChart(ctx context.Context, req ports.MarketFetchRequest) (string, error) {
	m.recordCall(MarketCall{Method: "FetchChart", Name: req.Name, Version: req.Version})
	if m.FetchChartFunc != nil {
		return m.FetchChartFunc(ctx, req)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Charts[req.Name+"@"+req.Version], nil
}

// Interface conformance guard.
var _ ports.MarketChartRepo = (*MarketChartRepo)(nil)
