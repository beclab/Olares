/*
Copyright 2019 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceKindUser                      = "User"
	ResourcesSingularUser                 = "user"
	ResourcesPluralUser                   = "users"
	ResourceKindGlobalRoleBinding         = "GlobalRoleBinding"
	ResourcesSingularGlobalRoleBinding    = "globalrolebinding"
	ResourcesPluralGlobalRoleBinding      = "globalrolebindings"
	ResourceKindClusterRoleBinding        = "ClusterRoleBinding"
	ResourcesSingularClusterRoleBinding   = "clusterrolebinding"
	ResourcesPluralClusterRoleBinding     = "clusterrolebindings"
	ResourceKindRoleBinding               = "RoleBinding"
	ResourcesSingularRoleBinding          = "rolebinding"
	ResourcesPluralRoleBinding            = "rolebindings"
	ResourceKindGlobalRole                = "GlobalRole"
	ResourcesSingularGlobalRole           = "globalrole"
	ResourcesPluralGlobalRole             = "globalroles"
	ResourceKindWorkspaceRoleBinding      = "WorkspaceRoleBinding"
	ResourcesSingularWorkspaceRoleBinding = "workspacerolebinding"
	ResourcesPluralWorkspaceRoleBinding   = "workspacerolebindings"
	ResourceKindWorkspaceRole             = "WorkspaceRole"
	ResourcesSingularWorkspaceRole        = "workspacerole"
	ResourcesPluralWorkspaceRole          = "workspaceroles"
	ResourceKindClusterRole               = "ClusterRole"
	ResourcesSingularClusterRole          = "clusterrole"
	ResourcesPluralClusterRole            = "clusterroles"
	ResourceKindRole                      = "Role"
	ResourcesSingularRole                 = "role"
	ResourcesPluralRole                   = "roles"
	RegoOverrideAnnotation                = "iam.kubesphere.io/rego-override"
	AggregationRolesAnnotation            = "iam.kubesphere.io/aggregation-roles"
	GlobalRoleAnnotation                  = "iam.kubesphere.io/globalrole"
	WorkspaceRoleAnnotation               = "iam.kubesphere.io/workspacerole"
	ClusterRoleAnnotation                 = "iam.kubesphere.io/clusterrole"
	GrantedClustersAnnotation             = "iam.kubesphere.io/granted-clusters"
	UninitializedAnnotation               = "iam.kubesphere.io/uninitialized"
	LastPasswordChangeTimeAnnotation      = "iam.kubesphere.io/last-password-change-time"
	RoleAnnotation                        = "iam.kubesphere.io/role"
	RoleTemplateLabel                     = "iam.kubesphere.io/role-template"
	ScopeLabelFormat                      = "scope.kubesphere.io/%s"
	UserReferenceLabel                    = "iam.kubesphere.io/user-ref"
	IdentifyProviderLabel                 = "iam.kubesphere.io/identify-provider"
	OriginUIDLabel                        = "iam.kubesphere.io/origin-uid"
	ServiceAccountReferenceLabel          = "iam.kubesphere.io/serviceaccount-ref"
	FieldEmail                            = "email"
	ExtraEmail                            = FieldEmail
	ExtraIdentityProvider                 = "idp"
	ExtraUID                              = "uid"
	ExtraUsername                         = "username"
	ExtraDisplayName                      = "displayName"
	ExtraUninitialized                    = "uninitialized"
	InGroup                               = "ingroup"
	NotInGroup                            = "notingroup"
	AggregateTo                           = "aggregateTo"
	ScopeWorkspace                        = "workspace"
	ScopeCluster                          = "cluster"
	ScopeNamespace                        = "namespace"
	ScopeDevOps                           = "devops"
	PlatformAdmin                         = "platform-admin"
	NamespaceAdmin                        = "admin"
	ClusterAdmin                          = "cluster-admin"
	PreRegistrationUser                   = "system:pre-registration"
	PreRegistrationUserGroup              = "pre-registration"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true

// User is the Schema for the users API
// +kubebuilder:printcolumn:name="Email",type="string",JSONPath=".spec.email"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:resource:categories="iam",scope="Cluster"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type User struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UserSpec `json:"spec"`
	// +optional
	Status UserStatus `json:"status,omitempty"`
}

type FinalizerName string

// UserSpec defines the desired state of User
type UserSpec struct {
	// Unique email address(https://www.ietf.org/rfc/rfc5322.txt).
	Email string `json:"email"`
	// InitialPassword only for the first user that need sync from here to lldap
	// +optional
	InitialPassword string `json:"initialPassword,omitempty"`
	// The preferred written or spoken language for the user.
	// +optional
	Lang string `json:"lang,omitempty"`
	// Description of the user.
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	DisplayName string `json:"displayName,omitempty"`
	// +optional
	Groups []string `json:"groups,omitempty"`
}

type UserState string

// These are the valid phases of a user.
const (
	// UserActive means the user is available.
	UserActive UserState = "Active"
	// UserDisabled means the user is disabled.
	UserDisabled UserState = "Disabled"
	// UserAuthLimitExceeded means restrict user login.
	UserAuthLimitExceeded UserState = "AuthLimitExceeded"

	AuthenticatedSuccessfully = "authenticated successfully"
)

// UserStatus defines the observed state of User
type UserStatus struct {
	// The user status
	// +optional
	State UserState `json:"state,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// Last login attempt timestamp
	// +optional
	LastLoginTime *metav1.Time `json:"lastLoginTime,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GlobalRole struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Rules holds all the PolicyRules for this GlobalRole
	// +optional
	Rules []rbacv1.PolicyRule `json:"rules" protobuf:"bytes,2,rep,name=rules"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GlobalRoleList contains a list of GlobalRole
type GlobalRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRole `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GlobalRoleBinding is the Schema for the globalrolebindings API
type GlobalRoleBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Subjects holds references to the objects the role applies to.
	// +optional
	Subjects []rbacv1.Subject `json:"subjects,omitempty" protobuf:"bytes,2,rep,name=subjects"`

	// RoleRef can only reference a GlobalRole.
	// If the RoleRef cannot be resolved, the Authorizer must return an error.
	RoleRef rbacv1.RoleRef `json:"roleRef" protobuf:"bytes,3,opt,name=roleRef"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GlobalRoleBindingList contains a list of GlobalRoleBinding
type GlobalRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRoleBinding `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RoleBase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:EmbeddedResource
	Role runtime.RawExtension `json:"role"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoleBaseList contains a list of RoleBase
type RoleBaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleBase `json:"items"`
}
