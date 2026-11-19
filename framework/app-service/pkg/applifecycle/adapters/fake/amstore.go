package fake

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// AMStore is an in-memory ports.AMStore that models the two
// invariants a real cluster provides:
//
//  1. **NotFound** — reads on a missing AM/Application return an
//     apierrors.NewNotFound error so caller code that does
//     `apierrors.IsNotFound(err)` works identically against the fake
//     and the real store.
//  2. **Optimistic concurrency** — every mutation is gated on the
//     caller-supplied `curResourceVersion`. A mismatch returns
//     apierrors.NewConflict, exactly what a real API server returns
//     when a stale ResourceVersion is used.
//
// Zero value is legal. All methods are safe for concurrent use.
type AMStore struct {
	mu sync.Mutex

	ams  map[string]*appv1.ApplicationManager // key = am.Name
	apps map[string]*appv1.Application        // key = app.Name

	// nextRV monotonically increases across ALL mutations to any
	// object in the store. Matches K8s behaviour (RV is global, not
	// per-object) which keeps `List --resource-version=X` sane.
	nextRV int64

	// mutations logs every write in call order. Tests use it to
	// assert both what changed and in what order.
	mutations []Mutation
}

// Mutation is a single write recorded on AMStore. Kind is one of
// "PatchStatus", "PatchSpecAndAnnotations", "UpdateFull", "AddAM",
// "AddApp". Fields other than Kind + Name are opportunistic — inspect
// them per kind.
type Mutation struct {
	Kind string
	Name string

	// StatusPatch is populated for Kind=="PatchStatus".
	StatusPatch *ports.StatusPatch

	// SpecPatch / AnnotationPatch are populated for Kind==
	// "PatchSpecAndAnnotations".
	SpecPatch       *ports.SpecPatch
	AnnotationPatch *ports.AnnotationPatch

	// Snapshot is a deep copy of the AM AFTER the mutation. nil
	// for pre-load Adds.
	Snapshot *appv1.ApplicationManager
}

// NewAMStore returns an empty store. Use Add / AddApp to seed test
// state.
func NewAMStore() *AMStore {
	return &AMStore{
		ams:  make(map[string]*appv1.ApplicationManager),
		apps: make(map[string]*appv1.Application),
	}
}

// Add pre-loads an ApplicationManager into the store. Assigns the
// object a fresh ResourceVersion (overwriting any existing one) so
// tests don't have to compute it. Returns the RV so tests that need
// to Patch immediately can capture it.
func (s *AMStore) Add(am *appv1.ApplicationManager) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextRV++
	am = am.DeepCopy()
	am.ResourceVersion = strconv.FormatInt(s.nextRV, 10)
	s.ams[am.Name] = am
	s.mutations = append(s.mutations, Mutation{Kind: "AddAM", Name: am.Name, Snapshot: am.DeepCopy()})
	return am.ResourceVersion
}

// Delete removes an ApplicationManager so a later Get returns NotFound.
// Tests use it to simulate a delete that arrives before Reconcile.
func (s *AMStore) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.ams, name)
}

// AddApp pre-loads an Application into the store. Same RV assignment
// semantics as Add.
func (s *AMStore) AddApp(app *appv1.Application) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextRV++
	app = app.DeepCopy()
	app.ResourceVersion = strconv.FormatInt(s.nextRV, 10)
	s.apps[app.Name] = app
	s.mutations = append(s.mutations, Mutation{Kind: "AddApp", Name: app.Name})
	return app.ResourceVersion
}

// Get implements ports.AMStore.
func (s *AMStore) Get(_ context.Context, name string) (*appv1.ApplicationManager, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	am, ok := s.ams[name]
	if !ok {
		return nil, newNotFoundAM(name)
	}
	return am.DeepCopy(), nil
}

// List implements ports.AMStore.
func (s *AMStore) List(_ context.Context) ([]appv1.ApplicationManager, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]appv1.ApplicationManager, 0, len(s.ams))
	for _, am := range s.ams {
		out = append(out, *am.DeepCopy())
	}
	return out, nil
}

// GetApplication implements ports.AMStore.
func (s *AMStore) GetApplication(_ context.Context, name string) (*appv1.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	app, ok := s.apps[name]
	if !ok {
		return nil, newNotFoundApp(name)
	}
	return app.DeepCopy(), nil
}

// PatchStatus implements ports.AMStore. Applies the patch under
// optimistic concurrency: returns Conflict if curResourceVersion !=
// stored RV; returns NotFound if the AM does not exist.
func (s *AMStore) PatchStatus(_ context.Context, name, curRV string, patch ports.StatusPatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	am, ok := s.ams[name]
	if !ok {
		return newNotFoundAM(name)
	}
	if am.ResourceVersion != curRV {
		return newConflictAM(name, am.ResourceVersion, curRV)
	}
	applyStatusPatch(&am.Status, patch)
	s.nextRV++
	am.ResourceVersion = strconv.FormatInt(s.nextRV, 10)
	s.mutations = append(s.mutations, Mutation{
		Kind: "PatchStatus", Name: name, StatusPatch: &patch, Snapshot: am.DeepCopy(),
	})
	return nil
}

// PatchSpecAndAnnotations implements ports.AMStore.
func (s *AMStore) PatchSpecAndAnnotations(_ context.Context, name, curRV string,
	spec *ports.SpecPatch, ann *ports.AnnotationPatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	am, ok := s.ams[name]
	if !ok {
		return newNotFoundAM(name)
	}
	if am.ResourceVersion != curRV {
		return newConflictAM(name, am.ResourceVersion, curRV)
	}
	if spec != nil {
		if spec.Config != "" {
			am.Spec.Config = spec.Config
		}
		if spec.OpType != "" {
			am.Spec.OpType = spec.OpType
		}
	}
	if ann != nil {
		if am.Annotations == nil && (len(ann.Set) > 0) {
			am.Annotations = make(map[string]string, len(ann.Set))
		}
		for k, v := range ann.Set {
			am.Annotations[k] = v
		}
		for _, k := range ann.Delete {
			delete(am.Annotations, k)
		}
	}
	s.nextRV++
	am.ResourceVersion = strconv.FormatInt(s.nextRV, 10)
	s.mutations = append(s.mutations, Mutation{
		Kind: "PatchSpecAndAnnotations", Name: name, SpecPatch: spec, AnnotationPatch: ann, Snapshot: am.DeepCopy(),
	})
	return nil
}

// UpdateFull implements ports.AMStore. Requires am.ResourceVersion
// to match the stored value; overwrites the entire object.
func (s *AMStore) UpdateFull(_ context.Context, am *appv1.ApplicationManager) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.ams[am.Name]
	if !ok {
		return newNotFoundAM(am.Name)
	}
	if cur.ResourceVersion != am.ResourceVersion {
		return newConflictAM(am.Name, cur.ResourceVersion, am.ResourceVersion)
	}
	replaced := am.DeepCopy()
	s.nextRV++
	replaced.ResourceVersion = strconv.FormatInt(s.nextRV, 10)
	s.ams[am.Name] = replaced
	s.mutations = append(s.mutations, Mutation{
		Kind: "UpdateFull", Name: am.Name, Snapshot: replaced.DeepCopy(),
	})
	return nil
}

// Mutations returns the write log in call order. Copy so callers
// can range without holding the store's mutex.
func (s *AMStore) Mutations() []Mutation {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Mutation, len(s.mutations))
	copy(out, s.mutations)
	return out
}

// applyStatusPatch merges the ports.StatusPatch fields into status.
// Empty scalar fields are preserved; nil pointer / nil map fields
// are preserved; only explicitly-set fields are overwritten. Matches
// the "field-level merge, not wholesale replace" semantic documented
// on ports.StatusPatch.
func applyStatusPatch(status *appv1.ApplicationManagerStatus, patch ports.StatusPatch) {
	if patch.State != "" {
		status.State = patch.State
	}
	if patch.Message != "" {
		status.Message = patch.Message
	}
	if patch.Reason != "" {
		status.Reason = patch.Reason
	}
	if patch.OpType != "" {
		status.OpType = patch.OpType
	}
	if patch.OpID != "" {
		status.OpID = patch.OpID
	}
	if patch.Progress != "" {
		status.Progress = patch.Progress
	}
	if patch.Payload != nil {
		// Contract (ports/amstore.go StatusPatch.Payload):
		//   * nil            → leave the current map untouched
		//   * empty non-nil  → clear the entire map (a handler
		//                      leaving a state uses this to drop
		//                      that state's phase key)
		//   * non-empty      → merge key-by-key
		//
		// The clear semantic matters because state-leaving handlers
		// (installing.finishInstall, downloading.handoffToInstalling)
		// rely on it to prevent stale phase Payload from bleeding
		// into the next state's decisions.
		if len(patch.Payload) == 0 {
			status.Payload = nil
		} else {
			if status.Payload == nil {
				status.Payload = make(map[string]string, len(patch.Payload))
			}
			for k, v := range patch.Payload {
				status.Payload[k] = v
			}
		}
	}
	for _, key := range patch.DeletePayloadKeys {
		delete(status.Payload, key)
	}
	if patch.OpRecord != nil {
		status.OpRecords = append([]appv1.OpRecord{*patch.OpRecord}, status.OpRecords...)
	}
	if patch.StatusTime != nil {
		status.StatusTime = patch.StatusTime
	}
	if patch.UpdateTime != nil {
		status.UpdateTime = patch.UpdateTime
	}
	_ = metav1.Now // keep the import live even if the type reshuffles
}

func newNotFoundAM(name string) error {
	return apierrors.NewNotFound(schema.GroupResource{
		Group: appv1.GroupVersion.Group, Resource: "applicationmanagers",
	}, name)
}

func newNotFoundApp(name string) error {
	return apierrors.NewNotFound(schema.GroupResource{
		Group: appv1.GroupVersion.Group, Resource: "applications",
	}, name)
}

func newConflictAM(name, stored, given string) error {
	err := apierrors.NewConflict(schema.GroupResource{
		Group: appv1.GroupVersion.Group, Resource: "applicationmanagers",
	}, name, fmt.Errorf("resource version conflict: stored=%s given=%s", stored, given))
	return err
}

// Interface conformance guard.
var _ ports.AMStore = (*AMStore)(nil)
