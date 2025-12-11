package kvrocksbakcup

import (
	"errors"
	"os"

	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/workload/kvrocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (c *controller) handler(action Action, obj interface{}) error {
	backup, ok := obj.(*v1alpha1.KVRocksBackup)
	if !ok {
		return errors.New("invalid object, not kvrocks backup")
	}

	switch action {
	case ADD:
		switch backup.Status.State {
		case v1alpha1.BackupStateError, v1alpha1.BackupStateRejected, v1alpha1.BackupStateReady:
			klog.Info("ignore kvrocks backup with finished state")
			return nil
		}
		if backup.Spec.ClusterName == "" {
			err := errors.New("kvrocks backup cluster name is empty")
			c.updateStatus(backup, v1alpha1.BackupStateError, err)

			return err
		}

		klog.Info("find kvrocks to backup, ", backup.Spec.ClusterName, ", ", backup.Namespace)
		cluster, err := c.aprClientSet.AprV1alpha1().RedixClusters(backup.Namespace).Get(c.ctx, backup.Spec.ClusterName, metav1.GetOptions{})
		if err != nil {
			klog.Error("find kvrocks to backup error, ", err, ", ", backup.Spec.ClusterName, ", ", backup.Namespace)
			return err
		}

		switch cluster.Spec.Type {
		default:
			klog.Info("ignore unsupported redix cluster type, ", cluster.Spec.Type)
			return nil
		case v1alpha1.RedisCluster:
			// TODO:
			return nil
		case v1alpha1.KVRocks:
		}

		// update status to running
		backup, err = c.updateStatus(backup, v1alpha1.BackupStateRunning, nil)
		if err != nil {
			klog.Errorf("update backup job running error %v", err)
			return err
		}

		err = kvrocks.BackupKVRocks(c.ctx, c.k8sClientSet, cluster, backup)
		if err != nil {
			_, e := c.updateStatus(backup, v1alpha1.BackupStateError, nil)
			if e != nil {
				klog.Errorf("update backup job running error %v", err)
			}

			return err
		}

		_, err = c.updateStatus(backup, v1alpha1.BackupStateReady, nil)
		if err != nil {
			klog.Error("update backup job ready error, ", err)
		}

	case DELETE:
		// clean up the backup files
		if backup.Status.BackupPath != "" {
			err := os.RemoveAll(backup.Status.BackupPath)
			if err != nil {
				klog.Warning("delete kvrocks backup files error, ", err, ", ", backup.Status.BackupPath)
			}
		}
	}

	return nil
}

func (c *controller) updateStatus(backup *v1alpha1.KVRocksBackup, state v1alpha1.BackupState, errMsg error) (*v1alpha1.KVRocksBackup, error) {
	backup.Status.State = state
	if errMsg != nil {
		backup.Status.Error = errMsg.Error()
	}

	now := metav1.Now()
	switch state {
	case v1alpha1.BackupStateError, v1alpha1.BackupStateRejected, v1alpha1.BackupStateReady:
		backup.Status.CompletedAt = &now
	case v1alpha1.BackupStateRunning:
		backup.Status.StartAt = &now
	}

	b, err := c.aprClientSet.AprV1alpha1().KVRocksBackups(backup.Namespace).UpdateStatus(c.ctx, backup, metav1.UpdateOptions{})
	if err != nil {
		klog.Error("update kvrocks backup status error, ", err, ", ", backup.Name, ", ", backup.Namespace)
		return nil, err
	}

	return b, nil
}
