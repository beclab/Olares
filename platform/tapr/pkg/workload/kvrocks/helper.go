package kvrocks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"
	"bytetrade.io/web3os/tapr/pkg/workload/utils"
	redis "github.com/go-redis/redis/v8"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

func GetKVRocksDefineByUser(ctx context.Context, client *kubernetes.Clientset,
	user, namespace string, kvrocksDef *v1alpha1.RedixCluster) (*appv1.StatefulSet, error) {
	if user == "" {
		return nil, errors.New("user name is empty")
	}

	if kvrocksDef.Spec.Type != v1alpha1.KVRocks {
		return nil, errors.New("wrong redix cluster type")
	}

	var (
		pvc string
		err error
	)
	if user == "system" {
		pvcRes, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, "kvrocks-data-pvc", metav1.GetOptions{})
		if err != nil {
			klog.Error("find kvrocks pvc error, ", err)
			return nil, err
		}

		pvRes, err := client.CoreV1().PersistentVolumes().Get(ctx, pvcRes.Spec.VolumeName, metav1.GetOptions{})
		if err != nil {
			klog.Error("find kvrocks pv error, ", err)
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

	sts := KVRocksStatefulSet.DeepCopy()

	sts.Namespace = namespace
	sts.Name = kvrocksDef.Name

	for i, c := range sts.Spec.Template.Spec.InitContainers {
		if c.Name == "init-kvrocks-cfg" {
			ptrC := &sts.Spec.Template.Spec.InitContainers[i]
			if kvrocksDef.Spec.KVRocks.Image != "" {
				ptrC.Image = kvrocksDef.Spec.KVRocks.Image
			}

			if kvrocksDef.Spec.KVRocks.ImagePullPolicy != "" {
				ptrC.ImagePullPolicy = kvrocksDef.Spec.KVRocks.ImagePullPolicy
			}
		}
	}

	for i, c := range sts.Spec.Template.Spec.Containers {
		if c.Name == "kvrocks" {
			ptrC := &sts.Spec.Template.Spec.Containers[i]
			if kvrocksDef.Spec.KVRocks.Image != "" {
				ptrC.Image = kvrocksDef.Spec.KVRocks.Image
			}

			if kvrocksDef.Spec.KVRocks.ImagePullPolicy != "" {
				ptrC.ImagePullPolicy = kvrocksDef.Spec.KVRocks.ImagePullPolicy
			}

			if kvrocksDef.Spec.KVRocks.Resources != nil {
				ptrC.Resources = *kvrocksDef.Spec.KVRocks.Resources
			}

			break
		}
	}

	for i, vol := range sts.Spec.Template.Spec.Volumes {
		if vol.Name == KVRocksVolumeName {
			sts.Spec.Template.Spec.Volumes[i].HostPath.Path = pvc + "/kvrdata"
		}

		if vol.Name == KVRocksBackupVolumeName {
			if kvrocksDef.Spec.KVRocks.BackupStorage != nil && *kvrocksDef.Spec.KVRocks.BackupStorage != "" {
				sts.Spec.Template.Spec.Volumes[i].HostPath.Path = *kvrocksDef.Spec.KVRocks.BackupStorage
			} else {
				sts.Spec.Template.Spec.Volumes[i].HostPath.Path = pvc + "/kvrbackup"
			}
		}
	}

	// apply config via command
	for k, v := range kvrocksDef.Spec.KVRocks.KVRocksConfig {
		sts.Spec.Template.Spec.Containers[0].Command =
			append(sts.Spec.Template.Spec.Containers[0].Command, []string{"--" + k, v}...)
	}

	// set admin password
	password, err := kvrocksDef.Spec.KVRocks.Password.GetVarValue(ctx, client, kvrocksDef.Namespace)
	if err != nil {
		return nil, err
	}

	if password != "" {
		sts.Spec.Template.Spec.Containers[0].Command =
			append(sts.Spec.Template.Spec.Containers[0].Command, []string{"--requirepass", password}...)
	}

	return sts, nil

}

func createKVRocks(ctx context.Context, client *kubernetes.Clientset,
	clusterDef *v1alpha1.RedixCluster) (*appv1.StatefulSet, error) {
	sts, err := client.AppsV1().StatefulSets(clusterDef.Namespace).Get(ctx, clusterDef.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	if apierrors.IsNotFound(err) {
		sts, err = GetKVRocksDefineByUser(ctx, client,
			clusterDef.Spec.KVRocks.Owner, clusterDef.Namespace, clusterDef)
		if err != nil {
			return nil, err
		}

		sts, err = client.AppsV1().StatefulSets(clusterDef.Namespace).Create(ctx, sts, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}

	}

	return sts, nil
}

func createKVRocksService(ctx context.Context, client *kubernetes.Clientset, clusterName, namespace string) error {
	svcName := getServiceName(clusterName)
	_, err := client.CoreV1().Services(namespace).Get(ctx, svcName, metav1.GetOptions{})

	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if apierrors.IsNotFound(err) {
		svc := KVRocksService.DeepCopy()
		svc.Namespace = namespace
		svc.Name = svcName

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
	err := wait.PollWithContext(ctx, time.Second, 10*time.Minute, func(ctx context.Context) (done bool, err error) {
		pod, err := client.CoreV1().Pods(namespace).Get(ctx, podname, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}

		for _, c := range pod.Status.ContainerStatuses {
			if c.Name == "kvrocks" && !c.Ready {
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

func CreateOrUpdateKVRocks(ctx context.Context,
	client *kubernetes.Clientset,
	clusterDef *v1alpha1.RedixCluster, isUpdated bool) (*appv1.StatefulSet, error) {
	var retSts *appv1.StatefulSet
	oldSts, err := client.AppsV1().StatefulSets(clusterDef.Namespace).Get(ctx, clusterDef.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	if apierrors.IsNotFound(err) {
		retSts, err = createKVRocks(ctx, client, clusterDef)
		if err != nil {
			klog.Error("create kvrocks error, ", err)
			return nil, err
		}
	}

	if oldSts != nil {
		oldTemplateVersion := oldSts.Spec.Template.Labels["pod-template-version"]
		newTemplateVersion := KVRocksStatefulSet.Spec.Template.Labels["pod-template-version"]
		// the kvrocks is existing, set the return sts to old one first
		// if no need to update, just return the old one
		retSts = oldSts

		if isUpdated || oldTemplateVersion != newTemplateVersion {
			klog.Info("kvrocks pod template version changed, need to update sts")
			// update sts
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				sts, err := client.AppsV1().StatefulSets(clusterDef.Namespace).Get(ctx, clusterDef.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}

				updateSts, err := GetKVRocksDefineByUser(ctx, client,
					clusterDef.Spec.KVRocks.Owner, clusterDef.Namespace, clusterDef)
				if err != nil {
					return err
				}

				sts.Spec.Template = updateSts.Spec.Template

				retSts, err = client.AppsV1().StatefulSets(clusterDef.Namespace).Update(ctx, sts, metav1.UpdateOptions{})
				return err
			})

			if err != nil {
				klog.Error("update kvrocks error, ", err)
				return nil, err
			}
		}
	}

	// create or update service
	klog.Info("creating or update kvrocks service")
	err = createKVRocksService(ctx, client, clusterDef.Name, clusterDef.Namespace)
	if err != nil {
		klog.Error("create or update kvrocks service error, ", err)
		return nil, err
	}

	return retSts, nil
}

func DeleteKVRocks(ctx context.Context,
	client *kubernetes.Clientset,
	clusterDef *v1alpha1.RedixCluster) error {

	// delete kvrocks service
	klog.Info("delete kvrocks service")
	svcName := getServiceName(clusterDef.Name)
	err := client.CoreV1().Services(clusterDef.Namespace).Delete(ctx, svcName, metav1.DeleteOptions{})

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error("delete kvrocks service error, ", err, ", ", svcName)
		return err
	}

	// delete kvrocks sts
	klog.Info("delete kvrocks sts")
	err = client.AppsV1().StatefulSets(clusterDef.Namespace).Delete(ctx, clusterDef.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Error("delete kvrock sts error, ", err)
		return err
	}

	return nil
}

func getServiceName(_ string) string {
	// return fmt.Sprintf("%s-svc", clusterName)

	return "redis-cluster-proxy"
}

func BackupKVRocks(ctx context.Context,
	client *kubernetes.Clientset,
	clusterDef *v1alpha1.RedixCluster,
	backup *v1alpha1.KVRocksBackup) error {

	if backup.Spec.ClusterName != clusterDef.Name {
		return fmt.Errorf("backup request is mismatch, %s , %s", backup.Spec.ClusterName, clusterDef.Name)
	}

	cli, err := GetKVRocksClient(ctx, client, clusterDef)
	if err != nil {
		return err
	}

	defer cli.Close()

	backupDir := KVRocksBackupDir + "/backup"

	klog.Info("remove prev kvrocks backup files")
	fs, err := os.Stat(backupDir)
	if err != nil && !os.IsNotExist(err) {
		klog.Error("check kvrocks prev backup files error, ", err)
		return err
	}

	if err == nil {
		if fs.IsDir() {
			err = os.RemoveAll(backupDir)
		} else {
			err = os.Remove(backupDir)
		}

		if err != nil {
			klog.Error("remove kvrocks prev backup dir error, ", err)
			return err
		}
	}

	klog.Info("send backup command to kvrocks")
	if err = cli.BgSave(ctx).Err(); err != nil {
		return fmt.Errorf("kvrocks backup error: %v", err)
	}

	klog.Info("waiting for kvrocks backup complete")
	err = wait.PollWithContext(ctx, time.Second, 30*time.Minute, func(ctx context.Context) (done bool, err error) {
		_, err = os.Stat(backupDir)
		if err != nil && !os.IsNotExist(err) {
			klog.Error("check kvrocks current backup files error, ", err)
			return false, err
		}

		if err == nil {
			return true, nil
		}

		return false, nil

	})

	if err != nil {
		return err
	}

	backup.Status.BackupPath = backupDir
	return nil
}

func RestoreKVRocks(ctx context.Context,
	client *kubernetes.Clientset,
	cluster *v1alpha1.RedixCluster,
	restore *v1alpha1.KVRocksRestore) error {
	backupDir := KVRocksBackupDir + "/backup"
	fs, err := os.Stat(backupDir)
	if err != nil {
		klog.Error("check kvrocks backup files error, ", err)
		return err
	}

	if !fs.IsDir() {
		err = errors.New("invalid backup files")

		klog.Error("check kvrocks backup files error, ", err)
		return err
	}

	// scale sts to 0,
	klog.Info("close kvrocks instance to restore")
	sts, err := client.AppsV1().StatefulSets(cluster.Namespace).GetScale(ctx, cluster.Name, metav1.GetOptions{})
	if err != nil {
		klog.Error("find kvrocks workload error, ", err)

		return err
	}

	sts.Spec.Replicas = 0
	sts, err = client.AppsV1().StatefulSets(sts.Namespace).UpdateScale(ctx, sts.Name, sts, metav1.UpdateOptions{})
	if err != nil {
		klog.Error("scale kvrocks error, ", err)
		return err
	}

	// waiting for kvrocks pods close
	err = wait.PollWithContext(ctx, time.Second, 5*time.Minute, func(ctx context.Context) (done bool, err error) {
		pods, err := client.CoreV1().Pods(sts.Namespace).List(ctx, metav1.ListOptions{LabelSelector: "app=kvrocks"})

		if err != nil {
			klog.Error("list kvrocks pods error, ", err)
			return false, nil
		}

		return len(pods.Items) == 0, nil
	})

	if err != nil {
		return err
	}

	// move backup files to db
	klog.Info("move backup files to db")
	dbDir := KVRocksDataDir + "/db"
	_, err = os.Stat(dbDir)
	if err != nil {
		klog.Error("check kvrocks db files error, ", err)
		return err
	}

	err = os.RemoveAll(dbDir)
	if err != nil {
		klog.Error("remove kvrocks db files error, ", err)
		return err
	}

	err = os.Rename(backupDir, dbDir)
	if err != nil {
		klog.Error("move kvrocks backup files to db error, ", err)
		return err
	}

	// restart kvrocks
	klog.Info("restart kvrocks instance")
	sts, err = client.AppsV1().StatefulSets(cluster.Namespace).GetScale(ctx, cluster.Name, metav1.GetOptions{})
	if err != nil {
		klog.Error("find kvrocks workload error, ", err)

		return err
	}

	sts.Spec.Replicas = 1
	sts, err = client.AppsV1().StatefulSets(sts.Namespace).UpdateScale(ctx, sts.Name, sts, metav1.UpdateOptions{})
	if err != nil {
		klog.Error("scale kvrocks error, ", err)
		return err
	}

	// waiting for kvrocks pods restart
	err = wait.PollWithContext(ctx, time.Second, 5*time.Minute, func(ctx context.Context) (done bool, err error) {
		pods, err := client.CoreV1().Pods(sts.Namespace).List(ctx, metav1.ListOptions{LabelSelector: "app=kvrocks"})

		if err != nil {
			klog.Error("list kvrocks pods error, ", err)
			return false, nil
		}

		return len(pods.Items) > 0, nil
	})

	if err != nil {
		klog.Error("restart kvrocks pods error, ", err)
		return err
	}

	return nil
}

func ForceCreateNewKVRocksBackup(ctx context.Context, client *aprclientset.Clientset, backup *v1alpha1.KVRocksBackup) error {
	currentBackup, err := client.AprV1alpha1().KVRocksBackups(backup.Namespace).Get(ctx, backup.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error("get prev kvrocks backup cr error, ", err)
		return err
	}

	if err == nil {
		if currentBackup.Status.State == v1alpha1.BackupStateRunning {
			klog.Error("prev kvrocks backup job is still running")
			return errors.New("duplicate kvrocks backup job")
		}

		klog.Info("remove prev kvrocks backup, ", backup.Name, ", ", backup.Namespace)
		err = client.AprV1alpha1().KVRocksBackups(backup.Namespace).Delete(ctx, backup.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.Error("delete prev kvrocks backup error, ", err)
			return err
		}

		// waiting for deletion complete
		wait.PollImmediateWithContext(ctx, 1*time.Second, time.Hour, func(ctx context.Context) (done bool, err error) {
			_, err = client.AprV1alpha1().KVRocksBackups(backup.Namespace).Get(ctx, backup.Name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					klog.Info("prev kvrocks backup removed")
					return true, nil
				}

				return false, err
			}

			return false, nil
		})
	} // end if backup found

	klog.Info("create a new kvrocks backup, ", backup.Name, ", ", backup.Namespace)
	backup.Status.State = v1alpha1.BackupStateNew
	_, err = client.AprV1alpha1().KVRocksBackups(backup.Namespace).Create(ctx, backup, metav1.CreateOptions{})
	if err != nil {
		klog.Error("create kvrocks backup error, ", err, ", ", backup.Name, ", ", backup.Namespace)
		return err
	}

	return err

}

func ForceCreateNewKVRocksRestore(ctx context.Context, client *aprclientset.Clientset,
	restore *v1alpha1.KVRocksRestore) error {
	currentRestore, err := client.AprV1alpha1().KVRocksRestores(restore.Namespace).Get(ctx, restore.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error("get prev kvrocks restore cr error, ", err)
		return err
	}

	if err == nil {
		if currentRestore.Status.State == v1alpha1.RestoreStateRunning {
			klog.Error("prev kvrocks restore job is still running")
			return errors.New("duplicate restore job")
		}

		klog.Info("remove prev kvrocks restore, ", ", ", restore.Name, ", ", restore.Namespace)
		err = client.AprV1alpha1().KVRocksRestores(restore.Namespace).Delete(ctx, restore.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.Error("delete prev kvrocks restore error, ", err)
			return err
		}
	}

	klog.Info("create a new kvrocks restore, ", ", ", restore.Name, ", ", restore.Namespace)
	restore.Status.State = v1alpha1.RestoreStateNew
	_, err = client.AprV1alpha1().KVRocksRestores(restore.Namespace).Create(ctx, restore, metav1.CreateOptions{})
	if err != nil {
		klog.Error("create kvrocks restore error, ", err, ", ", restore.Name, ", ", restore.Namespace)
	}

	return err
}

func WaitForAllBackupComplete(ctx context.Context, client *aprclientset.Clientset) error {
	return wait.PollWithContext(ctx, 5*time.Second, time.Hour,
		func(context.Context) (done bool, err error) {
			var errs []error
			incompleted := false
			backups, err := client.AprV1alpha1().KVRocksBackups("").List(ctx, metav1.ListOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}

				klog.Error("list kvrocks backup error, ", err)
				return false, err
			}

			for _, backup := range backups.Items {
				switch backup.Status.State {
				case v1alpha1.BackupStateRunning, v1alpha1.BackupStateWaiting:
					incompleted = true
				case v1alpha1.BackupStateRejected, v1alpha1.BackupStateError:
					errs = append(errs, fmt.Errorf("backup kvrocks failed, %s, %s, %s", backup.Status.Error, backup.Name, backup.Namespace))
				}
			}

			if len(errs) > 0 {
				err = utilerrors.NewAggregate(errs)
			}

			done = !incompleted

			return
		})
}

func WaitForAllRestoreComplete(ctx context.Context, client *aprclientset.Clientset) error {
	return wait.PollWithContext(ctx, 5*time.Second, time.Hour,
		func(context.Context) (done bool, err error) {
			var errs []error
			incompleted := false
			restores, err := client.AprV1alpha1().KVRocksRestores("").List(ctx, metav1.ListOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}

				klog.Error("list kvrocks restore error, ", err)
				return false, err
			}

			for _, restore := range restores.Items {
				switch restore.Status.State {
				case v1alpha1.RestoreStateRunning, v1alpha1.RestoreStateWaiting:
					incompleted = true
				case v1alpha1.RestoreStateRejected, v1alpha1.RestoreStateError:
					errs = append(errs, fmt.Errorf("restore kvrocks failed, %s, %s, %s", restore.Status.Error, restore.Name, restore.Namespace))
				}
			}

			if len(errs) > 0 {
				err = utilerrors.NewAggregate(errs)
			}

			done = !incompleted

			return
		})
}

// kvrocks commands
type kvrClient struct {
	*redis.Client
}

type Namespace struct {
	Name  string
	Token string
}

func GetKVRocksClient(ctx context.Context, client *kubernetes.Clientset,
	clusterDef *v1alpha1.RedixCluster) (*kvrClient, error) {
	klog.Info("find kvrocks password")
	password, err := clusterDef.Spec.KVRocks.Password.GetVarValue(ctx, client, clusterDef.Namespace)
	if err != nil {
		return nil, err
	}

	svcName := getServiceName(clusterDef.Name)
	svcPort := KVRocksService.Spec.Ports[0].Port
	klog.Info("find kvrocks service, ", svcName, ":", svcPort)

	cli := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", svcName, svcPort),
		Password: password,
		// other options with default
	})

	if err = cli.Ping(ctx).Err(); err != nil {
		cli.Close()
		return nil, fmt.Errorf("kvrocks connection error: %v", err)
	}

	return &kvrClient{cli}, nil
}

func (cli *kvrClient) Namespace(ctx context.Context, arg ...interface{}) *redis.Cmd {
	return cli.Do(ctx, append([]interface{}{"namespace"}, arg...)...)
}

func (cli *kvrClient) AddNamespace(ctx context.Context, namespace, token string) error {
	if err := cli.Namespace(ctx, "add", namespace, token).Err(); err != nil {
		return err
	}

	return cli.ConfigRewrite(ctx).Err()
}

func (cli *kvrClient) GetNamespace(ctx context.Context, namespace string) (*Namespace, error) {
	cmd := cli.Namespace(ctx, "get", namespace)
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil { // not found
			return nil, nil
		}
		return nil, cmd.Err()
	}

	return &Namespace{namespace, cmd.String()}, nil
}

func (cli *kvrClient) ListNamespace(ctx context.Context) ([]*Namespace, error) {
	cmd := cli.Namespace(ctx, "get", "*")
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil { // not found
			return nil, nil
		}
		return nil, cmd.Err()
	}

	var ns []*Namespace
	slice, err := cmd.StringSlice()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(slice); i += 2 {
		// default namespace token is admin password
		if slice[i] != "__namespace" {
			ns = append(ns, &Namespace{slice[i], slice[i+1]})
		}
	}

	return ns, nil
}

func (cli *kvrClient) UpdateNamespace(ctx context.Context, namespace, token string) error {
	if err := cli.Namespace(ctx, "set", namespace, token).Err(); err != nil {
		return err
	}

	return cli.ConfigRewrite(ctx).Err()

}

func (cli *kvrClient) DeleteNamespace(ctx context.Context, namespace string) error {
	if err := cli.Namespace(ctx, "del", namespace).Err(); err != nil {
		return err
	}

	return cli.ConfigRewrite(ctx).Err()
}
