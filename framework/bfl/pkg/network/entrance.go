package network

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"time"

	"bytetrade.io/web3os/bfl/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const entrancePollInterval = 30 * time.Second

// EntranceController observes machine addresses independently of users and
// cloud provisioning. Only public-IP mode uses the outbound address probe.
type EntranceController struct {
	Store         *Store
	ProbePublicIP func(context.Context) (string, error)
	nextProbe     time.Time
}

func (c *EntranceController) SetupWithManager(mgr ctrl.Manager) error {
	queue := handler.EnqueueRequestsFromMapFunc(func(context.Context, client.Object) []reconcile.Request {
		return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: ConfigName}}}
	})
	return ctrl.NewControllerManagedBy(mgr).Named("network-entrance").
		Watches(&corev1.ConfigMap{}, queue, builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetNamespace() == Namespace && obj.GetName() == ConfigName
		}))).
		Watches(&corev1.Pod{}, queue, builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetNamespace() == Namespace && obj.GetLabels()["app"] == "l4-bfl-proxy"
		}))).
		Watches(&corev1.Node{}, queue, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event.CreateEvent) bool { return true },
			DeleteFunc: func(event.DeleteEvent) bool { return true },
			UpdateFunc: func(e event.UpdateEvent) bool {
				old, current := e.ObjectOld.(*corev1.Node), e.ObjectNew.(*corev1.Node)
				return !reflect.DeepEqual(old.Status.Addresses, current.Status.Addresses) || !reflect.DeepEqual(old.DeletionTimestamp, current.DeletionTimestamp)
			},
		})).Complete(c)
}

func (c *EntranceController) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	cm, spec, _, err := c.Store.Config(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	entrance := entranceFromData(cm.Data)
	ip, err := (&Controller{Store: c.Store}).l4IP(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	entrance.L4Ready = ip != ""
	if ip != "" {
		entrance.LocalIP = ip
	}
	// Publish the local endpoint before running any outbound probe.
	if err := c.Store.SaveEntrance(ctx, entrance); err != nil {
		return ctrl.Result{}, err
	}
	if spec.Config.Mode != ModePublicIP {
		c.nextProbe = time.Time{}
		return ctrl.Result{RequeueAfter: entrancePollInterval}, nil
	}
	if !time.Now().Before(c.nextProbe) {
		probe := c.ProbePublicIP
		if probe == nil {
			probe = func(ctx context.Context) (string, error) {
				return utils.GetExternalIPAddr(ctx), nil
			}
		}
		publicIP, probeErr := probe(ctx)
		address := net.ParseIP(publicIP)
		if probeErr == nil && (address == nil || !address.IsGlobalUnicast() || address.IsPrivate()) {
			probeErr = fmt.Errorf("no public IP address found")
		}
		entrance.PublicIPError = ""
		if probeErr != nil {
			entrance.PublicIPError = probeErr.Error()
		} else {
			entrance.PublicIP = address.String()
		}
		c.nextProbe = time.Now().Add(entrancePollInterval)
		if err := c.Store.SaveEntrance(ctx, entrance); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{RequeueAfter: entrancePollInterval}, nil
}
