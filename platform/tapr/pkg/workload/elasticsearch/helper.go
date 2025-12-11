package elasticsearch

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/tapr/pkg/constants"
	kbappsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListRabbitMQClusters(ctx context.Context, ctrlClient client.Client, namespace string) (clusters []kbappsv1.Cluster, err error) {
	var clusterList kbappsv1.ClusterList
	err = ctrlClient.List(ctx, &clusterList)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusterList.Items {
		if cluster.Labels != nil && cluster.Labels[constants.ClusterInstanceNameKey] == "elasticsearch" {
			clusters = append(clusters, cluster)
		}
	}
	return clusters, nil
}

func FindElasticsearchAdminUser(ctx context.Context, k8sClient *kubernetes.Clientset, namespace string) (user, password string, err error) {
	secret, err := k8sClient.CoreV1().Secrets(namespace).Get(ctx, "elasticsearch-mdit-account-elastic", metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return string(secret.Data["username"]), string(secret.Data["password"]), nil
}

func GetIndexName(appNamespace, name string) string {
	return fmt.Sprintf("%s-%s", appNamespace, name)
}
