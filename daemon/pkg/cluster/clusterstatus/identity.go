package clusterstatus

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// kubeSystemNamespace is the object whose UID identifies this cluster.
const kubeSystemNamespace = "kube-system"

// ClusterID reports the identifier of the Olares installation this daemon
// belongs to: the UID of the kube-system namespace.
//
// It is the one identifier already scoped to exactly what a cluster id has to
// mean. Kubernetes mints it when the cluster is created and never changes it,
// so upgrades, reboots and nodes joining or leaving leave it alone, and
// rebuilding the cluster produces a different one — which is the behaviour
// asked of it, since that is a different installation.
//
// The alternatives are all somebody else's identifier. The Olares ID names the
// person, and survives them reinstalling onto new hardware. A TermiPass device
// id names one client's binding and changes when the client is re-paired. The
// cloud storage cluster-id belongs to a backup account. Reusing any of them
// would make two questions share one answer and drift apart later.
//
// Reading it is a get on an object the cluster already owns, so nothing here
// needs write access and there is no new record to keep in step.
func ClusterID(ctx context.Context, client kubernetes.Interface) (string, error) {
	ns, err := client.CoreV1().Namespaces().Get(ctx, kubeSystemNamespace, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("read the %s namespace: %w", kubeSystemNamespace, err)
	}
	// An empty UID is not an identifier. Returning it would give every
	// cluster that hits this the same one.
	if ns.UID == "" {
		return "", errors.New("the " + kubeSystemNamespace + " namespace carries no UID")
	}
	return string(ns.UID), nil
}
