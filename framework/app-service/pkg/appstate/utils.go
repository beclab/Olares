package appstate

import (
	"context"
	"encoding/json"
	"fmt"
	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			switch workload := w.(type) {
			case *appsv1.Deployment:
				if check(targetAppName, workload.Name) {
					if workload.Annotations == nil {
						workload.Annotations = make(map[string]string)
					}
					workload.Annotations[suspendAnnotation] = "app-service"
					workload.Annotations[suspendCauseAnnotation] = "user operate"
					workload.Spec.Replicas = &replicas
					workloadName = workload.Namespace + "/" + workload.Name
				}
			case *appsv1.StatefulSet:
				if check(targetAppName, workload.Name) {
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
			delete(am.Annotations, api.AppStopAllKey)
			delete(am.Annotations, api.AppResumeAllKey)
			if err := cli.Update(ctx, am); err != nil {
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

func findServerPods(cli client.Client, subChart []appcfg.Chart, owner string) ([]corev1.Pod, error) {
	pods := make([]corev1.Pod, 0)
	for _, c := range subChart {
		if !c.Shared {
			continue
		}
		ns := c.Namespace(owner)
		var podList corev1.PodList
		err := cli.List(context.TODO(), &podList, client.InNamespace(ns))
		if err != nil {
			klog.Errorf("failed to list pods %v", err)
			return nil, err
		}
		pods = append(pods, podList.Items...)
	}
	return pods, nil
}

func findV1OrClientPods(cli client.Client, ns string) ([]corev1.Pod, error) {
	var podList corev1.PodList
	err := cli.List(context.TODO(), &podList, client.InNamespace(ns))
	if err != nil {
		klog.Errorf("get ns:%s pods err %v", ns, err)
		return nil, err
	}
	return podList.Items, nil

}

func isStartUp(am *appv1alpha1.ApplicationManager, cli client.Client) (bool, error) {
	var appconfig appcfg.ApplicationConfig
	err := json.Unmarshal([]byte(am.Spec.Config), &appconfig)
	if err != nil {
		return false, err
	}
	if appconfig.IsV2() && appconfig.IsMultiCharts() {
		serverPods, err := findServerPods(cli, appconfig.SubCharts, appconfig.OwnerName)
		if err != nil {
			return false, err
		}
		podNames := make([]string, 0)
		for _, p := range serverPods {
			podNames = append(podNames, p.Name)
		}
		serverStarted, err := utils.CheckIfStartup(cli, serverPods, true)
		if err != nil {
			return false, err
		}
		if !serverStarted {
			return false, nil
		}

	}
	clientPods, err := findV1OrClientPods(cli, am.Namespace)
	if err != nil {
		return false, err
	}
	clientStarted, err := utils.CheckIfStartup(cli, clientPods, false)
	if err != nil {
		return false, err
	}
	return clientStarted, nil
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
