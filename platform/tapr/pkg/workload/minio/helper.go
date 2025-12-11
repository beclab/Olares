package minio

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/tapr/pkg/constants"

	kbappsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ListMinioClusters(ctx context.Context, ctrlClient client.Client, namespace string) (clusters []kbappsv1.Cluster, err error) {
	var clusterList kbappsv1.ClusterList
	err = ctrlClient.List(ctx, &clusterList)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusterList.Items {
		if cluster.Labels != nil && cluster.Labels[constants.ClusterInstanceNameKey] == "minio" {
			clusters = append(clusters, cluster)
		}
	}
	return clusters, nil
}

func FindMinioAdminUser(ctx context.Context, k8sClient *kubernetes.Clientset, namespace string) (user, password string, err error) {
	secret, err := k8sClient.CoreV1().Secrets(namespace).Get(ctx, "minio-minio-account-root", metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to find mongo user and password ")
		return
	}
	return string(secret.Data["username"]), string(secret.Data["password"]), nil
}

func GetBucketName(namespace, db string) string {
	return fmt.Sprintf("%s-%s", namespace, db)
}
