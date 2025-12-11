package middlewarerequest

import (
	"database/sql"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	wmariadb "bytetrade.io/web3os/tapr/pkg/workload/mariadb"
	wmysql "bytetrade.io/web3os/tapr/pkg/workload/mysql"

	_ "github.com/go-sql-driver/mysql"
)

const mysqlNamespace = "mysql-middleware"

func (c *controller) createOrUpdateMysqlRequest(req *aprv1.MiddlewareRequest) error {
	adminUser, adminPassword, err := wmysql.FindMysqlAdminUser(c.ctx, c.k8sClientSet, mysqlNamespace)
	if err != nil {
		klog.Errorf("failed to get mysql admin user %v", err)
		return err
	}
	userPassword, err := req.Spec.Mysql.Password.GetVarValue(c.ctx, c.k8sClientSet, req.Namespace)
	if err != nil {
		klog.Errorf("failed to get mysql user password %v", err)
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", adminUser, adminPassword, c.getMysqlHost())
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		klog.Errorf("failed to open mysql %v", err)
		return err
	}
	defer db.Close()

	// create user if not exists
	createUserSQL := fmt.Sprintf("CREATE USER IF NOT EXISTS `%s` IDENTIFIED BY '%s'", req.Spec.Mysql.User, userPassword)
	_, err = db.ExecContext(c.ctx, createUserSQL)
	if err != nil {
		klog.Errorf("failed to create user %v", err)
		return err
	}

	// create databases and grant privileges
	for _, d := range req.Spec.Mysql.Databases {
		dbName := wmariadb.GetDatabaseName(req.Spec.AppNamespace, d.Name)
		createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)
		if _, err = db.ExecContext(c.ctx, createDBSQL); err != nil {
			klog.Errorf("failed to execute create database %v", err)
			return err
		}
		grantSQL := fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO `%s`", dbName, req.Spec.Mysql.User)
		if _, err = db.ExecContext(c.ctx, grantSQL); err != nil {
			klog.Errorf("failed to grant database %s privileges %v", dbName, err)
			return err
		}
	}
	if _, err = db.ExecContext(c.ctx, "FLUSH PRIVILEGES"); err != nil {
		klog.Errorf("failed to flush user %s privileges %v", req.Spec.Mysql.User, err)
		return err
	}
	return nil
}

func (c *controller) deleteMysqlRequest(req *aprv1.MiddlewareRequest) error {
	adminUser, adminPassword, err := wmysql.FindMysqlAdminUser(c.ctx, c.k8sClientSet, mysqlNamespace)
	if err != nil {
		klog.Errorf("failed to get mysql admin user %v", err)
		if apierrors.IsNotFound(err) {
			// MySQL admin secret missing, service likely already removed. No-op.
			klog.Infof("mysql admin secret not found, skipping deletion for user %s", req.Spec.Mysql.User)
			return nil
		}
		return err
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", adminUser, adminPassword, c.getMysqlHost())
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		klog.Errorf("failed to open connection %v", err)
		return err
	}
	defer db.Close()
	dropUserSQL := fmt.Sprintf("DROP USER IF EXISTS `%s`", req.Spec.Mysql.User)
	_, err = db.ExecContext(c.ctx, dropUserSQL)
	if err != nil {
		klog.Errorf("failed to drop user %s %v", req.Spec.MariaDB.User, err)
		return err
	}

	for _, d := range req.Spec.Mysql.Databases {
		dbName := wmysql.GetDatabaseName(req.Spec.AppNamespace, d.Name)
		dropDBSQL := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName)
		_, err = db.ExecContext(c.ctx, dropDBSQL)
		if err != nil {
			klog.Errorf("failed to drop database %s, %v", dbName, err)
			return err
		}
	}
	return nil
}

func (c *controller) getMysqlHost() string {
	return fmt.Sprintf("mysql-mysql-headless.%s.svc.cluster.local:3306", "mysql-middleware")
}
