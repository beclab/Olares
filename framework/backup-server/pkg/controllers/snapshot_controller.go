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
	sysapiv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	v1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
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

// BackupReconciler reconciles a Snapshot object
type SnapshotReconciler struct {
	client.Client
	factory             k8sclient.Factory
	scheme              *runtime.Scheme
	handler             handlers.Interface
	controllerStartTime metav1.Time
}

func NewSnapshotController(c client.Client, factory k8sclient.Factory, schema *runtime.Scheme, handler handlers.Interface) *SnapshotReconciler {
	return &SnapshotReconciler{Client: c,
		factory:             factory,
		scheme:              schema,
		handler:             handler,
		controllerStartTime: metav1.Now(),
	}
}

//+kubebuilder:rbac:groups=sys.bytetrade.i,resources=snapshot,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=snapshot/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=snapshot/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Snapshot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *SnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Infof("received snapshot request, id: %s", req.Name)

	c, err := r.factory.Sysv1Client()
	if err != nil {
		log.Errorf("get sysv1 client error: %v, name: %s", err, req.Name)
		return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Second}, errors.WithStack(err)
	}

	snapshot, err := c.SysV1().Snapshots(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		log.Infof("snapshot not found, it may have been deleted, id: %s", req.Name)
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Errorf("get snapshot error: %v, id: %s", err, req.Name)
		return ctrl.Result{}, errors.WithStack(err)
	}

	log.Infof("received snapshot request, id: %s, phase: %s, extra: %s", req.Name, *snapshot.Spec.Phase, util.ToJSON(snapshot.Spec.Extra))

	// TODO check backup
	backup, err := r.getBackup(snapshot.Spec.BackupId)
	if err != nil && apierrors.IsNotFound(err) {
		log.Errorf("snapshot not found, it may have been deleted, id: %s", req.Name)
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Errorf("get snapshot error: %v, id: %s", err, req.Name)
		return ctrl.Result{}, errors.WithStack(err)
	}

	var phase = *snapshot.Spec.Phase

	switch phase {
	case constant.Pending.String():
		isNewlyCreated := snapshot.CreationTimestamp.After(r.controllerStartTime.Time)
		if isNewlyCreated {
			err := r.addToWorkerManager(backup, snapshot) // todo enhance
			if err != nil {
				log.Errorf("add snapshot to worker error: %v, id: %s", err, snapshot.Name)
				r.handler.GetSnapshotHandler().UpdatePhase(context.Background(), snapshot.Name, constant.Failed.String(), err.Error())
			}
		} else {
			r.handler.GetSnapshotHandler().UpdatePhase(context.Background(), snapshot.Name, constant.Failed.String(), constant.MessageBackupServerRestart)
		}
	case constant.Running.String():
		isNewlyCreated := snapshot.CreationTimestamp.After(r.controllerStartTime.Time)
		if !isNewlyCreated {
			r.handler.GetSnapshotHandler().UpdatePhase(context.Background(), snapshot.Name, constant.Failed.String(), constant.MessageBackupServerRestart)
		}
	case constant.Completed.String(), constant.Failed.String(), constant.Canceled.String(), constant.Rejected.String():
		if phase == constant.Canceled.String() {
			worker.GetWorkerPool().CancelSnapshot(backup.Spec.Owner, snapshot.Name)
		}
		if err := r.notifySnapshotResult(ctx, backup, snapshot); err != nil {
			log.Errorf("[notify] snapshot error: %v, id: %s, phase: %s", err, snapshot.Name, *snapshot.Spec.Phase)
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, fmt.Errorf("notify snapshot error: %v", err)
		} else {
			log.Infof("[notify] snapshot success, id: %s, phase: %s", snapshot.Name, *snapshot.Spec.Phase)
		}
		if err := r.handler.GetSnapshotHandler().UpdateNotifyResultState(context.Background(), snapshot); err != nil {
			log.Errorf("update snapshot notify state error: %v, id: %s", err, snapshot.Name)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	_, err := ctrl.NewControllerManagedBy(mgr).
		For(&sysapiv1.Snapshot{}, builder.WithPredicates(predicate.Funcs{
			GenericFunc: func(genericEvent event.GenericEvent) bool { return false },
			CreateFunc: func(e event.CreateEvent) bool {
				snapshot, ok := r.isSysSnapshot(e.Object)
				if !ok {
					log.Debugf("not a snapshot resource")
					return false
				}

				var reconcile bool
				var phase = *snapshot.Spec.Phase

				switch phase {
				case constant.Completed.String(), constant.Failed.String(), constant.Canceled.String(), constant.Rejected.String():
					snapshotNotified, err := handlers.CheckSnapshotNotifyState(snapshot, "result")
					if err != nil {
						log.Errorf("hit snapshot create event, check snapshot push state error: %v, id: %s", err, snapshot.Name)
						reconcile = false
					} else {
						reconcile = !snapshotNotified
					}
				default:
					reconcile = true
				}

				if reconcile {
					log.Infof("hit snapshot create event, id: %s, phase: %s, extra: %s", snapshot.Name, *snapshot.Spec.Phase, util.ToJSON(snapshot.Spec.Extra))
				}
				return reconcile
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				log.Info("hit snapshot update event")

				oldObj, newObj := updateEvent.ObjectOld, updateEvent.ObjectNew
				oldSnapshot, ok1 := r.isSysSnapshot(oldObj)
				newSnapshot, ok2 := r.isSysSnapshot(newObj)

				if !(ok1 && ok2) || reflect.DeepEqual(oldSnapshot.Spec, newSnapshot.Spec) {
					return false
				}

				if r.isRunningProgress(oldSnapshot, newSnapshot) {
					return false
				}

				snapshotNotified, err := handlers.CheckSnapshotNotifyState(newSnapshot, "result")
				if err != nil {
					log.Errorf("hit snapshot update event, check snapshot push state error: %v, id: %s", err, newSnapshot.Name)
					return false
				}

				if snapshotNotified {
					return false
				}

				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				log.Info("hit snapshot delete event")
				return false
			},
		})).Build(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *SnapshotReconciler) isSysSnapshot(obj client.Object) (*sysapiv1.Snapshot, bool) {
	b, ok := obj.(*sysapiv1.Snapshot)
	if !ok || b == nil {
		return nil, false
	}

	return b, true
}

func (r *SnapshotReconciler) getBackup(backupId string) (*v1.Backup, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	backup, err := r.handler.GetBackupHandler().GetById(ctx, backupId)
	if err != nil {
		return nil, err
	}

	return backup, nil
}

func (r *SnapshotReconciler) isRunningProgress(oldSnapshot *v1.Snapshot, newSnapshot *v1.Snapshot) bool {
	if *newSnapshot.Spec.Phase == *oldSnapshot.Spec.Phase && *newSnapshot.Spec.Phase == constant.Running.String() {
		return true
	}

	return false
}

func (r *SnapshotReconciler) isRunning(oldSnapshot *v1.Snapshot, newSnapshot *v1.Snapshot) bool {
	if *oldSnapshot.Spec.Phase == constant.Pending.String() && *newSnapshot.Spec.Phase == constant.Running.String() {
		return true
	}
	return false
}

func (r *SnapshotReconciler) isCanceled(newSnapshot *v1.Snapshot) bool {
	newPhase := *newSnapshot.Spec.Phase
	return newPhase == constant.Canceled.String()
}

func (r *SnapshotReconciler) addToWorkerManager(backup *sysapiv1.Backup, snapshot *sysapiv1.Snapshot) error {
	if err := r.notifySnapshot(backup, snapshot, constant.Pending.String()); err != nil {
		log.Errorf("[notify] snapshot error: %v, id: %s, phase: Pending", err, snapshot.Name)
		// return err
	} else {
		log.Infof("[notify] snapshot success, id: %s, phase: Pending", snapshot.Name)
	}

	if err := worker.GetWorkerPool().AddBackupTask(backup.Spec.Owner, backup.Name, snapshot.Name); err != nil {
		return err
	}

	return nil
}

func (r *SnapshotReconciler) checkSnapshotPushState(snapshot *sysapiv1.Snapshot) (bool, error) {
	return handlers.CheckSnapshotNotifyState(snapshot, "result")
}

func (r *SnapshotReconciler) notifySnapshot(backup *v1.Backup, snapshot *v1.Snapshot, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	locationConfig, _ := handlers.GetBackupLocationConfig(backup)
	location := locationConfig["location"]

	if location != constant.BackupLocationSpace.String() {
		return nil
	}

	locationConfigName := locationConfig["name"]
	olaresSpaceToken, err := integration.IntegrationManager().GetIntegrationSpaceToken(ctx, backup.Spec.Owner, locationConfigName)
	if err != nil {
		return err
	}

	var snapshotRecord = &notify.Snapshot{
		UserId:       olaresSpaceToken.OlaresDid,
		BackupId:     backup.Name,
		SnapshotId:   snapshot.Name,
		Size:         0,
		BackupSize:   0,
		Unit:         constant.DefaultSnapshotSizeUnit,
		SnapshotTime: snapshot.Spec.CreateAt.Unix(),
		Status:       status,
		Type:         handlers.ParseSnapshotTypeText(snapshot.Spec.SnapshotType),
	}

	if err := notify.NotifySnapshot(ctx, constant.SyncServerURL, snapshotRecord); err != nil {
		return err
	}

	return nil
}

func (r *SnapshotReconciler) notifySnapshotResult(ctx context.Context, backup *v1.Backup, snapshot *v1.Snapshot) error {
	locationConfig, _ := handlers.GetBackupLocationConfig(backup)
	location := locationConfig["location"]

	if location != constant.BackupLocationSpace.String() {
		return nil
	}

	locationConfigName := locationConfig["name"]
	spaceToken, err := integration.IntegrationManager().GetIntegrationSpaceToken(ctx, backup.Spec.Owner, locationConfigName)
	if err != nil {
		return fmt.Errorf("get space token error: %v", err)
	}

	var storageInfo, resticInfo = r.handler.GetSnapshotHandler().ParseSnapshotInfo(snapshot)

	var snapshotRecord = &notify.Snapshot{
		UserId:       spaceToken.OlaresDid,
		BackupId:     backup.Name,
		SnapshotId:   snapshot.Name,
		Unit:         constant.DefaultSnapshotSizeUnit,
		SnapshotTime: snapshot.Spec.CreateAt.Unix(),
		Type:         handlers.ParseSnapshotTypeText(snapshot.Spec.SnapshotType),
		Status:       *snapshot.Spec.Phase,
	}

	if storageInfo != nil {
		snapshotRecord.Url = storageInfo.Url
		snapshotRecord.CloudName = storageInfo.CloudName
		snapshotRecord.RegionId = storageInfo.RegionId
		snapshotRecord.Bucket = storageInfo.Bucket
		snapshotRecord.Prefix = storageInfo.Prefix
	}
	if resticInfo != nil {
		snapshotRecord.Message = util.ToJSON(resticInfo)
		snapshotRecord.Size = resticInfo.TotalBytesProcessed
		snapshotRecord.BackupSize = *backup.Spec.Size
		snapshotRecord.ResticSnapshotId = resticInfo.SnapshotID
	} else if snapshot.Spec.Message != nil {
		snapshotRecord.Message = *snapshot.Spec.Message
	}

	return notify.NotifySnapshot(ctx, constant.SyncServerURL, snapshotRecord)
}
