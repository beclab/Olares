package store

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// User copy from kubesphere
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

	// password will be encrypted by mutating admission webhook
	// :validation:MinLength=6
	// :validation:MaxLength=64
	// :validation:Pattern=`^(.*[a-z].*[A-Z].*[0-9].*)$|^(.*[a-z].*[0-9].*[A-Z].*)$|^(.*[A-Z].*[a-z].*[0-9].*)$|^(.*[A-Z].*[0-9].*[a-z].*)$|^(.*[0-9].*[a-z].*[A-Z].*)$|^(.*[0-9].*[A-Z].*[a-z].*)$|^(\$2[ayb]\$.{56})$`
	// Password pattern is tricky here.
	// The rule is simple: length between [6,64], at least one uppercase letter, one lowercase letter, one digit.
	// The regexp in console(javascript) is quite straightforward: ^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[^]{6,64}$
	// But in Go, we don't have ?= (back tracking) capability in regexp (also in CRD validation pattern)
	// So we adopted an alternative scheme to achieve.
	// Use 6 different regexp to combine to achieve the same effect.
	// These six schemes enumerate the arrangement of numbers, uppercase letters, and lowercase letters that appear for the first time.
	// - ^(.*[a-z].*[A-Z].*[0-9].*)$ stands for lowercase letter comes first, then followed by an uppercase letter, then a digit.
	// - ^(.*[a-z].*[0-9].*[A-Z].*)$ stands for lowercase letter comes first, then followed by a digit, then an uppercase leeter.
	// - ^(.*[A-Z].*[a-z].*[0-9].*)$ ...
	// - ^(.*[A-Z].*[0-9].*[a-z].*)$ ...
	// - ^(.*[0-9].*[a-z].*[A-Z].*)$ ...
	// - ^(.*[0-9].*[A-Z].*[a-z].*)$ ...
	// Last but not least, the bcrypt string is also included to match the encrypted password. ^(\$2[ayb]\$.{56})$
	EncryptedPassword string `json:"password,omitempty"`
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
