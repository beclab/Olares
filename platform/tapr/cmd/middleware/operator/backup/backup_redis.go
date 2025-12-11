package backup

import (
	"errors"

	rediscluster "bytetrade.io/web3os/tapr/pkg/workload/redis-cluster"
	"k8s.io/klog/v2"
)

func (w *Watcher) backupRedis() error {
	clusters, err := rediscluster.ListRedisClusters(w.ctx, w.dynamicClient, "")
	if err != nil {
		klog.Error("list redis clusters error, ", err)
		return err
	}

	klog.Info("start to backup all users' redis clusters, of ", len(clusters))
	for _, cluster := range clusters {
		if cluster.Status.Status != "Healthy" {
			err := errors.New("redis cluster not healthy")
			klog.Error("redis cluster status error, ", err, ", ", cluster.Name, ", ", cluster.Namespace)
			return err
		}

		klog.Info("create crd to backup cluster, ", cluster.Name, ", ", cluster.Namespace)
		backup := rediscluster.ClusterBackup.DeepCopy()
		backup.Namespace = cluster.Namespace

		klog.Info("find userspace hostpath, ", cluster.Namespace)
		backupPath, err := w.getMiddlewareBackupPath(cluster.Namespace)
		if err != nil {
			klog.Info("find userspace hostpath error, ", err)
			return err
		}
		backupPath += "/redis-backup"
		backup.Spec.Local.HostPath.Path = backupPath

		err = rediscluster.ForceCreateNewRedisClusterBackup(w.ctx, w.dynamicClient, backup)
		if err != nil {
			return err
		}
	}

	klog.Info("wait for all redis cluster backup complete")
	err = rediscluster.WaitForAllBackupComplete(w.ctx, w.dynamicClient)
	if err != nil {
		klog.Error("wait for backup complete error, ", err)
	}

	return err
}

func (w *Watcher) restoreRedis() error {
	if err := rediscluster.WaitForInitializeComplete(w.ctx, w.dynamicClient, w.k8sClientSet); err != nil {
		klog.Error("redis cluster initialize error, ", err)
		return err
	}

	clusters, err := rediscluster.ListRedisClusters(w.ctx, w.dynamicClient, "")
	if err != nil {
		klog.Error("list redis clusters error, ", err)
		return err
	}

	klog.Info("start to restore all users' redis clusters, of ", len(clusters))
	for _, cluster := range clusters {
		updateCluster := cluster
		if cluster.Status.Restore.Backup != nil {
			cluster.Status.Restore.Backup = nil
			updateCluster, err = rediscluster.UpdataClusterStatus(w.ctx, w.dynamicClient, updateCluster)
			if err != nil {
				klog.Errorf("update restore redis cluster %s status error %v", cluster.Name, err)
				return err
			}
		}

		if updateCluster.Spec.Init == nil {
			updateCluster.Spec.Init = &rediscluster.InitSpec{
				BackupSource: &rediscluster.BackupSourceSpec{
					Name:      rediscluster.ClusterBackup.Name,
					Namespace: cluster.Namespace,
				}}
			updateCluster.Status.Restore.Phase = rediscluster.RestorePhaseRunning
			_, err = rediscluster.UpdataCluster(w.ctx, w.dynamicClient, updateCluster)
			if err != nil {
				klog.Errorf("update restore redis cluster %s backup %s status %v", cluster.Name, rediscluster.ClusterBackup.Name, err)
				return err
			}
		}
	}

	return nil
}
