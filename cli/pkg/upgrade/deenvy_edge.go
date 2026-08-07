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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	deenvyGateNamespace = "os-framework"
	deenvyGateConfigMap = "olares-deenvy-steady-state"
	deenvyGateReady     = "Ready"

	conditionL4ProxyReady = "L4ProxyReady"

	cpAccept = "accept"
	cpCommit = "commit"
	cpDone   = "done"

	authKindLabel               = "gateway.olares.io/auth-kind"
	authKindEntranceExtAuth     = "entrance-ext-auth"
	authKindEntranceCookie      = "entrance-cookie"
	authKindEntranceProbeBypass = "entrance-probe-bypass"

	entranceExtAuthSuffix       = "-entrance-ext-auth"
	entranceCookieSuffix        = "-entrance-cookie"
	entranceProbeBypassSuffix   = "-probe-bypass"
	entranceProbePolicySuffix   = "-entrance-probe"
	securityPolicyJWTAuthnSuffix = "-jwt-authn"
)

var (
	edgeSecurityPolicyGVR = schema.GroupVersionResource{
		Group: "gateway.envoyproxy.io", Version: "v1alpha1", Resource: "securitypolicies",
	}
	edgeExtensionPolicyGVR = schema.GroupVersionResource{
		Group: "gateway.envoyproxy.io", Version: "v1alpha1", Resource: "envoyextensionpolicies",
	}
	edgeHTTPRouteGVR = schema.GroupVersionResource{
		Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes",
	}
)

// deenvyGateState mirrors app-service mesh.SteadyGateState JSON.
type deenvyGateState struct {
	Phase         string          `json:"phase"`
	TargetVersion string          `json:"targetVersion,omitempty"`
	Checkpoint    string          `json:"checkpoint,omitempty"`
	Conditions    map[string]bool `json:"conditions,omitempty"`
	Message       string          `json:"message,omitempty"`
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

// oesRecreateBatchLimit caps pod deletes per Accept attempt so controllers can
// reschedule without a thundering herd; Accept retries clear the remainder.
const oesRecreateBatchLimit = 50

func listPodsWithBusinessOES(ctx context.Context, kube kubernetes.Interface) ([]corev1.Pod, error) {
	pods, err := kube.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var out []corev1.Pod
	for i := range pods.Items {
		p := pods.Items[i]
		for _, c := range append(p.Spec.Containers, p.Spec.InitContainers...) {
			if isBusinessOESContainer(p, c) {
				out = append(out, p)
				break
			}
		}
	}
	return out, nil
}

// recreatePodsWithBusinessOES drains legacy business oes by deleting pods.
// Controllers recreate them; app-service webhook no longer injects oes once
// l4 Deployment is Ready. Preferred over machine reboot (power-on ready,
// smaller blast radius). Returns number deleted.
func recreatePodsWithBusinessOES(ctx context.Context, kube kubernetes.Interface) (int, error) {
	offenders, err := listPodsWithBusinessOES(ctx, kube)
	if err != nil {
		return 0, err
	}
	deleted := 0
	for i := range offenders {
		if deleted >= oesRecreateBatchLimit {
			break
		}
		p := offenders[i]
		err := kube.CoreV1().Pods(p.Namespace).Delete(ctx, p.Name, metav1.DeleteOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return deleted, fmt.Errorf("delete oes pod %s/%s: %w", p.Namespace, p.Name, err)
		}
		deleted++
		logger.Infof("deenvy: recreate (delete) oes pod %s/%s", p.Namespace, p.Name)
	}
	return deleted, nil
}

func l4ProxyDeploymentReady(ctx context.Context, kube kubernetes.Interface) bool {
	dep, err := kube.AppsV1().Deployments("os-network").Get(ctx, "l4-bfl-proxy", metav1.GetOptions{})
	if err != nil {
		return false
	}
	return dep.Status.ReadyReplicas >= 1
}

func isEntranceEGPEPObject(name string, labels map[string]string) bool {
	if labels != nil {
		switch labels[authKindLabel] {
		case authKindEntranceExtAuth, authKindEntranceCookie, authKindEntranceProbeBypass:
			return true
		}
	}
	n := strings.TrimSpace(name)
	if n == "" {
		return false
	}
	return strings.HasSuffix(n, entranceExtAuthSuffix) ||
		strings.HasSuffix(n, entranceCookieSuffix) ||
		strings.HasSuffix(n, entranceProbeBypassSuffix) ||
		strings.HasSuffix(n, entranceProbePolicySuffix)
}

func garbageCollectEntranceEGPEP(ctx context.Context, dc dynamic.Interface) (int, error) {
	if dc == nil {
		return 0, fmt.Errorf("dynamic client required")
	}
	deleted := 0
	for _, gvr := range []schema.GroupVersionResource{edgeSecurityPolicyGVR, edgeExtensionPolicyGVR, edgeHTTPRouteGVR} {
		list, err := dc.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return deleted, err
		}
		for i := range list.Items {
			obj := &list.Items[i]
			if gvr == edgeSecurityPolicyGVR && strings.HasSuffix(obj.GetName(), securityPolicyJWTAuthnSuffix) {
				continue
			}
			if !isEntranceEGPEPObject(obj.GetName(), obj.GetLabels()) {
				continue
			}
			ns, name := obj.GetNamespace(), obj.GetName()
			if err := dc.Resource(gvr).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return deleted, err
			}
			deleted++
			logger.Infof("deenvy: EG PEP GC deleted %s %s/%s", gvr.Resource, ns, name)
		}
	}
	return deleted, nil
}

func countEntranceEGPEPLeftovers(ctx context.Context, dc dynamic.Interface) (int, error) {
	if dc == nil {
		return 0, fmt.Errorf("dynamic client required")
	}
	total := 0
	for _, gvr := range []schema.GroupVersionResource{edgeSecurityPolicyGVR, edgeExtensionPolicyGVR, edgeHTTPRouteGVR} {
		list, err := dc.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return total, err
		}
		for i := range list.Items {
			obj := &list.Items[i]
			if gvr == edgeSecurityPolicyGVR && strings.HasSuffix(obj.GetName(), securityPolicyJWTAuthnSuffix) {
				continue
			}
			if isEntranceEGPEPObject(obj.GetName(), obj.GetLabels()) {
				total++
			}
		}
	}
	return total, nil
}

func ensureNoEntranceEGPEP(ctx context.Context, dc dynamic.Interface) (bool, string, error) {
	if _, err := garbageCollectEntranceEGPEP(ctx, dc); err != nil {
		return false, "", err
	}
	n, err := countEntranceEGPEPLeftovers(ctx, dc)
	if err != nil {
		return false, "", err
	}
	if n != 0 {
		return false, fmt.Sprintf("EG entrance PEP leftovers=%d", n), nil
	}
	return true, "no entrance EG PEP leftovers", nil
}

// evaluateEdgeAcceptSuitePassed is the EDGE accept gate:
// L4ProxyReady ∧ ZeroOesInventory ∧ NoEntranceEGExtAuth.
// Linkerd is not an Accept gate (new install stops oes via l4 probe; upgrade
// clears inventory by recreating pods).
func evaluateEdgeAcceptSuitePassed(conds map[string]bool) bool {
	if conds == nil {
		return false
	}
	for _, k := range []string{conditionL4ProxyReady, "ZeroOesInventory", "NoEntranceEGExtAuth"} {
		if !conds[k] {
			return false
		}
	}
	return true
}

func runEdgeAcceptSuite(ctx context.Context, kube kubernetes.Interface, dc dynamic.Interface) (map[string]bool, string, error) {
	conds := map[string]bool{}

	// M-L4: this branch ships verify+F2+F3; require l4 Deployment Ready.
	conds[conditionL4ProxyReady] = l4ProxyDeploymentReady(ctx, kube)

	invOK, invMsg := inventoryZeroUnauthorizedOES(ctx, kube)
	conds["ZeroOesInventory"] = invOK

	egOK := false
	egMsg := "dynamic client unavailable"
	if dc != nil {
		var err error
		egOK, egMsg, err = ensureNoEntranceEGPEP(ctx, dc)
		if err != nil {
			logger.Errorf("deenvy: EG PEP GC failed: %v", err)
			egOK = false
			egMsg = err.Error()
		}
	}
	conds["NoEntranceEGExtAuth"] = egOK

	passed := evaluateEdgeAcceptSuitePassed(conds)
	conds["AcceptSuitePassed"] = passed

	if !passed {
		parts := []string{}
		if !conds[conditionL4ProxyReady] {
			parts = append(parts, "L4ProxyReady=false")
		}
		if !invOK {
			parts = append(parts, invMsg)
		}
		if !egOK {
			parts = append(parts, egMsg)
		}
		return conds, strings.Join(parts, "; "), nil
	}
	return conds, "edge accept suite passed", nil
}

type deenvyEdgeAccept struct {
	common.KubeAction
	TargetVersion string
}

func (a *deenvyEdgeAccept) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return err
	}
	dc, _ := dynamicClientFromRuntime()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Upgrade path: recreate (delete) pods still carrying business oes so
	// controllers reschedule without oes. Prefer this over host reboot.
	if n, err := recreatePodsWithBusinessOES(ctx, kube); err != nil {
		return fmt.Errorf("recreate oes pods: %w", err)
	} else if n > 0 {
		logger.Infof("deenvy: deleted %d oes pods for recreate (batch cap %d)", n, oesRecreateBatchLimit)
		time.Sleep(5 * time.Second)
	}

	conds, msg, err := runEdgeAcceptSuite(ctx, kube, dc)
	if err != nil {
		return err
	}
	st := &deenvyGateState{
		Phase:         "Pending",
		TargetVersion: a.TargetVersion,
		Checkpoint:    cpAccept,
		Conditions:    conds,
		Message:       msg,
	}
	if conds["AcceptSuitePassed"] {
		st.Phase = deenvyGateReady
		st.Checkpoint = cpCommit
		st.Message = msg
	}
	if err := storeDeenvyGate(ctx, kube, st); err != nil {
		return fmt.Errorf("store SteadyGate: %w", err)
	}
	if !conds["AcceptSuitePassed"] {
		return fmt.Errorf("deenvy edge accept failed: %s", msg)
	}
	logger.Infof("deenvy: edge accept passed; L4ProxyReady written")
	return nil
}

type deenvyEdgeCommitGate struct {
	common.KubeAction
	TargetVersion string
}

func (a *deenvyEdgeCommitGate) Execute(runtime connector.Runtime) error {
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
	if !strings.EqualFold(st.Phase, deenvyGateReady) || !st.Conditions["AcceptSuitePassed"] {
		return fmt.Errorf("deenvy: refuse version write; phase=%s accept=%v", st.Phase, st.Conditions["AcceptSuitePassed"])
	}
	if !st.Conditions[conditionL4ProxyReady] {
		return fmt.Errorf("deenvy: refuse version write; L4ProxyReady=false")
	}
	st.TargetVersion = a.TargetVersion
	st.Checkpoint = cpDone
	st.Message = "SteadyStateGate Ready"
	return storeDeenvyGate(ctx, kube, st)
}

// deenvyEdgeUpgradeTasks returns Accept + CommitGate for EDGE (no EG ExtAuth suite).
// Accept deletes residual oes pods (recreate) before inventory; no host reboot.
func deenvyEdgeUpgradeTasks(targetVersion string) []task.Interface {
	return []task.Interface{
		&task.LocalTask{Name: "DeenvyEdgeAccept", Action: &deenvyEdgeAccept{TargetVersion: targetVersion}, Retry: 8, Delay: 20 * time.Second},
		&task.LocalTask{Name: "DeenvyEdgeCommitGate", Action: &deenvyEdgeCommitGate{TargetVersion: targetVersion}, Retry: 3},
	}
}
