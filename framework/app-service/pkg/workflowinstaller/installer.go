package workflowinstaller

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/beclab/Olares/framework/app-service/pkg/argo"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	v1 "github.com/beclab/Olares/framework/app-service/pkg/workflowinstaller/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const instanceID = "knowledge-shared"

// Install a workflow with helm client.
func Install(ctx context.Context, kubeConfig *rest.Config, workflow *WorkflowConfig) error {
	helmClient, err := v1.NewHelmClient(ctx, kubeConfig, workflow.Namespace)
	if err != nil {
		return err
	}

	if installed, err := helmClient.IsInstalled(workflow.WorkflowName); err != nil {
		klog.Errorf("Failed to get install history workflowName=%s err=%v", workflow.WorkflowName, err)
		return err
	} else if installed {
		klog.Errorf("workflowName=%s already installed", workflow.WorkflowName)
		return errors.New(workflow.WorkflowName + " is already installed")
	}

	vals, err := getSettings(ctx, kubeConfig, workflow)
	if err != nil {
		return err
	}

	err = helmClient.Install(workflow.WorkflowName, workflow.ChartsName, workflow.RepoURL, workflow.Namespace, vals)
	if err != nil {
		klog.Errorf("Failed to install workflow chart name=%s err=%v", workflow.WorkflowName, err)
		return err
	}
	klog.Infof("install workflow old instanceID: %s", instanceID)

	return argo.UpdateWorkflowInNamespace(ctx, kubeConfig, workflow.WorkflowName,
		workflow.Namespace, instanceID, workflow.OwnerName, workflow.Cfg.Metadata.Title, workflow.Cfg.Options.SyncProvider)
}

// Uninstall remove a helm release for workflow.
func Uninstall(ctx context.Context, kubeConfig *rest.Config, workflow *WorkflowConfig) error {
	helmClient, err := v1.NewHelmClient(ctx, kubeConfig, workflow.Namespace)
	if err != nil {
		return err
	}

	if installed, err := helmClient.IsInstalled(workflow.WorkflowName); err != nil {
		klog.Errorf("Failed to get install history workflowName=%s err=%v", workflow.WorkflowName, err)
		return err
	} else if !installed {
		return errors.New("workflow not installed")
	}

	err = helmClient.Uninstall(workflow.WorkflowName)
	if err != nil {
		klog.Errorf("Failed to uninstall workflow chart name=%s err=%v", workflow.WorkflowName, err)
		return err
	}

	return nil
}

// Upgrade upgrade a workflow
func Upgrade(ctx context.Context, kubeConfig *rest.Config, workflow *WorkflowConfig) error {
	helmClient, err := v1.NewHelmClient(ctx, kubeConfig, workflow.Namespace)
	if err != nil {
		return err
	}

	if installed, err := helmClient.IsInstalled(workflow.WorkflowName); err != nil {
		klog.Errorf("Failed to get install history workflowName=%s err=%v", workflow.WorkflowName, err)
		return err
	} else if !installed {
		return errors.New("workflow not installed")
	}

	vals, err := getSettings(ctx, kubeConfig, workflow)
	if err != nil {
		return err
	}

	err = helmClient.Upgrade(workflow.WorkflowName, workflow.ChartsName, workflow.RepoURL, workflow.Namespace, vals)
	if err != nil {
		klog.Errorf("Failed to upgrade workflow chart name=%s err=%v", workflow.WorkflowName, err)
		return err
	}

	klog.Infof("install workflow old instanceID: %s", instanceID)
	return argo.UpdateWorkflowInNamespace(ctx, kubeConfig, workflow.WorkflowName,
		workflow.Namespace, instanceID, workflow.OwnerName, workflow.Cfg.Metadata.Title, workflow.Cfg.Options.SyncProvider)
}

func getSettings(ctx context.Context, kubeConfig *rest.Config, workflow *WorkflowConfig) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	values["bfl"] = map[string]interface{}{
		"username": workflow.OwnerName,
	}

	values["apiUrl"] = fmt.Sprintf("http://knowledge-base-api.user-system-%s:3010", workflow.OwnerName)
	values["title"] = workflow.Cfg.Metadata.Title

	appData, appCache, userdata, err := getAppData(ctx, kubeConfig, workflow.OwnerName)
	if err != nil {
		klog.Errorf("Failed to get user appdata err=%v", err)
		return nil, err
	}
	userspce := make(map[string]interface{})
	if workflow.Cfg.Permission.AppData {
		userspce["appData"] = appData
	}
	if workflow.Cfg.Permission.AppCache {
		userspce["appCache"] = appCache
	}
	if len(workflow.Cfg.Permission.UserData) > 0 {
		userspce["userData"] = userdata
	}
	values["userspace"] = userspce
	return values, nil
}

func getAppData(ctx context.Context, kubeConfig *rest.Config, owner string) (applicationdata, appdata, userdata string, err error) {
	k8s, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return "", "", "", err
	}

	bfl, err := k8s.AppsV1().StatefulSets("user-space-"+owner).Get(ctx, "bfl",
		metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get bfl err=%v", err)
		return "", "", "", err
	}

	appdata = bfl.Annotations[constants.UserAppDataDirKey]
	if appdata == "" {
		return "", "", "", errors.New("appdata not found")
	}

	userspace, ok := bfl.Annotations[constants.UserSpaceDirKey]
	if !ok {
		return "", "", "", errors.New("userspace not found")
	}

	applicationdata = filepath.Join(userspace, "Data")
	userdata = filepath.Join(userspace, "Home")

	return
}
