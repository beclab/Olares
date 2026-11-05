package fake

import (
	"context"
	"sync"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// ImagePuller is an in-memory ports.ImagePuller. Progress reads can
// be programmed per-AM via SetProgress; every mutating call is
// recorded in Calls().
type ImagePuller struct {
	mu       sync.Mutex
	progress map[string]ports.ImageProgress
	calls    []ImageCall

	// DefaultProgress is reported for any AM that SetProgress has
	// not covered. NewImagePuller defaults it to "completed" so the
	// Runner's poll loop terminates in tests that only exercise the
	// state machine; set it to the zero value to model an AM whose
	// ImageManager CR does not exist yet.
	DefaultProgress ports.ImageProgress

	ResolveRefsFunc     func(ctx context.Context, am *appv1.ApplicationManager, appCfg any) ([]appv1.Ref, error)
	CreateJobFunc       func(ctx context.Context, am *appv1.ApplicationManager, refs []appv1.Ref) error
	CancelJobFunc       func(ctx context.Context, amName string) error
	UpdateJobStatusFunc func(ctx context.Context, amName, state, message string) error
}

// ImageCall is one recorded write-side invocation. Method is the
// port method name; Name is the AM name when relevant.
type ImageCall struct {
	Method  string
	Name    string
	State   string // UpdateJobStatus only
	Message string // UpdateJobStatus only
	Refs    []appv1.Ref
}

// NewImagePuller returns an ImagePuller whose pulls complete on the
// first Progress read. See DefaultProgress.
func NewImagePuller() *ImagePuller {
	return &ImagePuller{
		progress:        make(map[string]ports.ImageProgress),
		DefaultProgress: ports.ImageProgress{Percent: 100, Complete: true},
	}
}

// SetProgress programs the value Progress(amName) will return.
// Handlers that make progress-driven decisions can be walked
// through a series of `SetProgress → step reconcile → assert` cycles
// this way.
func (p *ImagePuller) SetProgress(amName string, prog ports.ImageProgress) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.progress == nil {
		p.progress = make(map[string]ports.ImageProgress)
	}
	p.progress[amName] = prog
}

// Calls returns a copy of the recorded call log.
func (p *ImagePuller) Calls() []ImageCall {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]ImageCall, len(p.calls))
	copy(out, p.calls)
	return out
}

func (p *ImagePuller) recordCall(c ImageCall) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, c)
}

func (p *ImagePuller) ResolveRefs(ctx context.Context, am *appv1.ApplicationManager, appCfg any) ([]appv1.Ref, error) {
	if p.ResolveRefsFunc != nil {
		return p.ResolveRefsFunc(ctx, am, appCfg)
	}
	return nil, nil
}

func (p *ImagePuller) CreateJob(ctx context.Context, am *appv1.ApplicationManager, refs []appv1.Ref) error {
	p.recordCall(ImageCall{Method: "CreateJob", Name: am.Name, Refs: refs})
	if p.CreateJobFunc != nil {
		return p.CreateJobFunc(ctx, am, refs)
	}
	return nil
}

func (p *ImagePuller) CancelJob(ctx context.Context, amName string) error {
	p.recordCall(ImageCall{Method: "CancelJob", Name: amName})
	if p.CancelJobFunc != nil {
		return p.CancelJobFunc(ctx, amName)
	}
	return nil
}

func (p *ImagePuller) UpdateJobStatus(ctx context.Context, amName, state, message string) error {
	p.recordCall(ImageCall{Method: "UpdateJobStatus", Name: amName, State: state, Message: message})
	if p.UpdateJobStatusFunc != nil {
		return p.UpdateJobStatusFunc(ctx, amName, state, message)
	}
	return nil
}

// Progress returns the value programmed by SetProgress for amName.
// An AM with nothing programmed reports DefaultProgress, so a test
// that does not care about the download phase can let pulls complete
// instead of hanging the Runner's poll loop.
func (p *ImagePuller) Progress(_ context.Context, amName string) (ports.ImageProgress, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if prog, ok := p.progress[amName]; ok {
		return prog, nil
	}
	return p.DefaultProgress, nil
}

// Interface conformance guard.
var _ ports.ImagePuller = (*ImagePuller)(nil)
