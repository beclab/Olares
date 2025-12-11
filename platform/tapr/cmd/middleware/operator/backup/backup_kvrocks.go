package backup

import (
	"context"
	"time"

	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/workload/kvrocks"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

func (w *Watcher) backupRedix() error {
	clusters, err := w.aprClientSet.AprV1alpha1().RedixClusters("").List(w.ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list redix clusters error, ", err)
		return err
	}

	klog.Info("start to backup all users' redix clusters, of ", len(clusters.Items))
	for _, cluster := range clusters.Items {
		klog.Info("create crd to backup redix cluster, ", cluster.Name, ", ", cluster.Namespace, ", ", cluster.Spec.Type)
		switch cluster.Spec.Type {
		case v1alpha1.KVRocks:
			backup := kvrocks.KVRocksBackup.DeepCopy()
			backup.Namespace = cluster.Namespace
			backup.Spec.ClusterName = cluster.Name

			err = kvrocks.ForceCreateNewKVRocksBackup(w.ctx, w.aprClientSet, backup)
			if err != nil {
				return err
			}

			klog.Info("wait for kvrocks backup complete")
			err = kvrocks.WaitForAllBackupComplete(w.ctx, w.aprClientSet)
			if err != nil {
				klog.Error("wait for kvrocks backup complete error, ", err)
			}
		case v1alpha1.RedisCluster:
			err = w.backupRedis()
		}
	}

	return err

}

func (w *Watcher) restoreRedix() error {
	clusters, err := w.aprClientSet.AprV1alpha1().RedixClusters("").List(w.ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list redix clusters error, ", err)
		return err
	}

	klog.Info("start to restore all users' redix clusters, of ", len(clusters.Items))
	for _, cluster := range clusters.Items {
		klog.Info("create crd to restore redix cluster, ", cluster.Name, ", ", cluster.Namespace, ", ", cluster.Spec.Type)
		switch cluster.Spec.Type {
		case v1alpha1.KVRocks:
			// wait for kvrocks sts deploy
			err = wait.PollWithContext(w.ctx, time.Second, 30*time.Minute, func(ctx context.Context) (done bool, err error) {
				_, err = w.k8sClientSet.AppsV1().
					StatefulSets(cluster.Namespace).Get(w.ctx, cluster.Name, metav1.GetOptions{})

				if err != nil {
					if apierrors.IsNotFound(err) {
						return false, nil
					}

					klog.Error("find kvrocks sts error, ", err)
					return false, err
				}

				return true, nil
			})

			if err != nil {
				return err
			}

			klog.Info("create kvrocks restore")
			restore := kvrocks.KVRocksRestore.DeepCopy()
			restore.Namespace = cluster.Namespace
			restore.Spec.ClusterName = cluster.Name

			err = kvrocks.ForceCreateNewKVRocksRestore(w.ctx, w.aprClientSet, restore)
			if err != nil {
				return err
			}

			klog.Info("wait for all kvrocks restore complete")
			err = kvrocks.WaitForAllRestoreComplete(w.ctx, w.aprClientSet)
			if err != nil {
				klog.Error("wait for kvrocks restore complete error, ", err)
			}

		case v1alpha1.RedisCluster:
			err = w.restoreRedis()
		}
	}

	return err
}
