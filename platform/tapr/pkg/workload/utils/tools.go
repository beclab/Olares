package utils

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetUserDBPVCName(ctx context.Context, client *kubernetes.Clientset, namespace string) (string, error) {
	claims, err := client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{LabelSelector: "bytetrade.io/user-storage=dbdata"})
	if err != nil {
		return "", err
	}
	if len(claims.Items) != 1 || claims.Items[0].DeletionTimestamp != nil {
		return "", fmt.Errorf("expected one dbdata PVC in %s", namespace)
	}
	path := claims.Items[0].Annotations["bytetrade.io/host-path"]
	if path == "" {
		return "", fmt.Errorf("dbdata PVC in %s has no host path", namespace)
	}
	return path, nil
}

func AnyPtr[T any](t T) *T {
	return &t
}
