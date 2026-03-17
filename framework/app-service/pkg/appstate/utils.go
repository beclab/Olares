package appstate

import (
	"context"
	"encoding/json"
	"fmt"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const suspendAnnotation = "bytetrade.io/suspend-by"
const suspendCauseAnnotation = "bytetrade.io/suspend-cause"

// suspendOrResumeApp suspends or resumes an application.
func suspendOrResumeApp(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager, replicas int32, stopOrResumeServer bool) error {
	suspendOrResume := func(list client.ObjectList, targetNamespace, targetAppName string) error {
		err := cli.List(ctx, list, client.InNamespace(targetNamespace))
		if err != nil {
			klog.Errorf("Failed to get workload namespace=%s err=%v", targetNamespace, err)
			return err
		}

		listObjects, err := apimeta.ExtractList(list)
		if err != nil {
			klog.Errorf("Failed to extract list namespace=%s err=%v", targetNamespace, err)
			return err
		}
		check := func(appName, deployName string) bool {
			if targetNamespace == fmt.Sprintf("user-space-%s", am.Spec.AppOwner) ||
				targetNamespace == fmt.Sprintf("user-system-%s", am.Spec.AppOwner) ||
				targetNamespace == "os-platform" ||
				targetNamespace == "os-framework" {
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
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				switch workload := w.(type) {
				case *appsv1.Deployment:
					var latest appsv1.Deployment
					if err := cli.Get(ctx, types.NamespacedName{Name: workload.Name, Namespace: workload.Namespace}, &latest); err != nil {
						return err
					}
					if !check(targetAppName, latest.Name) {
						return nil
					}
					if latest.Annotations == nil {
						latest.Annotations = make(map[string]string)
					}
					latest.Annotations[suspendAnnotation] = "app-service"
					latest.Annotations[suspendCauseAnnotation] = "user operate"
					latest.Spec.Replicas = &replicas
					workloadName = latest.Namespace + "/" + latest.Name
					return cli.Update(ctx, &latest)
				case *appsv1.StatefulSet:
					var latest appsv1.StatefulSet
					if err := cli.Get(ctx, types.NamespacedName{Name: workload.Name, Namespace: workload.Namespace}, &latest); err != nil {
						return err
					}
					if !check(targetAppName, latest.Name) {
						return nil
					}
					if latest.Annotations == nil {
						latest.Annotations = make(map[string]string)
					}
					latest.Annotations[suspendAnnotation] = "app-service"
					latest.Annotations[suspendCauseAnnotation] = "user operate"
					latest.Spec.Replicas = &replicas
					workloadName = latest.Namespace + "/" + latest.Name
					return cli.Update(ctx, &latest)
				}
				return nil
			})
			if err != nil {
				klog.Errorf("Failed to scale workload name=%s err=%v", workloadName, err)
				return err
			}
			if workloadName != "" {
				if replicas == 0 {
					klog.Infof("Try to suspend workload name=%s", workloadName)
				} else {
					klog.Infof("Try to resume workload name=%s", workloadName)
				}
				klog.Infof("Success to operate workload name=%s", workloadName)
			}
		} // end list object loop

		return nil
	} // end of suspend func

	var deploymentList appsv1.DeploymentList
	err := suspendOrResume(&deploymentList, am.Spec.AppNamespace, am.Spec.AppName)
	if err != nil {
		return err
	}

	var stsList appsv1.StatefulSetList
	err = suspendOrResume(&stsList, am.Spec.AppNamespace, am.Spec.AppName)
	if err != nil {
		return err
	}

	// If stopOrResumeServer is true, also suspend/resume shared server charts for V2 apps
	if stopOrResumeServer {
		var appCfg *appcfg.ApplicationConfig
		if err := json.Unmarshal([]byte(am.Spec.Config), &appCfg); err != nil {
			klog.Warningf("failed to unmarshal app config for stopServer check: %v", err)
			return err
		}

		if appCfg != nil && appCfg.IsV2() && appCfg.HasClusterSharedCharts() {
			for _, chart := range appCfg.SubCharts {
				if !chart.Shared {
					continue
				}
				ns := chart.Namespace(am.Spec.AppOwner)
				if replicas == 0 {
					klog.Infof("suspending shared chart %s in namespace %s", chart.Name, ns)
				} else {
					klog.Infof("resuming shared chart %s in namespace %s", chart.Name, ns)
				}

				var sharedDeploymentList appsv1.DeploymentList
				if err := suspendOrResume(&sharedDeploymentList, ns, chart.Name); err != nil {
					klog.Errorf("failed to scale deployments in shared chart %s namespace %s: %v", chart.Name, ns, err)
					return err
				}

				var sharedStsList appsv1.StatefulSetList
				if err := suspendOrResume(&sharedStsList, ns, chart.Name); err != nil {
					klog.Errorf("failed to scale statefulsets in shared chart %s namespace %s: %v", chart.Name, ns, err)
					return err
				}
			}
		}

		// Reset the stop-all/resume-all annotation after processing
		if am.Annotations != nil {
			if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				var latest appv1alpha1.ApplicationManager
				if err := cli.Get(ctx, types.NamespacedName{Name: am.Name}, &latest); err != nil {
					return err
				}
				if latest.Annotations == nil {
					return nil
				}
				delete(latest.Annotations, api.AppStopAllKey)
				delete(latest.Annotations, api.AppResumeAllKey)
				delete(latest.Annotations, api.AppStopByControllerDuePendingPod)
				return cli.Update(ctx, &latest)
			}); err != nil {
				klog.Warningf("failed to reset stop-all/resume-all annotations for app=%s owner=%s: %v", am.Spec.AppName, am.Spec.AppOwner, err)
				// Don't return error, operation already succeeded
			}
		}
	}

	return nil
}

func suspendV1AppOrV2Client(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager) error {
	return suspendOrResumeApp(ctx, cli, am, 0, false)
}

func suspendV2AppAll(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager) error {
	return suspendOrResumeApp(ctx, cli, am, 0, true)
}

func resumeV1AppOrV2AppClient(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager) error {
	return suspendOrResumeApp(ctx, cli, am, 1, false)
}

func resumeV2AppAll(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager) error {
	return suspendOrResumeApp(ctx, cli, am, 1, true)
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
