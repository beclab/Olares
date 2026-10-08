package clusterop

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/fanout"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpgradeStageTokenIsMintedOnceAndReReadAfterwards(t *testing.T) {
	client := fake.NewSimpleClientset()
	ctx := context.Background()

	first, err := ensureUpgradeToken(ctx, client, "op-20260101-000000-abcdef")
	if err != nil {
		t.Fatalf("ensureUpgradeToken: %v", err)
	}
	if first == "" {
		t.Fatal("no token was minted")
	}

	// A resumed run must arrive at the same value, or the workers a previous
	// run already dispatched to would stop recognizing it.
	second, err := ensureUpgradeToken(ctx, client, "op-20260101-000000-abcdef")
	if err != nil {
		t.Fatalf("second ensureUpgradeToken: %v", err)
	}
	if second != first {
		t.Errorf("a resumed run minted a new token: %q then %q", first, second)
	}
}

func TestUpgradeStageTokenVerification(t *testing.T) {
	client := fake.NewSimpleClientset()
	ctx := context.Background()
	const id = "op-20260101-000000-abcdef"

	token, err := ensureUpgradeToken(ctx, client, id)
	if err != nil {
		t.Fatalf("ensureUpgradeToken: %v", err)
	}

	if err := verifyUpgradeToken(ctx, client, id, token); err != nil {
		t.Errorf("the token this cluster minted was refused: %v", err)
	}
	if err := verifyUpgradeToken(ctx, client, id, "not-the-token"); err == nil {
		t.Error("a wrong token was accepted")
	}
	if err := verifyUpgradeToken(ctx, client, id, ""); err == nil {
		t.Error("an empty token was accepted")
	}
	// An operation this cluster never started authorizes nothing, even with a
	// token that is valid for another one.
	if err := verifyUpgradeToken(ctx, client, "op-20260101-000000-fedcba", token); err == nil {
		t.Error("a token was accepted for an operation this cluster is not running")
	}
}

// The id reaches this code from a request body and becomes part of a resource
// name, so anything that is not a plain id is refused rather than escaped.
func TestUpgradeStageSecretNameRefusesAHostileOperationID(t *testing.T) {
	for _, id := range []string{
		"", "  ", "../kube-root-ca.crt", "Op-With-Capitals", "op/with/slashes",
		"op with spaces", "op.with.dots",
	} {
		if _, err := upgradeSecretName(id); err == nil {
			t.Errorf("%q was accepted as an operation id", id)
		}
	}
	if _, err := upgradeSecretName("op-20260101-000000-abcdef"); err != nil {
		t.Errorf("a real operation id was refused: %v", err)
	}
}

// The token is a secret in the cluster, and it is scoped to the namespace that
// already identifies the cluster to both sides.
func TestUpgradeStageTokenLivesInKubeSystem(t *testing.T) {
	client := fake.NewSimpleClientset()
	ctx := context.Background()
	const id = "op-20260101-000000-abcdef"

	if _, err := ensureUpgradeToken(ctx, client, id); err != nil {
		t.Fatalf("ensureUpgradeToken: %v", err)
	}
	secrets, err := client.CoreV1().Secrets(upgradeSecretNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(secrets.Items) != 1 {
		t.Fatalf("secrets in %s = %d, want 1", upgradeSecretNamespace, len(secrets.Items))
	}
	if got := secrets.Items[0].Name; got != upgradeSecretPrefix+id {
		t.Errorf("secret name = %q", got)
	}
}

// The dispatch keeps the transport's classification: a node that never
// answered and a node that answered with a refusal are different problems, and
// the record has to say which.
func TestUpgradeStageDispatchKeepsTheTransportClassification(t *testing.T) {
	cases := []struct {
		name   string
		result fanout.NodeResult
		want   string
	}{
		{"unreachable", fanout.NodeResult{Status: fanout.StatusUnreachable, Err: "dial tcp: refused"}, CodeNodeUnreachable},
		{"timeout", fanout.NodeResult{Status: fanout.StatusTimeout, Err: "deadline exceeded"}, CodeNodeUnreachable},
		{"refused", fanout.NodeResult{Status: fanout.StatusError, Err: "node returned 403"}, CodeDispatchFailed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := stageResultOf([]fanout.NodeResult{tc.result})
			var dispatch *StageDispatchError
			if !errors.As(err, &dispatch) {
				t.Fatalf("err = %v, want a classified dispatch error", err)
			}
			if dispatch.Code != tc.want {
				t.Errorf("code = %s, want %s", dispatch.Code, tc.want)
			}
		})
	}
}

// A node that accepted the stage answers with its own record of it.
func TestUpgradeStageDispatchReadsTheNodesAnswer(t *testing.T) {
	body, err := json.Marshal(struct {
		Code int               `json:"code"`
		Data UpgradeStageState `json:"data"`
	}{
		Code: 200,
		Data: UpgradeStageState{OperationID: "op-1", Stage: "02-all-nodes", Phase: UpgradeStagePhaseRunning},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	state, err := stageResultOf([]fanout.NodeResult{{Status: fanout.StatusOK, Data: body}})
	if err != nil {
		t.Fatalf("stageResultOf: %v", err)
	}
	if state.Stage != "02-all-nodes" || state.Phase != UpgradeStagePhaseRunning {
		t.Errorf("state = %+v", state)
	}
}

// An answer that is not a stage state is a dispatch failure, not a stage that
// silently reports the zero value — which would read as a stage in no phase at
// all and be waited on until the deadline.
func TestUpgradeStageDispatchRefusesAnUndecodableAnswer(t *testing.T) {
	_, err := stageResultOf([]fanout.NodeResult{{Status: fanout.StatusOK, Data: []byte("not json")}})
	var dispatch *StageDispatchError
	if !errors.As(err, &dispatch) || dispatch.Code != CodeDispatchFailed {
		t.Fatalf("err = %v, want a dispatch failure", err)
	}
}

func TestUpgradePlanValidation(t *testing.T) {
	valid := UpgradePlan{
		Version: "1.12.8",
		Stages:  []UpgradeStage{{Name: "01-master", Placement: PlacementAdmin}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("a usable plan was refused: %v", err)
	}

	cases := map[string]UpgradePlan{
		"no version": {Stages: valid.Stages},
		"no stages":  {Version: "1.12.8"},
		"unnamed stage": {Version: "1.12.8",
			Stages: []UpgradeStage{{Placement: PlacementAdmin}}},
		"repeated stage": {Version: "1.12.8", Stages: []UpgradeStage{
			{Name: "01-master", Placement: PlacementAdmin},
			{Name: "01-master", Placement: PlacementWorkers},
		}},
		// A scope this daemon does not know how to schedule is refused while
		// the cluster is whole, not discovered part way through.
		"unknown fanout": {Version: "1.12.8",
			Stages: []UpgradeStage{{Name: "01-gpu-nodes", Placement: "gpu-nodes"}}},
	}
	for name, plan := range cases {
		if err := plan.Validate(); err == nil {
			t.Errorf("%s: was accepted", name)
		}
	}
}
