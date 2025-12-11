package kvrocksrestore

import (
	"errors"

	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/workload/kvrocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (c *controller) handler(action Action, obj interface{}) error {
	restore, ok := obj.(*v1alpha1.KVRocksRestore)
	if !ok {
		return errors.New("invalid object, not kvrocks restore")
	}

	switch action {
	case ADD:
		switch restore.Status.State {
		case v1alpha1.RestoreStateError, v1alpha1.RestoreStateRejected, v1alpha1.RestoreStateReady:
			klog.Info("ignore kvrocks restore with finished state")
			return nil
		}

		if restore.Spec.ClusterName == "" {
			err := errors.New("kvrocks restore cluster name is empty")
			c.updateStatus(restore.DeepCopy(), v1alpha1.RestoreStateError, err)

			return err
		}

		klog.Info("find kvrocks to restore, ", restore.Spec.ClusterName, ", ", restore.Namespace)
		cluster, err := c.aprClientSet.AprV1alpha1().RedixClusters(restore.Namespace).Get(c.ctx, restore.Spec.ClusterName, metav1.GetOptions{})
		if err != nil {
			klog.Error("find kvrocks to backup error, ", err, ", ", restore.Spec.ClusterName, ", ", restore.Namespace)
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
		restore, err = c.updateStatus(restore, v1alpha1.RestoreStateRunning, nil)
		if err != nil {
			klog.Errorf("update restore job running error %v", err)
			return err
		}

		err = kvrocks.RestoreKVRocks(c.ctx, c.k8sClientSet, cluster, restore)
		if err != nil {
			_, e := c.updateStatus(restore, v1alpha1.RestoreStateError, nil)
			if e != nil {
				klog.Errorf("update restore job running error %v", err)
			}

			return err
		}

		_, err = c.updateStatus(restore, v1alpha1.RestoreStateReady, nil)
		if err != nil {
			klog.Error("update restore job ready error, ", err)
		}

	}

	return nil
}

func (c *controller) updateStatus(restore *v1alpha1.KVRocksRestore, state v1alpha1.RestoreState, errMsg error) (*v1alpha1.KVRocksRestore, error) {
	restore.Status.State = state
	if errMsg != nil {
		restore.Status.Error = errMsg.Error()
	}

	now := metav1.Now()
	switch state {
	case v1alpha1.RestoreStateError, v1alpha1.RestoreStateRejected, v1alpha1.RestoreStateReady:
		restore.Status.CompletedAt = &now
	case v1alpha1.RestoreStateRunning:
		restore.Status.StartAt = &now
	}

	r, err := c.aprClientSet.AprV1alpha1().KVRocksRestores(restore.Namespace).UpdateStatus(c.ctx, restore, metav1.UpdateOptions{})
	if err != nil {
		klog.Error("update kvrocks restore status error, ", err, ", ", restore.Name, ", ", restore.Namespace)
		return nil, err
	}

	return r, nil
}
