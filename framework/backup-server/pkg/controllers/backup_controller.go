package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	sysv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	k8sclient "olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/handlers"
	"olares.com/backup-server/pkg/integration"
	"olares.com/backup-server/pkg/notify"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/worker"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// BackupReconciler reconciles a Backup object
type BackupReconciler struct {
	client.Client
	factory             k8sclient.Factory
	scheme              *runtime.Scheme
	handler             handlers.Interface
	controllerStartTime metav1.Time
}

func NewBackupController(c client.Client, factory k8sclient.Factory, schema *runtime.Scheme, handler handlers.Interface) *BackupReconciler {
	return &BackupReconciler{
		Client:              c,
		factory:             factory,
		scheme:              schema,
		handler:             handler,
		controllerStartTime: metav1.Now(),
	}
}

//+kubebuilder:rbac:groups=sys.bytetrade.i,resources=backup,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=backup/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=backup/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Backup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Infof("received backup request, id: %s", req.Name)

	c, err := r.factory.Sysv1Client()
	if err != nil {
		log.Errorf("get sysv1 client error: %v, id: %s", err, req.Name)
		return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Second}, errors.WithStack(err)
	}

	backup, err := c.SysV1().Backups(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		log.Errorf("backup not found, it may have been deleted, id: %s", req.Name)
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Errorf("get backup error: %v, id: %s", err, req.Name)
		return ctrl.Result{}, errors.WithStack(err)
	}

	log.Infof("received backup request, id: %s, name: %s, owner: %s, deleted: %v, enabled: %v", req.Name, backup.Spec.Name, backup.Spec.Owner, backup.Spec.Deleted, backup.Spec.BackupPolicy.Enabled)

	if r.isDeleted(backup) {
		log.Infof("received backup delete request, id: %s, event: deleted", req.Name)
		worker.GetWorkerPool().CancelBackup(backup.Spec.Owner, backup.Name)
		r.handler.GetSnapshotHandler().RemoveFromSchedule(ctx, backup)

		if err := r.deleteBackup(backup); err != nil {
			log.Errorf("delete backup error: %v, id: %s, name: %s, retry...", err, backup.Name, backup.Spec.Name)
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, errors.WithStack(err)
		}
		return ctrl.Result{}, nil
	}

	if !r.isNotified(backup) {
		err = r.notify(backup)
		if err != nil {
			log.Errorf("notify backup error: %v, id: %s, name: %s", err, backup.Name, backup.Spec.Name)
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		}
		log.Infof("notify backup success, id: %s, name: %s", backup.Name, backup.Spec.Name)
	}

	if !r.isEnabled(backup) {
		r.handler.GetSnapshotHandler().RemoveFromSchedule(ctx, backup)
		return ctrl.Result{}, nil
	}

	err = r.reconcileBackupPolicies(backup)
	if err != nil {
		log.Errorf("reconcile backup policies error: %v, id: %s, name: %s", err, backup.Name, backup.Spec.Name)
		return ctrl.Result{}, errors.WithStack(err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	_, err := ctrl.NewControllerManagedBy(mgr).
		For(&sysv1.Backup{}, builder.WithPredicates(predicate.Funcs{
			GenericFunc: func(e event.GenericEvent) bool { return false },
			CreateFunc: func(e event.CreateEvent) bool {
				log.Info("hit backup create event")
				return true
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				log.Info("hit backup update event")

				bc1, ok1 := e.ObjectOld.(*sysv1.Backup)
				bc2, ok2 := e.ObjectNew.(*sysv1.Backup)
				if !(ok1 || ok2) || reflect.DeepEqual(bc1.Spec, bc2.Spec) {
					log.Info("backup not changed")
					return false
				}

				if bc2.Spec.Deleted {
					return true
				}

				flag, err := r.isSizeUpdated(bc1, bc2)
				if err != nil {
					log.Errorf("backup size updated error: %v", err)
					return false
				}
				if flag {
					return true
				}
				return !r.isPolicyUpdated(bc1, bc2)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				log.Info("hit backup delete event")
				return false
			}})).
		Build(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *BackupReconciler) isNotified(backup *sysv1.Backup) bool {
	return backup.Spec.Notified
}

func (r *BackupReconciler) isDeleted(newBackup *sysv1.Backup) bool {
	return newBackup.Spec.Deleted
}

func (r *BackupReconciler) isEnabled(newBackup *sysv1.Backup) bool {
	return newBackup.Spec.BackupPolicy.Enabled
}

func (r *BackupReconciler) isSizeUpdated(oldBackupSpec, newBackupSpec *sysv1.Backup) (bool, error) {
	oldExtra := oldBackupSpec.Spec.Extra
	newExtra := newBackupSpec.Spec.Extra

	oldSizeUpdated, ok1 := oldExtra["size_updated"]
	newSizeUpdated, ok2 := newExtra["size_updated"]

	if !ok1 || !ok2 {
		return false, fmt.Errorf("backup size_updated invalid, old: %v, new: %v", ok1, ok2)
	}

	if oldSizeUpdated != newSizeUpdated {
		return true, nil
	}

	return false, nil
}

func (r *BackupReconciler) isPolicyUpdated(oldBackupSpec, newBackupSpec *sysv1.Backup) bool {
	return reflect.DeepEqual(oldBackupSpec.Spec.BackupPolicy, newBackupSpec.Spec.BackupPolicy)
}

func (r *BackupReconciler) reconcileBackupPolicies(backup *sysv1.Backup) error {
	ctx := context.Background()
	if backup.Spec.BackupPolicy != nil {
		cron, _ := util.ParseToCron(backup.Spec.BackupPolicy.SnapshotFrequency, backup.Spec.BackupPolicy.TimesOfDay, backup.Spec.BackupPolicy.DayOfWeek, backup.Spec.BackupPolicy.DateOfMonth)
		err := r.handler.GetSnapshotHandler().CreateSchedule(ctx, backup, cron, !backup.Spec.BackupPolicy.Enabled)
		if err != nil {
			return err
		}
		log.Debugf("schedule %s created: %q", backup.Spec.Name, cron)
	}
	return nil
}

func (r *BackupReconciler) notify(backup *sysv1.Backup) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	backupType := handlers.GetBackupType(backup)

	locationConfig, err := handlers.GetBackupLocationConfig(backup)
	if err != nil {
		return fmt.Errorf("get backup location config error: %v", err)
	}
	if locationConfig == nil {
		return fmt.Errorf("backup location config not exists")
	}
	var location = locationConfig["location"]
	var locationConfigName = locationConfig["name"]

	if location != constant.BackupLocationSpace.String() {
		return r.handler.GetBackupHandler().UpdateNotifyState(ctx, backup.Name, true)
	}

	olaresSpaceToken, err := integration.IntegrationManager().GetIntegrationSpaceToken(ctx, backup.Spec.Owner, locationConfigName)
	if err != nil {
		return err
	}

	var notifyBackupObj = &notify.Backup{
		UserId:         olaresSpaceToken.OlaresDid,
		Token:          olaresSpaceToken.AccessToken,
		BackupId:       backup.Name,
		Name:           backup.Spec.Name,
		BackupType:     backupType,
		BackupTime:     backup.Spec.CreateAt.Unix(),
		BackupPath:     handlers.GetBackupPath(backup),
		BackupLocation: location,
	}

	if err := notify.NotifyBackup(ctx, constant.SyncServerURL, notifyBackupObj); err != nil {
		return fmt.Errorf("[push] notify backup obj error: %v", err)
	}

	return r.handler.GetBackupHandler().UpdateNotifyState(ctx, backup.Name, true)
}

func (r *BackupReconciler) deleteBackup(backup *sysv1.Backup) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := r.handler.GetBackupHandler().DeleteBackup(ctx, backup); err != nil {
		return err
	}

	return nil
}
