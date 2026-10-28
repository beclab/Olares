package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
)

// fakeStageRunner records what it was asked to run without running anything.
type fakeStageRunner struct {
	started []clusterop.UpgradeStageRequest
	state   clusterop.UpgradeStageState
	found   bool
	err     error
}

func (f *fakeStageRunner) Start(_ context.Context, req clusterop.UpgradeStageRequest) (clusterop.UpgradeStageState, error) {
	if f.err != nil {
		return clusterop.UpgradeStageState{}, f.err
	}
	f.started = append(f.started, req)
	return clusterop.UpgradeStageState{
		OperationID: req.OperationID, Stage: req.Stage,
		Phase: clusterop.UpgradeStagePhaseRunning,
	}, nil
}

func (f *fakeStageRunner) Status(string, string) (clusterop.UpgradeStageState, bool) {
	return f.state, f.found
}

const testStageToken = "test-upgrade-token"

// withStageRunner installs a runner and a token check that accepts only the
// operation's own token. The handler's own checks all still run.
func withStageRunner(t *testing.T, runner clusterop.UpgradeStageRunner) *fakeStageRunner {
	t.Helper()
	prevRunner := upgradeStages
	prevVerify := verifyUpgradeToken
	upgradeStages = runner
	verifyUpgradeToken = func(_ context.Context, operationID, presented string) error {
		if operationID == "" {
			return errors.New("no operation")
		}
		if presented != testStageToken {
			return errors.New("wrong token")
		}
		return nil
	}
	t.Cleanup(func() {
		upgradeStages = prevRunner
		verifyUpgradeToken = prevVerify
	})
	if f, ok := runner.(*fakeStageRunner); ok {
		return f
	}
	return nil
}

func stageBody(t *testing.T, req clusterop.UpgradeStageRequest) string {
	t.Helper()
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(raw)
}

func stageHeaders(token string) map[string]string {
	return map[string]string{clusterop.UpgradeTokenHeader: token}
}

func sampleStageRequest() clusterop.UpgradeStageRequest {
	return clusterop.UpgradeStageRequest{
		OperationID: "op-1", Stage: "02-all-nodes",
		Version: "1.12.8", ClusterID: "cluster-test",
	}
}

func TestUpgradeStageRunsWithTheOperationToken(t *testing.T) {
	runner := withStageRunner(t, &fakeStageRunner{})
	asOwnerSignature(t) // supplies clusterIDOf
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/upgrade-node",
		stageBody(t, sampleStageRequest()), stageHeaders(testStageToken))

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, string(body))
	}
	if len(runner.started) != 1 || runner.started[0].Stage != "02-all-nodes" {
		t.Fatalf("runner was asked for %+v", runner.started)
	}
}

// The route is not behind an owner signature, so the token is the only thing
// standing in front of it and has to actually be checked.
func TestUpgradeStageRefusesAWrongToken(t *testing.T) {
	runner := withStageRunner(t, &fakeStageRunner{})
	asOwnerSignature(t)
	asWorker(t)

	for name, headers := range map[string]map[string]string{
		"wrong token": stageHeaders("not-the-token"),
		"no token":    {},
	} {
		resp, _ := callRegisteredMethod(t, http.MethodPost, "/command/upgrade-node",
			stageBody(t, sampleStageRequest()), headers)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s: status = %d, want 403", name, resp.StatusCode)
		}
	}
	if len(runner.started) != 0 {
		t.Errorf("a refused request still started %d stages", len(runner.started))
	}
}

// A token is only meaningful in the cluster that minted it, and this node
// reads its own kube-system UID rather than trusting the request's.
func TestUpgradeStageRefusesAnotherCluster(t *testing.T) {
	runner := withStageRunner(t, &fakeStageRunner{})
	asOwnerSignature(t)
	asWorker(t)

	req := sampleStageRequest()
	req.ClusterID = "some-other-cluster"
	resp, _ := callRegisteredMethod(t, http.MethodPost, "/command/upgrade-node",
		stageBody(t, req), stageHeaders(testStageToken))

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
	if len(runner.started) != 0 {
		t.Error("a stage for another cluster was started")
	}
}

func TestUpgradeStageRefusesARequestNamingNoStage(t *testing.T) {
	runner := withStageRunner(t, &fakeStageRunner{})
	asOwnerSignature(t)
	asWorker(t)

	resp, _ := callRegisteredMethod(t, http.MethodPost, "/command/upgrade-node",
		stageBody(t, clusterop.UpgradeStageRequest{OperationID: "op-1", ClusterID: "cluster-test"}),
		stageHeaders(testStageToken))

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	if len(runner.started) != 0 {
		t.Error("a request naming no stage was started")
	}
}

// A node with no runner installed refuses rather than pretending to work.
func TestUpgradeStageRefusesWhenTheNodeCannotRunStages(t *testing.T) {
	prev := upgradeStages
	upgradeStages = nil
	t.Cleanup(func() { upgradeStages = prev })
	asOwnerSignature(t)
	asWorker(t)

	resp, _ := callRegisteredMethod(t, http.MethodPost, "/command/upgrade-node",
		stageBody(t, sampleStageRequest()), stageHeaders(testStageToken))
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
}

func TestUpgradeStageStatusIsReadable(t *testing.T) {
	withStageRunner(t, &fakeStageRunner{
		found: true,
		state: clusterop.UpgradeStageState{
			OperationID: "op-1", Stage: "02-all-nodes",
			Phase: clusterop.UpgradeStagePhaseSucceeded,
		},
	})
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegistered(t,
		"/node/upgrade-stage?operationId=op-1&stage=02-all-nodes", stageHeaders(testStageToken))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, string(body))
	}

	var env struct {
		Data clusterop.UpgradeStageState `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Phase != clusterop.UpgradeStagePhaseSucceeded {
		t.Errorf("phase = %s", env.Data.Phase)
	}
}

// Reading a stage names the version being rolled out, so it needs the same
// authorization as starting one.
func TestUpgradeStageStatusRefusesAWrongToken(t *testing.T) {
	withStageRunner(t, &fakeStageRunner{found: true})
	asOwnerSignature(t)
	asWorker(t)

	resp, _ := callRegistered(t,
		"/node/upgrade-stage?operationId=op-1&stage=02-all-nodes", stageHeaders("nope"))
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
}

func TestUpgradeStageStatusIsNotFoundForAnUnknownStage(t *testing.T) {
	withStageRunner(t, &fakeStageRunner{found: false})
	asOwnerSignature(t)
	asWorker(t)

	resp, _ := callRegistered(t,
		"/node/upgrade-stage?operationId=op-1&stage=99-nope", stageHeaders(testStageToken))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}
