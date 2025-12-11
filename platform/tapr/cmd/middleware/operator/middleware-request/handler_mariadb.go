package middlewarerequest

import (
	"database/sql"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	wmariadb "bytetrade.io/web3os/tapr/pkg/workload/mariadb"

	_ "github.com/go-sql-driver/mysql"
)

const mariadbNamespace = "mariadb-middleware"

func (c *controller) createOrUpdateMariaDBRequest(req *aprv1.MiddlewareRequest) error {
	adminUser, adminPassword, err := wmariadb.FindMariaDBAdminUser(c.ctx, c.k8sClientSet, mariadbNamespace)
	if err != nil {
		klog.Errorf("failed to get admin user %v", err)
		return err
	}
	userPassword, err := req.Spec.MariaDB.Password.GetVarValue(c.ctx, c.k8sClientSet, req.Namespace)
	if err != nil {
		klog.Errorf("failed to mariadb ")
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", adminUser, adminPassword, c.getMariaDBHost())
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		klog.Errorf("failed to open mariadb %v", err)
		return err
	}
	defer db.Close()

	// create user if not exists
	createUserSQL := fmt.Sprintf("CREATE USER IF NOT EXISTS `%s` IDENTIFIED BY '%s'", req.Spec.MariaDB.User, userPassword)
	_, err = db.ExecContext(c.ctx, createUserSQL)
	if err != nil {
		klog.Errorf("failed to create user %v", err)
		return err
	}

	// create databases and grant privileges
	for _, d := range req.Spec.MariaDB.Databases {
		dbName := wmariadb.GetDatabaseName(req.Spec.AppNamespace, d.Name)
		createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)
		if _, err = db.ExecContext(c.ctx, createDBSQL); err != nil {
			klog.Errorf("failed to execute create database %v", err)
			return err
		}
		grantSQL := fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO `%s`", dbName, req.Spec.MariaDB.User)
		if _, err = db.ExecContext(c.ctx, grantSQL); err != nil {
			klog.Errorf("failed to grant database %s privileges %v", dbName, err)
			return err
		}
	}
	if _, err = db.ExecContext(c.ctx, "FLUSH PRIVILEGES"); err != nil {
		klog.Errorf("failed to flush user %s privileges %v", req.Spec.MariaDB.User, err)
		return err
	}
	return nil
}

func (c *controller) deleteMariaDBRequest(req *aprv1.MiddlewareRequest) error {
	adminUser, adminPassword, err := wmariadb.FindMariaDBAdminUser(c.ctx, c.k8sClientSet, mariadbNamespace)
	if err != nil {
		klog.Errorf("failed to get mariadb admin user %v", err)
		if apierrors.IsNotFound(err) {
			// MariaDB admin secret missing, service likely already removed. No-op.
			klog.Infof("mariadb admin secret not found, skipping deletion for user %s", req.Spec.MariaDB.User)
			return nil
		}
		return err
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", adminUser, adminPassword, c.getMariaDBHost())
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		klog.Errorf("failed to open connection %v", err)
		return err
	}
	defer db.Close()
	dropUserSQL := fmt.Sprintf("DROP USER IF EXISTS `%s`", req.Spec.MariaDB.User)
	_, err = db.ExecContext(c.ctx, dropUserSQL)
	if err != nil {
		klog.Errorf("failed to drop user %s %v", req.Spec.MariaDB.User, err)
		return err
	}

	for _, d := range req.Spec.MariaDB.Databases {
		dbName := wmariadb.GetDatabaseName(req.Spec.AppNamespace, d.Name)
		dropDBSQL := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName)
		_, err = db.ExecContext(c.ctx, dropDBSQL)
		if err != nil {
			klog.Errorf("failed to drop database %s, %v", dbName, err)
			return err
		}
	}
	return nil
}

func (c *controller) getMariaDBHost() string {
	return fmt.Sprintf("mariadb-mariadb-headless.%s.svc.cluster.local:3306", "mariadb-middleware")
}
