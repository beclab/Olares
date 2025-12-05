package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	tailScaleACLPolicyMd5Key       = "tailscale-acl-md5"
	tailScaleDeployOrContainerName = "tailscale"
	subnetRoutesEnv                = "TS_ROUTES"
)

var defaultACLs = []v1alpha1.ACL{
	{
		Action: "accept",
		Src:    []string{"*"},
		Proto:  "tcp",
		Dst:    []string{"*:443"},
	},
	{
		Action: "accept",
		Src:    []string{"*"},
		Proto:  "tcp",
		Dst:    []string{"*:18088"},
	},
	{
		Action: "accept",
		Src:    []string{"*"},
		Proto:  "",
		Dst:    []string{"*:53"},
	},
	{
		Action: "accept",
		Src:    []string{"*"},
		Proto:  "",
		Dst:    []string{"*:80"},
	},
}
var defaultSubRoutes = []string{"$(COREDNS_SVC)/32"}

type ACLPolicy struct {
	ACLs          []v1alpha1.ACL `json:"acls"`
	AutoApprovers AutoApprovers  `json:"autoApprovers"`
}

type AutoApprovers struct {
	Routes   map[string][]string `json:"routes"`
	ExitNode []string            `json:"exitNode"`
}

type TailScaleACLController struct {
	client.Client
}

func (r *TailScaleACLController) SetUpWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("app's tailscale acls manager controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&v1alpha1.Application{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, app *v1alpha1.Application) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name:      app.Name,
					Namespace: app.Spec.Owner,
				}}}
			}),
		predicate.TypedFuncs[*v1alpha1.Application]{
			CreateFunc: func(e event.TypedCreateEvent[*v1alpha1.Application]) bool {
				return true
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*v1alpha1.Application]) bool {
				return true
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*v1alpha1.Application]) bool {
				return true
			},
		},
	))
	if err != nil {
		return err
	}
	return nil
}

func (r *TailScaleACLController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	klog.Infof("reconcile tailscale acls subroutes request name=%v, owner=%v", req.Name, req.Namespace)
	owner := req.Namespace

	// for this request req.Namespace is owner
	// list all apps by owner and generate acls by owner
	var apps v1alpha1.ApplicationList
	err := r.List(ctx, &apps)
	if err != nil {
		klog.Errorf("list applications failed: %v", err)
		return ctrl.Result{}, err
	}
	filteredApps := make([]v1alpha1.Application, 0)
	for _, app := range apps.Items {
		if app.Spec.Owner != owner {
			continue
		}
		filteredApps = append(filteredApps, app)
	}

	sort.Slice(filteredApps, func(i, j int) bool {
		return filteredApps[j].CreationTimestamp.Before(&filteredApps[i].CreationTimestamp)
	})

	tailScaleACLConfig := "tailscale-acl"
	headScaleNamespace := fmt.Sprintf("user-space-%s", owner)

	// calculate acls
	acls := make([]v1alpha1.ACL, 0)
	subRoutes := make([]string, 0)
	routeSet := sets.NewString()

	subRoutes = append(subRoutes, defaultSubRoutes...)
	for _, app := range filteredApps {
		acls = append(acls, app.Spec.TailScale.ACLs...)
		// just to maintain compatibility with existing application
		acls = append(acls, app.Spec.TailScaleACLs...)
		for _, subRoute := range app.Spec.TailScale.SubRoutes {
			if routeSet.Has(subRoute) {
				continue
			}
			subRoutes = append(subRoutes, subRoute)
			routeSet.Insert(subRoute)
		}
	}

	tailScaleDeploy := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: tailScaleDeployOrContainerName, Namespace: headScaleNamespace}, tailScaleDeploy)
	if err != nil {
		klog.Errorf("get tailscale deploy failed: %v", err)
		return ctrl.Result{}, err
	}
	tailScaleRouteEnv := ""
	for _, container := range tailScaleDeploy.Spec.Template.Spec.Containers {
		if container.Name != tailScaleDeployOrContainerName {
			continue
		}
		for _, env := range container.Env {
			if env.Name == subnetRoutesEnv {
				tailScaleRouteEnv = env.Value
			}
		}
	}

	oldTailScaleRoutes := strings.Split(tailScaleRouteEnv, ",")
	klog.Infof("oldTailScaleRoutes: %v", oldTailScaleRoutes)
	klog.Infof("new sub Routes: %v", subRoutes)

	if !isTsRoutesEqual(oldTailScaleRoutes, subRoutes) {
		newTailScaleRoutesEnv := strings.Join(subRoutes, ",")
		containers := tailScaleDeploy.Spec.Template.Spec.Containers
		for i := range containers {
			if containers[i].Name != tailScaleDeployOrContainerName {
				continue
			}
			for j := range containers[i].Env {
				if containers[i].Env[j].Name == subnetRoutesEnv {
					containers[i].Env[j].Value = newTailScaleRoutesEnv
				}
			}
		}
		err = r.Update(ctx, tailScaleDeploy)
		if err != nil {
			klog.Errorf("update tailscale deploy failed %v", err)
			return ctrl.Result{}, err
		}
	}

	configMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: tailScaleACLConfig, Namespace: headScaleNamespace}, configMap)
	if err != nil {
		return ctrl.Result{}, err
	}

	// If no ACLs need to be applied and the ConfigMap tailscale-acl has not been updated by the Tailscale ACL controller,
	// there is no need to update.
	if len(acls) == 0 && (configMap.Annotations == nil || (configMap.Annotations != nil && configMap.Annotations[tailScaleACLPolicyMd5Key] == "")) {
		return ctrl.Result{}, nil
	}

	aclPolicyByte, err := makeACLPolicy(acls)
	if err != nil {
		return ctrl.Result{}, err
	}
	klog.Infof("aclPolicyByte:string: %s", string(aclPolicyByte))
	oldTailScaleACLPolicyMd5Sum := ""
	if configMap.Annotations != nil {
		oldTailScaleACLPolicyMd5Sum = configMap.Annotations[tailScaleACLPolicyMd5Key]
	}
	curTailScaleACLPolicyMd5Sum := utils.Md5String(string(aclPolicyByte))

	if curTailScaleACLPolicyMd5Sum != oldTailScaleACLPolicyMd5Sum {
		if configMap.Annotations == nil {
			configMap.Annotations = make(map[string]string)
		}
		if configMap.Data == nil {
			configMap.Data = make(map[string]string)
		}

		configMap.Annotations[tailScaleACLPolicyMd5Key] = curTailScaleACLPolicyMd5Sum
		configMap.Data["acl.json"] = string(aclPolicyByte)
		err = r.Update(ctx, configMap)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	deploy := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Namespace: headScaleNamespace, Name: "headscale"}, deploy)
	if err != nil {
		return ctrl.Result{}, err
	}
	headScaleACLMd5 := ""
	if deploy.Spec.Template.Annotations != nil {
		klog.Infof("headscaleaclmd5..: %s", deploy.Spec.Template.Annotations[tailScaleACLPolicyMd5Key])
		headScaleACLMd5 = deploy.Spec.Template.Annotations[tailScaleACLPolicyMd5Key]
	}
	klog.Infof("oldheadscaleACLmd5: %v, newmd5: %v", headScaleACLMd5, curTailScaleACLPolicyMd5Sum)
	if headScaleACLMd5 != curTailScaleACLPolicyMd5Sum {
		if deploy.Spec.Template.Annotations == nil {
			deploy.Spec.Template.Annotations = make(map[string]string)
		}

		// update headscale deploy template annotations for rolling update
		deploy.Spec.Template.Annotations[tailScaleACLPolicyMd5Key] = curTailScaleACLPolicyMd5Sum
		err = r.Update(ctx, deploy)
		if err != nil {
			klog.Errorf("update headscale deploy failed: %v", err)
			return ctrl.Result{}, err
		}
		klog.Infof("rolling update headscale...")
	}

	return ctrl.Result{}, nil
}

func makeACLPolicy(acls []v1alpha1.ACL) ([]byte, error) {
	acls = append(acls, defaultACLs...)
	for i := range acls {
		acls[i].Action = "accept"
		acls[i].Src = []string{"*"}
	}
	aclPolicy := ACLPolicy{
		ACLs: acls,
		AutoApprovers: AutoApprovers{
			Routes: map[string][]string{
				"10.0.0.0/8":     {"default"},
				"172.16.0.0/12":  {"default"},
				"192.168.0.0/16": {"default"},
			},
			ExitNode: []string{},
		},
	}
	aclPolicyByte, err := json.Marshal(aclPolicy)
	if err != nil {
		return nil, err
	}
	return aclPolicyByte, nil
}

func isTsRoutesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
