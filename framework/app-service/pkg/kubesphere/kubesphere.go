package kubesphere

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	userAnnotationCPULimitKey    = "bytetrade.io/user-cpu-limit"
	userAnnotationMemoryLimitKey = "bytetrade.io/user-memory-limit"
	userIndex                    = "bytetrade.io/user-index"
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

// GetAdminUserList returns admin list, an error if there is any.
func GetAdminUserList(ctx context.Context, kubeConfig *rest.Config) ([]UserInfo, error) {
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
		if role == "owner" || role == "admin" {
			adminUserList = append(adminUserList, UserInfo{Name: u.GetName(), Role: role})
		}
	}

	return adminUserList, nil
}

func IsAdmin(ctx context.Context, kubeConfig *rest.Config, owner string) (bool, error) {
	adminList, err := GetAdminUserList(ctx, kubeConfig)
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
	adminList, err := GetAdminUserList(ctx, kubeConfig)
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
