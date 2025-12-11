package backup

import (
	"bytetrade.io/web3os/tapr/pkg/workload/citus"
	"bytetrade.io/web3os/tapr/pkg/workload/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (w *Watcher) backupPostgres() error {
	clusters, err := w.aprClientSet.AprV1alpha1().PGClusters("").List(w.ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list pg clusters error, ", err)
		return err
	}

	klog.Info("start to backup all users' pg clusters, of ", len(clusters.Items))
	for _, cluster := range clusters.Items {
		klog.Info("create crd to backup cluster, ", cluster.Name, ", ", cluster.Namespace)
		backup := citus.ClusterBackup.DeepCopy()
		backup.Namespace = cluster.Namespace

		if cluster.Spec.BackupStorage != "" {
			backup.Spec.VolumeSpec.HostPath.Path = cluster.Spec.BackupStorage
		} else {
			klog.Info("find cluster hostpath, ", cluster.Namespace)
			var pvc string
			if cluster.Spec.Owner == "system" {
				pvcRes, err := w.k8sClientSet.CoreV1().PersistentVolumeClaims(cluster.Namespace).Get(w.ctx, "citus-data-pvc", metav1.GetOptions{})
				if err != nil {
					klog.Error("find pg pvc error, ", err)
					return err
				}

				pvRes, err := w.k8sClientSet.CoreV1().PersistentVolumes().Get(w.ctx, pvcRes.Spec.VolumeName, metav1.GetOptions{})
				if err != nil {
					klog.Error("backup find pg pv error, ", err)
					return err
				}

				pvc = pvRes.Spec.HostPath.Path

			} else {
				bflNamespace := "user-space-" + cluster.Spec.Owner
				pvc, err = utils.GetUserDBPVCName(w.ctx, w.k8sClientSet, bflNamespace)
				if err != nil {
					return err
				}
			}

			backup.Spec.VolumeSpec.HostPath.Path = pvc + "/pg_backup"
		}

		err = citus.ForceCreateNewPGClusterBackup(w.ctx, w.aprClientSet, backup)
		if err != nil {
			return err
		}
	}

	klog.Info("wait for all pg cluster backup complete")
	err = citus.WaitForAllBackupComplete(w.ctx, w.aprClientSet)
	if err != nil {
		klog.Error("wait for backup complete error, ", err)
	}

	return err
}

func (w *Watcher) restorePostgres() error {
	clusters, err := w.aprClientSet.AprV1alpha1().PGClusters("").List(w.ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list pg clusters error, ", err)
		return err
	}

	klog.Info("start to restore all users' pg clusters, of ", len(clusters.Items))
	for _, cluster := range clusters.Items {
		klog.Info("create crd to restore cluster, ", cluster.Name, ", ", cluster.Namespace)

		restore := citus.ClusterRestore.DeepCopy()
		restore.Namespace = cluster.Namespace

		err = citus.ForceCreateNewPGClusterRestore(w.ctx, w.aprClientSet, restore)
		if err != nil {
			return err
		}

	}

	klog.Info("wait for all pg cluster restore complete")
	err = citus.WaitForAllRestoreComplete(w.ctx, w.aprClientSet)
	if err != nil {
		klog.Error("wait for pg restore complete error, ", err)
	}

	return err
}
