package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"k8s.io/klog/v2"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

// MatchVersion check if the version satisfies the constraint.
func MatchVersion(version, constraint string) bool {
	if len(version) == 0 {
		return true
	}
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		klog.Errorf("Invalid constraint=%s err=%v, ", constraint, err)
		return false
	}
	v, err := semver.NewVersion(version)
	if err != nil {
		klog.Errorf("Invalid version=%s err=%v", version, err)
		return false
	}

	return c.Check(v)
}

// AppNamespace returns the namespace of an application.
func AppNamespace(app, owner, ns string) (string, error) {
	// can not get app namespace info, so have to list
	if len(ns) == 0 {
		client, err := GetClient()
		if err != nil {
			return "", err
		}
		appMgr, err := FindAppManager(context.TODO(), client, app, owner)
		if err == nil {
			return appMgr.Spec.AppNamespace, nil
		}
		if !apierrors.IsNotFound(err) {
			return "", err
		}
	}

	if strings.HasPrefix(ns, "user-space") {
		app = "user-space"
	} else {
		if ns != "" {
			return ns, nil
		}
	}
	if userspace.IsSysApp(app) {
		app = "user-space"
	}
	return fmt.Sprintf("%s-%s", app, owner), nil
}

// FindAppManager resolves a shared app before a legacy per-user registration.
// The system app name must stay intact during lookup; only the namespace
// fallback uses user-space when no registration exists.
func FindAppManager(ctx context.Context, client versioned.Interface, app, owner string) (*appv1alpha1.ApplicationManager, error) {
	managers, err := client.AppV1alpha1().ApplicationManagers().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var owned *appv1alpha1.ApplicationManager
	for i := range managers.Items {
		am := &managers.Items[i]
		if am.Spec.AppName != app {
			continue
		}
		if am.Labels[constants.AppSharedLabel] == constants.AppSharedTrue {
			return am, nil
		}
		if am.Spec.AppOwner == owner {
			owned = am
		}
	}
	if owned != nil {
		return owned, nil
	}
	return nil, apierrors.NewNotFound(schema.GroupResource{Group: "app.bytetrade.io", Resource: "applicationmanagers"}, app)
}

// GetClient returns versioned ClientSet.
func GetClient() (*versioned.Clientset, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	client, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func EnvOrDefault(name, def string) string {
	v := os.Getenv(name)

	if v == "" && def != "" {
		return def
	}
	return v
}
