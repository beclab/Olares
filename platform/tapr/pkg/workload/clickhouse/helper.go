package clickhouse

import (
	"context"
	"fmt"

	kbappsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ListClickHouseClusters(ctx context.Context, ctrlClient client.Client, namespace string) (clusters []kbappsv1.Cluster, err error) {
	var clusterList kbappsv1.ClusterList
	err = ctrlClient.List(ctx, &clusterList)
	if err != nil {
		return nil, err
	}
	for _, c := range clusterList.Items {
		if c.Labels != nil && (c.Labels["clusterdefinition.kubeblocks.io/name"] == "clickhouse" || (c.Name == "clickhouse" && c.Spec.ClusterDef == "clickhouse")) {
			clusters = append(clusters, c)
		}
	}
	return clusters, nil
}

func FindClickHouseAdminUser(ctx context.Context, k8sClient *kubernetes.Clientset, namespace string) (user, password string, err error) {
	secret, err := k8sClient.CoreV1().Secrets(namespace).Get(ctx, "clickhouse-account-info", metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return "admin", string(secret.Data["password"]), nil
}

func GetDatabaseName(appNamespace, name string) string {
	return fmt.Sprintf("%s_%s", appNamespace, name)
}
