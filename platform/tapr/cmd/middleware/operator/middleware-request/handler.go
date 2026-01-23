package middlewarerequest

import (
	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/klog/v2"
)

func (c *controller) handler(action Action, obj interface{}) error {
	request, ok := obj.(*aprv1.MiddlewareRequest)
	if !ok {
		return errors.New("invalid object")
	}
	switch action {
	case ADD, UPDATE:
		return c.handleCreateOrUpdate(action, request)
	case DELETE:
		return c.handleDeletion(request)
	}
	return nil
}

func (c *controller) handleCreateOrUpdate(action Action, request *aprv1.MiddlewareRequest) error {
	if err := c.addFinalizer(request); err != nil {
		klog.Errorf("failed to add finalizer to %s/%s: %v", request.Namespace, request.Name, err)
		return err
	}

	switch request.Spec.Middleware {
	case aprv1.TypePostgreSQL:
		// create app db user
		err := c.createOrUpdatePGRequest(request)
		if err != nil {
			return err
		}

		if action == UPDATE {
			// delete db if not in request
			err = c.deleteDatabaseIfNotExists(request)
			if err != nil {
				return err
			}
		}

	case aprv1.TypeMongoDB:
		if err := c.createOrUpdateMDBRequest(request); err != nil {
			return err
		}

	case aprv1.TypeRedis:
		if err := c.createOrUpdateRedixRequest(request, action == UPDATE); err != nil {
			return err
		}

	case aprv1.TypeNats:
		klog.Infof("create nat user name: %s", request.Name)
		if err := c.createOrUpdateNatsUser(request); err != nil {
			return err
		}

	case aprv1.TypeMinio:
		klog.Infof("create minio user name: %s", request.Name)
		if err := c.createOrUpdateMinioRequest(request); err != nil {
			klog.Errorf("failed to process minio create or update request %v", err)
			return err
		}

	case aprv1.TypeRabbitMQ:
		klog.Infof("create rabbitmq user name: %s", request.Name)
		if err := c.createOrUpdateRabbitMQRequest(request); err != nil {
			klog.Errorf("failed to process rabbitmq create or update request %v", err)
			return err
		}

	case aprv1.TypeElasticsearch:
		klog.Infof("create elasticsearch user name: %s", request.Name)
		if err := c.createOrUpdateElasticsearchRequest(request); err != nil {
			klog.Errorf("failed to process elasticsearch create or update request %v", err)
			return err
		}

	case aprv1.TypeMariaDB:
		klog.Infof("create mariadb user name: %s", request.Name)
		if err := c.createOrUpdateMariaDBRequest(request); err != nil {
			klog.Errorf("failed to process mariadb create or update request %v", err)
			return err
		}

	case aprv1.TypeMysql:
		klog.Infof("create mysql user name: %s", request.Name)
		if err := c.createOrUpdateMysqlRequest(request); err != nil {
			klog.Errorf("failed to process mysql create or update request %v", err)
			return err
		}

	case aprv1.TypeClickHouse:
		klog.Infof("create clickhouse user name: %s", request.Name)
		if err := c.createOrUpdateClickHouseRequest(request); err != nil {
			klog.Errorf("failed to process clickhouse create or update request %v", err)
			return err
		}
	}

	return nil
}

func (c *controller) handleDeletion(request *aprv1.MiddlewareRequest) error {
	if !containsString(request.ObjectMeta.Finalizers, middlewareRequestFinalizer) {
		return nil
	}

	klog.Infof("handling deletion for %s/%s", request.Namespace, request.Name)

	// Perform actual middleware resource cleanup based on type
	var err error
	switch request.Spec.Middleware {
	case aprv1.TypePostgreSQL:
		err = c.deletePGAll(request)
	case aprv1.TypeMongoDB:
		err = c.deleteMDBRequest(request)
	case aprv1.TypeRedis:
		err = c.deleteRedixRequest(request)
	case aprv1.TypeNats:
		err = c.deleteNatsUserAndStream(request)
	case aprv1.TypeMinio:
		err = c.deleteMinioRequest(request)
	case aprv1.TypeRabbitMQ:
		err = c.deleteRabbitMQRequest(request)
	case aprv1.TypeElasticsearch:
		err = c.deleteElasticsearchRequest(request)
	case aprv1.TypeMariaDB:
		err = c.deleteMariaDBRequest(request)
	case aprv1.TypeMysql:
		err = c.deleteMysqlRequest(request)
	case aprv1.TypeClickHouse:
		err = c.deleteClickHouseRequest(request)
	}

	if err != nil {
		klog.Errorf("failed to delete middleware resources for %s/%s: %v", request.Namespace, request.Name, err)
		return err
	}

	klog.Infof("middleware cleanup successful, removing finalizer from %s/%s", request.Namespace, request.Name)
	if err := c.removeFinalizer(request); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("failed to remove finalizer from %s/%s: %v", request.Namespace, request.Name, err)
		return err
	}

	klog.Infof("finalizer removed, resource %s/%s will be deleted", request.Namespace, request.Name)
	return nil
}
