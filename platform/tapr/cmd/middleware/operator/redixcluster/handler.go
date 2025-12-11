package redixcluster

import (
	"errors"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/workload/kvrocks"
	"k8s.io/klog/v2"
)

func (c *controller) handler(action Action, obj interface{}) error {
	cluster, ok := obj.(*aprv1.RedixCluster)
	if !ok {
		return errors.New("invalid object")
	}

	klog.Info("start to reconcile the redix cluster, ", cluster.Namespace, "/", cluster.Name)

	switch cluster.Spec.Type {
	case aprv1.KVRocks:
	case aprv1.RedisCluster:
		// TODO: sync cluster define to redis cluster operator
		return nil
	default:
		klog.Warning("Unsupported redix cluster type")
		return nil
	}

	switch action {
	case ADD:
		sts, err := kvrocks.CreateOrUpdateKVRocks(c.ctx, c.k8sClientSet, cluster, false)
		if err != nil {
			return err
		}

		if sts != nil {
			_, err = kvrocks.WaitForPodRunning(c.ctx, c.k8sClientSet, sts.Namespace, sts.Name+"-0")
		}
		return err

	case UPDATE:
		_, err := kvrocks.CreateOrUpdateKVRocks(c.ctx, c.k8sClientSet, cluster, true)
		if err != nil {
			return err
		}
	case DELETE:
		return kvrocks.DeleteKVRocks(c.ctx, c.k8sClientSet, cluster)
	}

	return nil
}
