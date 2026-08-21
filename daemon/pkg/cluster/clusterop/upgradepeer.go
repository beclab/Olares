package clusterop

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cli"
	"github.com/beclab/Olares/daemon/pkg/cluster/fanout"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// UpgradeStagePath is the node-local endpoint the control node hands an
	// upgrade stage to.
	UpgradeStagePath = "/command/upgrade-node"

	// UpgradeStageStatusPath is where the same node reports how it is going.
	UpgradeStageStatusPath = "/node/upgrade-stage"

	// UpgradeReadinessPath is where a node answers whether it can run a stage
	// of this upgrade at all. It sits on the upgrade's own surface, behind
	// the operation token, because that is the only credential the
	// orchestrator has; see UpgradeReadiness.
	UpgradeReadinessPath = "/node/upgrade-readiness"

	// UpgradeTokenHeader carries the per-operation authorization. It is not the
	// owner's signature; see UpgradeDeps.Auth.
	UpgradeTokenHeader = "X-Upgrade-Token"

	// stageDispatchTimeout bounds starting a stage, not running it. The node
	// answers as soon as it has recorded and launched the work.
	stageDispatchTimeout = 30 * time.Second

	// stageStatusTimeout bounds one status poll.
	stageStatusTimeout = 10 * time.Second

	// upgradeSecretNamespace is where the per-operation token lives. It is the
	// same namespace whose UID already identifies the cluster to both sides,
	// so this needs no trust the design did not already require.
	upgradeSecretNamespace = "kube-system"

	upgradeSecretPrefix = "olares-upgrade-"
	upgradeSecretKey    = "token"
)

// NewUpgradeDeps returns the upgrade seams wired to the real cluster. It is
// separate from NewDeps so that a daemon which cannot run upgrades still gets
// the power orchestrator: it passes nil here and everything else works.
func NewUpgradeDeps(local UpgradeStageRunner) *UpgradeDeps {
	return &UpgradeDeps{
		Plan:      ReadUpgradePlan,
		Local:     local,
		Auth:      EnsureUpgradeToken,
		Start:     startStageOnNode,
		Status:    stageStatusOfNode,
		Readiness: upgradeReadinessOfNode,
	}
}

// ReadUpgradePlan asks this node's olares-cli what an upgrade to the version
// it carries consists of.
//
// The daemon does not hold a plan of its own. Which tasks exist, and which
// nodes each one has to run on, is a property of the binary that implements
// them — it changes with every release, and a copy kept here would describe
// some other version's upgrade.
func ReadUpgradePlan(ctx context.Context) (UpgradePlan, error) {
	out, err := runCLI(ctx, "upgrade", "plan")
	if err != nil {
		return UpgradePlan{}, err
	}

	var plan UpgradePlan
	if err := json.Unmarshal(out, &plan); err != nil {
		return UpgradePlan{}, fmt.Errorf("decode the upgrade plan: %w", err)
	}
	return plan, nil
}

// RunUpgradeStage is the production executor: it runs one stage of the flow on
// this machine and returns when olares-cli exits.
//
// Only the stage name is passed through. Which tasks that name resolves to is
// this machine's olares-cli's answer, and the orchestrator has already checked
// that this machine holds the version whose answer it scheduled.
func RunUpgradeStage(ctx context.Context, req UpgradeStageRequest) error {
	args := []string{"upgrade", "--base-dir", commands.TERMINUS_BASE_DIR, "--stage", req.Stage}
	_, err := runCLI(ctx, args...)
	return err
}

// runCLI invokes olares-cli by the path the rest of the daemon uses, so a
// stage runs the same binary the daemon's other commands do rather than
// whatever happens to be on PATH.
func runCLI(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, cli.TERMINUS_CLI, args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("olares-cli %s: %w: %s",
				strings.Join(args, " "), err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("olares-cli %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

// EnsureUpgradeToken returns the token workers check this operation's stages
// against, creating it on first use.
//
// It lives in a Secret rather than in this process because both properties an
// upgrade needs are properties of the cluster, not of the run: it has to
// outlive the daemon that minted it — an upgrade restarts olaresd on purpose —
// and a resumed run has to arrive at the same value, or the workers a previous
// run already dispatched to would stop recognizing it.
//
// The trust boundary is unchanged from the rest of this design: a node that
// can read kube-system is a node of this cluster, which is already what
// clusterId assumes when both sides compare that namespace's UID.
func EnsureUpgradeToken(ctx context.Context, operationID string) (string, error) {
	client, err := utils.GetKubeClient()
	if err != nil {
		return "", err
	}
	return ensureUpgradeToken(ctx, client, operationID)
}

func ensureUpgradeToken(ctx context.Context, client kubernetes.Interface, operationID string) (string, error) {
	name, err := upgradeSecretName(operationID)
	if err != nil {
		return "", err
	}
	secrets := client.CoreV1().Secrets(upgradeSecretNamespace)

	if existing, err := secrets.Get(ctx, name, metav1.GetOptions{}); err == nil {
		if token := strings.TrimSpace(string(existing.Data[upgradeSecretKey])); token != "" {
			return token, nil
		}
		return "", fmt.Errorf("the upgrade token for %s is empty", operationID)
	} else if !apierrors.IsNotFound(err) {
		return "", err
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate an upgrade token: %w", err)
	}
	token := hex.EncodeToString(buf)

	_, err = secrets.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: upgradeSecretNamespace,
			Labels:    map[string]string{"olares.io/upgrade-operation": "true"},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{upgradeSecretKey: []byte(token)},
	}, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		// Another run of the same operation got there first; theirs is the
		// one the workers will have seen.
		existing, getErr := secrets.Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			return "", getErr
		}
		return strings.TrimSpace(string(existing.Data[upgradeSecretKey])), nil
	}
	if err != nil {
		return "", err
	}
	return token, nil
}

// VerifyUpgradeToken reports whether presented is this operation's token. It is
// what a worker calls, and it reads the same Secret the control node wrote.
func VerifyUpgradeToken(ctx context.Context, operationID, presented string) error {
	if strings.TrimSpace(presented) == "" {
		return errors.New("no upgrade token was presented")
	}
	client, err := utils.GetKubeClient()
	if err != nil {
		return err
	}
	return verifyUpgradeToken(ctx, client, operationID, presented)
}

func verifyUpgradeToken(ctx context.Context, client kubernetes.Interface, operationID, presented string) error {
	name, err := upgradeSecretName(operationID)
	if err != nil {
		return err
	}
	secret, err := client.CoreV1().Secrets(upgradeSecretNamespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return errors.New("this upgrade is not one this cluster is running")
		}
		return err
	}
	want := strings.TrimSpace(string(secret.Data[upgradeSecretKey]))
	// Constant time: the comparison is against a secret, and answering faster
	// for a wrong first byte is how one gets guessed.
	if want == "" || subtle.ConstantTimeCompare([]byte(want), []byte(strings.TrimSpace(presented))) != 1 {
		return errors.New("the upgrade token does not authorize this stage")
	}
	return nil
}

// upgradeSecretName derives a Secret name from an operation id. The id is
// generated by this daemon and is already name-safe, but it reaches here from
// a request body, so it is validated rather than assumed.
func upgradeSecretName(operationID string) (string, error) {
	id := strings.TrimSpace(operationID)
	if id == "" {
		return "", errors.New("no operation to authorize")
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
		default:
			return "", fmt.Errorf("%q is not a usable operation id", operationID)
		}
	}
	name := upgradeSecretPrefix + id
	if len(name) > 253 {
		return "", fmt.Errorf("%q is not a usable operation id", operationID)
	}
	return name, nil
}

// StageDispatchError is a failure to hand a stage to a node, carrying the code
// the transport classified it as so the operation record can keep a node that
// never answered apart from one that refused.
type StageDispatchError struct {
	Code string
	Err  error
}

func (e *StageDispatchError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *StageDispatchError) Unwrap() error { return e.Err }

// startStageOnNode hands one stage to another node's olaresd.
//
// It goes through the fan-out even though it addresses a single node. The
// concurrency is not what is wanted here — the orchestrator runs its own,
// because a stage has to be polled after it is started and a serial stage must
// not start on the next node at all — but everything else is: the URL
// convention, the per-call timeout, the response envelope, and above all the
// classification of what went wrong. Hand-rolling the call meant reporting a
// node that never answered and a node that refused with the same code, which
// is strictly worse than what the power path reports for the same failure.
func startStageOnNode(ctx context.Context, node inventory.Node, req UpgradeStageRequest,
	token string) (UpgradeStageState, error) {
	results := (&fanout.Dispatcher{
		PeerPath: UpgradeStagePath,
		Headers:  map[string]string{UpgradeTokenHeader: token},
		Timeout:  stageDispatchTimeout,
	}).Run(ctx, peerTargets([]inventory.Node{node}), func(fanout.NodeTarget) any { return req })

	return stageResultOf(results)
}

// stageResultOf turns one fan-out result into a stage state or a classified
// failure.
func stageResultOf(results []fanout.NodeResult) (UpgradeStageState, error) {
	if len(results) != 1 {
		return UpgradeStageState{}, &StageDispatchError{
			Code: CodeDispatchFailed,
			Err:  fmt.Errorf("expected one result, got %d", len(results)),
		}
	}
	r := results[0]
	if r.Status != fanout.StatusOK {
		code := CodeDispatchFailed
		if r.Status == fanout.StatusUnreachable || r.Status == fanout.StatusTimeout {
			code = CodeNodeUnreachable
		}
		return UpgradeStageState{}, &StageDispatchError{Code: code, Err: errors.New(r.Err)}
	}

	var env struct {
		Data UpgradeStageState `json:"data"`
	}
	if err := json.Unmarshal(r.Data, &env); err != nil {
		return UpgradeStageState{}, &StageDispatchError{
			Code: CodeDispatchFailed,
			Err:  fmt.Errorf("decode the node's answer: %w", err),
		}
	}
	return env.Data, nil
}

// upgradeReadinessOfNode asks one node whether it can run a stage of this
// upgrade. An olaresd from before stages existed has no such route, and the
// 404 it answers with is reported to the caller as an error — which is the
// right answer to the question.
func upgradeReadinessOfNode(ctx context.Context, node inventory.Node, operationID, token string) (UpgradeReadiness, error) {
	return getFromNode[UpgradeReadiness](ctx, node, UpgradeReadinessPath,
		url.Values{"operationId": {operationID}}, token)
}

// stageStatusOfNode asks another node how a stage is going.
//
// This one is not a fan-out call. The fan-out speaks one shape — POST a JSON
// body — and reading a stage is a read: making it a POST so that both halves
// could share a transport would hide that in the one place a reader looks for
// it. It is a short GET, and the polling loop above already treats a failed
// poll as "ask again", so the classification the dispatch needs has no use
// here.
func stageStatusOfNode(ctx context.Context, node inventory.Node,
	operationID, stageName, token string) (UpgradeStageState, error) {
	return getFromNode[UpgradeStageState](ctx, node, UpgradeStageStatusPath,
		url.Values{"operationId": {operationID}, "stage": {stageName}}, token)
}

// getFromNode reads one token-guarded value off another node's olaresd.
//
// Both upgrade reads are the same request with a different path and a
// different payload, so they are one function. A non-200 is an error rather
// than a zero value: these answers decide whether an upgrade may go ahead,
// and a node that will not say is not a node that said no.
func getFromNode[T any](ctx context.Context, node inventory.Node,
	path string, query url.Values, token string) (T, error) {
	var zero T
	if node.IP == "" {
		return zero, errors.New("node has no internal address")
	}

	reqCtx, cancel := context.WithTimeout(ctx, stageStatusTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet,
		nodeUpgradeURL(node.IP, path, query), nil)
	if err != nil {
		return zero, err
	}
	httpReq.Header.Set(UpgradeTokenHeader, token)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("node answered %s", resp.Status)
	}
	var env struct {
		Data T `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return zero, fmt.Errorf("decode the node's answer: %w", err)
	}
	return env.Data, nil
}

func nodeUpgradeURL(ip, path string, query url.Values) string {
	u := "http://" + net.JoinHostPort(ip, strconv.Itoa(fanout.OlaresdPort)) + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}
