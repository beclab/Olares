package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"bytetrade.io/web3os/tapr/pkg/constants"

	kbappsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	psmdbv1 "github.com/percona/percona-server-mongodb-operator/pkg/apis/psmdb/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ScalePerconaMongoNodes(ctx context.Context, dynamicClient *dynamic.DynamicClient, name, namespace string, nodes int32) error {
	cluster, err := dynamicClient.Resource(PSMDBClassGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		klog.Error("get percona mongo cluster error, ", err, ", ", name, ", ", namespace)
		return err
	}

	psmdb := psmdbv1.PerconaServerMongoDB{}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(cluster.Object, &psmdb); err != nil {
		klog.Error("convert mongo cluster crd error, ", err, ", ", name, ", ", namespace)
		return err
	}

	if len(psmdb.Spec.Replsets) > 0 {
		psmdb.Spec.Replsets[0].Size = nodes

		updateCluster, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&psmdb)
		if err != nil {
			klog.Error("to unstructured mongo cluster crd error, ", err, ", ", name, ", ", namespace)
			return err
		}

		_, err = dynamicClient.Resource(PSMDBClassGVR).Namespace(namespace).
			Update(ctx, &unstructured.Unstructured{Object: updateCluster}, metav1.UpdateOptions{})

		return err
	}

	return errors.New("invalid percona mongo cluster definition")
}

func CheckMongoRestoreStatus(ctx context.Context, dynamicClient *dynamic.DynamicClient) (bool, error) {
	// When the restore and mongodbs status is ready, if for some reason the pod restarts and triggers a MongoDB restore,
	// it will cause the MongoDB service to crash.
	mongoClusterStruct, err := dynamicClient.Resource(PSMDBClassGVR).Namespace(PSMDB_NAMESPACE).Get(ctx, PSMDB_NAME, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	if mongoClusterStruct == nil {
		return false, fmt.Errorf("mongo cluster crd not found")
	}

	var mongoCluster = psmdbv1.PerconaServerMongoDB{}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(mongoClusterStruct.Object, &mongoCluster); err != nil {
		klog.Error("convert mongo cluster crd error, ", err, ", ", mongoClusterStruct.GetName(), ", ", mongoClusterStruct.GetNamespace())
		return false, err
	}

	mongoRestoreStruct, err := dynamicClient.Resource(PSMDBRestoreClassGVR).Namespace(PSMDB_NAMESPACE).Get(ctx, PSMDB_RESTORE_NAME, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	if mongoRestoreStruct == nil {
		return false, fmt.Errorf("mongo restore crd not found")
	}

	var mongoRestore = psmdbv1.PerconaServerMongoDBRestore{}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(mongoRestoreStruct.Object, &mongoRestore); err != nil {
		klog.Error("convert mongo restore crd error, ", err, ", ", mongoRestoreStruct.GetName(), ", ", mongoRestoreStruct.GetNamespace())
		return false, err
	}

	return true, nil

	// if mongoCluster.Status.State == psmdbv1.AppStateReady && mongoRestore.Status.State == psmdbv1.RestoreStateReady {
	// 	return true, nil
	// }
	// switch mongoRestore.Status.State {
	// case psmdbv1.RestoreStateWaiting, psmdbv1.RestoreStateRequested, psmdbv1.RestoreStateRunning, psmdbv1.RestoreStateReady:
	// 	klog.Info("exists mongo restore: %s", mongoRestore.Status.State)
	// 	return true, nil
	// }

	// return false, nil
}

func WaitForInitializeComplete(ctx context.Context, dynamicClient *dynamic.DynamicClient, k8sClient *kubernetes.Clientset) error {
	return wait.PollWithContext(ctx, 5*time.Second, time.Hour,
		func(context.Context) (done bool, err error) {
			incompleted := false
			mongoClusterStruct, err := dynamicClient.Resource(PSMDBClassGVR).Namespace(PSMDB_NAMESPACE).Get(ctx, PSMDB_NAME, metav1.GetOptions{})

			if err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}

				klog.Error("get mongo cluster error, ", err)
				return false, err
			}

			var mongoCluster = psmdbv1.PerconaServerMongoDB{}

			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(mongoClusterStruct.Object, &mongoCluster); err != nil {
				klog.Error("convert mongo cluster crd error, ", err, ", ", mongoClusterStruct.GetName(), ", ", mongoClusterStruct.GetNamespace())
				return false, nil
			}

			if mongoCluster.Status.Mongos == nil || mongoCluster.Status.Mongos.Status != psmdbv1.AppStateReady {
				klog.Info("mongo cluster mongos not ready, waiting")
				return false, nil
			}

			for rsName, rs := range mongoCluster.Status.Replsets {
				if rs.Ready < int32(1) {
					klog.Infof("get mongo cluster replsets %s not ready, waiting", rsName)
					return false, nil
				}
			}

			done = !incompleted

			return
		})
}

func ForceCreateNewMongoClusterBackup(ctx context.Context, dynamicClient *dynamic.DynamicClient,
	backup *psmdbv1.PerconaServerMongoDBBackup, backupPath string) error {
	currentBackupData, err := dynamicClient.Resource(PSMDBBackupClassGVR).Namespace(backup.Namespace).Get(ctx, backup.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error("get prev mongo backup cr error, ", err)
		return err
	}

	if err == nil {
		var currentBackup psmdbv1.PerconaServerMongoDBBackup
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(currentBackupData.Object, &currentBackup)
		if err != nil {
			klog.Error("to mongo unstructured error, ", err, ", ", currentBackupData.GetName(), ", ", currentBackupData.GetNamespace())
			return err
		}

		if currentBackup.Status.State == psmdbv1.BackupStateRunning {
			klog.Error("prev mongo backup job is still running")
			return errors.New("duplicate mongo backup job")
		}

		klog.Info("remove prev backup, ", ", ", backup.Name, ", ", backup.Namespace)
		err = dynamicClient.Resource(PSMDBBackupClassGVR).Namespace(backup.Namespace).Delete(ctx, backup.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.Error("delete prev mongo backup error, ", err)
			return err
		}

		// waiting for deletion complete
		wait.PollImmediateWithContext(ctx, 1*time.Second, time.Hour, func(ctx context.Context) (done bool, err error) {
			_, err = dynamicClient.Resource(PSMDBBackupClassGVR).Namespace(backup.Namespace).Get(ctx, backup.Name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					klog.Info("prev mongo backup removed")
					return true, nil
				}

				return false, err
			}

			return false, nil
		})
	}

	klog.Info("create a new mongo backup, ", ", ", backup.Name, ", ", backup.Namespace)
	createBackup, err := runtime.DefaultUnstructuredConverter.ToUnstructured(backup)
	if err != nil {
		klog.Error("to mongo unstructured error, ", err, ", ", backup.Name, ", ", backup.Namespace)
		return err
	}

	_, err = dynamicClient.Resource(PSMDBBackupClassGVR).Namespace(backup.Namespace).Create(ctx,
		&unstructured.Unstructured{Object: createBackup}, metav1.CreateOptions{})

	if err != nil {
		klog.Error("create mongo backup error, ", err, ", ", backup.Name, ", ", backup.Namespace)
	}

	return err
}

func WaitForAllBackupComplete(ctx context.Context, dynamicClient *dynamic.DynamicClient) error {
	return wait.PollWithContext(ctx, 5*time.Second, time.Hour,
		func(context.Context) (done bool, err error) {
			var errs []error
			incompleted := false
			backups, err := dynamicClient.Resource(PSMDBBackupClassGVR).List(ctx, metav1.ListOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}

				klog.Error("list mongo backup error, ", err)
				return false, err
			}

			for _, b := range backups.Items {
				backup := psmdbv1.PerconaServerMongoDBBackup{}
				if err = runtime.DefaultUnstructuredConverter.FromUnstructured(b.Object, &backup); err != nil {
					klog.Error("convert mongo backup crd error, ", err, ", ", b.GetName(), ", ", b.GetNamespace())
					return false, err
				}

				switch backup.Status.State {
				case psmdbv1.BackupStateRunning, psmdbv1.BackupStateWaiting,
					psmdbv1.BackupStateNew, psmdbv1.BackupStateRequested:
					incompleted = true
				case psmdbv1.BackupStateRejected, psmdbv1.BackupStateError:
					errs = append(errs, fmt.Errorf("mongo backup failed, %s, %s, %s, %s", backup.Status.Error, backup.Status.State, backup.Name, backup.Namespace))
				}
			}

			if len(errs) > 0 {
				err = utilerrors.NewAggregate(errs)
			}

			done = !incompleted

			return
		})
}

func ForceCreateNewMongoClusterRestore(ctx context.Context, dynamicClient *dynamic.DynamicClient,
	restore *psmdbv1.PerconaServerMongoDBRestore) error {
	currentRestoreData, err := dynamicClient.Resource(PSMDBRestoreClassGVR).Namespace(restore.Namespace).Get(ctx, restore.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error("get prev mongo restore cr error, ", err)
		return err
	}

	if err == nil {
		var currentRestore psmdbv1.PerconaServerMongoDBRestore
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(currentRestoreData.Object, &currentRestore)
		if err != nil {
			klog.Error("to mongo unstructured error, ", err, ", ", currentRestoreData.GetName(), ", ", currentRestoreData.GetNamespace())
			return err
		}

		if currentRestore.Status.State == psmdbv1.RestoreStateRunning {
			klog.Error("prev mongo restore job is still running")
			return errors.New("duplicate mongo restore job")
		}

		klog.Info("remove prev mongo restore, ", ", ", restore.Name, ", ", restore.Namespace)
		err = dynamicClient.Resource(PSMDBRestoreClassGVR).Namespace(restore.Namespace).Delete(ctx, restore.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.Error("delete prev mongo restore error, ", err)
			return err
		}
	}

	klog.Info("create a new mongo restore, ", ", ", restore.Name, ", ", restore.Namespace)
	createRestore, err := runtime.DefaultUnstructuredConverter.ToUnstructured(restore)
	if err != nil {
		klog.Error("to mongo unstructured error, ", err, ", ", restore.Name, ", ", restore.Namespace)
		return err
	}

	_, err = dynamicClient.Resource(PSMDBRestoreClassGVR).Namespace(restore.Namespace).Create(ctx,
		&unstructured.Unstructured{Object: createRestore}, metav1.CreateOptions{})

	if err != nil {
		klog.Error("create mongo restore error, ", err, ", ", restore.Name, ", ", restore.Namespace)
	}

	return err
}

func WaitForAllRestoreComplete(ctx context.Context, dynamicClient *dynamic.DynamicClient) error {
	return wait.PollWithContext(ctx, 5*time.Second, time.Hour,
		func(context.Context) (done bool, err error) {
			var errs []error
			incompleted := false
			restores, err := dynamicClient.Resource(PSMDBRestoreClassGVR).List(ctx, metav1.ListOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return true, nil
				}

				klog.Error("list mongo restore error, ", err)
				return false, err
			}

			for _, b := range restores.Items {
				restore := psmdbv1.PerconaServerMongoDBRestore{}
				if err = runtime.DefaultUnstructuredConverter.FromUnstructured(b.Object, &restore); err != nil {
					klog.Error("convert mongo restore crd error, ", err, ", ", b.GetName(), ", ", b.GetNamespace())
					return false, err
				}

				switch restore.Status.State {
				case psmdbv1.RestoreStateRunning, psmdbv1.RestoreStateWaiting, psmdbv1.RestoreStateNew, psmdbv1.RestoreStateRequested:
					incompleted = true
				case psmdbv1.RestoreStateRejected, psmdbv1.RestoreStateError:
					errs = append(errs, fmt.Errorf("restore mongo failed, %s, %s, %s", restore.Status.Error, restore.Name, restore.Namespace))
				}
			}

			if len(errs) > 0 {
				err = utilerrors.NewAggregate(errs)
			}

			done = !incompleted

			return
		})
}

func GetDatabaseName(namespace, db string) string {
	return fmt.Sprintf("%s_%s", namespace, db)
}

func ListMongoClusters(ctx context.Context, ctrlClient client.Client, namespace string) (clusters []kbappsv1.Cluster, err error) {
	var clusterList kbappsv1.ClusterList
	err = ctrlClient.List(ctx, &clusterList)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusterList.Items {
		if cluster.Labels != nil && cluster.Labels[constants.ClusterInstanceNameKey] == "mongodb" {
			clusters = append(clusters, cluster)
		}
	}
	return clusters, nil
}

func FindMongoAdminUser(ctx context.Context, k8sClient *kubernetes.Clientset, namespace string) (user, password string, err error) {
	secret, err := k8sClient.CoreV1().Secrets(namespace).Get(ctx, "mongodb-mongodb-account-root", metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to find mongo user and password ")
		return
	}
	return string(secret.Data["username"]), string(secret.Data["password"]), nil
}
