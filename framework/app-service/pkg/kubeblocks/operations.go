package kubeblocks

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"

	"strings"

	kbappsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	kbopv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OperationOptions struct {
	cxt       context.Context
	Namespace string
	Name      string
	OpsType   kbopv1alpha1.OpsType
	Client    client.Client
}

func NewOperation(ctx context.Context, opsType kbopv1alpha1.OpsType, manager *appv1alpha1.ApplicationManager, ctrlClient client.Client) *OperationOptions {
	return &OperationOptions{
		cxt:       ctx,
		Name:      manager.Spec.AppName,
		Namespace: manager.Spec.AppNamespace,
		OpsType:   opsType,
		Client:    ctrlClient,
	}
}

func (op *OperationOptions) validate() error {
	var cluster kbappsv1.Cluster
	err := op.Client.Get(op.cxt, types.NamespacedName{Name: op.Name, Namespace: op.Namespace}, &cluster)
	if err != nil {
		return err
	}
	return nil
}

func (op *OperationOptions) Stop() error {
	err := op.do()
	if err != nil {
		klog.Errorf("failed to stop middleware %v", err)
		return err
	}
	return nil

}

func (op *OperationOptions) Start() error {
	err := op.do()
	if err != nil {
		klog.Errorf("failed to start middleware %v", err)
		return err
	}
	return nil
}

func (op *OperationOptions) do() error {
	err := op.validate()
	if err != nil {
		klog.Errorf("failed to validate middleware cluster %v", err)
		return err
	}
	name := fmt.Sprintf("%s-%s-%s", op.Namespace, strings.ToLower(string(op.OpsType)), uuid.NewString())
	opsRequest := &kbopv1alpha1.OpsRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: op.Namespace,
		},
		Spec: kbopv1alpha1.OpsRequestSpec{
			ClusterName: op.Name,
			Type:        op.OpsType,
		},
	}
	err = op.Client.Create(op.cxt, opsRequest)
	if err != nil {
		klog.Errorf("failed to create ops request %v", err)
		return err
	}
	return nil
}
