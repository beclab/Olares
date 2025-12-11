package middlewarerequest

import (
	"strconv"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/postgres"
	"bytetrade.io/web3os/tapr/pkg/workload/citus"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (c *controller) createOrUpdatePGRequest(req *aprv1.MiddlewareRequest) error {
	sts, adminUser, adminPwd, err := c.findClusterWorkloadAndAdminuserAndPassword()
	if err != nil {
		return err
	}

	var index int32 = 0
	for index < *sts.Spec.Replicas {
		nodeHost := sts.Name + "-" + strconv.Itoa(int(index)) + ".citus-headless." + sts.Namespace

		if err := func() error {
			nodeClient, err := postgres.NewClientBuidler(adminUser, adminPwd, nodeHost, postgres.PG_PORT).Build()
			if err != nil {
				klog.Error("connect to node error, ", err, ", ", nodeHost)
				return err
			}
			defer nodeClient.Close()

			klog.Info("create user, ", req.Spec.PostgreSQL.User)
			pwd, err := req.Spec.PostgreSQL.Password.GetVarValue(c.ctx, c.k8sClientSet, req.Namespace)
			if err != nil {
				return err
			}

			err = nodeClient.CreateOrUpdateUser(c.ctx, req.Spec.PostgreSQL.User, pwd)
			if err != nil {
				return err
			}

			for _, db := range req.Spec.PostgreSQL.Databases {

				klog.Info("create db for user, ", db.Name, ", ", db.IsDistributed(), ", ", req.Spec.PostgreSQL.User)
				dbRealName := citus.GetDatabaseName(req.Spec.AppNamespace, db.Name)
				if db.IsDistributed() || index == 0 {
					err = nodeClient.CreateDatabaseIfNotExists(c.ctx, dbRealName, req.Spec.PostgreSQL.User)
					if err != nil {
						return err
					}

					err = nodeClient.SwitchDatabase(dbRealName)
					if err != nil {
						return err
					}
					if db.IsDistributed() {
						err = nodeClient.CreateCitus(c.ctx)
						if err != nil {
							klog.Error("create citus error, ", err)
							return err
						}

						if index == 0 { // master node
							err = nodeClient.SetMasterNode(c.ctx, nodeHost, postgres.PG_PORT)
							if err != nil {
								klog.Error("set master node error, ", err, ", ", nodeHost)
								return err
							}
						}
					}
					if len(db.Extensions) > 0 {
						err = nodeClient.CreateExtensions(c.ctx, db.Extensions)
						if err != nil {
							klog.Errorf("failed to create extension err=%v", err)
							return err
						}
					}
					if len(db.Scripts) > 0 {
						err = nodeClient.ExecuteScript(c.ctx, dbRealName, req.Spec.PostgreSQL.User, db.Scripts)
						if err != nil {
							klog.Errorf("failed to execute script err=%v", err)
							return err
						}

					}
				}
			}
			return nil
		}(); err != nil {
			return err
		}

		index += 1
	} // end loop replicas

	return c.addWorkerNode(req)
}

func (c *controller) addWorkerNode(req *aprv1.MiddlewareRequest) error {
	sts, adminUser, adminPwd, err := c.findClusterWorkloadAndAdminuserAndPassword()
	if err != nil {
		return err
	}

	masterHost := sts.Name + "-0.citus-headless." + sts.Namespace
	masterClientBuilder := postgres.NewClientBuidler(adminUser, adminPwd, masterHost, postgres.PG_PORT)

	var index int32 = 1
	for index < *sts.Spec.Replicas {
		nodeHost := sts.Name + "-" + strconv.Itoa(int(index)) + ".citus-headless." + sts.Namespace
		for _, db := range req.Spec.PostgreSQL.Databases {
			if db.IsDistributed() {
				dbRealName := citus.GetDatabaseName(req.Spec.AppNamespace, db.Name)
				masterClient, err := masterClientBuilder.WithDatabase(dbRealName).Build()
				if err != nil {
					klog.Error("connect to master error, ", err, ", ", masterHost)
					return err
				}

				err = masterClient.AddWorkerNode(c.ctx, nodeHost, postgres.PG_PORT)
				masterClient.Close()
				if err != nil {
					return err
				}
			}
		}

		index += 1
	}

	return nil
}

func (c *controller) PGClusterRecreated(cluster *aprv1.PGCluster) {
	reqs, err := c.aprClientSet.AprV1alpha1().MiddlewareRequests("").List(c.ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list middleware requests error, ", err, ", ", cluster.Namespace, "/", cluster.Name)
		return
	}

	for _, r := range reqs.Items {
		klog.Info("reconcil middleware request, ", r.Namespace, "/", r.Name, ",", r.Spec.Middleware)
		err := c.createOrUpdatePGRequest(&r)
		if err != nil {
			klog.Error("create middleware request by cluster error, ", err, ",", r.Namespace)
		}
	}
}

func (c *controller) deleteDatabaseIfNotExists(req *aprv1.MiddlewareRequest) error {
	sts, adminUser, adminPwd, err := c.findClusterWorkloadAndAdminuserAndPassword()
	if err != nil {
		return err
	}

	masterHost := sts.Name + "-0.citus-headless." + sts.Namespace
	masterClient, err := postgres.NewClientBuidler(adminUser, adminPwd, masterHost, postgres.PG_PORT).Build()
	if err != nil {
		klog.Error("connect to master error, ", err, ", ", masterHost)
		return err
	}

	dbs, err := masterClient.ListDatabaseByOwner(c.ctx, req.Spec.PostgreSQL.User)
	masterClient.Close()
	if err != nil {
		klog.Error("list db by owner error, ", err, ", ", req.Spec.PostgreSQL.User)
		return err
	}

	for _, db := range dbs {

		if func() bool {
			for _, rdb := range req.Spec.PostgreSQL.Databases {
				if citus.GetDatabaseName(req.Spec.AppNamespace, rdb.Name) == db {
					return false
				}
			}

			return true
		}() {
			// not found in request, delete it
			var index int32 = 0
			for index < *sts.Spec.Replicas {
				nodeHost := sts.Name + "-" + strconv.Itoa(int(index)) + ".citus-headless." + sts.Namespace
				nodeClient, err := postgres.NewClientBuidler(adminUser, adminPwd, nodeHost, postgres.PG_PORT).Build()
				if err != nil {
					klog.Error("cannot connect to host, ", err, ", ", nodeHost)
					return err
				}

				err = nodeClient.DropDatabase(c.ctx, db)
				nodeClient.Close()
				if err != nil {
					return err
				}

				index += 1
			}
		}
	} // end loog dbs

	return nil
}

func (c *controller) deletePGAll(req *aprv1.MiddlewareRequest) error {
	sts, adminUser, adminPwd, err := c.findClusterWorkloadAndAdminuserAndPassword()
	if err != nil {
		return err
	}

	var index int32 = 0
	for index < *sts.Spec.Replicas {
		nodeHost := sts.Name + "-" + strconv.Itoa(int(index)) + ".citus-headless." + sts.Namespace
		nodeClient, err := postgres.NewClientBuidler(adminUser, adminPwd, nodeHost, postgres.PG_PORT).Build()
		if err != nil {
			klog.Error("cannot connect to host, ", err, ", ", nodeHost)
			return err
		}

		if err = func() error {
			defer nodeClient.Close()

			// delete db
			for _, db := range req.Spec.PostgreSQL.Databases {
				if db.IsDistributed() || index == 0 {
					dbRealName := citus.GetDatabaseName(req.Spec.AppNamespace, db.Name)
					err = nodeClient.DropDatabase(c.ctx, dbRealName)
					if err != nil {
						return err
					}
				}
			}

			// remove other db not int request to make sure user can be deleted
			dbs, err := nodeClient.ListDatabaseByOwner(c.ctx, req.Spec.PostgreSQL.User)
			if err != nil {
				klog.Error("list db by owner error, ", err, ", ", req.Spec.PostgreSQL.User)
				return err
			}

			for _, db := range dbs {
				err = nodeClient.DropDatabase(c.ctx, db)
				if err != nil {
					return err
				}
			}

			// delete user
			err = nodeClient.DeleteUser(c.ctx, req.Spec.PostgreSQL.User)

			return err
		}(); err != nil {
			return err
		}

		index += 1
	}

	return nil
}

func (c *controller) findClusterWorkloadAndAdminuserAndPassword() (sts *appsv1.StatefulSet, adminUser, adminPwd string, err error) {
	sts, err = c.k8sClientSet.AppsV1().StatefulSets(citus.PGClusterNamespace).Get(c.ctx, citus.PGClusterName, metav1.GetOptions{})
	if err != nil {
		return
	}

	adminUser, adminPwd, err = citus.GetPGClusterAdminUserAndPassword(c.ctx, c.aprClientSet, c.k8sClientSet, citus.PGClusterNamespace)
	if err != nil {
		klog.Error("find cluster admin user error, ", err)
		return
	}

	return
}
