package gateway

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

// SystemPlaneService describes one L4-parity system hostname → Service mapping.
type SystemPlaneService struct {
	Name       string
	Service    string
	Port       int32
	AuthMode   string // AuthModeExtAuth or AuthModeNone
	StampOwner bool   // stamp X-BFL-USER (auth/wizard)
	ApexHost   bool   // profile uses zone apex, not <name>.<zone>
}

// SystemPlaneServices mirrors l4-bfl-proxy systemServices + profile.
var SystemPlaneServices = []SystemPlaneService{
	{Name: "auth", Service: "authelia-svc", Port: 80, AuthMode: AuthModeNone, StampOwner: true},
	{Name: "desktop", Service: "edge-desktop", Port: 80, AuthMode: AuthModeExtAuth},
	{Name: "wizard", Service: "wizard", Port: 80, AuthMode: AuthModeNone, StampOwner: true},
	{Name: "profile", Service: "profile-service", Port: 3000, AuthMode: AuthModeExtAuth, ApexHost: true},
}

// SystemPlaneSRRName returns the SRR object name for a system plane service.
func SystemPlaneSRRName(svcName string) string {
	return "sys-" + strings.ToLower(strings.TrimSpace(svcName))
}

// SystemPlaneHostPatterns builds exact hostnames for a user zone (and local alias).
func SystemPlaneHostPatterns(svc SystemPlaneService, zone, user string, ephemeral bool) []string {
	zone = strings.ToLower(strings.TrimSpace(zone))
	user = strings.ToLower(strings.TrimSpace(user))
	if zone == "" {
		return nil
	}
	var hosts []string
	add := func(h string) {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "" {
			return
		}
		hosts = append(hosts, h)
	}
	localZone := ""
	if tokens := strings.Split(zone, "."); len(tokens) > 1 {
		localZone = tokens[0] + ".olares.local"
	}

	if svc.ApexHost {
		if ephemeral {
			return nil
		}
		add(zone)
		if localZone != "" {
			add(localZone)
		}
		return hosts
	}

	prefix := svc.Name
	if ephemeral {
		if svc.Name != "wizard" || user == "" {
			return nil
		}
		prefix = fmt.Sprintf("wizard-%s", user)
	}
	add(prefix + "." + zone)
	if localZone != "" {
		add(prefix + "." + localZone)
	}
	return hosts
}

// EnsureSystemPlaneRoutes upserts SRRs for auth/desktop/wizard/profile in userspaceNS.
func EnsureSystemPlaneRoutes(ctx context.Context, c client.Client, user, zone, userspaceNS string, ephemeral bool) error {
	if c == nil {
		return fmt.Errorf("client is nil")
	}
	user = strings.ToLower(strings.TrimSpace(user))
	zone = strings.ToLower(strings.TrimSpace(zone))
	userspaceNS = strings.TrimSpace(userspaceNS)
	if user == "" || zone == "" || userspaceNS == "" {
		return fmt.Errorf("user/zone/userspaceNS required")
	}

	desiredNames := map[string]struct{}{}
	for _, svc := range SystemPlaneServices {
		hosts := SystemPlaneHostPatterns(svc, zone, user, ephemeral)
		if len(hosts) == 0 {
			continue
		}
		name := SystemPlaneSRRName(svc.Name)
		desiredNames[name] = struct{}{}
		if err := upsertSystemPlaneSRR(ctx, c, userspaceNS, name, user, svc, hosts); err != nil {
			return err
		}
	}
	return pruneSystemPlaneSRRs(ctx, c, userspaceNS, desiredNames)
}

func upsertSystemPlaneSRR(ctx context.Context, c client.Client, ns, name, user string, svc SystemPlaneService, hosts []string) error {
	anns := map[string]string{
		AnnotationExtAuthUser: user,
		AnnotationAuthMode:    svc.AuthMode,
	}
	if svc.StampOwner {
		anns[AnnotationStampBFLUser] = user
	}
	spec := srrv1alpha1.SharedRouteRegistrySpec{
		RouteMode:     srrv1alpha1.RouteModeGateway,
		EntranceClass: srrv1alpha1.EntranceClassApplication,
		HostPatterns:  hosts,
		Upstream: srrv1alpha1.UpstreamRef{
			ServiceName: svc.Service,
			Port:        svc.Port,
		},
	}
	got := &srrv1alpha1.SharedRouteRegistry{}
	getErr := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, got)
	switch {
	case apierrors.IsNotFound(getErr):
		obj := &srrv1alpha1.SharedRouteRegistry{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   ns,
				Labels:      systemPlaneLabels(name),
				Annotations: anns,
			},
			Spec: spec,
		}
		if err := c.Create(ctx, obj); err != nil {
			return fmt.Errorf("create system SRR %s/%s: %w", ns, name, err)
		}
		return nil
	case getErr != nil:
		return fmt.Errorf("get system SRR %s/%s: %w", ns, name, getErr)
	}
	patched := got.DeepCopy()
	patched.Spec = spec
	if patched.Labels == nil {
		patched.Labels = map[string]string{}
	}
	for k, v := range systemPlaneLabels(name) {
		patched.Labels[k] = v
	}
	if patched.Annotations == nil {
		patched.Annotations = map[string]string{}
	}
	for k, v := range anns {
		patched.Annotations[k] = v
	}
	if err := c.Patch(ctx, patched, client.MergeFrom(got)); err != nil {
		return fmt.Errorf("patch system SRR %s/%s: %w", ns, name, err)
	}
	return nil
}

func systemPlaneLabels(srrName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "app-service",
		"app.kubernetes.io/instance":   SystemPlaneInstance,
		LabelSystemPlane:               LabelSystemPlaneValue,
		"gateway.olares.io/entrance":   strings.TrimPrefix(srrName, "sys-"),
	}
}

// DeleteSystemPlaneRoutes withdraws only app-service-managed system routes.
func DeleteSystemPlaneRoutes(ctx context.Context, c client.Client, ns string) error {
	return pruneSystemPlaneSRRs(ctx, c, ns, nil)
}

func pruneSystemPlaneSRRs(ctx context.Context, c client.Client, ns string, keep map[string]struct{}) error {
	list := &srrv1alpha1.SharedRouteRegistryList{}
	if err := c.List(ctx, list, client.InNamespace(ns), client.MatchingLabels{
		LabelSystemPlane:               LabelSystemPlaneValue,
		"app.kubernetes.io/managed-by": "app-service",
		"app.kubernetes.io/instance":   SystemPlaneInstance,
	}); err != nil {
		return fmt.Errorf("list system SRRs in %s: %w", ns, err)
	}
	for i := range list.Items {
		obj := &list.Items[i]
		if _, ok := keep[obj.Name]; ok {
			continue
		}
		if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete stale system SRR %s/%s: %w", obj.Namespace, obj.Name, err)
		}
	}
	return nil
}
