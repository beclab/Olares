package helmops

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const suspendAnnotation = "bytetrade.io/suspend-by"
const suspendCauseAnnotation = "bytetrade.io/suspend-cause"

// SuspendOrResumeApp suspends or resumes an application by patching
// Deployment/StatefulSet replicas. Restored from module-appservice
// pkg/appstate/utils.go — call sites must use this helper rather
// than reimplementing the patch loop.
func SuspendOrResumeApp(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager, replicas int32, stopOrResumeServer bool) error {
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
		}

		return nil
	}

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
				ns := appcfg.ChartNamespace(&chart, am.Spec.AppOwner)
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
			}
		}
	}

	return nil
}

// SuspendV1AppOrV2Client patches only the app's own workloads to 0.
func SuspendV1AppOrV2Client(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager) error {
	return SuspendOrResumeApp(ctx, cli, am, 0, false)
}

// SuspendV2AppAll patches the app plus shared server charts to 0.
func SuspendV2AppAll(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager) error {
	return SuspendOrResumeApp(ctx, cli, am, 0, true)
}

// ResumeV1AppOrV2AppClient patches only the app's own workloads to 1.
func ResumeV1AppOrV2AppClient(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager) error {
	return SuspendOrResumeApp(ctx, cli, am, 1, false)
}

// ResumeV2AppAll patches the app plus shared server charts to 1.
func ResumeV2AppAll(ctx context.Context, cli client.Client, am *appv1alpha1.ApplicationManager) error {
	return SuspendOrResumeApp(ctx, cli, am, 1, true)
}
