package app

import (
	"context"
	"fmt"
	"reflect"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// CreateSysAppMgr registers a platform-managed workload in its actual namespace.
// No User resource is needed, so shared frontends can register before onboarding.
func CreateSysAppMgr(ctx context.Context, client versioned.Interface, app *v1alpha1.Application) error {
	if app.Spec.Namespace == "" || app.Spec.Name == "" || app.Spec.Owner == "" {
		return fmt.Errorf("system application %s has incomplete identity", app.Name)
	}
	name := fmt.Sprintf("%s-%s", app.Spec.Namespace, app.Spec.Name)
	managers := client.AppV1alpha1().ApplicationManagers()
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		existing, err := managers.Get(ctx, name, metav1.GetOptions{})
		create := apierrors.IsNotFound(err)
		if err != nil && !create {
			return err
		}
		if create {
			existing = &v1alpha1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: name}}
		}
		desired := existing.DeepCopy()
		desired.Spec.AppName = app.Spec.Name
		desired.Spec.RawAppName = app.Spec.RawAppName
		if desired.Spec.RawAppName == "" {
			desired.Spec.RawAppName = app.Spec.Name
		}
		desired.Spec.AppNamespace = app.Spec.Namespace
		desired.Spec.AppOwner = app.Spec.Owner
		desired.Spec.Source = "system"
		desired.Spec.Type = v1alpha1.App
		for _, key := range []string{constants.AppApiVersionLabel, constants.AppSharedLabel} {
			if value, ok := app.Labels[key]; ok {
				if desired.Labels == nil {
					desired.Labels = make(map[string]string)
				}
				desired.Labels[key] = value
			} else {
				delete(desired.Labels, key)
			}
		}
		// Existing lifecycle state belongs to the manager controller. Only initialize
		// a new/empty registration; repeated reconciles must not reset operations.
		if desired.Status.State == "" {
			now := metav1.Now()
			desired.Spec.OpType = v1alpha1.InstallOp
			desired.Status = v1alpha1.ApplicationManagerStatus{
				OpType: v1alpha1.InstallOp, State: v1alpha1.Running,
				Message: "sys app install completed", UpdateTime: &now, StatusTime: &now,
			}
		}
		if create {
			_, err = managers.Create(ctx, desired, metav1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				// Another reconcile won creation; retry the read/update path.
				return apierrors.NewConflict(v1alpha1.Resource("applicationmanagers"), name, err)
			}
			return err
		}
		if reflect.DeepEqual(existing, desired) {
			return nil
		}
		_, err = managers.Update(ctx, desired, metav1.UpdateOptions{})
		return err
	})
}
