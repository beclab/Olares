package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// multusNetworksAnnotation is the Multus network selection annotation.
	multusNetworksAnnotation = "k8s.v1.cni.cncf.io/networks"
	// underlayNetworkName / underlayNetworkNamespace identify the
	// NetworkAttachmentDefinition Overlay Gateway Pods attach through. A Pod
	// selecting it holds a DHCP lease in the cni-dhcp daemon.
	underlayNetworkName      = "underlay-macvlan"
	underlayNetworkNamespace = "kube-system"

	// overlayRecreateBatchLimit bounds how many Pods are deleted before the
	// loop pauses, so a large fleet is not recreated all at once.
	overlayRecreateBatchLimit = 50
	overlayRecreateTimeout    = 5 * time.Minute
)

// overlayRecreateBatchPause is the wait between two delete batches.
const overlayRecreateBatchPause = 5 * time.Second

// overlayRecreateBatchWait waits out the pause between two delete batches and
// reports why it stopped: nil when the pause elapsed, the context's error when
// the deadline passed first. It is a variable so tests drive the batching
// without sleeping and without cancelling the context the deletes run under.
var overlayRecreateBatchWait = func(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// cniBinDir is shared: besides the plugins from the official cni-plugins
// archive it also holds binaries owned by other components (calico, multus)
// and plugins that earlier upgraders replaced with newer builds. An upgrade
// must therefore never extract the archive over this directory.
const cniBinDir = "/opt/cni/bin"

// cniDhcpStageDir is where the plugin is unpacked before it is installed. It
// is a sibling of cniBinDir rather than a directory inside it, so nothing that
// is not a plugin is ever written into the plugin directory. Both normally live
// on one filesystem, which makes installing the plugin a rename: the running
// daemon keeps the old inode until it is restarted, and a half-written file can
// never be left in place. Should they ever be separate mounts, mv unlinks the
// destination and copies, which still avoids writing into the running binary.
const cniDhcpStageDir = "/opt/cni/.dhcp-upgrade"

// syncCniPluginsArchive copies the cni-plugins archive pinned in the manifest
// to the node and returns its path there.
func syncCniPluginsArchive(runtime connector.Runtime, manifestPath string) (string, error) {
	m, err := manifest.ReadAll(manifestPath)
	if err != nil {
		return "", fmt.Errorf("read manifest: %w", err)
	}
	binary, err := m.Get("cni-plugins")
	if err != nil {
		return "", fmt.Errorf("get cni-plugins binary info failed: %w", err)
	}

	src := binary.FilePath(runtime.GetBaseDir())
	dst := filepath.Join(common.TmpDir, binary.Filename)
	logger.Debugf("cni-plugins: copy archive from %s to %s", src, dst)
	if err := runtime.GetRunner().Scp(src, dst); err != nil {
		return "", errors.Wrap(errors.WithStack(err), "sync cni-plugins archive failed")
	}
	return dst, nil
}

// nodeCommand is one step of a node-side installation, with the text used in
// the error message when it fails.
type nodeCommand struct {
	desc string
	cmd  string
	// logOutput reports the command's first output line, for steps whose
	// output is evidence worth keeping in the upgrade log.
	logOutput bool
}

// cniDhcpInstallCommands stages the dhcp plugin out of the archive, checks that
// it runs on this platform, moves it into place and restarts the daemon. Only
// the dhcp member is extracted, so every other file in cniBinDir keeps the
// build it already has (the archive itself is the official upstream release
// with dhcp replaced, but the node's other plugins may be newer). The staged
// file carries the archive's own mode and ownership, which is what a fresh
// install would leave behind.
func cniDhcpInstallCommands(archive string) []nodeCommand {
	staged := cniDhcpStageDir + "/dhcp"
	return []nodeCommand{
		{desc: "prepare the staging directory", cmd: fmt.Sprintf("rm -rf %s && mkdir -p %s", cniDhcpStageDir, cniDhcpStageDir)},
		// the archive is packed from a directory, so members are named ./dhcp;
		// accept a plain dhcp entry as well
		{desc: "extract the dhcp plugin", cmd: fmt.Sprintf("tar -zxf %s -C %s ./dhcp || tar -zxf %s -C %s dhcp", archive, cniDhcpStageDir, archive, cniDhcpStageDir)},
		{desc: "run the staged dhcp plugin", cmd: staged + " --version", logOutput: true},
		{desc: "install the dhcp plugin", cmd: fmt.Sprintf("mv -f %s %s/dhcp", staged, cniBinDir)},
		{desc: "restart the cni-dhcp daemon", cmd: "systemctl restart cni-dhcp"},
	}
}

// upgradeCniDhcpBinary replaces only /opt/cni/bin/dhcp so the new plugin
// (stable client identifier, ipam.sendRelease) is in effect. The generic
// upgrade task set never replaces CNI binaries. Restarting the daemon drops
// its in-memory leases, so the upgrader recreates the Overlay Gateway Pods
// right after the NAD is re-rendered.
type upgradeCniDhcpBinary struct {
	common.KubeAction
}

func (u *upgradeCniDhcpBinary) Execute(runtime connector.Runtime) error {
	archive, err := syncCniPluginsArchive(runtime, u.KubeConf.Arg.Manifest)
	if err != nil {
		return err
	}
	runner := runtime.GetRunner()
	defer func() {
		if _, err := runner.SudoCmd("rm -rf "+cniDhcpStageDir, false, false); err != nil {
			logger.Warnf("cni-dhcp: leftover staging directory %s: %v", cniDhcpStageDir, err)
		}
	}()
	for _, step := range cniDhcpInstallCommands(archive) {
		out, err := runner.SudoCmd(step.cmd, false, false)
		if err != nil {
			return errors.Wrap(err, step.desc+" failed")
		}
		if step.logOutput {
			logger.Infof("cni-dhcp: staged plugin reports %s", strings.SplitN(strings.TrimSpace(out), "\n", 2)[0])
		}
	}
	return nil
}

func cniDhcpBinaryUpgradeTasks() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeCniDhcpBinary",
			Desc:   "Install the dhcp plugin and restart cni-dhcp",
			Action: new(upgradeCniDhcpBinary),
			Retry:  3,
		},
	}
}

// networkSelection is one entry of the Multus selection annotation in its
// JSON form; only the fields needed to identify the network are decoded.
type networkSelection struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// parseNetworkSelections decodes the Multus selection annotation. It accepts
// the JSON form (array or single object) and the short form
// "[namespace/]name[@interface][, ...]". A network without a namespace lives
// in the Pod's own namespace, as Multus resolves it.
func parseNetworkSelections(value, podNamespace string) []networkSelection {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if strings.HasPrefix(value, "[") || strings.HasPrefix(value, "{") {
		var list []networkSelection
		if err := json.Unmarshal([]byte(value), &list); err != nil {
			var single networkSelection
			if err := json.Unmarshal([]byte(value), &single); err != nil {
				return nil
			}
			list = []networkSelection{single}
		}
		for i := range list {
			if list[i].Namespace == "" {
				list[i].Namespace = podNamespace
			}
		}
		return list
	}

	var out []networkSelection
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if at := strings.Index(item, "@"); at >= 0 {
			item = item[:at]
		}
		sel := networkSelection{Namespace: podNamespace}
		if slash := strings.Index(item, "/"); slash >= 0 {
			sel.Namespace, sel.Name = item[:slash], item[slash+1:]
		} else {
			sel.Name = item
		}
		out = append(out, sel)
	}
	return out
}

// podSelectsUnderlay reports whether the Pod's Multus selection references the
// kube-system/underlay-macvlan network in either annotation syntax.
func podSelectsUnderlay(p *corev1.Pod) bool {
	for _, sel := range parseNetworkSelections(p.Annotations[multusNetworksAnnotation], p.Namespace) {
		if sel.Name == underlayNetworkName && sel.Namespace == underlayNetworkNamespace {
			return true
		}
	}
	return false
}

func listOverlayGatewayPods(ctx context.Context, kube kubernetes.Interface) ([]corev1.Pod, error) {
	pods, err := kube.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var out []corev1.Pod
	for i := range pods.Items {
		p := pods.Items[i]
		if p.DeletionTimestamp != nil || !podSelectsUnderlay(&p) {
			continue
		}
		// Deleting is only a recreate when a controller owns the Pod. A
		// standalone Pod would be gone for good, so leave it alone and say so.
		if len(p.OwnerReferences) == 0 {
			logger.Warnf("cni-dhcp: overlay gateway pod %s/%s has no controller and is left running; restart it by hand to move its lease to the new daemon", p.Namespace, p.Name)
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

// recreateOverlayGatewayPods deletes every Overlay Gateway Pod so its
// controller recreates it against the restarted cni-dhcp daemon and the
// re-rendered NAD: the new leases are then renewed by the daemon and carry
// sendRelease=false. Pods are deleted in batches with a pause in between.
// It returns how many Pods were deleted and how many were left when the
// context ended or an error occurred; the caller reports both.
func recreateOverlayGatewayPods(ctx context.Context, kube kubernetes.Interface) (deleted, remaining int, err error) {
	pods, err := listOverlayGatewayPods(ctx, kube)
	if err != nil {
		return 0, 0, err
	}
	for i := range pods {
		if i > 0 && i%overlayRecreateBatchLimit == 0 {
			if err := overlayRecreateBatchWait(ctx, overlayRecreateBatchPause); err != nil {
				return deleted, len(pods) - i, err
			}
		}
		p := pods[i]
		if err := kube.CoreV1().Pods(p.Namespace).Delete(ctx, p.Name, metav1.DeleteOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return deleted, len(pods) - i, fmt.Errorf("delete overlay pod %s/%s: %w", p.Namespace, p.Name, err)
		}
		deleted++
		logger.Infof("cni-dhcp: recreate (delete) overlay gateway pod %s/%s", p.Namespace, p.Name)
	}
	return deleted, 0, nil
}

// overlayGatewayPodsRecreate is best-effort: a failure is reported loudly but
// never fails the upgrade, because the cluster is already on the new binary
// and the NAD and a Pod that is not recreated only loses LAN reachability once
// its router lease expires.
type overlayGatewayPodsRecreate struct {
	common.KubeAction
}

func (a *overlayGatewayPodsRecreate) Execute(_ connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		logger.Warnf("cni-dhcp: overlay gateway pods NOT recreated (kube client: %v); they lose LAN reachability when their DHCP lease expires unless restarted", err)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), overlayRecreateTimeout)
	defer cancel()

	deleted, remaining, err := recreateOverlayGatewayPods(ctx, kube)
	switch {
	case err != nil:
		logger.Warnf("cni-dhcp: overlay gateway pod recreate stopped after %d pods, %d left (%v); restart the remaining Overlay apps to keep their LAN address renewed", deleted, remaining, err)
	case deleted == 0:
		logger.Infof("cni-dhcp: no overlay gateway pods to recreate")
	default:
		logger.Infof("cni-dhcp: deleted %d overlay gateway pods for recreate", deleted)
	}
	return nil
}

func overlayGatewayRecreateTasks() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "RecreateOverlayGatewayPods",
			Desc:   "Recreate Overlay Gateway pods against the upgraded cni-dhcp",
			Action: new(overlayGatewayPodsRecreate),
		},
	}
}
