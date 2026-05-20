// Package v1alpha1 mirrors the gateway.olares.io/v1alpha1 SharedRouteRegistry
// types so the app-service-routecontrol compiles without dragging in app-service's
// dependency tree. The CRD schema (gateway.olares.io_sharedrouteregistries.yaml)
// is the single source of truth shared by both modules.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var GroupVersion = schema.GroupVersion{Group: "gateway.olares.io", Version: "v1alpha1"}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&SharedRouteRegistry{},
		&SharedRouteRegistryList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}

type RouteMode string

const (
	RouteModeGateway RouteMode = "gateway"
	RouteModeDirect  RouteMode = "direct"
)

type AuthzDefaultAction string

const (
	AuthzDefaultAllow AuthzDefaultAction = "allow"
	AuthzDefaultDeny  AuthzDefaultAction = "deny"
)

type UpstreamRef struct {
	ServiceName      string `json:"serviceName"`
	ServiceNamespace string `json:"serviceNamespace,omitempty"`
	Port             int32  `json:"port,omitempty"`
	PortName         string `json:"portName,omitempty"`
}

type AuthzRef struct {
	DefaultAction AuthzDefaultAction `json:"defaultAction,omitempty"`
}

type SharedRouteRegistrySpec struct {
	RouteMode    RouteMode   `json:"routeMode,omitempty"`
	HostPatterns []string    `json:"hostPatterns"`
	Upstream     UpstreamRef `json:"upstream"`
	AuthzRef     *AuthzRef   `json:"authzRef,omitempty"`
}

type SharedRouteRegistryStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	HTTPRouteName      string             `json:"httpRouteName,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type SharedRouteRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SharedRouteRegistrySpec   `json:"spec,omitempty"`
	Status            SharedRouteRegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type SharedRouteRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SharedRouteRegistry `json:"items"`
}
