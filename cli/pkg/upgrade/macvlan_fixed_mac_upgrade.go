package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/plugins/network"
	"github.com/beclab/Olares/cli/pkg/terminus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
)

const fixedMACStateFile = "/var/lib/olares/upgrades/overlay-gateway-v1.json"
const fixedMACLabel = "applications.app.bytetrade.io/macvlan-init"
const fixedMACNetworks = "k8s.v1.cni.cncf.io/networks"

var fixedMACNAD = schema.GroupVersionResource{Group: "k8s.cni.cncf.io", Version: "v1", Resource: "network-attachment-definitions"}
var fixedMACApplications = schema.GroupVersionResource{Group: "app.bytetrade.io", Version: "v1alpha1", Resource: "applications"}
var fixedMACAllocations = schema.GroupVersionResource{Group: "app.bytetrade.io", Version: "v1alpha1", Resource: "overlaymacallocations"}

type fixedMACInstance struct {
	Namespace      string
	Name           string
	OldUID         types.UID
	Kind           string
	Workload       string
	WorkloadUID    types.UID
	ReplacementUID types.UID
}
type fixedMACProgress struct {
	Instances []fixedMACInstance
	DaemonPID string
	Complete  bool
	DHCPPhase string `json:",omitempty"`
}

func fixedMACTasks() []task.Interface {
	return []task.Interface{
		&task.LocalTask{Name: "WaitFixedMACSystemComponents", Action: new(terminus.CheckSystemComponentsReady), Retry: 60, Delay: 5 * time.Second},
		&task.LocalTask{Name: "UpgradeFixedMACDHCP", Desc: "Upgrade the existing DHCP service and recover each affected instance", Action: new(upgradeFixedMAC), Retry: 3, Delay: 10 * time.Second},
	}
}

type upgradeFixedMAC struct{ common.KubeAction }

func shellWord(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
func saveFixedMACProgress(p *fixedMACProgress) error {
	if err := os.MkdirAll(filepath.Dir(fixedMACStateFile), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(fixedMACStateFile), ".fixed-mac-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(name, fixedMACStateFile); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(fixedMACStateFile))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
func loadFixedMACProgress() (*fixedMACProgress, error) {
	b, err := os.ReadFile(fixedMACStateFile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var p fixedMACProgress
	if err = json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func fixedMACReady(ctx context.Context, kube kubernetes.Interface, dc dynamic.Interface) error {
	for _, name := range []string{"macvlan-init-webhook", "macvlan-annotation-webhook"} {
		var annotations map[string]string
		if name == "macvlan-init-webhook" {
			v, e := kube.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, name, metav1.GetOptions{})
			if e != nil {
				return e
			}
			annotations = v.Annotations
		} else {
			v, e := kube.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(ctx, name, metav1.GetOptions{})
			if e != nil {
				return e
			}
			annotations = v.Annotations
		}
		if annotations["app.bytetrade.io/overlay-gateway-version"] != "1" {
			return fmt.Errorf("%s has not been upgraded to fixed MAC admission", name)
		}
	}
	_, err := dc.Resource(fixedMACAllocations).List(ctx, metav1.ListOptions{Limit: 1})
	return err
}

// Resolve both Multus annotation forms. Unqualified names belong to the Pod namespace.
func fixedMACSelections(raw, ns string) ([]map[string]interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var list []map[string]interface{}
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &list); err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(raw, "{") {
		var one map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &one); err != nil {
			return nil, err
		}
		list = append(list, one)
	} else {
		for _, token := range strings.Split(raw, ",") {
			name := strings.SplitN(strings.TrimSpace(token), "@", 2)[0]
			list = append(list, map[string]interface{}{"name": name})
		}
	}
	for _, n := range list {
		name, ok := n["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("invalid network selection")
		}
		if strings.Contains(name, "/") {
			parts := strings.Split(name, "/")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid network name")
			}
			n["namespace"], n["name"] = parts[0], parts[1]
		}
		if n["namespace"] == nil || n["namespace"] == "" {
			n["namespace"] = ns
		}
	}
	return list, nil
}
func containsDHCP(v interface{}) bool {
	switch x := v.(type) {
	case map[string]interface{}:
		if x["type"] == "dhcp" {
			return true
		}
		for _, v := range x {
			if containsDHCP(v) {
				return true
			}
		}
	case []interface{}:
		for _, v := range x {
			if containsDHCP(v) {
				return true
			}
		}
	}
	return false
}
func snapshotFixedMAC(ctx context.Context, kube kubernetes.Interface, dc dynamic.Interface, node string) (*fixedMACProgress, error) {
	pods, err := kube.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	p := &fixedMACProgress{}
	for i := range pods.Items {
		pod := &pods.Items[i]

		affected := false
		for _, key := range []string{fixedMACNetworks, "v1.multus-cni.io/default-network"} {
			selections, err := fixedMACSelections(pod.Annotations[key], pod.Namespace)
			if err != nil {
				return nil, err
			}
			for _, s := range selections {
				ns, _ := s["namespace"].(string)
				name, _ := s["name"].(string)
				if ns == "kube-system" && name == "underlay-macvlan" {
					if key != fixedMACNetworks {
						return nil, fmt.Errorf("pod %s/%s uses underlay as its default network", pod.Namespace, pod.Name)
					}
					affected = true
					continue
				}
				nad, err := dc.Resource(fixedMACNAD).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
				if err != nil {
					return nil, fmt.Errorf("inspect %s/%s before DHCP restart: %w", ns, name, err)
				}
				raw, _, _ := unstructured.NestedString(nad.Object, "spec", "config")
				var config interface{}
				if err = json.Unmarshal([]byte(raw), &config); err != nil {
					return nil, err
				}
				if containsDHCP(config) && (pod.Spec.NodeName == "" || pod.Spec.NodeName == node) {
					return nil, fmt.Errorf("pod %s/%s uses another DHCP network %s/%s; migrate it before restarting shared cni-dhcp", pod.Namespace, pod.Name, ns, name)
				}
			}
		}
		if !affected {
			continue
		}
		if pod.Spec.NodeName != "" && pod.Spec.NodeName != node {
			return nil, fmt.Errorf("fixed MAC upgrade requires overlay pods on local master; %s/%s is on %s", pod.Namespace, pod.Name, pod.Spec.NodeName)
		}
		if pod.DeletionTimestamp != nil {
			return nil, fmt.Errorf("wait for terminating overlay pod %s/%s before upgrade", pod.Namespace, pod.Name)
		}
		if pod.Labels[fixedMACLabel] != "true" {
			return nil, fmt.Errorf("pod %s/%s lacks platform macvlan authorization", pod.Namespace, pod.Name)
		}
		owner := metav1.GetControllerOf(pod)
		if owner == nil {
			return nil, fmt.Errorf("pod %s/%s has no controller; cannot safely recreate", pod.Namespace, pod.Name)
		}
		item := fixedMACInstance{Namespace: pod.Namespace, Name: pod.Name, OldUID: pod.UID, Kind: owner.Kind, Workload: owner.Name, WorkloadUID: owner.UID}
		switch owner.Kind {
		case "ReplicaSet":
			rs, err := kube.AppsV1().ReplicaSets(pod.Namespace).Get(ctx, owner.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			d := metav1.GetControllerOf(rs)
			if rs.UID != owner.UID || d == nil || d.Kind != "Deployment" {
				return nil, fmt.Errorf("unverified ReplicaSet owner")
			}
			deployment, err := kube.AppsV1().Deployments(pod.Namespace).Get(ctx, d.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			if deployment.UID != d.UID || (deployment.Spec.Replicas != nil && *deployment.Spec.Replicas > 1) || deployment.Spec.Template.Labels[fixedMACLabel] != "true" {
				return nil, fmt.Errorf("Deployment %s/%s is not a supported fixed MAC singleton", pod.Namespace, d.Name)
			}
			item.Kind, item.Workload, item.WorkloadUID = "Deployment", d.Name, d.UID
		case "StatefulSet":
			sts, err := kube.AppsV1().StatefulSets(pod.Namespace).Get(ctx, owner.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			if sts.UID != owner.UID || sts.Spec.Template.Labels[fixedMACLabel] != "true" {
				return nil, fmt.Errorf("unverified StatefulSet owner")
			}
		default:
			return nil, fmt.Errorf("unsupported fixed MAC workload %s", owner.Kind)
		}
		for _, existing := range p.Instances {
			if sameFixedMACInstance(existing, item) {
				return nil, fmt.Errorf("workload %s/%s still has overlapping pods; wait for rollout before fixed MAC upgrade", item.Namespace, item.Workload)
			}
		}
		p.Instances = append(p.Instances, item)
	}
	if err := includeDesiredFixedMAC(ctx, kube, dc, p); err != nil {
		return nil, err
	}
	return p, nil
}

func prepareFixedMACDeployments(ctx context.Context, kube kubernetes.Interface, p *fixedMACProgress) error {
	for _, item := range p.Instances {
		if item.Kind != "Deployment" {
			continue
		}
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			d, e := kube.AppsV1().Deployments(item.Namespace).Get(ctx, item.Workload, metav1.GetOptions{})
			if e != nil {
				return e
			}
			if d.UID != item.WorkloadUID {
				return fmt.Errorf("Deployment UID changed")
			}
			d.Spec.Strategy = appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType}
			_, e = kube.AppsV1().Deployments(item.Namespace).Update(ctx, d, metav1.UpdateOptions{})
			return e
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func upgradeFixedMACNAD(ctx context.Context, dc dynamic.Interface) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nad, err := dc.Resource(fixedMACNAD).Namespace("kube-system").Get(ctx, "underlay-macvlan", metav1.GetOptions{})
		if err != nil {
			return err
		}
		raw, _, _ := unstructured.NestedString(nad.Object, "spec", "config")
		var c map[string]interface{}
		if err = json.Unmarshal([]byte(raw), &c); err != nil {
			return err
		}
		if c["type"] != "macvlan" || (c["master"] != "br-olares" && c["master"] != network.OverlayParentAltname) {
			return fmt.Errorf("overlay upgrade expects the platform bridge or wired-parent macvlan NAD")
		}
		ipam, ok := c["ipam"].(map[string]interface{})
		if !ok || ipam["type"] != "dhcp" {
			return fmt.Errorf("underlay is not DHCP")
		}
		c["master"] = network.OverlayParentAltname
		ipam["sendRelease"] = false
		b, err := json.Marshal(c)
		if err != nil {
			return err
		}
		if err = unstructured.SetNestedField(nad.Object, string(b), "spec", "config"); err != nil {
			return err
		}
		_, err = dc.Resource(fixedMACNAD).Namespace("kube-system").Update(ctx, nad, metav1.UpdateOptions{})
		return err
	})
}

func fixedMACPodReady(ctx context.Context, dc dynamic.Interface, pod *corev1.Pod) (bool, error) {
	if pod.DeletionTimestamp != nil || pod.Status.Phase != corev1.PodRunning {
		return false, nil
	}
	ready := false
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			ready = true
		}
	}
	if !ready {
		return false, nil
	}
	selections, err := fixedMACSelections(pod.Annotations[fixedMACNetworks], pod.Namespace)
	if err != nil {
		return false, err
	}
	mac := ""
	for _, s := range selections {
		if s["name"] == "underlay-macvlan" && s["namespace"] == "kube-system" {
			mac, _ = s["mac"].(string)
		}
	}
	if _, err := net.ParseMAC(mac); err != nil {
		return false, nil
	}
	var statuses []struct {
		Interface string   `json:"interface"`
		IPs       []string `json:"ips"`
		MAC       string   `json:"mac"`
	}
	if err = json.Unmarshal([]byte(pod.Annotations["k8s.v1.cni.cncf.io/network-status"]), &statuses); err != nil {
		return false, nil
	}
	lan := false
	for _, s := range statuses {
		if s.Interface == "net1" && strings.EqualFold(s.MAC, mac) {
			for _, ip := range s.IPs {
				if net.ParseIP(ip) != nil {
					lan = true
				}
			}
		}
	}
	if !lan {
		return false, nil
	}
	claim, err := dc.Resource(fixedMACAllocations).Get(ctx, strings.ReplaceAll(mac, ":", ""), metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	phase, _, _ := unstructured.NestedString(claim.Object, "spec", "phase")
	ref, _, _ := unstructured.NestedString(claim.Object, "spec", "applicationRef")
	uid, _, _ := unstructured.NestedString(claim.Object, "spec", "applicationUID")
	app, err := dc.Resource(fixedMACApplications).Get(ctx, ref, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	namespace, _, _ := unstructured.NestedString(app.Object, "spec", "namespace")
	appName, _, _ := unstructured.NestedString(app.Object, "spec", "name")
	if string(app.GetUID()) != uid || namespace != pod.Namespace || appName != pod.Labels["applications.app.bytetrade.io/name"] {
		return false, fmt.Errorf("recreated Pod MAC belongs to another application")
	}
	return phase == "Bound", nil
}

func replacementFixedMAC(ctx context.Context, kube kubernetes.Interface, dc dynamic.Interface, item fixedMACInstance) (*corev1.Pod, error) {
	// The old UID must have disappeared even if another pod is already Ready.
	if item.OldUID != "" {
		old, err := kube.CoreV1().Pods(item.Namespace).Get(ctx, item.Name, metav1.GetOptions{})
		if err == nil && old.UID == item.OldUID {
			return nil, nil
		}
		if err != nil && !apierrors.IsNotFound(err) {
			return nil, err
		}
	}
	pods, err := kube.CoreV1().Pods(item.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for i := range pods.Items {
		pod := &pods.Items[i]
		owner := metav1.GetControllerOf(pod)
		if owner == nil || pod.UID == item.OldUID {
			continue
		}
		matches := false
		if item.Kind == "StatefulSet" {
			matches = owner.Kind == "StatefulSet" && owner.UID == item.WorkloadUID && pod.Name == item.Name
		} else if owner.Kind == "ReplicaSet" {
			rs, e := kube.AppsV1().ReplicaSets(item.Namespace).Get(ctx, owner.Name, metav1.GetOptions{})
			if e != nil {
				return nil, e
			}
			d := metav1.GetControllerOf(rs)
			matches = rs.UID == owner.UID && d != nil && d.UID == item.WorkloadUID
		}
		if matches {
			ready, e := fixedMACPodReady(ctx, dc, pod)
			if e != nil {
				return nil, e
			}
			if ready {
				return pod, nil
			}
		}
	}
	return nil, nil
}
func recoverFixedMAC(ctx context.Context, kube kubernetes.Interface, dc dynamic.Interface, p *fixedMACProgress, save func(*fixedMACProgress) error) error {
	for i := range p.Instances {
		item := &p.Instances[i]
		if item.ReplacementUID != "" {
			pod, err := replacementFixedMAC(ctx, kube, dc, *item)
			if err != nil {
				return err
			}
			if pod != nil {
				continue
			}
			item.ReplacementUID = ""
			if err = save(p); err != nil {
				return err
			}
		}
		if item.OldUID != "" {
			if err := evictFixedMACPod(ctx, kube, *item, 2*time.Second); err != nil {
				return err
			}
		}
		deadline, cancel := context.WithTimeout(ctx, 5*time.Minute)
		for {
			replacement, e := replacementFixedMAC(deadline, kube, dc, *item)
			if e != nil {
				cancel()
				return e
			}
			if replacement != nil {
				item.ReplacementUID = replacement.UID
				if e = save(p); e != nil {
					cancel()
					return e
				}
				break
			}
			select {
			case <-deadline.Done():
				cancel()
				return fmt.Errorf("waiting for recovered fixed MAC instance %s/%s: %w", item.Namespace, item.Name, deadline.Err())
			case <-time.After(2 * time.Second):
			}
		}
		cancel()
	}
	return nil
}

// Stage only the published DHCP binary; other CNI executables remain untouched.
func stageFixedMACDHCP(runtime connector.Runtime, manifestPath string) (string, error) {
	hashes := map[string]string{"amd64": "fd946fea15d15cce9e62e295c974288954a0e283fe55225c43e26f3157fba083", "arm64": "855e62230e7622641853559c038164fc6d046c87157d331d8f0d337ad1fa7bae"}
	hash := hashes[runtime.GetSystemInfo().GetOsArch()]
	if hash == "" {
		return "", fmt.Errorf("unsupported fixed MAC DHCP architecture")
	}
	m, err := manifest.ReadAll(manifestPath)
	if err != nil {
		return "", err
	}
	binary, err := m.Get("cni-plugins")
	if err != nil {
		return "", err
	}
	if !strings.Contains(binary.Filename, "v1.6.2-olares2") {
		return "", fmt.Errorf("manifest must contain published v1.6.2-olares2")
	}
	dst := filepath.Join(common.TmpDir, binary.Filename)
	if err = runtime.GetRunner().Scp(binary.FilePath(runtime.GetBaseDir()), dst); err != nil {
		return "", err
	}
	staged := dhcpStaged
	// Copy to a root-only directory BEFORE verifying; /tmp is not trusted.
	secureArchive := "/opt/cni/.fixed-mac-v1/archive.tgz"
	if _, err = runtime.GetRunner().SudoCmd("test ! -L /opt/cni/.fixed-mac-v1 && install -d -o 0 -g 0 -m 0700 /opt/cni/.fixed-mac-v1 && rm -f "+secureArchive+" && install -o 0 -g 0 -m 0600 "+shellWord(dst)+" "+secureArchive, false, false); err != nil {
		return "", err
	}
	output, err := runtime.GetRunner().SudoCmd("sha256sum "+secureArchive, false, false)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(output)
	if len(fields) == 0 || fields[0] != hash {
		return "", fmt.Errorf("cni-plugins archive SHA-256 differs from published v1.6.2-olares2")
	}
	verify, err := secureDHCPVerifyCommand(runtime.GetSystemInfo().GetOsArch())
	if err != nil {
		return "", err
	}
	// Stream the member into a new root-owned inode, never restore archive ownership.
	commands := []string{secureDHCPExtractCommand(secureArchive), verify, staged + " --version"}
	for _, cmd := range commands {
		if _, err = runtime.GetRunner().SudoCmd(dhcpPrivilegedCommand(cmd), false, false); err != nil {
			return "", err
		}
	}
	return staged, nil
}
func setFixedMACGate(ctx context.Context, kube kubernetes.Interface, phase string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cm, err := kube.CoreV1().ConfigMaps("kube-system").Get(ctx, "olares-fixed-mac-upgrade", metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			_, err = kube.CoreV1().ConfigMaps("kube-system").Create(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "olares-fixed-mac-upgrade", Namespace: "kube-system"}, Data: map[string]string{"phase": phase}}, metav1.CreateOptions{})
			return err
		}
		if err != nil {
			return err
		}
		cm.Data = map[string]string{"phase": phase}
		_, err = kube.CoreV1().ConfigMaps("kube-system").Update(ctx, cm, metav1.UpdateOptions{})
		return err
	})
}
func daemonGeneration(runtime connector.Runtime) (string, error) {
	out, err := runtime.GetRunner().SudoCmd("systemctl is-active --quiet cni-dhcp && systemctl show cni-dhcp -p InvocationID --value", false, false)
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", fmt.Errorf("cni-dhcp invocation ID is unavailable")
	}
	return out, nil
}
func (a *upgradeFixedMAC) Execute(runtime connector.Runtime) error {
	if err := network.ResumeOverlayMigration(runtime); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return err
	}
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}
	p, err := loadFixedMACProgress()
	if err != nil {
		return err
	}
	// This is a versioned migration, not a daemon watchdog. A completed migration
	// never restarts DHCP or deletes applications again on a later release upgrade.
	if p != nil && p.Complete {
		if err := (&network.ClearOverlayMigrationMarker{}).Execute(runtime); err != nil {
			return err
		}
		return setFixedMACGate(ctx, kube, "complete")
	}
	if err = fixedMACReady(ctx, kube, dc); err != nil {
		return err
	}
	// Continue applying the target version after an interrupted attempt.
	if p != nil && p.DHCPPhase == "prepared" {
		if err = upgradeDHCPBinary(ctx, p, runtimeDHCPCommands(runtime), saveFixedMACProgress); err != nil {
			return err
		}
	}
	if p != nil && p.DHCPPhase == "" && p.DaemonPID == "" {
		p = nil
	}
	if p == nil {
		// Reject unsupported topology before changing admission or the DHCP service.
		preview, e := snapshotFixedMAC(ctx, kube, dc, runtime.GetSystemInfo().GetHostname())
		if e != nil {
			return e
		}
		if err = checkFixedMACLegacy(ctx, kube, dc, preview); err != nil {
			return err
		}
		if err = checkFixedMACEvictions(ctx, kube, preview); err != nil {
			return err
		}
		_, e = stageFixedMACDHCP(runtime, a.KubeConf.Arg.Manifest)
		if e != nil {
			return e
		}
		if err = setFixedMACGate(ctx, kube, "paused"); err != nil {
			return err
		}
		interrupted := false
		defer func() {
			if !interrupted {
				cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				_ = setFixedMACGate(cleanup, kube, "preflight-failed")
			}
		}()
		// Drain admission already in flight (both MAC webhooks have a 30s deadline).
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(35 * time.Second):
		}
		p, err = snapshotFixedMAC(ctx, kube, dc, runtime.GetSystemInfo().GetHostname())
		if err != nil {
			return err
		}
		if err = saveFixedMACProgress(p); err != nil {
			return err
		}
		// Only the durable topology phase commits the migration.
		defer func() { interrupted = p.DHCPPhase != "" }()
	}
	if p.DHCPPhase == "" {
		if err = setFixedMACGate(ctx, kube, "paused"); err != nil {
			return err
		}
		if err = checkFixedMACLegacy(ctx, kube, dc, p); err != nil {
			return err
		}
		if err = checkFixedMACEvictions(ctx, kube, p); err != nil {
			return err
		}
		if err = prepareFixedMACDeployments(ctx, kube, p); err != nil {
			return err
		}
		p.DHCPPhase = "topology"
		if err = saveFixedMACProgress(p); err != nil {
			return err
		}
	}
	if p.DHCPPhase == "topology" {
		if err = a.migrateOverlayParent(runtime); err != nil {
			return err
		}
		if len(p.Instances) > 0 {
			if _, err = runtime.GetRunner().SudoCmd("ip -o link show "+network.OverlayParentAltname, false, false); err != nil {
				return fmt.Errorf("overlay parent is not ready: %w", err)
			}
		}
		if err = upgradeFixedMACNAD(ctx, dc); err != nil {
			return err
		}
		if err = upgradeDHCPBinary(ctx, p, runtimeDHCPCommands(runtime), saveFixedMACProgress); err != nil {
			return err
		}

	} else {
		generation, e := ensureUpgradeDHCPRunning(ctx, runtimeDHCPCommands(runtime))
		if e != nil {
			return e
		}
		if generation != p.DaemonPID {
			// A crash/restart loses the leases of completed items too. Re-enrol the
			// current instances; do not leave a hidden, delayed lease-expiration fault.
			current, e := snapshotFixedMAC(ctx, kube, dc, runtime.GetSystemInfo().GetHostname())
			if e != nil {
				return e
			}
			// Keep expected instances absent between old deletion and replacement creation.
			for _, previous := range p.Instances {
				found := false
				for _, now := range current.Instances {
					if sameFixedMACInstance(previous, now) {
						found = true
						break
					}
				}
				if !found {
					previous.ReplacementUID = ""
					current.Instances = append(current.Instances, previous)
				}
			}
			p = current
			p.DaemonPID = generation
			p.DHCPPhase = "ready"
			if err = saveFixedMACProgress(p); err != nil {
				return err
			}
		}
	}
	if err = setFixedMACGate(ctx, kube, "recovering"); err != nil {
		return err
	}
	if err = recoverFixedMAC(ctx, kube, dc, p, saveFixedMACProgress); err != nil {
		return err
	}
	// Recheck completed instances, including those restored by an earlier attempt.
	for _, item := range p.Instances {
		pod, e := replacementFixedMAC(ctx, kube, dc, item)
		if e != nil {
			return e
		}
		if pod == nil {
			return fmt.Errorf("instance %s/%s is no longer in fixed MAC steady state", item.Namespace, item.Name)
		}
	}
	generation, err := daemonGeneration(runtime)
	if err != nil {
		return err
	}
	if generation != p.DaemonPID {
		return fmt.Errorf("DHCP restarted during recovery; retry to recover all current leases")
	}
	p.Complete = true
	if err = saveFixedMACProgress(p); err != nil {
		return err
	}
	if err := (&network.ClearOverlayMigrationMarker{}).Execute(runtime); err != nil {
		return err
	}
	return setFixedMACGate(ctx, kube, "complete")
}

func sameFixedMACInstance(a, b fixedMACInstance) bool {
	return a.Namespace == b.Namespace && a.Kind == b.Kind && a.WorkloadUID == b.WorkloadUID && (a.Kind == "Deployment" || a.Name == b.Name)
}

// Check disruption policy before restarting the sole DHCP service. This checks
// admission/PDB semantics without deleting Pods or consuming a disruption.
func checkFixedMACEvictions(ctx context.Context, kube kubernetes.Interface, p *fixedMACProgress) error {
	if err := checkFixedMACDisruptionBudget(ctx, kube, p); err != nil {
		return err
	}
	for _, item := range p.Instances {
		if item.OldUID == "" {
			continue
		}
		uid := item.OldUID
		err := kube.PolicyV1().Evictions(item.Namespace).Evict(ctx, &policyv1.Eviction{ObjectMeta: metav1.ObjectMeta{Name: item.Name, Namespace: item.Namespace}, DeleteOptions: &metav1.DeleteOptions{DryRun: []string{metav1.DryRunAll}, Preconditions: &metav1.Preconditions{UID: &uid}}})
		if err != nil {
			return fmt.Errorf("fixed MAC preflight cannot safely evict %s/%s: %w", item.Namespace, item.Name, err)
		}
	}
	return nil
}

// Plan enabled desired instances even if their replacement Pod does not exist yet.
func includeDesiredFixedMAC(ctx context.Context, kube kubernetes.Interface, dc dynamic.Interface, p *fixedMACProgress) error {
	apps, err := dc.Resource(fixedMACApplications).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	enabled := map[string]bool{}
	for _, app := range apps.Items {
		on, _, _ := unstructured.NestedString(app.Object, "spec", "settings", "enableOverlayGateway")
		if on != "true" {
			continue
		}
		ns, _, _ := unstructured.NestedString(app.Object, "spec", "namespace")
		name, _, _ := unstructured.NestedString(app.Object, "spec", "name")
		enabled[ns+"/"+name] = true
	}
	wanted := func(ns string, t corev1.PodTemplateSpec) bool {
		return t.Labels[fixedMACLabel] == "true" && enabled[ns+"/"+t.Labels["applications.app.bytetrade.io/name"]]
	}
	add := func(item fixedMACInstance) {
		for _, existing := range p.Instances {
			if sameFixedMACInstance(existing, item) {
				return
			}
		}
		p.Instances = append(p.Instances, item)
	}
	deployments, err := kube.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, d := range deployments.Items {
		if !wanted(d.Namespace, d.Spec.Template) {
			continue
		}
		replicas := int32(1)
		if d.Spec.Replicas != nil {
			replicas = *d.Spec.Replicas
		}
		if replicas == 0 {
			continue
		}
		if replicas != 1 {
			return fmt.Errorf("fixed MAC Deployment %s/%s has %d replicas", d.Namespace, d.Name, replicas)
		}
		add(fixedMACInstance{Namespace: d.Namespace, Kind: "Deployment", Workload: d.Name, WorkloadUID: d.UID})
	}
	sets, err := kube.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, sts := range sets.Items {
		if !wanted(sts.Namespace, sts.Spec.Template) {
			continue
		}
		replicas := int32(1)
		if sts.Spec.Replicas != nil {
			replicas = *sts.Spec.Replicas
		}
		start := int32(0)
		if sts.Spec.Ordinals != nil {
			start = sts.Spec.Ordinals.Start
		}
		for i := start; i < start+replicas; i++ {
			add(fixedMACInstance{Namespace: sts.Namespace, Name: sts.Name + "-" + strconv.Itoa(int(i)), Kind: "StatefulSet", Workload: sts.Name, WorkloadUID: sts.UID})
		}
	}
	return nil
}

// The admission gate and durable instance snapshot precede all host network changes.
func (a *upgradeFixedMAC) migrateOverlayParent(runtime connector.Runtime) error {
	if err := (&network.MigrateBridgeToDirect{}).Execute(runtime); err != nil {
		return err
	}
	migrated, err := network.OverlayMigratedFromBridge(runtime)
	if err != nil {
		return err
	}
	if migrated {
		hosts := &terminus.UpdateKubeKeyHosts{KubeAction: a.KubeAction}
		if err := hosts.Execute(runtime); err != nil {
			return err
		}
	}
	if err := (&network.EnsureOverlayAltname{}).Execute(runtime); err != nil {
		return err
	}
	return (&network.WriteOverlayDesiredIfMigrated{}).Execute(runtime)
}

// Restore only an already-started migration before Kubernetes health prechecks.
type resumeFixedMACNetwork struct{ common.KubeAction }

func (a *resumeFixedMACNetwork) Execute(runtime connector.Runtime) error {
	return network.ResumeOverlayMigration(runtime)
}
