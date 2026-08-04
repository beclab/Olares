// Package v1alpha1 contains the gateway.olares.io/v1alpha1 API types.
// SharedRouteRegistry (SRR) declares how a v3 shared app is exposed: app-service
// writes the spec and reconciles an HTTPRoute in the app namespace. No service mesh.
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
	// RouteModeGateway routes via Envoy Gateway (HTTPRoute).
	RouteModeGateway RouteMode = "gateway"
	// RouteModeDirect keeps the legacy l4-bfl-proxy -> backend Service path.
	RouteModeDirect RouteMode = "direct"
)

// EntranceClass marks which entrance category produced this SRR.
type EntranceClass string

const (
	// EntranceClassApplication marks app-level entrance records.
	EntranceClassApplication EntranceClass = "application"
	// EntranceClassShared marks shared entrance records.
	EntranceClassShared EntranceClass = "shared"
)

// UpstreamRef points at a Service in (by default) the same namespace as the SRR.
type UpstreamRef struct {
	// ServiceName is the Kubernetes Service short name (no domain).
	ServiceName string `json:"serviceName"`
	// ServiceNamespace defaults to the SRR namespace ({app}-shared).
	// +optional
	ServiceNamespace string `json:"serviceNamespace,omitempty"`
	// Port is the TCP target port (1-65535). Either Port or PortName must be set.
	// +optional
	Port int32 `json:"port,omitempty"`
	// PortName is an alternative to Port; mutually exclusive.
	// +optional
	PortName string `json:"portName,omitempty"`
}

// SharedRouteRegistrySpec is the declarative input for a single shared route.
type SharedRouteRegistrySpec struct {
	// RouteMode selects which data path serves this SRR. app-service only
	// reconciles the HTTPRoute when RouteMode=gateway.
	// +kubebuilder:validation:Enum=gateway;direct
	// +kubebuilder:default=gateway
	RouteMode RouteMode `json:"routeMode,omitempty"`

	// EntranceClass marks whether this SRR came from an application/shared entrance.
	// +kubebuilder:validation:Enum=application;shared
	// +kubebuilder:default=shared
	EntranceClass EntranceClass `json:"entranceClass,omitempty"`

	// HostPatterns must be lowercase DNS names without a port.
	// +kubebuilder:validation:MinItems=1
	HostPatterns []string `json:"hostPatterns"`

	// Upstream points at the backing Service in the same namespace.
	Upstream UpstreamRef `json:"upstream"`
}

// SharedRouteRegistryStatus reports the shared-route reconcile outcome.
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
