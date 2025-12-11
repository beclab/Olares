package controllers

import (
	"context"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	sysapiv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	v1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	bclient "olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/handlers"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/worker"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type RestoreReconciler struct {
	client.Client
	factory             bclient.Factory
	scheme              *runtime.Scheme
	handler             handlers.Interface
	controllerStartTime metav1.Time
}

func NewRestoreController(c client.Client, factory bclient.Factory, schema *runtime.Scheme, handler handlers.Interface) *RestoreReconciler {
	return &RestoreReconciler{Client: c,
		factory:             factory,
		scheme:              schema,
		handler:             handler,
		controllerStartTime: metav1.Now(),
	}
}

//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=restore,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=restore/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=restore/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Restore object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *RestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	_, err := ctrl.NewControllerManagedBy(mgr).
		For(&sysapiv1.Restore{}, builder.WithPredicates(predicate.Funcs{
			GenericFunc: func(genericEvent event.GenericEvent) bool { return false },
			CreateFunc: func(e event.CreateEvent) bool {
				log.Info("hit restore create event")

				restore, ok := r.isSysRestore(e.Object)
				if !ok {
					log.Debugf("not a restore resource")
					return false
				}

				backupType, err := handlers.GetRestoreType(restore)
				if err != nil {
					log.Error(err)
					return false
				}

				restoreType, err := handlers.ParseRestoreType(backupType, restore)
				if err != nil {
					log.Errorf("restore %s type %v invalid, error: %v", restore.Name, restore.Spec.RestoreType, err)
					r.handler.GetRestoreHandler().SetRestorePhase(restore.Name, constant.Failed)
					return false
				}

				log.Infof("hit restore create event %s, phase: %s, restoreType: %s", restore.Name, *restore.Spec.Phase, restoreType.Type)

				var phase = *restore.Spec.Phase
				switch phase {
				case constant.Pending.String():
					isNewlyCreated := restore.CreationTimestamp.After(r.controllerStartTime.Time)
					if isNewlyCreated {
						err := worker.GetWorkerPool().AddRestoreTask(restore.Spec.Owner, restore.Name)
						if err != nil {
							log.Errorf("add restore to worker error: %v, id: %s", err, restore.Name)
							if err = r.handler.GetRestoreHandler().SetRestorePhase(restore.Name, constant.Failed); err != nil {
								log.Errorf("update restore %s phase %s to Rejected error: %v", restore.Name, phase, err)
							} else {
								log.Infof("restore %s, type: %s, add to queue success", restore.Name, restoreType.Type)
							}
						}

						return false
					}
					fallthrough
				case constant.Running.String():
					if err := r.handler.GetRestoreHandler().SetRestorePhase(restore.Name, constant.Failed); err != nil {
						log.Errorf("update restore %s phase %s to Failed error: %v", restore.Name, phase, err)
					}
				}
				return false
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				log.Info("hit restore update event")

				oldObj, newObj := updateEvent.ObjectOld, updateEvent.ObjectNew
				oldRestore, ok1 := r.isSysRestore(oldObj)
				newRestore, ok2 := r.isSysRestore(newObj)

				if !(ok1 && ok2) || reflect.DeepEqual(oldRestore.Spec, newRestore.Spec) {
					return false
				}

				log.Infof("hit restore update event, name: %s, phase: %s", newRestore.Name, *newRestore.Spec.Phase)

				if r.isRunningProgress(oldRestore, newRestore) {
					return false
				}

				if *newRestore.Spec.Phase == constant.Failed.String() {
					return false
				}

				if *newRestore.Spec.Phase == constant.Canceled.String() {
					worker.GetWorkerPool().CancelRestore(newRestore.Spec.Owner, newRestore.Name)
				}

				if *newRestore.Spec.Phase == constant.Deleted.String() {
					r.handler.GetRestoreHandler().DeleteRestore(newRestore.Name)
				}

				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				log.Info("hit restore delete event")
				return false
			},
		})).Build(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *RestoreReconciler) isRunningProgress(oldRestore *v1.Restore, newRestore *v1.Restore) bool {
	if *oldRestore.Spec.Phase == *newRestore.Spec.Phase && *newRestore.Spec.Phase == constant.Running.String() {
		return true
	}

	return false
}

func (r *RestoreReconciler) isSysRestore(obj client.Object) (*sysapiv1.Restore, bool) {
	b, ok := obj.(*sysapiv1.Restore)
	if !ok || b == nil {
		return nil, false
	}

	return b, true
}
