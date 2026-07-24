package security

import (
	"testing"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsRouteControlManagedNP(t *testing.T) {
	tests := []struct {
		name string
		np   *netv1.NetworkPolicy
		want bool
	}{
		{
			name: "nil",
			np:   nil,
			want: false,
		},
		{
			name: "no labels",
			np:   &netv1.NetworkPolicy{},
			want: false,
		},
		{
			name: "managed-by only",
			np: &netv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{RouteControlManagedByLabel: RouteControlManagedByValue},
				},
			},
			want: false,
		},
		{
			name: "routecontrol owned",
			np: &netv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						RouteControlManagedByLabel: RouteControlManagedByValue,
						RouteControlComponentLabel: RouteControlComponentValue,
					},
				},
			},
			want: true,
		},
		{
			name: "caller-jwt not routecontrol",
			np: &netv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						RouteControlManagedByLabel: RouteControlManagedByValue,
						RouteControlComponentLabel: CallerJWTComponentValue,
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRouteControlManagedNP(tt.np); got != tt.want {
				t.Fatalf("IsRouteControlManagedNP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCallerJWTManagedNPAndExternal(t *testing.T) {
	callerJWT := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				RouteControlManagedByLabel: RouteControlManagedByValue,
				RouteControlComponentLabel: CallerJWTComponentValue,
			},
		},
	}
	routeControl := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				RouteControlManagedByLabel: RouteControlManagedByValue,
				RouteControlComponentLabel: RouteControlComponentValue,
			},
		},
	}
	if !IsCallerJWTManagedNP(callerJWT) {
		t.Fatal("expected caller-jwt NP")
	}
	if IsCallerJWTManagedNP(routeControl) {
		t.Fatal("route-control must not match caller-jwt")
	}
	if !IsAppServiceManagedExternalNP(callerJWT) || !IsAppServiceManagedExternalNP(routeControl) {
		t.Fatal("both external writers must be skipped by prune")
	}
	if IsAppServiceManagedExternalNP(&netv1.NetworkPolicy{}) {
		t.Fatal("empty NP must not be external-managed")
	}
}
