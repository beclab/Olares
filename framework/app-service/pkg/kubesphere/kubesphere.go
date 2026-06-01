package kubesphere

import (
	"context"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	annotationGroup              = "bytetrade.io"
	userAnnotationZoneKey        = fmt.Sprintf("%s/zone", annotationGroup)
	userAnnotationOwnerRole      = fmt.Sprintf("%s/owner-role", annotationGroup)
	userAnnotationWizardStatus   = fmt.Sprintf("%s/wizard-status", annotationGroup)
	userAnnotationCPULimitKey    = "bytetrade.io/user-cpu-limit"
	userAnnotationMemoryLimitKey = "bytetrade.io/user-memory-limit"
	userIndex                    = "bytetrade.io/user-index"
)

const (
	// wizardStatusCompleted marks a user as fully activated. A user is
	// only considered activated when this annotation equals "completed"
	// AND the User's Status.State is "Created".
	wizardStatusCompleted = "completed"
	// userStateCreated is the iam.kubesphere.io User Status.State value
	// that indicates the user record has been fully created.
	userStateCreated = "Created"
)

const (
	userOwnerRole = "owner"
	userAdminRole = "admin"
)

type Options struct {
	JwtSecret string `yaml:"jwtSecret"`
}

type Config struct {
	AuthenticationOptions *Options `yaml:"authentication,omitempty"`
}

type Type string

// GetUserZone returns user zone, an error if there is any.
func GetUserZone(ctx context.Context, username string) (string, error) {
	return GetUserAnnotation(ctx, username, userAnnotationZoneKey)
}

// GetUserRole returns user role, an error if there is any.
func GetUserRole(ctx context.Context, username string) (string, error) {
	return GetUserAnnotation(ctx, username, userAnnotationOwnerRole)
}

// GetUserAnnotation returns user annotation, an error if there is any.
func GetUserAnnotation(ctx context.Context, username, annotation string) (string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	config, err := ctrl.GetConfig()
	if err != nil {
		return "", err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", err
	}
	data, err := client.Resource(gvr).Get(ctx, username, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get user=%s err=%v", username, err)
		return "", err
	}

	a, ok := data.GetAnnotations()[annotation]
	if !ok {
		return "", fmt.Errorf("user annotation %s not found", annotation)
	}

	return a, nil
}

// GetUserCPULimit returns user cpu limit value, an error if there is any.
func GetUserCPULimit(ctx context.Context, username string) (string, error) {
	return GetUserAnnotation(ctx, username, userAnnotationCPULimitKey)
}

// GetUserMemoryLimit returns user memory limit value, an error if there is any.
func GetUserMemoryLimit(ctx context.Context, username string) (string, error) {
	return GetUserAnnotation(ctx, username, userAnnotationMemoryLimitKey)
}

// GetAdminUsername returns admin username, an error if there is any.
func GetAdminUsername(ctx context.Context, kubeConfig *rest.Config) (string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	client, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return "", err
	}
	data, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get user list err=%v", err)
		return "", err
	}

	var admin string
	for _, u := range data.Items {
		if u.Object == nil {
			continue
		}
		annotations := u.GetAnnotations()
		role := annotations[userAnnotationOwnerRole]
		if role == "owner" || role == "admin" {
			admin = u.GetName()
			break
		}
	}

	return admin, nil
}

func GetUserIndexByName(ctx context.Context, name string) (string, error) {
	return GetUserAnnotation(ctx, name, userIndex)
}

type UserInfo struct {
	Name string
	Role string
}

// GetOwnerOrAdminList returns owner/admin list, an error if there is any.
func GetOwnerOrAdminList(ctx context.Context, kubeConfig *rest.Config) ([]UserInfo, error) {
	adminUserList := make([]UserInfo, 0)

	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	client, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return adminUserList, err
	}
	data, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get user list err=%v", err)
		return adminUserList, err
	}

	for _, u := range data.Items {
		if u.Object == nil {
			continue
		}
		annotations := u.GetAnnotations()
		role := annotations[userAnnotationOwnerRole]
		if role == userOwnerRole || role == userAdminRole {
			adminUserList = append(adminUserList, UserInfo{Name: u.GetName(), Role: role})
		}
	}

	return adminUserList, nil
}

// GetActivatedUsers returns the names of users that have completed
// activation. A user is considered activated when their
// `bytetrade.io/wizard-status` annotation equals "completed" AND their
// User Status.State equals "Created". v3 / shared app event fan-out
// targets exactly this set so unactivated users (mid-wizard, ephemeral
// shells, etc.) never receive app lifecycle messages they cannot act on.
func GetActivatedUsers(ctx context.Context, kubeConfig *rest.Config) ([]string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	client, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	data, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get user list err=%v", err)
		return nil, err
	}

	users := make([]string, 0, len(data.Items))
	for _, u := range data.Items {
		if u.Object == nil {
			continue
		}
		if u.GetAnnotations()[userAnnotationWizardStatus] != wizardStatusCompleted {
			continue
		}
		// `status.state` is a string on the iam User CR (e.g. "Created",
		// "Creating", "Deleting"). Read it via the unstructured object
		// since we already have it on hand.
		state, _, _ := unstructured.NestedString(u.Object, "status", "state")
		if state != userStateCreated {
			continue
		}
		users = append(users, u.GetName())
	}
	return users, nil
}

func IsAdmin(ctx context.Context, kubeConfig *rest.Config, owner string) (bool, error) {
	adminList, err := GetOwnerOrAdminList(ctx, kubeConfig)
	if err != nil {
		return false, err
	}
	for _, user := range adminList {
		if user.Name == owner {
			return true, nil
		}
	}
	return false, nil
}

func GetOwner(ctx context.Context, kubeConfig *rest.Config) (string, error) {
	adminList, err := GetOwnerOrAdminList(ctx, kubeConfig)
	if err != nil {
		return "", err
	}
	for _, user := range adminList {
		if user.Role == "owner" {
			return user.Name, nil
		}
	}
	return "", errors.New("user with role owner not found")
}

// GetClusterOwner returns the cluster's primary owner user (the one
// carrying bytetrade.io/owner-role=owner) using ambient controller
// runtime kubeconfig. Convenience wrapper around GetOwner for callers
// that do not already hold a *rest.Config — see
// (*appcfg.ApplicationConfig).GetOwnerName, which delegates here for
// v3 / cluster-shared apps.
func GetClusterOwner(ctx context.Context) (string, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return "", err
	}
	return GetOwner(ctx, config)
}
