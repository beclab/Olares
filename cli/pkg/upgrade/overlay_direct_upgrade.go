package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/plugins/network"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	multusNetworkStatusAnnotation = "k8s.v1.cni.cncf.io/network-status"
	overlayLANInterface           = "net1"

	// overlayNet1WaitTimeout bounds how long the upgrade waits for the recreated
	// overlay Pods to come back with their LAN interface on the new parent.
	overlayNet1WaitTimeout  = 3 * time.Minute
	overlayNet1PollInterval = 5 * time.Second
)

// overlayNet1PollWait is the pause between two checks of the recreated Pods.
// It is a variable so tests can run the loop without sleeping.
var overlayNet1PollWait = func(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// podHasNet1 reports whether Multus recorded the LAN interface for the Pod.
func podHasNet1(p *corev1.Pod) bool {
	raw := p.Annotations[multusNetworkStatusAnnotation]
	if raw == "" {
		return false
	}
	var statuses []struct {
		Interface string `json:"interface"`
	}
	if err := json.Unmarshal([]byte(raw), &statuses); err != nil {
		return false
	}
	for _, s := range statuses {
		if s.Interface == overlayLANInterface {
			return true
		}
	}
	return false
}

// overlayPodsMissingNet1 lists the overlay Pods that are not yet running with
// their LAN interface. A Pod that is still being created counts as missing.
func overlayPodsMissingNet1(ctx context.Context, kube kubernetes.Interface) ([]string, error) {
	pods, err := listOverlayGatewayPods(ctx, kube)
	if err != nil {
		return nil, err
	}
	var missing []string
	for i := range pods {
		p := &pods[i]
		if p.Status.Phase != corev1.PodRunning || !podHasNet1(p) {
			missing = append(missing, p.Namespace+"/"+p.Name)
		}
	}
	return missing, nil
}

// waitOverlayGatewayPodsNet1 blocks until every overlay Pod runs with net1 or
// the context ends; the error names the Pods that were still missing it.
func waitOverlayGatewayPodsNet1(ctx context.Context, kube kubernetes.Interface) error {
	for {
		missing, err := overlayPodsMissingNet1(ctx, kube)
		if err != nil {
			return err
		}
		if len(missing) == 0 {
			return nil
		}
		if err := overlayNet1PollWait(ctx, overlayNet1PollInterval); err != nil {
			return fmt.Errorf("overlay gateway pods still without %s: %v", overlayLANInterface, missing)
		}
	}
}

// recreateOverlayGatewayPodsAfterMigration deletes the overlay Pods after the
// bridge was torn down so their controllers recreate them on the new parent.
// Unlike the best-effort recreate used for a daemon restart, this step is
// required: the old macvlan interfaces vanished with the bridge, so a Pod that
// is not recreated has no LAN presence at all. Waiting for the recreated Pods
// is a separate task, so a slow node retries the wait without deleting Pods
// that already came back with their LAN interface.
type recreateOverlayGatewayPodsAfterMigration struct {
	common.KubeAction
}

func (a *recreateOverlayGatewayPodsAfterMigration) Execute(runtime connector.Runtime) error {
	if !network.OverlayMigratedFromBridge(runtime) {
		logger.Infof("overlay-parent: no bridge was migrated, overlay gateway pods left as they are")
		return nil
	}
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return errors.Wrap(err, "kube client for overlay gateway pod recreate")
	}
	ctx, cancel := context.WithTimeout(context.Background(), overlayRecreateTimeout)
	defer cancel()

	deleted, remaining, err := recreateOverlayGatewayPods(ctx, kube)
	if err != nil {
		return errors.Wrapf(err, "recreate overlay gateway pods (%d deleted, %d left)", deleted, remaining)
	}
	logger.Infof("overlay-parent: deleted %d overlay gateway pods for recreate on the wired parent", deleted)
	return nil
}

// waitOverlayGatewayPodsNet1AfterMigration waits until every recreated overlay
// Pod runs with its LAN interface on the new parent.
type waitOverlayGatewayPodsNet1AfterMigration struct {
	common.KubeAction
}

func (a *waitOverlayGatewayPodsNet1AfterMigration) Execute(runtime connector.Runtime) error {
	if !network.OverlayMigratedFromBridge(runtime) {
		return nil
	}
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return errors.Wrap(err, "kube client for overlay gateway pod wait")
	}
	ctx, cancel := context.WithTimeout(context.Background(), overlayNet1WaitTimeout)
	defer cancel()
	if err := waitOverlayGatewayPodsNet1(ctx, kube); err != nil {
		logger.Errorf("overlay-parent: %v", err)
		return err
	}
	logger.Infof("overlay-parent: every overlay gateway pod runs with %s again", overlayLANInterface)
	return nil
}

// overlayBridgeMigrated gates the node-side steps on the bridge having been
// torn down by this upgrade, so a node with the overlay gateway off is not
// touched at all.
type overlayBridgeMigrated struct {
	common.KubePrepare
}

func (p *overlayBridgeMigrated) PreCheck(runtime connector.Runtime) (bool, error) {
	return network.OverlayMigratedFromBridge(runtime), nil
}

// overlayDirectPreTasks run before the NAD is re-rendered: the bridge must be
// gone and the alternative name in place before any Pod resolves the new master.
func overlayDirectPreTasks() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "MigrateOverlayBridgeToDirect",
			Desc:   "Move the overlay parent from the legacy bridge to the wired NIC",
			Action: new(network.MigrateBridgeToDirect),
		},
		// The daemon's host address check may have pointed the host name at
		// another interface while the NIC was between the bridge and its own
		// address; write the entry the upgrade started with again.
		&task.LocalTask{
			Name:    "ReassertHostsAfterMigration",
			Desc:    "Point the host name at the wired NIC address again",
			Prepare: new(overlayBridgeMigrated),
			Action:  new(terminus.UpdateKubeKeyHosts),
			Retry:   5,
		},
		&task.LocalTask{
			Name:    "EnsureOverlayAltname",
			Desc:    "Add the overlay parent alternative name to the NIC that left the bridge",
			Prepare: new(overlayBridgeMigrated),
			Action:  new(network.EnsureOverlayAltname),
			Retry:   3,
		},
	}
}

// overlayDirectPostTasks run after the NAD points at the alternative name.
func overlayDirectPostTasks() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "WriteOverlayDesiredState",
			Desc:   "Keep the overlay gateway enabled on nodes that had the bridge",
			Action: new(network.WriteOverlayDesiredIfMigrated),
			Retry:  3,
		},
		&task.LocalTask{
			Name:   "RecreateOverlayGatewayPodsAfterMigration",
			Desc:   "Recreate overlay gateway pods on the wired parent",
			Action: new(recreateOverlayGatewayPodsAfterMigration),
		},
		&task.LocalTask{
			Name:   "WaitOverlayGatewayPodsNet1",
			Desc:   "Wait for the overlay gateway pods to run with their LAN interface",
			Action: new(waitOverlayGatewayPodsNet1AfterMigration),
			Retry:  5,
			Delay:  10 * time.Second,
		},
		&task.LocalTask{
			Name:   "ClearOverlayMigrationMarker",
			Desc:   "Finish the overlay parent migration",
			Action: new(network.ClearOverlayMigrationMarker),
		},
	}
}
