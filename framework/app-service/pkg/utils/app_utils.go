package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/framework/app-service/pkg/generated/clientset/versioned"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
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
	if userspace.IsSysApp(app) {
		app = "user-space"
	}
	// can not get app namespace info, so have to list
	if len(ns) == 0 {
		client, err := GetClient()
		if err != nil {
			return "", err
		}
		appMgr, err := client.AppV1alpha1().ApplicationManagers().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		for _, a := range appMgr.Items {
			if a.Spec.AppName == app && a.Spec.AppOwner == owner {
				return a.Spec.AppNamespace, nil
			}
		}
	}

	if strings.HasPrefix(ns, "user-space") {
		app = "user-space"
	} else if strings.HasPrefix(ns, "user-system") {
		app = "user-system"
	} else {
		if ns != "" {
			return ns, nil
		}
	}
	return fmt.Sprintf("%s-%s", app, owner), nil
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
