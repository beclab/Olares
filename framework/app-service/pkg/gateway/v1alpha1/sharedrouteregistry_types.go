// Package v1alpha1 contains the gateway.olares.io/v1alpha1 API types.
// SharedRouteRegistry (SRR) is the hand-off contract between app-service
// (writer) and app-service-routecontrol (reader). See:
//   - archdoc/方案/shared应用/Shared外部访问Phase-A开发实现规格-2026-05-19.md §5
//   - archdoc/方案/shared应用/Shared外部访问Phase-A评审决议-2026-05-19.md (F-1～F-6)
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersion is the canonical Group/Version for SharedRouteRegistry.
var GroupVersion = schema.GroupVersion{Group: "gateway.olares.io", Version: "v1alpha1"}

// SchemeBuilder collects the type registrations for this API group.
var SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

// AddToScheme registers the types in this API group with the given scheme.
var AddToScheme = SchemeBuilder.AddToScheme

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&SharedRouteRegistry{},
		&SharedRouteRegistryList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}

// RouteMode controls whether ingress traffic for the shared app flows through
// the Envoy Gateway data plane or stays on the legacy direct L4 path.
type RouteMode string

const (
	// RouteModeGateway routes via Envoy Gateway (Phase A target).
	RouteModeGateway RouteMode = "gateway"
	// RouteModeDirect keeps the legacy l4-bfl-proxy -> backend Service path.
	RouteModeDirect RouteMode = "direct"
)

// AuthzDefaultAction is the fallback decision when no per-route policy matches.
type AuthzDefaultAction string

const (
	AuthzDefaultAllow AuthzDefaultAction = "allow"
	AuthzDefaultDeny  AuthzDefaultAction = "deny"
)

// UpstreamRef points at a Service in (by default) the same namespace as the SRR.
type UpstreamRef struct {
	// ServiceName is the Kubernetes Service short name (no domain).
	ServiceName string `json:"serviceName"`
	// ServiceNamespace defaults to the SRR's metadata.namespace ({app}-shared, F-2/F-4).
	// +optional
	ServiceNamespace string `json:"serviceNamespace,omitempty"`
	// Port is the TCP target port (1-65535). Either Port or PortName must be set.
	// +optional
	Port int32 `json:"port,omitempty"`
	// PortName is an alternative to Port; mutually exclusive.
	// +optional
	PortName string `json:"portName,omitempty"`
}

// AuthzRef carries Phase A defaults. Phase B will replace this with a
// reference to a per-route policy object.
type AuthzRef struct {
	// +kubebuilder:validation:Enum=allow;deny
	DefaultAction AuthzDefaultAction `json:"defaultAction,omitempty"`
}

// SharedRouteRegistrySpec is the declarative input for a single shared route.
type SharedRouteRegistrySpec struct {
	// RouteMode selects which data path serves this SRR. app-service-routecontrol only
	// reconciles HTTPRoute + NetworkPolicy when RouteMode=gateway.
	// +kubebuilder:validation:Enum=gateway;direct
	// +kubebuilder:default=gateway
	RouteMode RouteMode `json:"routeMode,omitempty"`

	// HostPatterns must be lowercase and contain no port (F-3).
	// +kubebuilder:validation:MinItems=1
	HostPatterns []string `json:"hostPatterns"`

	// Upstream points at the backing Service in the same namespace.
	Upstream UpstreamRef `json:"upstream"`

	// AuthzRef is optional in Phase A; defaults to allow.
	// +optional
	AuthzRef *AuthzRef `json:"authzRef,omitempty"`
}

// SharedRouteRegistryStatus reports the app-service-routecontrol's outcome.
type SharedRouteRegistryStatus struct {
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// HTTPRouteName is the gateway.networking.k8s.io HTTPRoute created in the
	// same namespace as this SRR.
	// +optional
	HTTPRouteName string `json:"httpRouteName,omitempty"`
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=srr,categories=olares
// +kubebuilder:subresource:status

// SharedRouteRegistry is the per-app declaration of "expose me through the
// shared Envoy Gateway data plane".
type SharedRouteRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SharedRouteRegistrySpec   `json:"spec,omitempty"`
	Status SharedRouteRegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SharedRouteRegistryList is the list wrapper for SharedRouteRegistry.
type SharedRouteRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SharedRouteRegistry `json:"items"`
}
