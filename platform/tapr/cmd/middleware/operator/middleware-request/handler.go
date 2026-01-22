package middlewarerequest

import (
	"errors"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"

	"k8s.io/klog/v2"
)

func (c *controller) handler(action Action, obj interface{}) error {
	request, ok := obj.(*aprv1.MiddlewareRequest)
	if !ok {
		return errors.New("invalid object")
	}

	switch request.Spec.Middleware {
	case aprv1.TypePostgreSQL:
		switch action {
		case ADD, UPDATE:
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

		case DELETE:
			err := c.deletePGAll(request)
			if err != nil {
				return err
			}
		}
	case aprv1.TypeMongoDB:
		switch action {
		case ADD, UPDATE:
			if err := c.createOrUpdateMDBRequest(request); err != nil {
				return err
			}

		case DELETE:
			if err := c.deleteMDBRequest(request); err != nil {
				return err
			}
		}
	case aprv1.TypeRedis:
		switch action {
		case ADD, UPDATE:
			if err := c.createOrUpdateRedixRequest(request, action == UPDATE); err != nil {
				return err
			}

		case DELETE:
			if err := c.deleteRedixRequest(request); err != nil {
				return err
			}
		}
	case aprv1.TypeNats:
		switch action {
		case ADD, UPDATE:
			klog.Infof("create nat user name: %s", request.Name)
			if err := c.createOrUpdateNatsUser(request); err != nil {
				return err
			}
		case DELETE:
			if err := c.deleteNatsUserAndStream(request); err != nil {
				return err
			}
		}
	case aprv1.TypeMinio:
		switch action {
		case ADD, UPDATE:
			klog.Infof("create minio user name: %s", request.Name)
			if err := c.createOrUpdateMinioRequest(request); err != nil {
				klog.Errorf("failed to process minio create or update request %v", err)
				return err
			}
		case DELETE:
			if err := c.deleteMinioRequest(request); err != nil {
				klog.Errorf("failed to process minio delete request %v", err)
				return err
			}
		}
	case aprv1.TypeRabbitMQ:
		switch action {
		case ADD, UPDATE:
			klog.Infof("create rabbitmq user name: %s", request.Name)
			if err := c.createOrUpdateRabbitMQRequest(request); err != nil {
				klog.Errorf("failed to process rabbitmq create or update request %v", err)
				return err
			}
		case DELETE:
			if err := c.deleteRabbitMQRequest(request); err != nil {
				klog.Errorf("failed to process rabbitmq delete request %v", err)
				return err
			}
		}
	case aprv1.TypeElasticsearch:
		switch action {
		case ADD, UPDATE:
			klog.Infof("create elasticsearch user name: %s", request.Name)
			if err := c.createOrUpdateElasticsearchRequest(request); err != nil {
				klog.Errorf("failed to process elasticsearch create or update request %v", err)
				return err
			}
		case DELETE:
			if err := c.deleteElasticsearchRequest(request); err != nil {
				klog.Errorf("failed to process elasticsearch delete request %v", err)
				return err
			}
		}
	case aprv1.TypeMariaDB:
		switch action {
		case ADD, UPDATE:
			klog.Infof("create mariadb user name: %s", request.Name)
			if err := c.createOrUpdateMariaDBRequest(request); err != nil {
				klog.Errorf("failed to process mariadb create or update request %v", err)
				return err
			}
		case DELETE:
			if err := c.deleteMariaDBRequest(request); err != nil {
				klog.Errorf("failed to process mariadb delete request %v", err)
				return err
			}
		}
	case aprv1.TypeMysql:
		switch action {
		case ADD, UPDATE:
			klog.Infof("create mysql user name: %s", request.Name)
			if err := c.createOrUpdateMysqlRequest(request); err != nil {
				klog.Errorf("failed to process mysql create or update request %v", err)
				return err
			}
		case DELETE:
			if err := c.deleteMysqlRequest(request); err != nil {
				klog.Errorf("failed to process mysql delete request %v", err)
				return err
			}
		}
	case aprv1.TypeClickHouse:
		switch action {
		case ADD, UPDATE:
			klog.Infof("create clickhouse user name: %s", request.Name)
			if err := c.createOrUpdateClickHouseRequest(request); err != nil {
				klog.Errorf("failed to process clickhouse create or update request %v", err)
				return err
			}
		case DELETE:
			if err := c.deleteClickHouseRequest(request); err != nil {
				klog.Errorf("failed to process clickhouse delete request %v", err)
				return err
			}
		}
	}

	return nil
}
