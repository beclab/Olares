package utils

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	UserDBDataPVC = "dbdata_hostpath"
)

func GetUserDBPVCName(ctx context.Context, client *kubernetes.Clientset, namespace string) (string, error) {
	bflSts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, "bfl", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	pvc, ok := bflSts.Annotations[UserDBDataPVC]
	if !ok {
		return "", fmt.Errorf("cannot find user db data pvc, %s", UserDBDataPVC)
	}

	return pvc, nil
}

func AnyPtr[T any](t T) *T {
	return &t
}
