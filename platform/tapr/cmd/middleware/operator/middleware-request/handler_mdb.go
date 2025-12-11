package middlewarerequest

import (
	"errors"
	"fmt"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/mongo"
	"bytetrade.io/web3os/tapr/pkg/workload/mongodb"

	kbappsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func (c *controller) createOrUpdateMDBRequest(req *aprv1.MiddlewareRequest) error {
	pwd, err := req.Spec.MongoDB.Password.GetVarValue(c.ctx, c.k8sClientSet, req.Namespace)
	if err != nil {
		return err
	}

	client, err := c.connectToCluster(req)
	if err != nil {
		klog.Errorf("failed to connect to mongodb cluster %v", err)
		return err
	}
	defer client.Close(c.ctx)

	return client.CreateOrUpdateUserWithDatabase(c.ctx, req.Spec.MongoDB.User, pwd, dbRealNames(req.Spec.AppNamespace, req.Spec.MongoDB.Databases))
}

func (c *controller) deleteMDBRequest(req *aprv1.MiddlewareRequest) error {
	client, err := c.connectToCluster(req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// MongoDB cluster or admin secret missing, service likely already removed. No-op.
			klog.Infof("mongodb admin credentials or cluster not found, skipping deletion for user %s", req.Spec.MongoDB.User)
			return nil
		}
		return err
	}
	defer client.Close(c.ctx)

	return client.DropUserAndDatabase(c.ctx, req.Spec.MongoDB.User, dbRealNames(req.Spec.AppNamespace, req.Spec.MongoDB.Databases))
}

func (c *controller) connectToCluster(req *aprv1.MiddlewareRequest) (*mongo.MongoClient, error) {
	host, err := c.getMongoClusterHost()
	if err != nil {
		return nil, err
	}

	user, pwd, err := c.getMongoClusterAdminUser(req)
	if err != nil {
		return nil, err
	}

	client := &mongo.MongoClient{
		User:     user,
		Password: pwd,
		Addr:     host + ":27017",
	}

	err = client.Connect(c.ctx)
	if err != nil {
		klog.Error("connect mongodb error, ", err, ", ", host)
		return nil, err
	}

	return client, nil
}

func (c *controller) getMongoClusterHost() (string, error) {
	var cluster kbappsv1.Cluster
	err := c.ctrlClient.Get(c.ctx, types.NamespacedName{Namespace: "mongodb-middleware", Name: "mongodb"}, &cluster)
	if err != nil {
		klog.Errorf("failed to find mongo cluster %v", err)
		return "", err
	}
	if cluster.Status.Phase != "Running" {
		return "", errors.New("cluster mongo is not running")
	}
	return fmt.Sprintf("%s-mongodb-headless.%s", cluster.Name, cluster.Namespace), nil
}

func (c *controller) getMongoClusterAdminUser(req *aprv1.MiddlewareRequest) (user, password string, err error) {
	return mongodb.FindMongoAdminUser(c.ctx, c.k8sClientSet, "mongodb-middleware")
}

func dbRealNames(namespace string, dbs []aprv1.MongoDatabase) []aprv1.MongoDatabase {
	ret := make([]aprv1.MongoDatabase, 0, len(dbs))
	for _, db := range dbs {
		ret = append(ret, aprv1.MongoDatabase{
			Name:    mongodb.GetDatabaseName(namespace, db.Name),
			Scripts: db.Scripts,
		})
	}

	return ret
}
