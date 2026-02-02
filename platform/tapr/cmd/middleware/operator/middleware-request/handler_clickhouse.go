package middlewarerequest

import (
	"context"
	"fmt"
	"time"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	wck "bytetrade.io/web3os/tapr/pkg/workload/clickhouse"

	"github.com/ClickHouse/clickhouse-go/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

const clickhouseNamespace = "clickhouse-middleware"

func (c *controller) createOrUpdateClickHouseRequest(req *aprv1.MiddlewareRequest) error {
	adminUser, adminPassword, err := wck.FindClickHouseAdminUser(c.ctx, c.k8sClientSet, clickhouseNamespace)
	if err != nil {
		klog.Errorf("failed to get clickhouse admin user %v", err)
		return err
	}
	userPassword, err := req.Spec.ClickHouse.Password.GetVarValue(c.ctx, c.k8sClientSet, req.Namespace)
	if err != nil {
		klog.Errorf("failed to get clickhouse user password %v", err)
		return err
	}

	conn, err := c.newClickHouseConn(c.ctx, adminUser, adminPassword)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	// create user if not exists
	createUserSQL := fmt.Sprintf("CREATE USER IF NOT EXISTS `%s` IDENTIFIED WITH plaintext_password BY '%s'", req.Spec.ClickHouse.User, userPassword)
	if err := conn.Exec(c.ctx, createUserSQL); err != nil {
		klog.Errorf("failed to create clickhouse user %v", err)
		return err
	}

	// create databases and grant privileges
	for _, d := range req.Spec.ClickHouse.Databases {
		dbName := wck.GetDatabaseName(req.Spec.AppNamespace, d.Name)
		createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)
		if err := conn.Exec(c.ctx, createDBSQL); err != nil {
			klog.Errorf("failed to create database %s %v", dbName, err)
			return err
		}
		grantSQL := fmt.Sprintf("GRANT ALL ON `%s`.* TO `%s`", dbName, req.Spec.ClickHouse.User)
		if err := conn.Exec(c.ctx, grantSQL); err != nil {
			klog.Errorf("failed to grant privileges on %s %v", dbName, err)
			return err
		}
	}
	return nil
}

func (c *controller) deleteClickHouseRequest(req *aprv1.MiddlewareRequest) error {
	adminUser, adminPassword, err := wck.FindClickHouseAdminUser(c.ctx, c.k8sClientSet, clickhouseNamespace)
	if err != nil {
		klog.Errorf("failed to get clickhouse admin user %v", err)
		if apierrors.IsNotFound(err) {
			// ClickHouse admin secret missing, service likely already removed. No-op.
			klog.Infof("clickhouse admin secret not found, skipping deletion for user %s", req.Spec.ClickHouse.User)
			return nil
		}
		return err
	}

	conn, err := c.newClickHouseConn(c.ctx, adminUser, adminPassword)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	// drop user
	dropUserSQL := fmt.Sprintf("DROP USER IF EXISTS `%s`", req.Spec.ClickHouse.User)
	if err := conn.Exec(c.ctx, dropUserSQL); err != nil {
		klog.Errorf("failed to drop user %s %v", req.Spec.ClickHouse.User, err)
		return err
	}

	// drop databases
	for _, d := range req.Spec.ClickHouse.Databases {
		dbName := wck.GetDatabaseName(req.Spec.AppNamespace, d.Name)
		dropDBSQL := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName)
		if err := conn.Exec(c.ctx, dropDBSQL); err != nil {
			klog.Errorf("failed to drop database %s %v", dbName, err)
			return err
		}
	}
	return nil
}

func (c *controller) newClickHouseConn(ctx context.Context, user, password string) (clickhouse.Conn, error) {
	addr := fmt.Sprintf("clickhouse-svc.%s.svc.cluster.local:9000", clickhouseNamespace)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: user,
			Password: password,
		},
		DialTimeout:      10 * time.Second,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		klog.Errorf("open clickhouse native connection error %v", err)
		return nil, err
	}
	if pingErr := conn.Ping(ctx); pingErr != nil {
		klog.Errorf("clickhouse ping error %v", pingErr)
		_ = conn.Close()
		return nil, pingErr
	}
	return conn, nil
}
