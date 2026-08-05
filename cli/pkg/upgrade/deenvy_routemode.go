package upgrade

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/beclab/Olares/cli/pkg/core/logger"
)

var deenvyApplicationGVR = schema.GroupVersionResource{
	Group: "app.bytetrade.io", Version: "v1alpha1", Resource: "applications",
}

const (
	deenvyRouteModeAnn     = "gateway.olares.io/route-mode"
	deenvyRouteModeGateway = "gateway"
	deenvyAppSharedLabel   = "app.bytetrade.io/app-shared"
	deenvyAppSharedTrue    = "true"
)

type entranceRouteModeTarget struct {
	Namespace string
	AppName   string
	Mode      string
}

func assignRouteModeDepConditions(conds map[string]bool, ready bool) {
	if conds == nil {
		return
	}
	conds["RouteModeGateway"] = ready
}

func evaluateEntranceRouteModeGatewayReady(targets []entranceRouteModeTarget) bool {
	for _, t := range targets {
		if strings.TrimSpace(t.Namespace) == "" || strings.TrimSpace(t.AppName) == "" {
			return false
		}
		if strings.ToLower(strings.TrimSpace(t.Mode)) != deenvyRouteModeGateway {
			return false
		}
	}
	return true
}

func annotateEntranceAppsRouteModeGateway(ctx context.Context, dc dynamic.Interface) error {
	if dc == nil {
		return fmt.Errorf("dynamic client required")
	}
	list, err := dc.Resource(deenvyApplicationGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for i := range list.Items {
		app := &list.Items[i]
		if !applicationNeedsGatewayRouteMode(app) {
			continue
		}
		ann := app.GetAnnotations()
		if ann == nil {
			ann = map[string]string{}
		}
		if strings.EqualFold(strings.TrimSpace(ann[deenvyRouteModeAnn]), deenvyRouteModeGateway) {
			continue
		}
		// Never overwrite an explicit direct pin.
		if strings.EqualFold(strings.TrimSpace(ann[deenvyRouteModeAnn]), "direct") {
			continue
		}
		ann[deenvyRouteModeAnn] = deenvyRouteModeGateway
		app.SetAnnotations(ann)
		if _, err := dc.Resource(deenvyApplicationGVR).Namespace(app.GetNamespace()).Update(ctx, app, metav1.UpdateOptions{}); err != nil {
			logger.Warnf("deenvy: set route-mode=gateway on Application %s/%s: %v", app.GetNamespace(), app.GetName(), err)
			return err
		}
		logger.Infof("deenvy: set route-mode=gateway on Application %s/%s", app.GetNamespace(), app.GetName())
	}
	return nil
}

func hasPartialEntranceRouteModeGateway(targets []entranceRouteModeTarget) bool {
	var sawGateway, sawOther bool
	for _, t := range targets {
		if strings.ToLower(strings.TrimSpace(t.Mode)) == deenvyRouteModeGateway {
			sawGateway = true
		} else {
			sawOther = true
		}
	}
	return sawGateway && sawOther
}

func applicationNeedsGatewayRouteMode(obj *unstructured.Unstructured) bool {
	if obj == nil {
		return false
	}
	labels := obj.GetLabels()
	if labels != nil && strings.EqualFold(labels[deenvyAppSharedLabel], deenvyAppSharedTrue) {
		return true
	}
	entrances, found, _ := unstructured.NestedSlice(obj.Object, "spec", "entrances")
	if found && len(entrances) > 0 {
		return true
	}
	sharedEntrances, found, _ := unstructured.NestedSlice(obj.Object, "spec", "sharedEntrances")
	return found && len(sharedEntrances) > 0
}

func listEntranceRouteModeTargets(ctx context.Context, dc dynamic.Interface) ([]entranceRouteModeTarget, error) {
	if dc == nil {
		return nil, nil
	}
	list, err := dc.Resource(deenvyApplicationGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]entranceRouteModeTarget, 0, len(list.Items))
	for i := range list.Items {
		app := &list.Items[i]
		if !applicationNeedsGatewayRouteMode(app) {
			continue
		}
		mode := ""
		if ann := app.GetAnnotations(); ann != nil {
			mode = ann[deenvyRouteModeAnn]
		}
		out = append(out, entranceRouteModeTarget{
			Namespace: app.GetNamespace(),
			AppName:   app.GetName(),
			Mode:      mode,
		})
	}
	return out, nil
}

func probeEntranceRouteModeGatewayReady(ctx context.Context, dc dynamic.Interface) (bool, error) {
	targets, err := listEntranceRouteModeTargets(ctx, dc)
	if err != nil {
		logger.Errorf("deenvy: list Applications for route-mode failed: %v", err)
		return false, err
	}
	if hasPartialEntranceRouteModeGateway(targets) {
		logger.Errorf("deenvy: partial route-mode=gateway among entrance Applications")
	}
	return evaluateEntranceRouteModeGatewayReady(targets), nil
}

func ensureRouteModeGatewayOK(ctx context.Context, dc dynamic.Interface, conds map[string]bool) bool {
	if dc == nil {
		assignRouteModeDepConditions(conds, false)
		return false
	}
	ok, err := probeEntranceRouteModeGatewayReady(ctx, dc)
	if err != nil {
		logger.Errorf("deenvy: RouteModeGateway probe failed: %v", err)
		ok = false
	}
	assignRouteModeDepConditions(conds, ok)
	return ok
}
