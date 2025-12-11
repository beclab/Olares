package controllers

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var UsersGVR = schema.GroupVersionResource{
	Group:    "iam.kubesphere.io",
	Version:  "v1alpha2",
	Resource: "users",
}

func getAdminUser(ctx context.Context, client dynamic.Interface) (string, error) {
	unstructuredUsers, err := client.Resource(UsersGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", errors.WithStack(err)
	}

	for _, u := range unstructuredUsers.Items {
		if a, ok := u.GetAnnotations()["bytetrade.io/owner-role"]; ok && (a == "owner" || a == "admin") {
			return u.GetName(), nil
		}
	}

	return "", errors.New("admin user not found")
}
