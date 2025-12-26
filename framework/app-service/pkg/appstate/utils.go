package appstate

import (
	"context"
	"fmt"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const suspendAnnotation = "bytetrade.io/suspend-by"
const suspendCauseAnnotation = "bytetrade.io/suspend-cause"

func suspendOrResumeApp(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager, replicas int32) error {
	suspend := func(list client.ObjectList) error {
		namespace := am.Spec.AppNamespace
		err := cli.List(ctx, list, client.InNamespace(namespace))
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("Failed to get workload namespace=%s err=%v", namespace, err)
			return err
		}

		listObjects, err := apimeta.ExtractList(list)
		if err != nil {
			klog.Errorf("Failed to extract list namespace=%s err=%v", namespace, err)
			return err
		}
		check := func(appName, deployName string) bool {
			if namespace == fmt.Sprintf("user-space-%s", am.Spec.AppOwner) ||
				namespace == fmt.Sprintf("user-system-%s", am.Spec.AppOwner) ||
				namespace == "os-platform" ||
				namespace == "os-framework" {
				if appName == deployName {
					return true
				}
			} else {
				return true
			}
			return false
		}

		//var zeroReplica int32 = 0
		for _, w := range listObjects {
			workloadName := ""
			switch workload := w.(type) {
			case *appsv1.Deployment:
				if check(am.Spec.AppName, workload.Name) {
					if workload.Annotations == nil {
						workload.Annotations = make(map[string]string)
					}
					workload.Annotations[suspendAnnotation] = "app-service"
					workload.Annotations[suspendCauseAnnotation] = "user operate"
					workload.Spec.Replicas = &replicas
					workloadName = workload.Namespace + "/" + workload.Name
				}
			case *appsv1.StatefulSet:
				if check(am.Spec.AppName, workload.Name) {
					if workload.Annotations == nil {
						workload.Annotations = make(map[string]string)
					}
					workload.Annotations[suspendAnnotation] = "app-service"
					workload.Annotations[suspendCauseAnnotation] = "user operate"
					workload.Spec.Replicas = &replicas
					workloadName = workload.Namespace + "/" + workload.Name
				}
			}
			if replicas == 0 {
				klog.Infof("Try to suspend workload name=%s", workloadName)
			} else {
				klog.Infof("Try to resume workload name=%s", workloadName)
			}
			err := cli.Update(ctx, w.(client.Object))
			if err != nil {
				klog.Errorf("Failed to scale workload name=%s err=%v", workloadName, err)
				return err
			}

			klog.Infof("Success to operate workload name=%s", workloadName)
		} // end list object loop

		return nil
	} // end of suspend func

	var deploymentList appsv1.DeploymentList
	err := suspend(&deploymentList)
	if err != nil {
		return err
	}

	var stsList appsv1.StatefulSetList
	err = suspend(&stsList)

	return err
}

func isStartUp(am *appv1alpha1.ApplicationManager, cli client.Client) (bool, error) {
	var labelSelector string
	var deployment appsv1.Deployment

	err := cli.Get(context.TODO(), types.NamespacedName{Name: am.Spec.AppName, Namespace: am.Spec.AppNamespace}, &deployment)

	if err == nil {
		labelSelector = metav1.FormatLabelSelector(deployment.Spec.Selector)
	}

	if apierrors.IsNotFound(err) {
		var sts appsv1.StatefulSet
		err = cli.Get(context.TODO(), types.NamespacedName{Name: am.Spec.AppName, Namespace: am.Spec.AppNamespace}, &sts)
		if err != nil {
			return false, err

		}
		labelSelector = metav1.FormatLabelSelector(sts.Spec.Selector)
	}
	var pods corev1.PodList
	//pods, err := h.client.KubeClient.Kubernetes().CoreV1().Pods(h.app.Namespace).
	//	List(h.ctx, metav1.ListOptions{LabelSelector: labelSelector})
	selector, _ := labels.Parse(labelSelector)
	err = cli.List(context.TODO(), &pods, &client.ListOptions{Namespace: am.Spec.AppNamespace, LabelSelector: selector})
	if len(pods.Items) == 0 {
		return false, errors.New("no pod found..")
	}
	for _, pod := range pods.Items {
		totalContainers := len(pod.Spec.Containers)
		startedContainers := 0
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]
			if *container.Started == true {
				startedContainers++
			}
		}
		if startedContainers == totalContainers {
			return true, nil
		}
	}
	return false, nil
}

func makeRecord(am *appv1alpha1.ApplicationManager, status appv1alpha1.ApplicationManagerState, message string) *appv1alpha1.OpRecord {
	if am == nil {
		return nil
	}
	now := metav1.Now()
	return &appv1alpha1.OpRecord{
		OpType:    am.Status.OpType,
		OpID:      am.Status.OpID,
		Source:    am.Spec.Source,
		Version:   am.Annotations[api.AppVersionKey],
		Message:   message,
		Status:    status,
		StateTime: &now,
	}
}
