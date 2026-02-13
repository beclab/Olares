package external_network_switch

import (
	"context"
	"fmt"
	"errors"
	"time"

	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	settingsv1alpha1 "bytetrade.io/web3os/bfl/pkg/apis/settings/v1alpha1"
	external_network "bytetrade.io/web3os/bfl/pkg/apis/settings/v1alpha1/external_network"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/watchers"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

var GVR = schema.GroupVersionResource{
	Group: "", Version: "v1", Resource: "configmaps",
}

type Subscriber struct {
	*watchers.Watchers
}

func NewSubscriber(w *watchers.Watchers) *Subscriber {
	return &Subscriber{Watchers: w}
}

func (s *Subscriber) Handler() cache.ResourceEventHandler {
	handleFunc := func(obj interface{}) {
		s.Watchers.Enqueue(
			watchers.EnqueueObj{
				Subscribe: s,
				Obj:       obj,
				Action:    watchers.UPDATE,
			},
		)
	}
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			cm, ok := obj.(*corev1.ConfigMap)
			if !ok {
				klog.Error("not configmap resource, invalid obj")
				return false
			}
			return cm.Namespace == constants.OSSystemNamespace && cm.Name == constants.ExternalNetworkSwitchConfigMapName
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: handleFunc,
			UpdateFunc: func(_, new interface{}) {
				handleFunc(new)
			},
			DeleteFunc: func(_ interface{}) {},
		},
	}
}

func (s *Subscriber) Do(ctx context.Context, obj interface{}, _ watchers.Action) error {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok || cm == nil {
		klog.Error("external network switch: invalid object type")
		return fmt.Errorf("invalid object type")
	}
	cfg, err := external_network.ParseConfigMap(cm)
	if err != nil {
		klog.Errorf("external network switch: parse configmap error: %v", err)
		return err
	}
	return s.reconcile(ctx, cfg)
}

func (s *Subscriber) reconcile(ctx context.Context, cfg *external_network.SwitchConfig) (retErr error) {
	if cfg == nil {
		return nil
	}
	defer func() {
		if retErr == nil {
			return
		}
		var rq *requeueError
		if errors.As(retErr, &rq) {
			return
		}
		_, _ = external_network.Upsert(ctx, func(sw *external_network.SwitchConfig) error {
			sw.Status.Message = retErr.Error()
			return nil
		})
	}()
	klog.Infof("external network switch reconcile: spec.disabled=%v phase=%s startedAt=%s taskId=%s",
		cfg.Spec.Disabled, cfg.Status.Phase, cfg.Status.StartedAt, cfg.Status.TaskID)
	// do not retry automatically once marked failed; wait for a new request to change spec/phase
	if cfg.Status.Phase == external_network.PhaseFailed {
		klog.Infof("external network switch: phase=FAILED, stop retrying")
		return nil
	}
	if cfg.Status.Phase == external_network.PhaseCompleted {
		klog.Infof("external network switch: phase=COMPLETED, already converged to spec.disabled=%v", cfg.Spec.Disabled)
		return nil
	}
	start := cfg.Status.StartedAt
	if start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			if time.Since(t) > 15*time.Minute {
				klog.Errorf("external network switch: operation timeout (>10m), marking failed. spec.disabled=%v startedAt=%s",
					cfg.Spec.Disabled, start)
				_, _ = external_network.Upsert(ctx, func(sw *external_network.SwitchConfig) error {
					sw.Spec.Disabled = cfg.Spec.Disabled
					sw.Status.Phase = external_network.PhaseFailed
					sw.Status.Message = "operation timed out"
					return nil
				})
				return nil
			}
		} else {
			klog.Warningf("external network switch: invalid startedAt %q: %v", start, err)
		}
	}

	userOp, err := operator.NewUserOperatorWithContext(ctx)
	if err != nil {
		klog.Errorf("external network switch: new user operator error: %v", err)
		return err
	}
	// fetch integration account token/userid on-demand for external public-dns APIs
	ownerUser, err := userOp.GetUser("")
	if err != nil {
		klog.Errorf("external network switch: get owner user error: %v", err)
		return err
	}
	terminusName := userOp.GetUserAnnotation(ownerUser, constants.UserAnnotationTerminusNameKey)
	acc, err := external_network.GetIntegrationAccount(ctx, ownerUser.Name, terminusName)
	if err != nil {
		klog.Errorf("external network switch: get integration account error: %v", err)
		return err
	}
	auth := external_network.PublicDNSAuthPayload{Token: acc.AccessToken, UserID: acc.UserID}

	if !cfg.Spec.Disabled {
		if err := external_network.RecoverPublicDNS(ctx, auth); err != nil {
			klog.Errorf("external network switch: public dns recover error: %v", err)
			return err
		}
		klog.Infof("external network switch: public dns recover completed")
		if err := applyReverseProxyExternalOff(ctx, userOp, cfg.Spec.Disabled); err != nil {
			return err
		}
		_, err := external_network.Upsert(ctx, func(sw *external_network.SwitchConfig) error {
			sw.Spec.Disabled = false
			sw.Status.Phase = external_network.PhaseCompleted
			sw.Status.Message = ""
			sw.Status.TaskID = ""
			return nil
		})
		if err == nil {
			klog.Infof("external network switch: enable completed")
		}
		return err
	}

	if cfg.Status.TaskID == "" {
		hasRecord, err := external_network.GetPublicDNSInfo(ctx, auth)
		if err != nil {
			klog.Errorf("external network switch: get public dns info error: %v", err)
			return err
		}
		if !hasRecord {
			klog.Infof("external network switch: no public dns record, skipping delete task")
			if err := applyReverseProxyExternalOff(ctx, userOp, cfg.Spec.Disabled); err != nil {
				return err
			}
			_, err := external_network.Upsert(ctx, func(sw *external_network.SwitchConfig) error {
				sw.Spec.Disabled = true
				sw.Status.Phase = external_network.PhaseCompleted
				sw.Status.Message = ""
				sw.Status.TaskID = ""
				return nil
			})
			if err == nil {
				klog.Infof("external network switch: disable completed (no public dns record)")
			}
			return err
		}
		klog.Infof("external network switch: disabling flow, creating public dns delete task")
		taskID, err := external_network.CreatePublicDNSDeleteTask(ctx, auth)
		if err != nil {
			// todo: parse specific error and determine if retry is possible
			klog.Errorf("external network switch: create public dns delete task error: %v", err)
			return err
		}
		klog.Infof("external network switch: public dns delete task created: taskId=%s", taskID)
		_, err = external_network.Upsert(ctx, func(sw *external_network.SwitchConfig) error {
			sw.Spec.Disabled = true
			sw.Status.TaskID = taskID
			sw.Status.Phase = external_network.PhaseProcessing
			sw.Status.Message = "delete task created"
			return nil
		})
		if err != nil {
			return err
		}
		return newRequeueError(fmt.Errorf("delete task created, waiting"))
	}

	taskResp, err := external_network.GetPublicDNSTask(ctx, auth, cfg.Status.TaskID)
	if err != nil {
		// todo: parse specific error and determine if retry is possible (if task not exists, it needs to be created again)
		klog.Errorf("external network switch: get public dns task error: taskId=%s err=%v", cfg.Status.TaskID, err)
		return err
	}
	done, msg := taskResp.AllTasksCompleted()
	if done {
		if err := applyReverseProxyExternalOff(ctx, userOp, cfg.Spec.Disabled); err != nil {
			return err
		}
		_, err := external_network.Upsert(ctx, func(sw *external_network.SwitchConfig) error {
			sw.Spec.Disabled = true
			sw.Status.Phase = external_network.PhaseCompleted
			sw.Status.Message = ""
			return nil
		})
		if err == nil {
			klog.Infof("external network switch: disable completed. taskId=%s", cfg.Status.TaskID)
		}
		return err
	}
	if msg != "" {
		// task response contains failed tasks -> mark failed and stop retrying
		klog.Errorf("external network switch: public dns delete tasks failed: taskId=%s msg=%s", cfg.Status.TaskID, msg)
		_, err := external_network.Upsert(ctx, func(sw *external_network.SwitchConfig) error {
			sw.Spec.Disabled = true
			sw.Status.Phase = external_network.PhaseFailed
			sw.Status.Message = msg
			return nil
		})
		return err
	}
	_, _ = external_network.Upsert(ctx, func(sw *external_network.SwitchConfig) error {
		sw.Spec.Disabled = true
		sw.Status.Phase = external_network.PhaseProcessing
		sw.Status.Message = "waiting public dns delete tasks"
		return nil
	})
	klog.Infof("external network switch: waiting public dns delete tasks. taskId=%s", cfg.Status.TaskID)
	return newRequeueError(fmt.Errorf("waiting public dns delete tasks"))
}

type requeueError struct {
	err error
}

func (e *requeueError) Error() string {
	return e.err.Error()
}

func newRequeueError(err error) error {
	if err == nil {
		return nil
	}
	return &requeueError{err: err}
}

func upsertReverseProxyConfigExternalOff(ctx context.Context, client kubernetes.Interface, namespace string, off bool) error {
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, constants.ReverseProxyConfigMapName, metav1.GetOptions{})
	isNew := false
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		isNew = true
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.ReverseProxyConfigMapName,
				Namespace: namespace,
			},
			Data: map[string]string{},
		}
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	want := ""
	if off {
		want = settingsv1alpha1.ReverseProxyConfigValueTrue
	}
	if cm.Data[settingsv1alpha1.ReverseProxyConfigKeyExternalNetworkOff] == want {
		return nil
	}
	cm.Data[settingsv1alpha1.ReverseProxyConfigKeyExternalNetworkOff] = want
	if isNew {
		_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
		return err
	}
	_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

func applyReverseProxyExternalOff(ctx context.Context, userOp *operator.UserOperator, off bool) error {
	users, err := userOp.ListUsers()
	if err != nil {
		klog.Errorf("external network switch: list users error: %v", err)
		return err
	}
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		klog.Errorf("external network switch: new kube client error: %v", err)
		return err
	}
	kube := kc.Kubernetes()

	for _, u := range users {
		ns := fmt.Sprintf(constants.UserspaceNameFormat, u.Name)
		if err := upsertReverseProxyConfigExternalOff(ctx, kube, ns, off); err != nil {
			klog.Errorf("external network switch: update reverse-proxy-config off=%v in ns=%s error: %v",
				off, ns, err)
			return err
		}
	}
	return nil
}
