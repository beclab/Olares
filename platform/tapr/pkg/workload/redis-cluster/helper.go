package rediscluster

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"
	"bytetrade.io/web3os/tapr/pkg/generated/listers/apr/v1alpha1"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func GetUserAuthSections(passwords [][]string) string {
	section := AuthSectionBegin
	for _, p := range passwords {
		section += fmt.Sprintf(AuthUserBegin, p[0])
		section += AuthUserModeWrite
		if len(p) > 2 && p[2] != "" {
			section += fmt.Sprintf(AuthUserNamespace, GetDatabaseName(p[1], p[2]))
		}
		section += AuthUserEnd
	}
	section += AuthSectionEnd

	return section
}

func UpdateProxyConfig(ctx context.Context, client *kubernetes.Clientset,
	aprclient *aprclientset.Clientset, dynamicClient *dynamic.DynamicClient) error {
	mreqs, err := aprclient.AprV1alpha1().MiddlewareRequests("").List(ctx, metav1.ListOptions{})

	if err != nil {
		klog.Error("list middleware request error, ", err)
		return err
	}

	klog.Info("list redis middleware request")
	var passwords [][]string
	for _, mr := range mreqs.Items {
		if mr.Spec.Middleware == aprv1.TypeRedis {
			pwd, err := mr.Spec.Redis.Password.GetVarValue(ctx, client, mr.Namespace)
			if err != nil {
				klog.Error("get middleware request password error, ", err, ", ", mr.Name, ", ", mr.Namespace)
				return err
			}

			pwdAndDb := []string{pwd, mr.Spec.AppNamespace}
			if len(mr.Spec.Redis.Namespace) > 0 {
				pwdAndDb = append(pwdAndDb, mr.Spec.Redis.Namespace)
			}

			passwords = append(passwords, pwdAndDb)
		}
	}

	klog.Info("find predixy config map")
	drcs, err := ListRedisClusters(ctx, dynamicClient, "")
	if err != nil {
		klog.Error("find redis cluster error, ", err)
		return err
	}

	if len(drcs) == 0 {
		err = errors.New("redis cluster not found")
		klog.Error("find redis cluster error, ", err)
		return err
	}

	klog.Info("add admin password")
	pwd, err := FindRedisClusterPassword(ctx, client, drcs[0].Namespace)
	if err != nil {
		klog.Error("cannot get redis cluster admin password")
		return err
	}

	passwords = append(passwords, []string{pwd})

	if len(passwords) > 0 {

		namespace := drcs[0].Namespace // cluster namespace, only one cluster supported
		cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, RedisclusterProxyConfig, metav1.GetOptions{})
		if err != nil {
			klog.Error("get predixy config map error, ", err)
			return err
		}

		config := cm.Data["predixy.conf"]

		configNew := regexp.MustCompile(AuthSectionExpr).ReplaceAllStringFunc(config, func(string) string {
			return AuthSectionExprPrev + "\n" + GetUserAuthSections(passwords) + "\n" + AuthSectionExprNext
		})

		cm.Data["predixy.conf"] = configNew
		_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
		if err != nil {
			klog.Error("update config map error, ", err)
			return err
		}

		klog.Info("restart redis proxy")
		proxy, err := client.AppsV1().Deployments(namespace).Get(ctx, RedisProxy, metav1.GetOptions{})
		if err != nil {
			klog.Error("find redis proxy error, ", err)
			return err
		}

		proxy.Spec.Template.ObjectMeta.Labels["timestamp"] = strconv.Itoa(int(time.Now().Unix()))
		_, err = client.AppsV1().Deployments(namespace).Update(ctx, proxy, metav1.UpdateOptions{})
		if err != nil {
			klog.Error("restart redis proxy error, ", err)
			return err
		}
	}

	return nil
}

func WaitForInitializeComplete(ctx context.Context, dynamicClient *dynamic.DynamicClient, k8sClient *kubernetes.Clientset) error {
	return wait.PollWithContext(ctx, 5*time.Second, time.Hour,
		func(context.Context) (done bool, err error) {
			incompleted := false

			redisClusterStruct, err := dynamicClient.Resource(RedisClusterClassGVR).Namespace(REDISCLUSTER_NAMESPACE).Get(ctx, REDISCLUSTER_NAME, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}

				klog.Error("get redis cluster error, ", err)
				return false, err
			}

			var redisCluster = DistributedRedisCluster{}
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(redisClusterStruct.Object, &redisCluster); err != nil {
				klog.Error("convert redis cluster crd error, ", err, ", ", redisClusterStruct.GetName(), ", ", redisClusterStruct.GetNamespace())
				return false, nil
			}

			if &redisCluster.Status == nil || redisCluster.Status.Nodes == nil || len(redisCluster.Status.Nodes) == 0 {
				klog.Error("redis cluster not ready, nodes is 0, waiting")
				return false, nil
			}

			if redisCluster.Status.Status != ClusterStatusOK && redisCluster.Status.Reason != "OK" {
				klog.Errorf("redis cluster not ready, status %s %s, waiting", redisCluster.Status.Status, redisCluster.Status.Reason)
				return false, nil
			}

			l, err := k8sClient.AppsV1().StatefulSets(REDISCLUSTER_NAMESPACE).List(ctx, metav1.ListOptions{LabelSelector: "managed-by=redis-cluster-operator"})
			if err != nil {
				return false, nil
			}

			if l.Items == nil || len(l.Items) == 0 {
				return false, nil
			}

			done = !incompleted

			return
		})
}

func UpdateRedisClusterDeployment(ctx context.Context, k8sClient *kubernetes.Clientset) error {
	klog.Info("restart redis proxy")
	proxy, err := k8sClient.AppsV1().Deployments(REDISCLUSTER_NAMESPACE).Get(ctx, RedisProxy, metav1.GetOptions{})
	if err != nil {
		klog.Error("find redis proxy error, ", err)
		return err
	}

	proxy.Spec.Template.ObjectMeta.Labels["timestamp"] = strconv.Itoa(int(time.Now().Unix()))
	_, err = k8sClient.AppsV1().Deployments(REDISCLUSTER_NAMESPACE).Update(ctx, proxy, metav1.UpdateOptions{})
	if err != nil {
		klog.Error("restart redis proxy error, ", err)
		return err
	}

	return nil
}

func ListKvRocks(reidsLister v1alpha1.RedixClusterLister) ([]*aprv1.RedixCluster, error) {
	list, err := reidsLister.List(labels.Everything())
	if err != nil {
		klog.Error("list kvrocks error, ", err)
		return nil, err
	}

	return list, nil
}

func ListRedisClusters(ctx context.Context, dynamicClient *dynamic.DynamicClient, namespace string) ([]*DistributedRedisCluster, error) {
	clusters, err := dynamicClient.Resource(RedisClusterClassGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list redis cluster crd error, ", err)
		return nil, err
	}

	var drcList []*DistributedRedisCluster

	for _, cluster := range clusters.Items {
		drc := DistributedRedisCluster{}
		if err = runtime.DefaultUnstructuredConverter.FromUnstructured(cluster.Object, &drc); err != nil {
			klog.Error("convert redis cluster crd error, ", err, ", ", drc)
			return nil, err
		}

		drcList = append(drcList, &drc)
	}

	return drcList, nil
}

func FindRedisClusterPassword(ctx context.Context, k8sClient *kubernetes.Clientset, namespace string) (string, error) {
	secret, err := k8sClient.CoreV1().Secrets(namespace).Get(ctx, "redix-cluster-admin", metav1.GetOptions{})
	if err != nil {
		klog.Error("find redis cluster secret error, ", err)
		return "", err
	}

	return string(secret.Data["kvrocks_password"]), err
}

func FindRedisClusterProxyInfo(ctx context.Context, k8sClient *kubernetes.Clientset,
	namespace string) (servicePort, replicas int32, err error) {
	service, err := k8sClient.CoreV1().Services(namespace).Get(ctx, RedisClusterService, metav1.GetOptions{})
	if err != nil {
		return
	}

	for _, port := range service.Spec.Ports {
		if port.Name == "proxy" {
			servicePort = port.Port
			break
		}
	}

	pods, err := k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(labels.Set(service.Spec.Selector))),
	})

	if err != nil {
		return
	}

	replicas = int32(len(funk.Filter(pods.Items, func(p corev1.Pod) bool {
		return p.Status.Phase == corev1.PodRunning
	}).([]corev1.Pod)))

	return
}

func ScaleRedisClusterNodes(ctx context.Context, dynamicClient *dynamic.DynamicClient, name, namespace string, nodes int32) error {
	cluster, err := dynamicClient.Resource(RedisClusterClassGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		klog.Error("get redis cluster error, ", err, ", ", name, ", ", namespace)
		return err
	}

	drc := DistributedRedisCluster{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(cluster.Object, &drc); err != nil {
		klog.Error("convert redis cluster crd error, ", err, ", ", name, ", ", namespace)
		return err
	}

	drc.Spec.MasterSize = nodes
	updateCluster, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&drc)
	if err != nil {
		klog.Error("to unstructured error, ", err, ", ", name, ", ", namespace)
		return err
	}

	_, err = dynamicClient.Resource(RedisClusterClassGVR).Namespace(namespace).Update(ctx,
		&unstructured.Unstructured{Object: updateCluster}, metav1.UpdateOptions{})

	return err
}

func ForceCreateNewRedisClusterBackup(ctx context.Context, dynamicClient *dynamic.DynamicClient, backup *RedisClusterBackup) error {
	currentBackupData, err := dynamicClient.Resource(RedisClusterBackupGVR).
		Namespace(backup.Namespace).Get(ctx, backup.Name, metav1.GetOptions{})

	if err != nil && !apierrors.IsNotFound(err) {
		klog.Error("get prev redis backup cr error, ", err)
		return err
	}

	if err == nil {
		var currentBackup RedisClusterBackup
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(currentBackupData.Object, &currentBackup)
		if err != nil {
			klog.Error("to unstructured error, ", err, ", ", currentBackupData.GetName(), ", ", currentBackupData.GetNamespace())
			return err
		}

		if currentBackup.Status.Phase == BackupPhaseRunning {
			klog.Error("prev redis backup job is still running")
			return errors.New("duplicate redis backup job")
		}

		klog.Info("remove prev redis backup, ", ", ", backup.Name, ", ", backup.Namespace)
		err = dynamicClient.Resource(RedisClusterBackupGVR).Namespace(backup.Namespace).Delete(ctx, backup.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.Error("delete prev redis backup error, ", err)
			return err
		}

		err = wait.PollWithContext(ctx, time.Second, time.Minute, func(ctx context.Context) (done bool, err error) {
			_, err = dynamicClient.Resource(RedisClusterBackupGVR).
				Namespace(backup.Namespace).Get(ctx, backup.Name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					klog.Info("prev redis backup removed")
					return true, nil
				}

				return false, err
			}

			return false, nil
		})

		if err != nil {
			return err
		}

		backupPath := backup.Spec.Local.HostPath.Path
		klog.Info("remove prev redis backup dir, ", ", ", backup.Name, ", ", backup.Namespace, ", ", backupPath)
		if dir, err := os.Stat(backupPath); err == nil {
			if dir.IsDir() {
				if err = os.RemoveAll(backupPath); err == nil {
					err = os.Mkdir(backupPath, 0755)
					if err != nil {
						klog.Error("re-make backup dir error, ", err, ", ", backupPath)
					}
				} else {
					klog.Error("remove prev backup dir error, ", err, ", ", backupPath)
				}
			}
		} else {
			klog.Error("get dir stat error, ", err, ", ", backupPath)
		}

	}

	klog.Info("create a new redis backup, ", ", ", backup.Name, ", ", backup.Namespace)
	createBackup, err := runtime.DefaultUnstructuredConverter.ToUnstructured(backup)
	if err != nil {
		klog.Error("to unstructured error, ", err, ", ", backup.Name, ", ", backup.Namespace)
		return err
	}

	_, err = dynamicClient.Resource(RedisClusterBackupGVR).Namespace(backup.Namespace).Create(ctx,
		&unstructured.Unstructured{Object: createBackup}, metav1.CreateOptions{})

	if err != nil {
		klog.Error("create redis backup error, ", err, ", ", backup.Name, ", ", backup.Namespace)
	}

	return err
}

func WaitForAllBackupComplete(ctx context.Context, dynamicClient *dynamic.DynamicClient) error {
	return wait.PollWithContext(ctx, 5*time.Second, time.Hour,
		func(context.Context) (done bool, err error) {
			var errs []error
			incompleted := false
			backups, err := dynamicClient.Resource(RedisClusterBackupGVR).List(ctx, metav1.ListOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}

				klog.Error("list redis backup error, ", err)
				return false, err
			}

			for _, b := range backups.Items {
				backup := RedisClusterBackup{}
				if err = runtime.DefaultUnstructuredConverter.FromUnstructured(b.Object, &backup); err != nil {
					klog.Error("convert redis backup crd error, ", err, ", ", b.GetName(), ", ", b.GetNamespace())
					return false, err
				}

				switch backup.Status.Phase {
				case BackupPhaseRunning:
					incompleted = true
				case BackupPhaseFailed:
					errs = append(errs, fmt.Errorf("backup failed, %s, %s, %s", backup.Status.Reason, backup.Name, backup.Namespace))
				}
			}

			if len(errs) > 0 {
				err = utilerrors.NewAggregate(errs)
			}

			done = !incompleted

			return
		})
}

func UpdataClusterStatus(ctx context.Context, dynamicClient *dynamic.DynamicClient, cluster *DistributedRedisCluster) (*DistributedRedisCluster, error) {
	return updataCluster(ctx, dynamicClient, cluster, func(data map[string]interface{}) (*unstructured.Unstructured, error) {
		c, err := dynamicClient.Resource(RedisClusterClassGVR).Namespace(cluster.Namespace).UpdateStatus(ctx, &unstructured.Unstructured{Object: data}, metav1.UpdateOptions{})
		if err != nil {
			klog.Error("update cluster status error, ", err, ", ", cluster.Name, ", ", cluster.Namespace)
			return nil, err
		}

		return c, nil
	})
}

func UpdataCluster(ctx context.Context, dynamicClient *dynamic.DynamicClient, cluster *DistributedRedisCluster) (*DistributedRedisCluster, error) {
	return updataCluster(ctx, dynamicClient, cluster, func(data map[string]interface{}) (*unstructured.Unstructured, error) {
		c, err := dynamicClient.Resource(RedisClusterClassGVR).Namespace(cluster.Namespace).Update(ctx, &unstructured.Unstructured{Object: data}, metav1.UpdateOptions{})
		if err != nil {
			klog.Error("update cluster data error, ", err, ", ", cluster.Name, ", ", cluster.Namespace)
			return nil, err
		}

		return c, nil
	})
}

func updataCluster(ctx context.Context, dynamicClient *dynamic.DynamicClient,
	cluster *DistributedRedisCluster, updateFn func(data map[string]interface{}) (*unstructured.Unstructured, error),
) (*DistributedRedisCluster, error) {
	updateCluster, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cluster)
	if err != nil {
		klog.Error("unstructured redis cluster error, ", err, ", ", cluster.Name, ", ", cluster.Namespace)
		return nil, err
	}

	c, err := updateFn(updateCluster)
	if err != nil {
		klog.Error("update redis cluster error, ", err, ", ", cluster.Name, ", ", cluster.Namespace)
		return nil, err
	}

	var retCluster DistributedRedisCluster
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(c.Object, &retCluster)
	if err != nil {
		klog.Error("convert redis cluster error, ", err, ", ", cluster.Name, ", ", cluster.Namespace)
		return nil, err
	}
	return &retCluster, nil
}

func GetDatabaseName(namespace, db string) string {
	return fmt.Sprintf("%s_%s", namespace, db)
}
