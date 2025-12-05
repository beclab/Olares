package argo

import (
	"context"

	"bytetrade.io/web3os/app-service/pkg/constants"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

// UpdateWorkflowInNamespace update workflow namespace info.
func UpdateWorkflowInNamespace(ctx context.Context, kubeConfig *rest.Config, workflowName, namespace string, instanceID string, owner, title string, data []map[string]interface{}) error {
	gvr := schema.GroupVersionResource{
		Group:    workflow.Group,
		Version:  workflow.Version,
		Resource: workflow.CronWorkflowPlural,
	}
	client, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		klog.Errorf("Failed to create dynamic client err=%v", err)
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Errorf("Failed to create k8s client err=%v", err)
		return err
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {

		err = createFakes3SecretAndRole(kubeClient, namespace)
		if err != nil {
			klog.Errorf("Failed to create fakes3secret and role namespace=%s err=%v", namespace, err)
			return err
		}

		err = saveProvidersToConfigMap(kubeClient, namespace, data)
		if err != nil {
			klog.Errorf("Failed to saveProvidersToConfigMap namespace=%s err=%v", namespace, err)
			return err
		}

		objs, err := client.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to list cronworkflow namespace=%s err=%v", namespace, err)
			return err
		}

		for _, obj := range objs.Items {
			labels := obj.GetLabels()
			labels[constants.InstanceIDLabel] = instanceID
			labels[constants.WorkflowOwnerLabel] = owner
			labels[constants.ApplicationClusterDep] = workflowName
			obj.SetLabels(labels)

			_, err := client.Resource(gvr).Namespace(namespace).Update(ctx, &obj, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("Failed to update cronworkflow instanceID=%s namespace=%s name=%s err=%v", instanceID, namespace, obj.GetName(), err)
				return err
			}

			klog.Infof("Success to update cronworkflow instanceID=%s namespace=%s name=%s", instanceID, namespace, obj.GetName())
		}

		ns, err := kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get namespace=%s err=%v", namespace, err)
			return err
		}

		ns.Labels[constants.WorkflowNameLabel] = workflowName
		ns.Labels[constants.ApplicationClusterDep] = workflowName
		ns.Labels[constants.ApplicationGroupClusterDep] = "workflow"
		annotations := ns.Annotations
		if annotations == nil {
			annotations = make(map[string]string)
		}

		annotations[constants.WorkflowTitleAnnotation] = title
		ns.Annotations = annotations

		_, err = kubeClient.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Failed to update namespace label namespace=%s err=%v", namespace, err)
			return err
		}

		return nil
	})
}
