package pgcluster

import (
	"errors"
	"strconv"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/postgres"
	"bytetrade.io/web3os/tapr/pkg/workload/citus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
)

func (c *controller) handler(action Action, obj interface{}) error {
	cluster, ok := obj.(*aprv1.PGCluster)
	if !ok {
		return errors.New("invalid object")
	}

	klog.Info("start to reconcile the cluster, ", cluster.Namespace, "/", cluster.Name)

	currentCluster, err := c.aprClientSet.AprV1alpha1().PGClusters(cluster.Namespace).Get(c.ctx, cluster.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	clusterDefNotFound := apierrors.IsNotFound(err)
	currentSts, err := c.k8sClientSet.AppsV1().StatefulSets(cluster.Namespace).Get(c.ctx, citus.PGClusterName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	var newCluster bool = false
	switch {
	case clusterDefNotFound:
		// delete cluster
		if err == nil || !apierrors.IsNotFound(err) {
			err := c.k8sClientSet.AppsV1().StatefulSets(cluster.Namespace).Delete(c.ctx, citus.PGClusterName, metav1.DeleteOptions{})
			if err != nil {
				klog.Error("delete pg cluster sts error, ", err)
				return err
			}
		}

		klog.Info("pg cluster deleted")

		// do nothing more
		return nil
	case apierrors.IsNotFound(err):
		// create cluster:
		currentSts, err = citus.CreatePGClusterForUser(c.ctx, c.k8sClientSet, c.aprClientSet, c.lister, currentCluster.Spec.Owner, currentCluster)
		if err != nil {
			klog.Error("create pg cluster error, ", err)
			return err
		}

		newCluster = true
		klog.Info("a new pg cluster created")
	default:
		if currentCluster.Spec.AdminUser == "" {
			// get default user
			secret, err := c.k8sClientSet.CoreV1().Secrets(cluster.Namespace).Get(c.ctx, citus.CitusAdminSecretName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			currentCluster.Spec.Password = aprv1.PasswordVar{
				Value: string(secret.Data["password"]),
			}
			currentCluster.Spec.AdminUser = string(secret.Data["user"])
		}
	}

	// update db admin user if necessary
	must, olduser, oldpwd, err := citus.MustUpdateClusterAdminUser(c.ctx,
		c.k8sClientSet,
		currentSts.Namespace,
		currentSts,
		currentCluster)
	if err != nil {
		return err
	}

	if must {
		klog.Info("update pg cluster admin user")
		if err = func() error {
			newpwd, err := citus.GetPGClusterDefinedPassword(c.ctx, c.k8sClientSet, currentSts.Namespace, currentCluster)
			if err != nil {
				return err
			}

			if !newCluster {
				// patch sts
				for i, c := range currentSts.Spec.Template.Spec.Containers {
					if c.Name == "postgres" {
						for n, env := range c.Env {
							switch env.Name {
							case "POSTGRES_USER":
								currentSts.Spec.Template.Spec.Containers[i].Env[n].Value = currentCluster.Spec.AdminUser
							case "POSTGRES_PASSWORD":
								if currentCluster.Spec.Password.ValueFrom != nil {
									currentSts.Spec.Template.Spec.Containers[i].Env[n].ValueFrom = &corev1.EnvVarSource{
										SecretKeyRef: currentCluster.Spec.Password.ValueFrom.SecretKeyRef.DeepCopy(),
									}
								} else if currentCluster.Spec.Password.Value != "" {
									currentSts.Spec.Template.Spec.Containers[i].Env[n].Value = currentCluster.Spec.Password.Value
								}
							}
						} // end envs
					} // end container found
				} // end container loop

				_, err = c.k8sClientSet.AppsV1().StatefulSets(currentSts.Namespace).Update(c.ctx, currentSts, metav1.UpdateOptions{})
				if err != nil {
					klog.Error("update sts password env error, ", err)
					return err
				}

			}

			// update db admin password on all nodes of current replicas
			var nodeIndex int32 = 0
			for nodeIndex < *currentSts.Spec.Replicas {
				podName := currentSts.Name + "-" + strconv.Itoa(int(nodeIndex))
				// host := podName + ".citus-headless." + currentSts.Namespace
				ip, err := citus.WaitForPodRunning(c.ctx, c.k8sClientSet, currentSts.Namespace, podName)
				if err != nil {
					klog.Error("wait for pod running error, ", err, ", ", podName)
					return err
				}

				pgclient, err := postgres.NewClientBuidler(olduser, oldpwd,
					ip, postgres.PG_PORT).Build()
				if err != nil {
					klog.Error("connect to pg master error, ", err, ", ", olduser, ", ", oldpwd)
					return err
				}

				err = pgclient.ChangeAdminUser(c.ctx, olduser, currentCluster.Spec.AdminUser, newpwd)
				if err != nil {
					pgclient.Close()
					klog.Error("change admin user error, ", err, ", ", olduser, ", ", currentCluster.Spec.AdminUser, ", ", newpwd)
					return err
				}

				pgclient.Close()
				nodeIndex += 1
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	// notify until admin user updated
	// set master node for distributed databse requests
	if newCluster {
		c.notifyClusterCreated(currentCluster)
	}

	// scale cluster if necessary
	if currentCluster.Spec.Replicas != *currentSts.Spec.Replicas {
		pwd, err := citus.GetPGClusterDefinedPassword(c.ctx, c.k8sClientSet, currentSts.Namespace, currentCluster)
		if err != nil {
			return err
		}

		effected, err := citus.ScalePGClusterNodes(c.ctx, c.k8sClientSet, currentSts.Namespace,
			currentCluster.Spec.Replicas, currentCluster.Spec.AdminUser, pwd)
		if err != nil {
			return err
		}

		switch {
		case effected > 0:
			// scale up
			// find out all distributed db request
			// connect to the master node
			masterClientBuilder := postgres.NewClientBuidler(currentCluster.Spec.AdminUser, pwd,
				postgres.PG_MASTER_HOST+"."+currentSts.Namespace, postgres.PG_PORT)

			requests, err := c.requestLister.MiddlewareRequests(currentSts.Namespace).List(labels.Everything())
			if err != nil {
				klog.Error("list all middleware request error, ", err)
				return err
			}

			for _, req := range requests {
				if req.Spec.Middleware == aprv1.TypePostgreSQL {
					klog.Info("sync middleware postgres request, ", req.Spec.PostgreSQL.Databases)
					for _, db := range req.Spec.PostgreSQL.Databases {
						if db.IsDistributed() {
							var newReplicas int32 = *currentSts.Spec.Replicas
							for newReplicas < currentCluster.Spec.Replicas {
								// waiting for replica running
								replicaPodName := currentSts.Name + "-" + strconv.Itoa(int(newReplicas))
								ip, err := citus.WaitForPodRunning(c.ctx, c.k8sClientSet, currentSts.Namespace, replicaPodName)
								if err != nil {
									return err
								}

								// create distributed db
								nodeHost := replicaPodName + ".citus-headless." + currentSts.Namespace
								klog.Info("creating distributed database on ", nodeHost, ", ", db.Name)
								if err = func() error {
									nodeClient, err := postgres.NewClientBuidler(currentCluster.Spec.AdminUser, pwd,
										ip, postgres.PG_PORT).Build()
									if err != nil {
										klog.Error("connect to pg node error, ", err, ", ", currentCluster.Spec.AdminUser, ", ", pwd)
										return err
									}
									defer nodeClient.Close()

									klog.Info("create or update user, ", req.Spec.PostgreSQL.User)
									pwd, err := req.Spec.PostgreSQL.Password.GetVarValue(c.ctx, c.k8sClientSet, req.Namespace)
									if err != nil {
										return err
									}

									err = nodeClient.CreateOrUpdateUser(c.ctx, req.Spec.PostgreSQL.User, pwd)
									if err != nil {
										return err
									}

									err = nodeClient.CreateDatabaseIfNotExists(c.ctx, db.Name, req.Spec.PostgreSQL.User)
									if err != nil {
										klog.Error("create database error, ", err, ", ", nodeHost, ", ", db.Name)
										return err
									}

									err = nodeClient.SwitchDatabase(db.Name)
									if err != nil {
										klog.Error("switch database error, ", err, ", ", nodeHost, ", ", db.Name)
										return err
									}

									err = nodeClient.CreateCitus(c.ctx)
									if err != nil {
										klog.Error("create citus error, ", err, ", ", nodeHost, ", ", db.Name)
										return err
									}

									masterClient, err := masterClientBuilder.WithDatabase(db.Name).Build()
									if err != nil {
										klog.Error("connect to master error, ", err, ", ", db.Name)
										return err
									}

									err = masterClient.AddWorkerNode(c.ctx, nodeHost, postgres.PG_PORT)
									if err != nil {
										klog.Error("add worker to master error, ", err, ", ", nodeHost, ", ", db.Name)
										masterClient.Close()
										return err
									}

									masterClient.Close()

									return nil
								}(); err != nil {
									return err
								}

								klog.Info("success to create distributed database on ", nodeHost, ", ", db.Name)
								newReplicas += 1
							} // end all scaled nodes

							// rebalance all nodes
							masterClient, err := masterClientBuilder.WithDatabase(db.Name).Build()
							if err != nil {
								klog.Error("connect to master error, ", err, ", ", db.Name)
								return err
							}

							err = masterClient.Rebalance(c.ctx)
							masterClient.Close()
							if err != nil {
								klog.Error("rebalance cluster, ", err, ", ", db.Name)
								return err
							}

						} // end distributed db
					} // end  databases loop
				}
			}
		case effected < 0:
			// scale down

		}
	}

	return nil
}
