package clusterop

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

const stageRecordSuffix = ".stage"

// upgradeStageStore keeps one node's record of the upgrade stages it was asked to
// run, as one small JSON file each, written atomically the same way operation
// records are.
//
// It exists because an upgrade stage restarts olaresd — installing the new
// daemon is part of the work — so the process that knows how a stage is going
// is the process the stage replaces. Anything held only in memory is lost
// exactly when the control node comes back to ask about it.
type upgradeStageStore struct {
	dir string
	mu  sync.Mutex
}

// newUpgradeStageStore prepares the directory and returns a store rooted at it.
func newUpgradeStageStore(dir string) (*upgradeStageStore, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("upgrade stage store: empty directory")
	}
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return nil, fmt.Errorf("create upgrade stage dir: %w", err)
	}
	if err := os.Chmod(dir, dirMode); err != nil {
		return nil, fmt.Errorf("secure upgrade stage dir: %w", err)
	}
	return &upgradeStageStore{dir: dir}, nil
}

// path names the file for one stage.
//
// The name is a hash rather than the two ids joined, because both arrive from
// the network and either could otherwise name a file outside this directory.
// The ids are inside the record, so nothing is lost by not having them in the
// name.
func (s *upgradeStageStore) path(operationID, stageName string) string {
	sum := sha256.Sum256([]byte(operationID + "\x00" + stageName))
	return filepath.Join(s.dir, hex.EncodeToString(sum[:])+stageRecordSuffix)
}

// Save writes a stage record, replacing any previous version atomically.
func (s *upgradeStageStore) Save(state UpgradeStageState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode stage record: %w", err)
	}
	final := s.path(state.OperationID, state.Stage)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := writeRecord(s.dir, ".stage.*.tmp", final, data); err != nil {
		return fmt.Errorf("write upgrade stage record: %w", err)
	}
	return nil
}

// Load reads one stage record.
func (s *upgradeStageStore) Load(operationID, stageName string) (UpgradeStageState, bool, error) {
	data, err := os.ReadFile(s.path(operationID, stageName))
	if err != nil {
		if os.IsNotExist(err) {
			return UpgradeStageState{}, false, nil
		}
		return UpgradeStageState{}, false, err
	}
	var state UpgradeStageState
	if err := json.Unmarshal(data, &state); err != nil {
		return UpgradeStageState{}, false, fmt.Errorf("decode upgrade stage record: %w", err)
	}
	return state, true, nil
}

// List reads every stage record. An unreadable one is skipped and logged:
// losing one stage's record must not hide the others.
func (s *upgradeStageStore) List() ([]UpgradeStageState, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	states := make([]UpgradeStageState, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), stageRecordSuffix) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			klog.Warningf("clusterop: skipping unreadable stage record %s: %v", e.Name(), err)
			continue
		}
		var state UpgradeStageState
		if err := json.Unmarshal(data, &state); err != nil {
			klog.Warningf("clusterop: skipping undecodable stage record %s: %v", e.Name(), err)
			continue
		}
		states = append(states, state)
	}
	return states, nil
}

// UpgradeStageExec runs one upgrade stage on this machine and returns when it is
// done. In production it is olares-cli; a test replaces it.
type UpgradeStageExec func(ctx context.Context, req UpgradeStageRequest) error

// LocalUpgradeStageRunner runs upgrade stages on the machine it is part of.
//
// It is the same implementation on the control node and on every compute node.
// The control node uses it directly for its own stages and reaches the others
// over HTTP, but what runs and what is recorded is identical, so an upgrade
// that has to be resumed reads one kind of record everywhere.
type LocalUpgradeStageRunner struct {
	store *upgradeStageStore
	exec  UpgradeStageExec
	base  context.Context
	now   func() time.Time

	// mu guards starts so that two requests for the same stage cannot both
	// decide it is not running yet.
	mu sync.Mutex
}

var _ UpgradeStageRunner = (*LocalUpgradeStageRunner)(nil)

// stageRecordsDir is where a node keeps its stage records, under the same
// directory the operation records live in.
const stageRecordsDir = "upgrade-stages"

// NewLocalUpgradeStageRunner opens this node's stage records under the
// operations directory and settles anything left running by a previous
// process.
//
// A stage runs on context.Background() rather than on anything the caller
// holds, for the reason Deps.Base gives: the request that asked for it is long
// gone, and the daemon's own shutdown context is wrong too, because on the
// control node olaresd is stopped by the very work the stage is doing.
func NewLocalUpgradeStageRunner(operationsDir string, exec UpgradeStageExec) (*LocalUpgradeStageRunner, error) {
	if exec == nil {
		return nil, errors.New("upgrade stage runner: no executor")
	}
	store, err := newUpgradeStageStore(filepath.Join(operationsDir, stageRecordsDir))
	if err != nil {
		return nil, err
	}
	r := &LocalUpgradeStageRunner{store: store, exec: exec, base: context.Background(), now: time.Now}
	if err := r.settleInterrupted(); err != nil {
		return nil, err
	}
	return r, nil
}

// settleInterrupted marks stages that were running when this process's
// predecessor stopped.
//
// The stage's olares-cli is a child of olaresd and goes down with it, so a
// record still saying "running" describes work that stopped part way. Reporting
// it as failed is what lets the control node re-dispatch the stage — which is
// safe because upgrade tasks are reentrant, and is the only way to make
// progress, since nothing about the half-finished run can be recovered.
func (r *LocalUpgradeStageRunner) settleInterrupted() error {
	states, err := r.store.List()
	if err != nil {
		return err
	}
	for _, state := range states {
		if state.Phase != UpgradeStagePhaseRunning {
			continue
		}
		at := time.Now()
		state.Phase = UpgradeStagePhaseFailed
		state.Code = CodeDaemonRestarted
		state.Error = "olaresd restarted while this stage was running"
		state.FinishedAt = &at
		if err := r.store.Save(state); err != nil {
			return fmt.Errorf("settle interrupted stage %s: %w", state.Stage, err)
		}
		klog.Infof("clusterop: upgrade stage %s of %s was interrupted by an olaresd restart",
			state.Stage, state.OperationID)
	}
	return nil
}

// runningElsewhere reports a stage this machine is running for some other
// (operation, stage) pair, which is the one thing that must not be joined by a
// second one.
func (r *LocalUpgradeStageRunner) runningElsewhere(operationID, stage string) (UpgradeStageState, bool, error) {
	states, err := r.store.List()
	if err != nil {
		return UpgradeStageState{}, false, err
	}
	for _, s := range states {
		if s.Phase != UpgradeStagePhaseRunning {
			continue
		}
		if s.OperationID == operationID && s.Stage == stage {
			continue
		}
		return s, true, nil
	}
	return UpgradeStageState{}, false, nil
}

// Start begins a stage, or reports the one already under way.
func (r *LocalUpgradeStageRunner) Start(_ context.Context, req UpgradeStageRequest) (UpgradeStageState, error) {
	if strings.TrimSpace(req.OperationID) == "" || strings.TrimSpace(req.Stage) == "" {
		return UpgradeStageState{}, errors.New("an upgrade stage needs an operation and a stage to run")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, found, err := r.store.Load(req.OperationID, req.Stage)
	if err != nil {
		return UpgradeStageState{}, err
	}
	if found {
		switch existing.Phase {
		case UpgradeStagePhaseSucceeded:
			// Already done. Saying so rather than doing it again is what
			// makes a resumed upgrade skip the work it already finished.
			return existing, nil
		case UpgradeStagePhaseRunning:
			return existing, nil
		}
		// Failed: fall through and run it again. Only the orchestrator
		// decides whether a stage is worth retrying, and it has asked.
	}

	// One stage at a time on this machine, whichever operation asked for it.
	//
	// Idempotency per (operation, stage) is not enough on its own. A stage the
	// control node gave up waiting for is still running here — a timeout ends
	// the waiting, not the work — and the retry that follows carries a new
	// operation id, so without this it would start a second olares-cli
	// alongside the first. Two of those on one machine is two helm upgrades,
	// two containerd restarts, or two driver installs at once.
	//
	// The hold is released by the run itself finishing, and by
	// settleInterrupted if olaresd goes down while it is held, so it cannot
	// outlive the work it describes.
	if busy, ok, err := r.runningElsewhere(req.OperationID, req.Stage); err != nil {
		return UpgradeStageState{}, err
	} else if ok {
		klog.Warningf("clusterop: refusing stage %s of %s: stage %s of %s is still running here",
			req.Stage, req.OperationID, busy.Stage, busy.OperationID)
		at := r.now()
		return UpgradeStageState{
			OperationID: req.OperationID,
			Stage:       req.Stage,
			Version:     req.Version,
			Phase:       UpgradeStagePhaseFailed,
			Code:        CodeStageBusy,
			Error:       reasonFor(CodeStageBusy),
			StartedAt:   at,
			FinishedAt:  &at,
		}, nil
	}

	state := UpgradeStageState{
		OperationID: req.OperationID,
		Stage:       req.Stage,
		Version:     req.Version,
		Phase:       UpgradeStagePhaseRunning,
		StartedAt:   r.now(),
	}
	if err := r.store.Save(state); err != nil {
		return UpgradeStageState{}, err
	}

	// Detached from the caller's request: the stage outlives the HTTP call
	// that asked for it by minutes, and the control node follows it by
	// polling rather than by holding the connection open.
	go r.run(req, state)

	return state, nil
}

func (r *LocalUpgradeStageRunner) run(req UpgradeStageRequest, state UpgradeStageState) {
	klog.Infof("clusterop: running upgrade stage %s of %s (target %s)",
		req.Stage, req.OperationID, req.Version)

	err := r.exec(r.base, req)

	at := r.now()
	state.FinishedAt = &at
	if err != nil {
		state.Phase = UpgradeStagePhaseFailed
		state.Code = CodeStageFailed
		// The detail stays on this machine. What the control node records and
		// serves is the code and its fixed text, for the same reason the power
		// operations suppress theirs: an upgrade failure carries file paths,
		// helm release names and apiserver addresses.
		state.Error = reasonFor(CodeStageFailed)
		klog.Errorf("clusterop: upgrade stage %s of %s failed: %v", req.Stage, req.OperationID, err)
	} else {
		state.Phase = UpgradeStagePhaseSucceeded
		klog.Infof("clusterop: upgrade stage %s of %s succeeded", req.Stage, req.OperationID)
	}

	if err := r.store.Save(state); err != nil {
		// Nothing else can be done here: the work is finished, and the only
		// record of it could not be written. The control node will time the
		// stage out rather than be told a wrong answer.
		klog.Errorf("clusterop: record upgrade stage %s of %s: %v", req.Stage, req.OperationID, err)
	}
}

// Status reports a stage this node was asked to run.
func (r *LocalUpgradeStageRunner) Status(operationID, stageName string) (UpgradeStageState, bool) {
	state, found, err := r.store.Load(operationID, stageName)
	if err != nil {
		klog.Errorf("clusterop: read upgrade stage %s of %s: %v", stageName, operationID, err)
		return UpgradeStageState{}, false
	}
	return state, found
}
