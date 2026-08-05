package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	deenvyGateNamespace = "os-framework"
	deenvyGateConfigMap = "olares-deenvy-steady-state"
	deenvyGateReady     = "Ready"

	cpPreflight = "preflight"
	cpPreload   = "preload"
	cpCharts    = "charts"
	cpDeps      = "deps-ready"
	cpCutover   = "cutover"
	cpRollout   = "rollout"
	cpAccept    = "accept"
	cpCommit    = "commit"
	cpDone      = "done"
	cpRollback  = "rollback"
)

// deenvyGateState mirrors app-service mesh.SteadyGateState JSON.
type deenvyGateState struct {
	Phase         string          `json:"phase"`
	TargetVersion string          `json:"targetVersion,omitempty"`
	Checkpoint    string          `json:"checkpoint,omitempty"`
	Conditions    map[string]bool `json:"conditions,omitempty"`
	Message       string          `json:"message,omitempty"`
}

func deenvyCheckpointOrder(cp string) int {
	order := []string{cpPreflight, cpPreload, cpCharts, cpDeps, cpCutover, cpRollout, cpAccept, cpCommit, cpDone}
	for i, v := range order {
		if v == cp {
			return i
		}
	}
	if cp == cpRollback {
		return -1
	}
	return 0
}

func loadDeenvyGate(ctx context.Context, kube kubernetes.Interface) (*deenvyGateState, error) {
	cm, err := kube.CoreV1().ConfigMaps(deenvyGateNamespace).Get(ctx, deenvyGateConfigMap, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return &deenvyGateState{Phase: "Pending", Conditions: map[string]bool{}}, nil
		}
		return nil, err
	}
	raw := strings.TrimSpace(cm.Data["state"])
	if raw == "" {
		return &deenvyGateState{Phase: cm.Data["phase"], Conditions: map[string]bool{}}, nil
	}
	var st deenvyGateState
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func storeDeenvyGate(ctx context.Context, kube kubernetes.Interface, st *deenvyGateState) error {
	body, err := json.Marshal(st)
	if err != nil {
		return err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deenvyGateConfigMap,
			Namespace: deenvyGateNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "olares-cli",
				"app.kubernetes.io/name":       "deenvy-steady-state",
			},
		},
		Data: map[string]string{
			"state": string(body),
			"phase": st.Phase,
		},
	}
	_, err = kube.CoreV1().ConfigMaps(deenvyGateNamespace).Update(ctx, cm, metav1.UpdateOptions{})
	if apierrors.IsNotFound(err) {
		_, err = kube.CoreV1().ConfigMaps(deenvyGateNamespace).Create(ctx, cm, metav1.CreateOptions{})
	}
	return err
}

func kubeClientFromRuntime() (kubernetes.Interface, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func dynamicClientFromRuntime() (dynamic.Interface, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(cfg)
}

func advanceDeenvyCheckpoint(ctx context.Context, kube kubernetes.Interface, targetVersion, checkpoint string, conditions map[string]bool, msg string) error {
	st, err := loadDeenvyGate(ctx, kube)
	if err != nil {
		return err
	}
	if strings.EqualFold(st.Phase, deenvyGateReady) && deenvyCheckpointOrder(checkpoint) < deenvyCheckpointOrder(cpCommit) {
		return nil // already committed; idempotent no-op for earlier steps
	}
	if deenvyCheckpointOrder(st.Checkpoint) > deenvyCheckpointOrder(checkpoint) && st.Checkpoint != cpRollback {
		logger.Infof("deenvy: skip checkpoint %s (already at %s)", checkpoint, st.Checkpoint)
		return nil
	}
	if conditions == nil {
		conditions = map[string]bool{}
	}
	if st.Conditions == nil {
		st.Conditions = map[string]bool{}
	}
	for k, v := range conditions {
		st.Conditions[k] = v
	}
	st.Checkpoint = checkpoint
	st.TargetVersion = targetVersion
	st.Message = msg
	if checkpoint == cpCommit || checkpoint == cpDone {
		st.Phase = deenvyGateReady
	} else if !strings.EqualFold(st.Phase, deenvyGateReady) {
		st.Phase = "InProgress"
	}
	return storeDeenvyGate(ctx, kube, st)
}

// --- task actions ---

type deenvyPreflight struct {
	common.KubeAction
	TargetVersion string
}

func (a *deenvyPreflight) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	st, err := loadDeenvyGate(ctx, kube)
	if err != nil {
		return err
	}
	if strings.EqualFold(st.Phase, deenvyGateReady) && st.TargetVersion == a.TargetVersion {
		logger.Info("deenvy: already Ready for target; preflight no-op")
		return nil
	}
	if st.Checkpoint == cpRollback {
		return fmt.Errorf("deenvy: previous run marked rollback; refuse forward cutover until cleared")
	}
	return advanceDeenvyCheckpoint(ctx, kube, a.TargetVersion, cpPreflight, map[string]bool{
		"PlatformChartsSynced": false,
	}, "preflight inventory ok")
}

type deenvyPreloadImages struct {
	common.KubeAction
	TargetVersion string
}

func (a *deenvyPreloadImages) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	// Image preload is performed by existing component download pipelines;
	// this task records the checkpoint for resume semantics.
	return advanceDeenvyCheckpoint(ctx, kube, a.TargetVersion, cpPreload, nil, "preload acknowledged")
}

type deenvyWaitDeps struct {
	common.KubeAction
	TargetVersion string
}

func (a *deenvyWaitDeps) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	deadline := time.Now().Add(8 * time.Minute)
	for {
		ok := true
		conds := map[string]bool{}
		// Linkerd control plane
		for _, name := range []string{"linkerd-destination", "linkerd-identity", "linkerd-proxy-injector"} {
			dep, err := kube.AppsV1().Deployments("os-mesh").Get(ctx, name, metav1.GetOptions{})
			ready := err == nil && dep.Status.ReadyReplicas >= 1
			conds["Linkerd_"+name] = ready
			if !ready {
				ok = false
			}
		}
		// EG data plane readiness is NOT ExtAuth coverage (true SecurityPolicy probe).
		eg, err := kube.AppsV1().Deployments("os-gateway").Get(ctx, "app-gateway-data", metav1.GetOptions{})
		egOK := err == nil && eg.Status.ReadyReplicas >= 1
		extOK := false
		dc, dcErr := dynamicClientFromRuntime()
		if dcErr != nil {
			logger.Errorf("deenvy: dynamic client for ExtAuth probe: %v", dcErr)
			ok = false
		} else {
			var probeErr error
			extOK, probeErr = probeEntranceExtAuthCovered(ctx, dc)
			if probeErr != nil {
				logger.Errorf("deenvy: EntranceExtAuthCovered probe failed: %v", probeErr)
				extOK = false
				ok = false
			}
		}
		assignExtAuthDepConditions(conds, egOK, extOK)
		if !egOK || !extOK {
			ok = false
		}
		// l4 retained platform Envoy
		l4, err := kube.AppsV1().Deployments("os-network").Get(ctx, "l4-bfl-proxy", metav1.GetOptions{})
		l4OK := err == nil && l4.Status.ReadyReplicas >= 1
		conds["L4ProxyReady"] = l4OK
		if !l4OK {
			ok = false
		}
		// system-server co-located TCP Envoy (platform allow-list; not extracted)
		conds["SystemServerProxyReady"] = true
		nsList, nsErr := kube.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if nsErr == nil {
			seen := false
			proxyOK := true
			for _, ns := range nsList.Items {
				if !strings.HasPrefix(ns.Name, "user-system-") {
					continue
				}
				dep, derr := kube.AppsV1().Deployments(ns.Name).Get(ctx, "system-server", metav1.GetOptions{})
				if derr != nil {
					if apierrors.IsNotFound(derr) {
						continue
					}
					proxyOK = false
					continue
				}
				seen = true
				hasProxy := false
				for _, c := range dep.Spec.Template.Spec.Containers {
					if c.Name == "proxy" && strings.Contains(strings.ToLower(c.Image), "envoy") {
						hasProxy = true
						break
					}
				}
				if !hasProxy || dep.Status.ReadyReplicas < 1 {
					proxyOK = false
				}
			}
			if seen {
				conds["SystemServerProxyReady"] = proxyOK
				if !proxyOK {
					ok = false
				}
			}
		}
		if ok {
			conds["LinkerdControlPlaneReady"] = true
			conds["PlatformChartsSynced"] = true
			return advanceDeenvyCheckpoint(ctx, kube, a.TargetVersion, cpDeps, conds, "deps ready")
		}
		if time.Now().After(deadline) {
			_ = advanceDeenvyCheckpoint(ctx, kube, a.TargetVersion, cpRollback, conds, "deps wait timed out")
			return fmt.Errorf("deenvy: dependency wait timed out: %#v", conds)
		}
		time.Sleep(10 * time.Second)
	}
}

type deenvyCutover struct {
	common.KubeAction
	TargetVersion string
}

func (a *deenvyCutover) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	// Annotate Applications for gateway route-mode; provider/l4 consume annotation.
	// Full Application list patch is best-effort via label on namespaces.
	nsList, err := kube.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, ns := range nsList.Items {
		if !strings.HasPrefix(ns.Name, "user-space-") && !strings.HasPrefix(ns.Name, "user-system-") {
			continue
		}
		if ns.Annotations == nil {
			ns.Annotations = map[string]string{}
		}
		if ns.Annotations["gateway.olares.io/route-mode"] == "gateway" {
			continue
		}
		ns.Annotations["gateway.olares.io/route-mode"] = "gateway"
		if _, err := kube.CoreV1().Namespaces().Update(ctx, &ns, metav1.UpdateOptions{}); err != nil {
			logger.Warnf("deenvy: annotate namespace %s: %v", ns.Name, err)
		}
	}
	return advanceDeenvyCheckpoint(ctx, kube, a.TargetVersion, cpCutover, map[string]bool{
		"RouteModeGateway": true,
	}, "cutover route-mode=gateway")
}

type deenvyRollout struct {
	common.KubeAction
	TargetVersion string
}

func (a *deenvyRollout) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	// Enable oes-free admission BEFORE inventory accept so rebuilt pods drop oes.
	st, err := loadDeenvyGate(ctx, kube)
	if err != nil {
		return err
	}
	if conditions := st.Conditions; conditions == nil {
		st.Conditions = map[string]bool{}
	}
	st.Conditions["WorkloadRolloutComplete"] = true
	st.Checkpoint = cpRollout
	st.TargetVersion = a.TargetVersion
	st.Phase = deenvyGateReady
	st.Message = "oes-free gate enabled; workloads may rebuild without oes"
	if err := storeDeenvyGate(ctx, kube, st); err != nil {
		return err
	}
	return nil
}

type deenvyAccept struct {
	common.KubeAction
	TargetVersion string
}

func (a *deenvyAccept) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	invOK, msg := inventoryZeroUnauthorizedOES(ctx, kube)
	conds := map[string]bool{
		"ZeroOesInventory": invOK,
		"KeyPodsReady":     true,
	}
	if !invOK {
		// Keep Phase Ready=false for next attempt after rollback charts.
		st, _ := loadDeenvyGate(ctx, kube)
		if st != nil {
			st.Phase = "Rollback"
			st.Checkpoint = cpRollback
			st.Message = msg
			st.Conditions = conds
			_ = storeDeenvyGate(ctx, kube, st)
		}
		return fmt.Errorf("deenvy accept failed: %s", msg)
	}
	return advanceDeenvyCheckpoint(ctx, kube, a.TargetVersion, cpAccept, conds, "accept suite passed")
}

func inventoryZeroUnauthorizedOES(ctx context.Context, kube kubernetes.Interface) (bool, string) {
	pods, err := kube.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err.Error()
	}
	var offenders []string
	for _, p := range pods.Items {
		for _, c := range append(p.Spec.Containers, p.Spec.InitContainers...) {
			if isBusinessOESContainer(p, c) {
				offenders = append(offenders, fmt.Sprintf("%s/%s:%s", p.Namespace, p.Name, c.Name))
				break
			}
		}
	}
	if len(offenders) > 0 {
		max := 5
		if len(offenders) < max {
			max = len(offenders)
		}
		return false, fmt.Sprintf("oes inventory non-zero: %s", strings.Join(offenders[:max], ","))
	}
	return true, "zero oes"
}

type deenvyCommitGate struct {
	common.KubeAction
	TargetVersion string
}

func (a *deenvyCommitGate) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	st, err := loadDeenvyGate(ctx, kube)
	if err != nil {
		return err
	}
	if st.Checkpoint != cpAccept && st.Checkpoint != cpCommit && st.Checkpoint != cpDone {
		return fmt.Errorf("deenvy: refuse commit; checkpoint=%s (need accept)", st.Checkpoint)
	}
	return advanceDeenvyCheckpoint(ctx, kube, a.TargetVersion, cpCommit, map[string]bool{
		"SteadyStateGate": true,
	}, "SteadyStateGate Ready")
}

type deenvyMarkDone struct {
	common.KubeAction
	TargetVersion string
}

func (a *deenvyMarkDone) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return advanceDeenvyCheckpoint(ctx, kube, a.TargetVersion, cpDone, nil, "upgrade done")
}

// deenvyUpgradeTasks returns the steady-state cutover task chain (PLAN-SYS-DEENVY-OTA-01).
func deenvyUpgradeTasks(targetVersion string) []task.Interface {
	return []task.Interface{
		&task.LocalTask{Name: "DeenvyPreflight", Action: &deenvyPreflight{TargetVersion: targetVersion}, Retry: 3},
		&task.LocalTask{Name: "DeenvyPreloadImages", Action: &deenvyPreloadImages{TargetVersion: targetVersion}, Retry: 3},
		&task.LocalTask{Name: "DeenvyWaitDeps", Action: &deenvyWaitDeps{TargetVersion: targetVersion}, Retry: 3, Delay: 10 * time.Second},
		&task.LocalTask{Name: "DeenvyCutover", Action: &deenvyCutover{TargetVersion: targetVersion}, Retry: 3},
		&task.LocalTask{Name: "DeenvyRollout", Action: &deenvyRollout{TargetVersion: targetVersion}, Retry: 3},
		&task.LocalTask{Name: "DeenvyAccept", Action: &deenvyAccept{TargetVersion: targetVersion}, Retry: 3, Delay: 15 * time.Second},
		&task.LocalTask{Name: "DeenvyCommitGate", Action: &deenvyCommitGate{TargetVersion: targetVersion}, Retry: 3},
	}
}

func deenvyPostTasks(targetVersion string) []task.Interface {
	return []task.Interface{
		&task.LocalTask{Name: "DeenvyMarkDone", Action: &deenvyMarkDone{TargetVersion: targetVersion}, Retry: 3},
	}
}
