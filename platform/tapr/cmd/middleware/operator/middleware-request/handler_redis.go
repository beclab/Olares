package middlewarerequest

import (
	"fmt"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/workload/kvrocks"
	rediscluster "bytetrade.io/web3os/tapr/pkg/workload/redis-cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (c *controller) reconcileRedisPassword(_ *aprv1.MiddlewareRequest) error {

	// TODO: redis-cluster
	return rediscluster.UpdateProxyConfig(c.ctx, c.k8sClientSet, c.aprClientSet, c.dynamicClient)
}

func (c *controller) createOrUpdateRedixRequest(req *aprv1.MiddlewareRequest, isUpdate bool) error {
	// assume middleware request in system namespace only currently
	clusters, err := c.aprClientSet.AprV1alpha1().RedixClusters(constants.PlatformNamespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("find redix cluster error, ", err)
		return err
	}

	// assume only one cluster in system namespace
	if len(clusters.Items) == 0 {
		klog.Warning("redix cluster not found")
		return nil
	}

	cluster := clusters.Items[0]

	switch cluster.Spec.Type {
	case aprv1.RedisCluster:
		return c.reconcileRedisPassword(req)
	case aprv1.KVRocks:
		return c.createOrUpdateKVRocksRequest(req, &cluster, isUpdate)
	}

	return nil
}

func (c *controller) deleteRedixRequest(req *aprv1.MiddlewareRequest) error {
	// assume middleware request in system namespace only currently
	clusters, err := c.aprClientSet.AprV1alpha1().RedixClusters(constants.PlatformNamespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("find redix cluster error, ", err)
		return err
	}

	// assume only one cluster in system namespace
	if len(clusters.Items) == 0 {
		klog.Warning("redix cluster not found")
		return nil
	}

	cluster := clusters.Items[0]

	switch cluster.Spec.Type {
	case aprv1.RedisCluster:
		return c.reconcileRedisPassword(req)
	case aprv1.KVRocks:
		return c.deleteKVRocksRequest(req, &cluster)
	}

	return nil
}

func (c *controller) createOrUpdateKVRocksRequest(req *aprv1.MiddlewareRequest, cluster *aprv1.RedixCluster, isUpdate bool) error {
	cli, err := kvrocks.GetKVRocksClient(c.ctx, c.k8sClientSet, cluster)
	if err != nil {
		klog.Error("get kvrocks client error, ", err)
		return err
	}
	defer cli.Close()

	// TODO: redis db support
	requestNamespace := GetKVRocksNamespaceName(req.Namespace, req.Spec.Redis.Namespace)
	token, err := req.Spec.Redis.Password.GetVarValue(c.ctx, c.k8sClientSet, req.Namespace)
	if err != nil {
		klog.Error("get redis request password error, ", err, ", ", req.Name, ", ", req.Namespace)
		return err
	}

	ns, err := cli.GetNamespace(c.ctx, requestNamespace)
	if err != nil {
		return err
	}

	if ns == nil {
		if isUpdate {
			// update requets, but namespace is a new one. we must check if the token exists.
			// if token exists, remove the old namesapce and create the new namespace
			// if not, just create a new namespace
			allns, err := cli.ListNamespace(c.ctx)
			if err != nil {
				klog.Error("list kvrocks namespace error, ", err, ", ", req.Name, ", ", req.Namespace)
				return err
			}

			for _, n := range allns {
				if n.Token == token {
					err = cli.DeleteNamespace(c.ctx, n.Name)
					if err != nil {
						klog.Warning("delete kvrocks namespace error, ", err, ", ", n.Name)
					}
				}
			}

			// FIXME: if both the namespace's name and token are changed, the old namespace
			// shoud be deleted
		}

		// create namespace
		err = cli.AddNamespace(c.ctx, requestNamespace, token)
		if err != nil {
			klog.Error("create kvrocks namespace error, ", err, ", ", req.Name, ", ", req.Namespace)
			return err
		}
	} else {
		err = cli.UpdateNamespace(c.ctx, requestNamespace, token)
		if err != nil {
			klog.Error("update kvrocks namespace error, ", err, ", ", req.Name, ", ", req.Namespace)
			return err
		}
	}

	return nil
}

func (c *controller) deleteKVRocksRequest(req *aprv1.MiddlewareRequest, cluster *aprv1.RedixCluster) error {
	cli, err := kvrocks.GetKVRocksClient(c.ctx, c.k8sClientSet, cluster)
	if err != nil {
		klog.Error("get kvrocks client error, ", err)
		return err
	}
	defer cli.Close()

	// TODO: redis db support
	requestNamespace := GetKVRocksNamespaceName(req.Namespace, req.Spec.Redis.Namespace)
	ns, err := cli.GetNamespace(c.ctx, requestNamespace)
	if err != nil {
		klog.Error("get kvrocks namespace error, ", err, ", ", req.Name, ", ", req.Namespace)
		return err
	}

	if ns == nil {
		klog.Info("kvrocks namespace not exists, ", req.Name, ", ", req.Namespace)
		return nil
	}

	err = cli.DeleteNamespace(c.ctx, requestNamespace)
	if err != nil {
		klog.Error("delete kvrocks namespace error, ", err, ", ", req.Name, ", ", req.Namespace)
		return err
	}

	// TODO: delete the keys of this namespace

	return nil
}

func GetKVRocksNamespaceName(reqNamespace, dbNamespace string) string {
	return fmt.Sprintf("%s_%s", reqNamespace, dbNamespace)
}
