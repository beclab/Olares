package k8sutil

import (
	"context"
	"fmt"
	"net"
	"time"

	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"
	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	applyCorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyMetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/pointer"
)

func GetL4ProxyNodeIP(ctx context.Context, waitTimeout time.Duration) (*string, error) {
	namespace := utils.EnvOrDefault("L4_PROXY_NAMESPACE", constants.OSSystemNamespace)
	return GetPodHostIPWithLabelSelector(ctx, waitTimeout, namespace, "app=l4-bfl-proxy")
}

func GetPodHostIPWithLabelSelector(ctx context.Context, waitTimeout time.Duration, namespace, labelSelector string) (*string, error) {
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return nil, errors.Errorf("new kube client in cluster: %v", err)
	}

	var nodeIP *string
	var observations int32

	err = wait.PollImmediate(time.Second, waitTimeout, func() (bool, error) {
		var podList *corev1.PodList
		podList, err = kc.Kubernetes().CoreV1().Pods(namespace).List(ctx,
			metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil && apierrors.IsNotFound(err) {
			return false, nil
		} else if err != nil {
			return false, errors.WithStack(err)
		}

		if podList != nil && len(podList.Items) > 0 {
			pod := podList.Items[0]
			if pod.Status.HostIP != "" {
				nodeIP = pointer.String(pod.Status.HostIP)
				observations++
			}
		}

		if observations > 2 {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return nodeIP, err
}

func GetMasterExternalIP(ctx context.Context) *string {
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		log.Warnf("new kube client: %v", err)
		return nil
	}

	var users iamV1alpha2.UserList
	err = kc.CtrlClient().List(ctx, &users)
	if err != nil {
		log.Warnf("list users: %v", err)
		return nil
	}

	var externalIP string

	for _, user := range users.Items {
		if role, ok := user.Annotations[constants.UserAnnotationOwnerRole]; ok && role == constants.RoleOwner {
			ip, ok1 := user.Annotations[constants.UserAnnotationPublicDomainIp]
			if ok1 && ip != "" {
				if _ip := net.ParseIP(ip); _ip != nil {
					externalIP = ip
					break
				}
			}
		}
	}

	if externalIP == "" {
		externalIP = utils.GetMyExternalIPAddr()
	}

	return pointer.String(externalIP)
}

func GetConfigMapData(ctx context.Context, ns, name string) (map[string]string, error) {
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		log.Warnf("new kube client: %v", err)
		return nil, err
	}
	cm, err := kc.Kubernetes().CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cm.Data, nil
}

func WriteConfigMapData(ctx context.Context, ns, name string, data map[string]string) error {
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		log.Warnf("new kube client: %v", err)
		return err
	}
	cmApply := &applyCorev1.ConfigMapApplyConfiguration{
		TypeMetaApplyConfiguration: applyMetav1.TypeMetaApplyConfiguration{
			Kind:       pointer.String("ConfigMap"),
			APIVersion: pointer.String(corev1.SchemeGroupVersion.String()),
		},
		ObjectMetaApplyConfiguration: &applyMetav1.ObjectMetaApplyConfiguration{
			Name:      pointer.String(name),
			Namespace: pointer.String(ns),
		},
		Data: data,
	}
	_, err = kc.Kubernetes().CoreV1().ConfigMaps(ns).Apply(
		ctx,
		cmApply,
		metav1.ApplyOptions{FieldManager: constants.ApplyPatchFieldManager},
	)
	return err
}

func RolloutRestartDeployment(ctx context.Context, ns, name string) error {
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		log.Warnf("new kube client: %v", err)
		return err
	}
	patchData := fmt.Sprintf(
		`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`,
		time.Now().Format(time.RFC3339))
	_, err = kc.Kubernetes().AppsV1().Deployments(ns).Patch(
		ctx,
		name,
		apitypes.StrategicMergePatchType,
		[]byte(patchData),
		metav1.PatchOptions{})
	if err != nil {
		return errors.Errorf("failed to patch deployment %s: %v", name, err)
	}
	return nil
}
