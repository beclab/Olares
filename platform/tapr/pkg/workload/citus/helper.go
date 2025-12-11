package citus

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"
	listerv1alpha1 "bytetrade.io/web3os/tapr/pkg/generated/listers/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/postgres"
	taprutils "bytetrade.io/web3os/tapr/pkg/utils"
	"bytetrade.io/web3os/tapr/pkg/workload/utils"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func GetPGClusterDefineByUser(ctx context.Context, client *kubernetes.Clientset, user, namespace string, clusterDef *v1alpha1.PGCluster) (*appv1.StatefulSet, error) {
	if user == "" {
		return nil, errors.New("user name is empty")
	}

	var (
		pvc string
		err error
	)
	if user == "system" {
		pvcRes, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, "citus-data-pvc", metav1.GetOptions{})
		if err != nil {
			klog.Error("find pg pvc error, ", err)
			return nil, err
		}

		pvRes, err := client.CoreV1().PersistentVolumes().Get(ctx, pvcRes.Spec.VolumeName, metav1.GetOptions{})
		if err != nil {
			klog.Error("find pg pv error, ", err)
			return nil, err
		}

		pvc = pvRes.Spec.HostPath.Path

	} else {
		bflNamespace := "user-space-" + user
		pvc, err = utils.GetUserDBPVCName(ctx, client, bflNamespace)
		if err != nil {
			return nil, err
		}
	}

	sts := CitusStatefulset.DeepCopy()

	sts.Namespace = namespace

	for i, vol := range sts.Spec.Template.Spec.Volumes {
		if vol.Name == CitusVolumeName {
			sts.Spec.Template.Spec.Volumes[i].HostPath.Path = pvc + "/pgdata"
		}

		if vol.Name == CitusBackupVolumeName {
			if clusterDef.Spec.BackupStorage != "" {
				sts.Spec.Template.Spec.Volumes[i].HostPath.Path = clusterDef.Spec.BackupStorage
			} else {
				sts.Spec.Template.Spec.Volumes[i].HostPath.Path = pvc + "/pgbackup"
			}
		}
	}
	sts.Spec.Template.Spec.PriorityClassName = "system-cluster-critical"

	return sts, nil
}

func GetPGClusterDBUserSecretDefine(namespace, user, password string) (*corev1.Secret, error) {
	secret := CitusAdminSecret.DeepCopy()
	secret.Namespace = namespace

	secret.StringData["user"] = user

	var err error
	if password == "" {
		password, err = genPassword()
		if err != nil {
			return nil, err
		}
	}

	secret.StringData["password"] = password

	return secret, nil
}

func CreatePGClusterForUser(ctx context.Context,
	client *kubernetes.Clientset,
	aprclient *aprclientset.Clientset,
	lister listerv1alpha1.PGClusterLister,
	user string,
	clusterDef *v1alpha1.PGCluster) (*appv1.StatefulSet, error) {
	namespace := clusterDef.Namespace

	// create a new postgres cluster
	var err error
	// 1. create secret
	if clusterDef.Spec.AdminUser == "" {
		// create default admin user
		password, err := createSecretForPG(ctx, client, namespace)
		if err != nil {
			return nil, err
		}

		clusterDef.Spec.AdminUser = DefaultPGAdminUser
		clusterDef.Spec.Password = v1alpha1.PasswordVar{ValueFrom: &v1alpha1.PasswordVarSource{SecretKeyRef: password}}
	}

	// 2. create sts
	sts, err := createPGCluster(ctx, client, user, namespace, clusterDef)
	if err != nil {
		return nil, err
	}

	// 3. create headless service
	err = createPGClusterService(ctx, client, namespace)
	if err != nil {
		return nil, err
	}

	// citus cluster define in database scope
	// postpone cluster setup until app's databases are created

	// 4. set master
	// masterPod := PGClusterName + "-0"
	// err = waitForPodRunning(ctx, client, namespace, masterPod)
	// if err != nil {
	// 	klog.Error("waiting for pod running error, ", err, ", ", masterPod)
	// 	return err
	// }

	// 5. scale cluster if necessary, and add nodes to master, then rebalance
	return sts, nil
}

func genPassword() (string, error) {
	randPwd := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, randPwd); err != nil {
		klog.Error("generate random password error, ", err)
		return "", err
	}

	return taprutils.Hex(randPwd), nil
}

func createSecretForPG(ctx context.Context, client *kubernetes.Clientset, namespace string) (password *corev1.SecretKeySelector, err error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, CitusAdminSecretName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return
	}

	if apierrors.IsNotFound(err) {
		secret, err = GetPGClusterDBUserSecretDefine(namespace, DefaultPGAdminUser, "")
		if err != nil {
			klog.Error("get secret define error, ", err)
			return
		}

		_, err = client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return
		}

	}

	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
		Key:                  "password",
	}, nil
}

func createPGCluster(ctx context.Context, client *kubernetes.Clientset,
	user, namespace string, clusterDef *v1alpha1.PGCluster) (*appv1.StatefulSet, error) {
	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, PGClusterName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	if apierrors.IsNotFound(err) {
		sts, err = GetPGClusterDefineByUser(ctx, client, user, namespace, clusterDef)
		if err != nil {
			return nil, err
		}
		if len(clusterDef.Spec.CitusImage) > 0 {
			sts.Spec.Template.Spec.Containers[0].Image = clusterDef.Spec.CitusImage
		}

		if clusterDef.Spec.AdminUser != "" {
			for i, c := range sts.Spec.Template.Spec.Containers {
				if c.Name == "postgres" {
					for n, env := range c.Env {
						switch env.Name {
						case "POSTGRES_USER":
							sts.Spec.Template.Spec.Containers[i].Env[n].Value = clusterDef.Spec.AdminUser
							sts.Spec.Template.Spec.Containers[i].Env[n].ValueFrom = nil
						case "POSTGRES_PASSWORD":
							if clusterDef.Spec.Password.ValueFrom != nil {
								sts.Spec.Template.Spec.Containers[i].Env[n].ValueFrom = &corev1.EnvVarSource{
									SecretKeyRef: clusterDef.Spec.Password.ValueFrom.SecretKeyRef.DeepCopy(),
								}
								sts.Spec.Template.Spec.Containers[i].Env[n].Value = ""
							} else if clusterDef.Spec.Password.Value != "" {
								sts.Spec.Template.Spec.Containers[i].Env[n].Value = clusterDef.Spec.Password.Value
								sts.Spec.Template.Spec.Containers[i].Env[n].ValueFrom = nil
							}
						}
					} // end envs
				} // end container naame
			} // end for loop
		}

		sts, err = client.AppsV1().StatefulSets(namespace).Create(ctx, sts, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}

	}

	return sts, nil
}

func createPGClusterService(ctx context.Context, client *kubernetes.Clientset, namespace string) error {
	_, err := client.CoreV1().Services(namespace).Get(ctx, CitusHeadlessServiceName, metav1.GetOptions{})

	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if apierrors.IsNotFound(err) {
		svc := CitusHeadlessService.DeepCopy()
		svc.Namespace = namespace

		_, err := client.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func WaitForPodRunning(ctx context.Context, client *kubernetes.Clientset, namespace, podname string) (string, error) {
	// wait for pod to restart
	time.Sleep(2 * time.Second)

	var podIP string
	err := wait.PollImmediate(time.Second, 10*time.Minute, func() (bool, error) {
		pod, err := client.CoreV1().Pods(namespace).Get(ctx, podname, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}

		for _, c := range pod.Status.ContainerStatuses {
			if c.Name == "postgres" && !c.Ready {
				return false, nil
			}
		}

		// wait for db system available
		time.Sleep(5 * time.Second)

		klog.Info("pod is running, ", namespace, "/", podname, "  --  ", pod.Status.PodIP)

		podIP = pod.Status.PodIP
		return true, nil
	})

	return podIP, err
}

func ScalePGClusterNodes(ctx context.Context, client *kubernetes.Clientset,
	namespace string, replicas int32, admin, pwd string) (effected int32, err error) {
	cluster, err := client.AppsV1().StatefulSets(namespace).GetScale(ctx, PGClusterName, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	index := cluster.Spec.Replicas
	effected = replicas - cluster.Spec.Replicas
	if effected == 0 {
		return effected, nil
	}

	if effected < 0 {
		return effected, errors.New("scale down cluster not supported")
	}

	cluster.Spec.Replicas = replicas

	_, err = client.AppsV1().StatefulSets(namespace).UpdateScale(ctx, PGClusterName, cluster, metav1.UpdateOptions{})
	if err != nil {
		return 0, err
	}

	if effected > 0 {
		// update hba
		masterNode := PGClusterName + "-0.citus-headless." + namespace + ".svc.cluster.local"
		for index < cluster.Spec.Replicas {
			podName := PGClusterName + "-" + strconv.Itoa(int(index))
			klog.Info("update node hba config, ", podName)
			ip, err := WaitForPodRunning(ctx, client, namespace, podName)
			if err != nil {
				return 0, err
			}

			nodeClient, err := postgres.NewClientBuidler(admin, pwd,
				ip, postgres.PG_PORT).Build()
			if err != nil {
				klog.Error("connect to pg node error, ", err, ", ", admin, ", ", pwd)
				return 0, err
			}

			if err = func() error {
				defer nodeClient.Close()

				hba := PGNodeHBATrust +
					"\n" +
					"host all all " + masterNode + " trust" +
					"\n" +
					PGNodeHBAScram

				hbas := strings.Split(hba, "\n")

				_, err = nodeClient.DB.ExecContext(ctx, "drop table if exists hba")
				if err != nil {
					klog.Error("drop hba table error, ", err)
					return err
				}

				configFile := "/var/lib/postgresql/data/" + podName + "/pg_hba.conf"
				res, err := nodeClient.DB.QueryxContext(ctx, "select setting from pg_settings where name like '%hba%'")
				if err != nil {
					klog.Error("find hba config file error, ", err)
					return err
				}

				path := struct {
					Settings string `db:"setting"`
				}{}
				if res.Next() {
					err = res.StructScan(&path)
					res.Close()
					if err != nil {
						return err
					}

					configFile = path.Settings
				}

				if _, err = nodeClient.DB.ExecContext(ctx, "create table if not exists hba(lines text)"); err != nil {
					klog.Error("create hba table error, ", err)
					return err
				}

				tx, err := nodeClient.DB.Begin()
				if err != nil {
					return err
				}
				for _, h := range hbas {
					if _, err = nodeClient.DB.NamedExecContext(ctx, "insert into hba values(:line)", map[string]interface{}{
						"line": h,
					}); err != nil {
						tx.Rollback()
						return err
					}
				}
				if err = tx.Commit(); err != nil {
					tx.Rollback()
					return err
				}

				if _, err = nodeClient.DB.ExecContext(ctx, fmt.Sprintf("copy hba to '%s'", configFile)); err != nil {
					klog.Error("save pg_hba.conf error, ", err)
					return err
				}

				if _, err = nodeClient.DB.ExecContext(ctx, "SELECT pg_reload_conf()"); err != nil {
					klog.Error("reload pg config error, ", err)
					return err
				}

				return nil
			}(); err != nil {
				return 0, err
			}

			index += 1

		}
	}

	return
}

func MustUpdateClusterAdminUser(ctx context.Context, client *kubernetes.Clientset, namespace string,
	clusterSts *appv1.StatefulSet, clusterDef *v1alpha1.PGCluster) (need bool, oldadmin, oldpwd string, err error) {
	if clusterDef.Spec.AdminUser == "" {
		return false, "", "", nil
	}

	var userEnv, pwdEnv *corev1.EnvVar
	for _, c := range clusterSts.Spec.Template.Spec.Containers {
		if c.Name == "postgres" {
			for _, env := range c.Env {
				switch env.Name {
				case "POSTGRES_USER":
					userEnv = env.DeepCopy()
				case "POSTGRES_PASSWORD":
					pwdEnv = env.DeepCopy()
				}
			}
		}
	}

	if userEnv == nil || (userEnv.Value == "" && userEnv.ValueFrom == nil) {
		return false, "", "", errors.New("invalid pg cluster db definition, user is empty")
	}

	if pwdEnv == nil || (pwdEnv.Value == "" && pwdEnv.ValueFrom == nil) {
		return false, "", "", errors.New("invalid pg cluster db definition, password is empty")
	}

	getValueFromSecret := func(secretSel *corev1.SecretKeySelector) (string, error) {
		secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretSel.Name, metav1.GetOptions{})
		if err != nil {
			klog.Error("get pg cluster value secret ref error, ", err, ", ", secretSel.Name)
			return "", err
		}

		return string(secret.Data[secretSel.Key]), nil
	}

	// if admin user changed or not
	if userEnv.Value != "" && userEnv.Value != clusterDef.Spec.AdminUser {
		pwd := pwdEnv.Value
		if pwd == "" {
			pwd, err = getValueFromSecret(pwdEnv.ValueFrom.SecretKeyRef)
			if err != nil {
				return false, "", "", err
			}
		}
		return true, userEnv.Value, pwd, nil
	}

	if userEnv.Value == "" {
		user, err := getValueFromSecret(userEnv.ValueFrom.SecretKeyRef)
		if err != nil {
			return false, "", "", err
		}

		if user != clusterDef.Spec.AdminUser {
			pwd, err := getValueFromSecret(pwdEnv.ValueFrom.SecretKeyRef)
			if err != nil {
				return false, "", "", err
			}

			return true, user, pwd, nil
		}
	}

	// if password changed or not
	// get clusterdef password
	newPwd, err := GetPGClusterDefinedPassword(ctx, client, namespace, clusterDef)
	if err != nil {
		klog.Error("get pg cluster defined password error, ", err)
		return false, "", "", err
	}

	if pwdEnv.Value != "" && pwdEnv.Value != newPwd {
		user := userEnv.Value
		if user == "" {
			user, err = getValueFromSecret(userEnv.ValueFrom.SecretKeyRef)
			if err != nil {
				return false, "", "", err
			}
		}

		return true, user, pwdEnv.Value, nil
	}

	if pwdEnv.Value == "" {
		oldpwd, err := getValueFromSecret(pwdEnv.ValueFrom.SecretKeyRef)
		if err != nil {
			return false, "", "", err
		}

		if oldpwd != newPwd {
			user := userEnv.Value
			if user == "" {
				user, err = getValueFromSecret(userEnv.ValueFrom.SecretKeyRef)
				if err != nil {
					return false, "", "", err
				}
			}
			return true, user, oldpwd, nil
		}
	}

	return false, "", "", nil
}

func GetPGClusterDefinedPassword(ctx context.Context, client *kubernetes.Clientset, namespace string,
	clusterDef *v1alpha1.PGCluster) (string, error) {
	return clusterDef.Spec.Password.GetVarValue(ctx, client, namespace)
}

func GetPGClusterAdminUserAndPassword(ctx context.Context, aprclient *aprclientset.Clientset,
	client *kubernetes.Clientset, namespace string) (user, pwd string, err error) {
	// FIXME: pgcluster.name
	cluster, err := aprclient.AprV1alpha1().PGClusters(namespace).Get(ctx, PGClusterName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	var secret *corev1.Secret
	if cluster.Spec.AdminUser == "" {
		// get default user
		secret, err = client.CoreV1().Secrets(namespace).Get(ctx, CitusAdminSecretName, metav1.GetOptions{})
		if err != nil {
			return
		}

		cluster.Spec.Password = v1alpha1.PasswordVar{
			Value: string(secret.Data["password"]),
		}
		cluster.Spec.AdminUser = string(secret.Data["user"])
	}

	pwd, err = GetPGClusterDefinedPassword(ctx, client, namespace, cluster)
	if err != nil {
		return "", "", err
	}

	return cluster.Spec.AdminUser, pwd, nil
}

func ForceCreateNewPGClusterBackup(ctx context.Context, client *aprclientset.Clientset, backup *v1alpha1.PGClusterBackup) error {
	currentBackup, err := client.AprV1alpha1().PGClusterBackups(backup.Namespace).Get(ctx, backup.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error("get prev pg backup cr error, ", err)
		return err
	}

	if err == nil {
		if currentBackup.Status.State == v1alpha1.BackupStateRunning {
			klog.Error("prev pg backup job is still running")
			return errors.New("duplicate pg backup job")
		}

		klog.Info("remove prev pg backup, ", backup.Name, ", ", backup.Namespace)
		err = client.AprV1alpha1().PGClusterBackups(backup.Namespace).Delete(ctx, backup.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.Error("delete prev pg backup error, ", err)
			return err
		}

		// waiting for deletion complete
		wait.PollImmediateWithContext(ctx, 1*time.Second, time.Hour, func(ctx context.Context) (done bool, err error) {
			_, err = client.AprV1alpha1().PGClusterBackups(backup.Namespace).Get(ctx, backup.Name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					klog.Info("prev pg backup removed")
					return true, nil
				}

				return false, err
			}

			return false, nil
		})

	}

	klog.Info("create a new pg backup, ", backup.Name, ", ", backup.Namespace)
	backup.Status.State = v1alpha1.BackupStateNew
	_, err = client.AprV1alpha1().PGClusterBackups(backup.Namespace).Create(ctx, backup, metav1.CreateOptions{})
	if err != nil {
		klog.Error("create pg backup error, ", err, ", ", backup.Name, ", ", backup.Namespace)
		return err
	}

	return err

}

func WaitForAllBackupComplete(ctx context.Context, client *aprclientset.Clientset) error {
	return wait.PollWithContext(ctx, 5*time.Second, time.Hour,
		func(context.Context) (done bool, err error) {
			var errs []error
			incompleted := false
			backups, err := client.AprV1alpha1().PGClusterBackups("").List(ctx, metav1.ListOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}

				klog.Error("list pg backup error, ", err)
				return false, err
			}

			for _, backup := range backups.Items {
				switch backup.Status.State {
				case v1alpha1.BackupStateRunning, v1alpha1.BackupStateWaiting:
					incompleted = true
				case v1alpha1.BackupStateRejected, v1alpha1.BackupStateError:
					errs = append(errs, fmt.Errorf("backup pg failed, %s, %s, %s", backup.Status.Error, backup.Name, backup.Namespace))
				}
			}

			if len(errs) > 0 {
				err = utilerrors.NewAggregate(errs)
			}

			done = !incompleted

			return
		})
}

func ForceCreateNewPGClusterRestore(ctx context.Context, client *aprclientset.Clientset,
	restore *v1alpha1.PGClusterRestore) error {
	currentRestore, err := client.AprV1alpha1().PGClusterRestores(restore.Namespace).Get(ctx, restore.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error("get prev pg restore cr error, ", err)
		return err
	}

	if err == nil {
		if currentRestore.Status.State == v1alpha1.RestoreStateRunning {
			klog.Error("prev pg restore job is still running")
			return errors.New("duplicate restore job")
		}

		klog.Info("remove prev pg restore, ", ", ", restore.Name, ", ", restore.Namespace)
		err = client.AprV1alpha1().PGClusterRestores(restore.Namespace).Delete(ctx, restore.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.Error("delete prev pg restore error, ", err)
			return err
		}
	}

	klog.Info("create a new restore, ", ", ", restore.Name, ", ", restore.Namespace)
	restore.Status.State = v1alpha1.RestoreStateNew
	_, err = client.AprV1alpha1().PGClusterRestores(restore.Namespace).Create(ctx, restore, metav1.CreateOptions{})
	if err != nil {
		klog.Error("create pg restore error, ", err, ", ", restore.Name, ", ", restore.Namespace)
	}

	return err
}

func WaitForAllRestoreComplete(ctx context.Context, client *aprclientset.Clientset) error {
	return wait.PollWithContext(ctx, 5*time.Second, time.Hour,
		func(context.Context) (done bool, err error) {
			var errs []error
			incompleted := false
			restores, err := client.AprV1alpha1().PGClusterRestores("").List(ctx, metav1.ListOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}

				klog.Error("list pg restore error, ", err)
				return false, err
			}

			for _, restore := range restores.Items {
				switch restore.Status.State {
				case v1alpha1.RestoreStateRunning, v1alpha1.RestoreStateWaiting:
					incompleted = true
				case v1alpha1.RestoreStateRejected, v1alpha1.RestoreStateError:
					errs = append(errs, fmt.Errorf("restore pg failed, %s, %s, %s", restore.Status.Error, restore.Name, restore.Namespace))
				}
			}

			if len(errs) > 0 {
				err = utilerrors.NewAggregate(errs)
			}

			done = !incompleted

			return
		})
}

func GetDatabaseName(namespace, db string) string {
	return strings.ReplaceAll(fmt.Sprintf("%s_%s", namespace, db), "-", "_")
}
