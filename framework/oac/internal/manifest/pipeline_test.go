package manifest

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

// fakeManifest is a minimal Manifest implementation for strategy tests.
type fakeManifest struct {
	apiVersion    string
	configVersion string
	raw           any
}

func (f *fakeManifest) APIVersion() string            { return f.apiVersion }
func (f *fakeManifest) ConfigVersion() string         { return f.configVersion }
func (f *fakeManifest) ConfigType() string            { return "app" }
func (f *fakeManifest) AppName() string               { return "x" }
func (f *fakeManifest) AppVersion() string            { return "0.0.0" }
func (f *fakeManifest) Entrances() []EntranceInfo     { return nil }
func (f *fakeManifest) OptionsImages() []string       { return nil }
func (f *fakeManifest) PermissionAppData() bool       { return false }
func (f *fakeManifest) Raw() any                      { return f.raw }

// fakeStrategy records Parse/Validate invocations so tests can assert what
// the pipeline actually called. Validate returns valErr; Parse returns
// (mfst, parseErr).
type fakeStrategy struct {
	mu         sync.Mutex
	parseCalls [][]byte
	validCalls []Manifest
	parseErr   error
	valErr     error
	manifest   Manifest
}

func (f *fakeStrategy) Parse(raw []byte) (Manifest, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.parseCalls = append(f.parseCalls, append([]byte(nil), raw...))
	if f.parseErr != nil {
		return nil, f.parseErr
	}
	if f.manifest != nil {
		return f.manifest, nil
	}
	return &fakeManifest{apiVersion: "v1"}, nil
}

func (f *fakeStrategy) Validate(m Manifest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.validCalls = append(f.validCalls, m)
	return f.valErr
}

func TestIsLegacyVersion(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"not-a-version", false},
		{"0.11.999", true},
		{"0.12.0", false},  // boundary: inclusive on modern side
		{"0.12.1", false},
		{"0.13.0", false},
		{"1.0.0", false},
		{"0.0.1", true},
	}
	for _, tc := range cases {
		if got := isLegacyVersion(tc.in); got != tc.want {
			t.Errorf("isLegacyVersion(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestNewPipeline_Dispatch(t *testing.T) {
	strat := &fakeStrategy{}

	cases := []struct {
		name       string
		olaresVer  string
		wantLegacy bool
	}{
		{"legacy 0.8.1", "0.8.1", true},
		{"legacy 0.11.999", "0.11.999", true},
		{"modern 0.12.0 (boundary)", "0.12.0", false},
		{"modern 0.12.1", "0.12.1", false},
		{"modern 1.0.0", "1.0.0", false},
		{"empty treated as modern", "", false},
		{"malformed treated as modern", "not-a-version", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := NewPipeline(tc.olaresVer, strat)
			if tc.wantLegacy {
				if _, ok := got.(dualOwnerPipeline); !ok {
					t.Fatalf("expected dualOwnerPipeline, got %T", got)
				}
			} else {
				if _, ok := got.(singlePipeline); !ok {
					t.Fatalf("expected singlePipeline, got %T", got)
				}
			}
			if got.Strategy() != strat {
				t.Fatalf("Pipeline must retain the passed-in Strategy")
			}
		})
	}
}

func TestSinglePipeline_IgnoresRenderer(t *testing.T) {
	s := &fakeStrategy{}
	p := singlePipeline{strat: s}

	raw := []byte("hello")
	rendererCalled := false
	render := func(_ []byte, _, _ string) ([]byte, error) {
		rendererCalled = true
		return []byte("should-not-be-used"), nil
	}

	if _, err := p.Parse(raw, render, "user", "admin"); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if rendererCalled {
		t.Fatal("singlePipeline.Parse must not invoke the renderer")
	}
	if len(s.parseCalls) != 1 || string(s.parseCalls[0]) != "hello" {
		t.Fatalf("Parse should forward raw verbatim, got %q", s.parseCalls)
	}

	if err := p.Validate(raw, nil, render, "user", "admin"); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if rendererCalled {
		t.Fatal("singlePipeline.Validate must not invoke the renderer")
	}
	if len(s.validCalls) != 1 {
		t.Fatalf("Validate should call Strategy.Validate exactly once, got %d", len(s.validCalls))
	}
}

// TestSinglePipeline_ValidateReusesParsedManifest is the regression test for
// P1: when Validate is handed the Manifest produced by a prior Parse, it
// must NOT ask the strategy to parse the raw bytes a second time. Before
// this optimization, Lint/ValidateManifestContent parsed the modern
// manifest twice end-to-end.
func TestSinglePipeline_ValidateReusesParsedManifest(t *testing.T) {
	s := &fakeStrategy{}
	p := singlePipeline{strat: s}

	parsed := &fakeManifest{apiVersion: "v1"}
	if err := p.Validate([]byte("raw"), parsed, nil, "", ""); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(s.parseCalls) != 0 {
		t.Fatalf("expected zero Strategy.Parse calls when a parsed manifest is supplied, got %d", len(s.parseCalls))
	}
	if len(s.validCalls) != 1 || s.validCalls[0] != parsed {
		t.Fatalf("Strategy.Validate must receive the supplied parsed manifest, got %+v", s.validCalls)
	}
}

// TestSinglePipeline_ValidateFallsBackToParse keeps the old behaviour alive
// for callers that don't (yet) have a pre-parsed Manifest handy.
func TestSinglePipeline_ValidateFallsBackToParse(t *testing.T) {
	s := &fakeStrategy{}
	p := singlePipeline{strat: s}

	if err := p.Validate([]byte("hello"), nil, nil, "", ""); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(s.parseCalls) != 1 || string(s.parseCalls[0]) != "hello" {
		t.Fatalf("expected Strategy.Parse on the raw bytes, got %q", s.parseCalls)
	}
	if len(s.validCalls) != 1 {
		t.Fatalf("Strategy.Validate should run once, got %d", len(s.validCalls))
	}
}

func TestSinglePipeline_Strategy(t *testing.T) {
	s := &fakeStrategy{}
	p := singlePipeline{strat: s}
	if p.Strategy() != s {
		t.Fatal("Strategy() must return the wrapped strategy")
	}
}

func TestDualOwnerPipeline_ParseUsesCallerInputs(t *testing.T) {
	s := &fakeStrategy{manifest: &fakeManifest{apiVersion: "v1", configVersion: "0.8.1"}}
	p := dualOwnerPipeline{strat: s}

	var capturedOwner, capturedAdmin string
	render := func(raw []byte, owner, admin string) ([]byte, error) {
		capturedOwner, capturedAdmin = owner, admin
		return []byte("rendered:" + string(raw)), nil
	}

	m, err := p.Parse([]byte("payload"), render, "alice", "root")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if m == nil {
		t.Fatal("expected a Manifest")
	}
	if capturedOwner != "alice" || capturedAdmin != "root" {
		t.Fatalf("Parse forwarded (owner=%q, admin=%q), want (alice, root)", capturedOwner, capturedAdmin)
	}
	if len(s.parseCalls) != 1 || string(s.parseCalls[0]) != "rendered:payload" {
		t.Fatalf("Parse should feed rendered bytes into strategy, got %q", s.parseCalls)
	}
}

func TestDualOwnerPipeline_ParseNilRenderer(t *testing.T) {
	p := dualOwnerPipeline{strat: &fakeStrategy{}}
	_, err := p.Parse([]byte("x"), nil, "", "")
	if !errors.Is(err, ErrNilRenderer) {
		t.Fatalf("expected ErrNilRenderer, got %v", err)
	}
}

func TestDualOwnerPipeline_ParseRendererError(t *testing.T) {
	p := dualOwnerPipeline{strat: &fakeStrategy{}}
	boom := errors.New("boom")
	_, err := p.Parse([]byte("x"), func([]byte, string, string) ([]byte, error) {
		return nil, boom
	}, "", "")
	if !errors.Is(err, boom) {
		t.Fatalf("expected error to wrap boom, got %v", err)
	}
}

func TestDualOwnerPipeline_ValidateRunsBothScenarios(t *testing.T) {
	s := &fakeStrategy{manifest: &fakeManifest{apiVersion: "v1"}}
	p := dualOwnerPipeline{strat: s}

	type call struct{ owner, admin string }
	var calls []call
	render := func(_ []byte, owner, admin string) ([]byte, error) {
		calls = append(calls, call{owner, admin})
		return []byte("ok"), nil
	}

	if err := p.Validate([]byte("raw"), nil, render, "alice", "root"); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected renderer to be called twice, got %d", len(calls))
	}
	first, second := calls[0], calls[1]
	if first.owner != "root" || first.admin != "root" {
		t.Fatalf("first scenario must be admin=owner=admin, got %+v", first)
	}
	if second.owner != "alice" || second.admin != "root" {
		t.Fatalf("second scenario must be admin!=owner with caller values, got %+v", second)
	}

	if len(s.parseCalls) != 2 {
		t.Fatalf("Strategy.Parse should run twice, got %d", len(s.parseCalls))
	}
	if len(s.validCalls) != 2 {
		t.Fatalf("Strategy.Validate should run twice, got %d", len(s.validCalls))
	}
}

func TestDualOwnerPipeline_ValidateFallbackDefaults(t *testing.T) {
	s := &fakeStrategy{manifest: &fakeManifest{apiVersion: "v1"}}
	p := dualOwnerPipeline{strat: s}

	var calls [][2]string
	render := func(_ []byte, owner, admin string) ([]byte, error) {
		calls = append(calls, [2]string{owner, admin})
		return []byte("ok"), nil
	}
	if err := p.Validate([]byte("raw"), nil, render, "", ""); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 render calls, got %d", len(calls))
	}
	if calls[0][0] == calls[1][0] && calls[0][1] == calls[1][1] {
		t.Fatalf("scenarios must differ; both rendered with %+v", calls[0])
	}
	if calls[0][0] != calls[0][1] {
		t.Fatalf("scenario A should be owner==admin, got %+v", calls[0])
	}
	if calls[1][0] == calls[1][1] {
		t.Fatalf("scenario B should be owner!=admin, got %+v", calls[1])
	}
}

func TestDualOwnerPipeline_ValidateAggregatesErrors(t *testing.T) {
	boom := errors.New("boom")
	s := &fakeStrategy{
		manifest: &fakeManifest{apiVersion: "v1"},
		valErr:   boom,
	}
	p := dualOwnerPipeline{strat: s}
	render := func(_ []byte, _, _ string) ([]byte, error) {
		return []byte("ok"), nil
	}
	err := p.Validate([]byte("raw"), nil, render, "alice", "root")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "admin=owner") || !strings.Contains(msg, "admin!=owner") {
		t.Fatalf("expected both scenario labels in error, got: %s", msg)
	}
}

func TestDualOwnerPipeline_ValidateContinuesPastParseError(t *testing.T) {
	s := &fakeStrategy{manifest: &fakeManifest{apiVersion: "v1"}}
	p := dualOwnerPipeline{strat: s}

	n := 0
	render := func(_ []byte, _, _ string) ([]byte, error) {
		n++
		if n == 1 {
			return nil, errors.New("render-A-failed")
		}
		return []byte("ok"), nil
	}
	err := p.Validate([]byte("raw"), nil, render, "alice", "root")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "render-A-failed") {
		t.Fatalf("expected scenario A render error, got: %v", err)
	}
	if len(s.validCalls) != 1 {
		t.Fatalf("expected scenario B to run (1 Validate call), got %d", len(s.validCalls))
	}
}

func TestDualOwnerPipeline_ValidateNilRenderer(t *testing.T) {
	p := dualOwnerPipeline{strat: &fakeStrategy{}}
	err := p.Validate([]byte("x"), nil, nil, "", "")
	if !errors.Is(err, ErrNilRenderer) {
		t.Fatalf("expected ErrNilRenderer, got %v", err)
	}
}

// TestDualOwnerPipeline_ValidateIgnoresParsedManifest documents that the
// legacy pipeline cannot reuse the caller's parsed manifest: Validate runs
// two scenarios with synthesized (owner, admin) values of its own, so both
// scenarios must still go through render → Parse → Validate from raw.
func TestDualOwnerPipeline_ValidateIgnoresParsedManifest(t *testing.T) {
	s := &fakeStrategy{manifest: &fakeManifest{apiVersion: "v1"}}
	p := dualOwnerPipeline{strat: s}

	render := func(_ []byte, _, _ string) ([]byte, error) {
		return []byte("ok"), nil
	}
	parsed := &fakeManifest{apiVersion: "v1"}
	if err := p.Validate([]byte("raw"), parsed, render, "alice", "root"); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(s.parseCalls) != 2 {
		t.Fatalf("legacy pipeline must still Parse once per scenario, got %d", len(s.parseCalls))
	}
}
