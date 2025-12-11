package backup

import (
	"context"
	"sync"
	"time"

	backupPkg "bytetrade.io/web3os/tapr/pkg/backup"
	"bytetrade.io/web3os/tapr/pkg/constants"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"
	"bytetrade.io/web3os/tapr/pkg/workload/mongodb"
	psmdbv1 "github.com/percona/percona-server-mongodb-operator/pkg/apis/psmdb/v1"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	veleroclientset "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

type Watcher struct {
	veleroClient  *veleroclientset.Clientset
	ctx           context.Context
	dynamicClient *dynamic.DynamicClient
	k8sClientSet  *kubernetes.Clientset
	aprClientSet  *aprclientset.Clientset
}

func NewWatcher(kubeconfig *rest.Config, ctx context.Context) *Watcher {
	veleroClient := veleroclientset.NewForConfigOrDie(kubeconfig)

	return &Watcher{
		veleroClient:  veleroClient,
		ctx:           ctx,
		dynamicClient: dynamic.NewForConfigOrDie(kubeconfig),
		k8sClientSet:  kubernetes.NewForConfigOrDie(kubeconfig),
		aprClientSet:  aprclientset.NewForConfigOrDie(kubeconfig),
	}
}

func (w *Watcher) Start() {
	go func() {
		var (
			backupWatcherOK        = false
			restoreWatcherOK       = false
			backupWatcherRecreate  = true
			restoreWatcherRecreate = true
			backupWatcher          watch.Interface
			restoreWatcher         watch.Interface
			event                  watch.Event
			err                    error
		)
		for {

			if backupWatcherRecreate {
				klog.Info("start to watch velero backup event")
				backupWatcher, err = w.dynamicClient.Resource(backupPkg.BackupGVR).Namespace(constants.FrameworkNamespace).Watch(w.ctx, metav1.ListOptions{})
				if err != nil {
					klog.Error("watch backup crd error, ", err)
					time.Sleep(time.Second)
					continue
				} else {
					backupWatcherRecreate = false
				}

			}

			if restoreWatcherRecreate {
				klog.Info("start to watch velero restore event")
				restoreWatcher, err = w.veleroClient.VeleroV1().Restores(constants.FrameworkNamespace).Watch(w.ctx, metav1.ListOptions{})
				if err != nil {
					klog.Error("watch restore crd error, ", err)
					time.Sleep(time.Second)
					continue
				} else {
					restoreWatcherRecreate = false
				}

			}

			select {
			case <-w.ctx.Done():
				backupWatcher.Stop()
				restoreWatcher.Stop()
				klog.Info("backup watcher stopped")

			case event, backupWatcherOK = <-backupWatcher.ResultChan():
				if !backupWatcherOK {
					klog.Error("backup watcher broken")
					backupWatcher.Stop()
					backupWatcherRecreate = true
				} else {
					if event.Type == watch.Added || event.Type == watch.Modified {
						data, ok := event.Object.(*unstructured.Unstructured)
						if !ok {
							klog.Error("invalid event with unexpected object ( not backup ) received")
							continue
						}
						var backup backupPkg.Backup
						err := runtime.DefaultUnstructuredConverter.FromUnstructured(data.Object, &backup)
						if err != nil {
							klog.Error("invalid unstructured data ( not backup ) received, ", err)
							continue
						}

						if backup.Spec.Phase == nil || *backup.Spec.Phase != backupPkg.BackupStart {
							klog.Info("backup is not started, ", *backup.Spec.Phase)
							continue
						}

						if state := backup.Spec.MiddleWarePhase; state != nil {
							switch *state {
							case "Failed", "Success":
								klog.Info("ignore finished middleware backup")
								continue
							case "Running":
								klog.Info("ignore middleware backup, a running backup exists")
								continue
							}
						}

						klog.Info("start to backup middleware")

						klog.Info("update backup middleware status to Running")
						if err := w.updateBackupState(&backup, "Running", nil); err != nil {
							klog.Error("update backup status error, ", err, ", ", backup.Name, ", ", backup.Namespace)
						}

						var state string
						var errMsg *string
						if err := w.backup(); err != nil {
							klog.Error("backup error, ", err)
							state = "Failed"
							e := err.Error()
							errMsg = &e
						} else {
							klog.Info("middleware backup success")
							state = "Success"
						}

						if err := w.updateBackupState(&backup, state, errMsg); err != nil {
							klog.Error("update backup status error, ", err, ", ", backup.Name, ", ", backup.Namespace)
						}
					} else {
						klog.Info("ignore backup event, ", event.Type)
					} // end of event type

				}

			case event, restoreWatcherOK = <-restoreWatcher.ResultChan():
				if !restoreWatcherOK {
					klog.Error("restore watcher broken")
					restoreWatcher.Stop()
					restoreWatcherRecreate = true
				} else {
					if event.Type == watch.Added {
						klog.Info("start to restore middleware")

						_, ok := event.Object.(*velerov1.Restore)
						if !ok {
							klog.Error("invalid event with unexpected object ( not restore ) received")
							continue
						}

						if err := w.restore(); err != nil {
							klog.Error("restore error, ", err)
						} else {
							klog.Info("middleware restore success                                                                                                       ")
						}
					} else {
						klog.Info("ignore restore event, ", event.Type)
					} // end of event type
				}
			}
		}
	}()
}

func (w *Watcher) updateBackupState(backup *backupPkg.Backup, state string, errMsg *string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		b, err := w.dynamicClient.Resource(backupPkg.BackupGVR).Namespace(backup.Namespace).Get(w.ctx, backup.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		var updateBackup backupPkg.Backup
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(b.Object, &updateBackup)
		if err != nil {
			klog.Error("convert sysbackups unstructured error, ", err)
			return err
		}

		mb, err := w.dynamicClient.Resource(mongodb.PSMDBBackupClassGVR).Namespace(backup.Namespace).Get(w.ctx, mongodb.PerconaMongoClusterBackup, metav1.GetOptions{})
		if err != nil {
			return err
		}

		var mongoBackup psmdbv1.PerconaServerMongoDBBackup
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(mb.Object, &mongoBackup)
		if err != nil {
			klog.Error("convert mongo backup unstructured error, ", err)
			return err
		}

		// FIXME: user patch instead
		updateBackup.Spec.MiddleWarePhase = &state
		updateBackup.Spec.MiddleWareFailedMessage = errMsg
		if updateBackup.Annotations == nil {
			updateBackup.Annotations = make(map[string]string)
		}
		updateBackup.Annotations[mongodb.PerconaMongoClusterLastBackupPBMName] = mongoBackup.Status.PBMname
		updateData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&updateBackup)
		if err != nil {
			klog.Error("convert to unstructured error, ", err, ", ", updateBackup.Name, ", ", updateBackup.Namespace)
			return err
		}

		_, err = w.dynamicClient.Resource(backupPkg.BackupGVR).Namespace(backup.Namespace).Update(w.ctx,
			&unstructured.Unstructured{Object: updateData}, metav1.UpdateOptions{})
		return err
	})
}

func (w *Watcher) backup() error {
	wg := sync.WaitGroup{}
	var errs []error
	runBackup := func(f func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := f()
			if err != nil {
				errs = append(errs, err)
			}
		}()
	}

	runBackup(w.backupRedix)
	//runBackup(w.backupMongo)
	runBackup(w.backupPostgres)

	wg.Wait()

	if len(errs) > 0 {
		return errors.NewAggregate(errs)
	}

	return nil
}

/**
 * runRestore(w.restoreMongo)
 * The functionality of restoring the mongo database has been moved to the restore script;
 * this leads to unexplainable errors when manipulating CRD resources at this location.
 */
func (w *Watcher) restore() error {
	wg := sync.WaitGroup{}
	var errs []error
	runRestore := func(f func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := f()
			if err != nil {
				errs = append(errs, err)
			}
		}()
	}

	runRestore(w.restorePostgres)
	runRestore(w.restoreRedix)

	wg.Wait()

	if len(errs) > 0 {
		return errors.NewAggregate(errs)
	}

	return nil
}
