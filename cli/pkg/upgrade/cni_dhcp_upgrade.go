package upgrade

import (
	"context"
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
	// underlayNetworkName is the NetworkAttachmentDefinition Overlay Gateway
	// Pods attach through; its presence in the selection means the Pod holds
	// a DHCP lease in the cni-dhcp daemon.
	underlayNetworkName = "underlay-macvlan"
	// overlayRecreateBatchLimit caps deletes per upgrade run.
	overlayRecreateBatchLimit = 50
	overlayRecreateTimeout    = 2 * time.Minute
)

// upgradeCniDhcpBinary installs the cni-plugins archive pinned in the manifest
// into /opt/cni/bin and restarts the cni-dhcp daemon so the new dhcp binary
// (stable client identifier, ipam.sendRelease) is in effect. The generic
// upgrade task set only re-renders the NAD and the systemd unit; the binary
// itself is only replaced by a version upgrader that includes this task.
type upgradeCniDhcpBinary struct {
	common.KubeAction
}

func (u *upgradeCniDhcpBinary) Execute(runtime connector.Runtime) error {
	m, err := manifest.ReadAll(u.KubeConf.Arg.Manifest)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	binary, err := m.Get("cni-plugins")
	if err != nil {
		return fmt.Errorf("get cni-plugins binary info failed: %w", err)
	}

	src := binary.FilePath(runtime.GetBaseDir())
	dst := filepath.Join(common.TmpDir, binary.Filename)
	logger.Debugf("cni-dhcp: copy cni-plugins from %s to %s", src, dst)
	if err := runtime.GetRunner().Scp(src, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync cni-plugins archive failed")
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar -zxf %s -C /opt/cni/bin", dst), false, false); err != nil {
		return errors.Wrap(err, "extract cni-plugins archive failed")
	}
	// The daemon and the plugin are the same binary; a restart picks up the
	// new daemon. In-memory leases are lost (takeover is a later WI), which is
	// why PostUpgrade recreates the Overlay Pods.
	if _, err := runtime.GetRunner().SudoCmd("systemctl restart cni-dhcp", false, false); err != nil {
		return errors.Wrap(err, "restart cni-dhcp failed")
	}
	return nil
}

func cniDhcpBinaryUpgradeTasks() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeCniDhcpBinary",
			Desc:   "Install cni-plugins archive and restart cni-dhcp",
			Action: new(upgradeCniDhcpBinary),
			Retry:  3,
		},
	}
}

// podSelectsUnderlay reports whether the Pod's Multus selection references the
// underlay-macvlan network, in either the JSON or the short-form syntax.
func podSelectsUnderlay(p *corev1.Pod) bool {
	sel, ok := p.Annotations[multusNetworksAnnotation]
	if !ok {
		return false
	}
	return strings.Contains(sel, underlayNetworkName)
}

func listOverlayGatewayPods(ctx context.Context, kube kubernetes.Interface) ([]corev1.Pod, error) {
	pods, err := kube.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var out []corev1.Pod
	for i := range pods.Items {
		p := pods.Items[i]
		if p.DeletionTimestamp != nil {
			continue
		}
		if podSelectsUnderlay(&p) {
			out = append(out, p)
		}
	}
	return out, nil
}

// recreateOverlayGatewayPods deletes Overlay Gateway Pods so their controllers
// recreate them against the restarted cni-dhcp daemon and the re-rendered NAD:
// the new leases are then owned by the daemon (renewals continue) and carry
// sendRelease=false. Without this, the old Pods keep their address but stop
// renewing and lose LAN reachability when the router lease expires.
func recreateOverlayGatewayPods(ctx context.Context, kube kubernetes.Interface) (int, error) {
	pods, err := listOverlayGatewayPods(ctx, kube)
	if err != nil {
		return 0, err
	}
	deleted := 0
	for i := range pods {
		if deleted >= overlayRecreateBatchLimit {
			break
		}
		p := pods[i]
		if err := kube.CoreV1().Pods(p.Namespace).Delete(ctx, p.Name, metav1.DeleteOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return deleted, fmt.Errorf("delete overlay pod %s/%s: %w", p.Namespace, p.Name, err)
		}
		deleted++
		logger.Infof("cni-dhcp: recreate (delete) overlay gateway pod %s/%s", p.Namespace, p.Name)
	}
	return deleted, nil
}

// overlayGatewayPodsRecreate is best-effort and never gates the upgrade.
type overlayGatewayPodsRecreate struct {
	common.KubeAction
}

func (a *overlayGatewayPodsRecreate) Execute(_ connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		logger.Errorf("cni-dhcp: skip overlay pod recreate; kube client: %v", err)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), overlayRecreateTimeout)
	defer cancel()

	n, err := recreateOverlayGatewayPods(ctx, kube)
	if err != nil {
		logger.Errorf("cni-dhcp: overlay pod recreate failed (upgrade continues): %v", err)
		return nil
	}
	if n > 0 {
		logger.Infof("cni-dhcp: deleted %d overlay gateway pods for recreate (batch cap %d)", n, overlayRecreateBatchLimit)
	} else {
		logger.Infof("cni-dhcp: no overlay gateway pods to recreate")
	}
	return nil
}

func overlayGatewayRecreateTasks() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "RecreateOverlayGatewayPods",
			Desc:   "Recreate Overlay Gateway pods against the upgraded cni-dhcp",
			Action: new(overlayGatewayPodsRecreate),
			Retry:  1,
		},
	}
}
