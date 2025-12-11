package controllers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"bytetrade.io/web3os/osnode-init/pkg/log"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// NodeInitController reconciles a BackupConfig object
type NodeInitController struct {
	client.Client
	scheme *runtime.Scheme
	cron   *cron.Cron
}

func NewNodeInitController(c client.Client, schema *runtime.Scheme, config *rest.Config) *NodeInitController {
	nic := &NodeInitController{Client: c, scheme: schema, cron: cron.New()}
	schedule := os.Getenv("SCHEDULE")
	if schedule == "" {
		// random timer schedule, to avoid too much concurrent api requets
		schedule = fmt.Sprintf("%d */8 * * *", time.Now().Minute())
	}

	if isMaster, _, err := nic.isMasterNode(config); err != nil {
		log.Fatal("get master node info error, ", err)
		panic(err)
	} else {
		if isMaster {
			// add job to master node
			nic.cron.AddFunc(schedule, func() {
				ctx := context.Background()
				dynamicClient, err := dynamic.NewForConfig(config)
				if err != nil {
					klog.Error("create kube client error, ", err)
					return
				}

				bucket := os.Getenv("S3_BUCKET")
				if bucket == "" {
					klog.Error("bucket is unknown")
					return
				}

				if bucket == "none" {
					return
				}

				klog.Info("get refresh session token from cloud")
				account, err := GetAwsAccountFromCloud(ctx, dynamicClient, bucket)
				if err != nil {
					klog.Error("get token from cloud error, ", err)
					return
				}

				// exec juice command
				klog.Info("find jiucefs redis ip and password")
				ip, pwd, err := getRedisIpAndPassword()
				if err != nil {
					klog.Error(err)
					return
				}

				args := []string{
					"config",
					"redis://:" + pwd + "@" + ip + ":6379/1",
					"--access-key",
					account.Key,
					"--secret-key",
					account.Secret,
					"--session-token",
					account.Token,
				}

				klog.Info("refresh juicefs config with ", args)
				out, err := exec.Command("/usr/local/bin/juicefs", args...).CombinedOutput()
				if err != nil {
					klog.Errorf("%s, %v", string(out), err)
				} else {
					klog.Info("refresh succeed, update temrinus s3 labels")
					updateAwsAccount(ctx, dynamicClient, account)
				}

			})
		}
	}

	nic.cron.Start()

	return nic
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BackupConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *NodeInitController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Infof("received nodeinit request, namespace: %q, name: %q", req.Namespace, req.Name)

	var (
		err      error
		nodeList corev1.NodeList
	)

	if err = r.List(ctx, &nodeList); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}, errors.WithStack(err)
	}
	if !r.isCurrentNode(nodeList) {
		log.Warnf("not current node %q, ignore", NodeIP)
		return ctrl.Result{}, nil
	}
	nodeList.Items = nil

	var statefulSets appsv1.StatefulSetList

	if err = r.List(ctx, &statefulSets, client.MatchingLabels{"tier": "bfl"}); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}, errors.WithStack(err)
	} else {
		for _, sts := range statefulSets.Items {
			if !isUserNamespaceBfl(sts.Namespace, sts.Name) {
				continue
			}
			log.Debugf("creating %q bfl userdata dirs", sts.Namespace)
			if !hasAllAnnotations(sts.Annotations, BflAnnotationAppCache, BflAnnotationDbData) {
				return ctrl.Result{}, errors.Errorf("namespace %q bfl has no userdata annotation", sts.Namespace)
			}
			if err = createDataDirs(&sts); err != nil {
				return ctrl.Result{}, errors.Errorf("creating %q bfl userdata dirs, %v", sts.Namespace, err)
			}
		}
	}

	return ctrl.Result{}, nil
}

func createDataDirs(sts *appsv1.StatefulSet) error {
	var err error

	mkDirs := func(perm []int, dirs ...string) (err error) {
		for _, dir := range dirs {
			if !filePathExists(dir) {
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					return errors.WithStack(err)
				}
			}

			var di os.FileInfo
			di, err = os.Stat(dir)
			if err != nil {
				return errors.WithStack(err)
			} else {
				stat := di.Sys().(*syscall.Stat_t)
				if stat.Uid == uint32(perm[0]) && stat.Gid == uint32(perm[1]) {
					return
				}
				if err = os.Chown(dir, perm[0], perm[1]); err != nil {
					return errors.WithStack(err)
				}
				log.Debugf("%q created, and set uid: %v, gid: %v ", dir, perm[0], perm[1])
			}
		}
		return
	}

	appDir, dbDir := sts.Annotations[BflAnnotationAppCache], sts.Annotations[BflAnnotationDbData]

	// appdata
	for name, perm := range AppSubDirs {
		if err = mkDirs(perm, filepath.Join(appDir, name)); err != nil {
			return err
		}
	}

	// create dbData sub dirs and set permissions
	for name, perm := range DbDataSubDirs {
		if err = mkDirs(perm, filepath.Join(dbDir, name)); err != nil {
			return err
		}
	}

	return nil
}

func (r *NodeInitController) isCurrentNode(nodeList corev1.NodeList) bool {
	if nodeList.Items == nil || len(nodeList.Items) == 0 {
		return false
	}

	for _, node := range nodeList.Items {
		if node.Status.Addresses != nil {
			for _, addr := range node.Status.Addresses {
				if addr.Type == corev1.NodeInternalIP && addr.Address == NodeIP {
					return true
				}
			}
		}
	}
	return false
}

func (r *NodeInitController) isMasterNode(config *rest.Config) (bool, string, error) {
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return false, "", err
	}

	var nodeLists *corev1.NodeList
	labels := client.HasLabels{"node-role.kubernetes.io/master"}
	var opts client.ListOptions
	labels.ApplyToList(&opts)

	if nodeLists, err = kubeClient.CoreV1().Nodes().List(context.TODO(),
		v1.ListOptions{LabelSelector: opts.LabelSelector.String()}); err != nil {
		return false, "", err
	}

	for _, node := range nodeLists.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP && addr.Address != "" && addr.Address == NodeIP {
				return true, addr.Address, nil
			}
		}
	}
	return false, "", nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeInitController) SetupWithManager(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).For(&corev1.Node{},
		builder.WithPredicates(newCreateOnlyPredicate(nil))).Build(r)
	if err != nil {
		return errors.WithStack(err)
	}

	return c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}},
		handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
			return []reconcile.Request{{NamespacedName: types.NamespacedName{
				Namespace: o.GetNamespace(),
				Name:      o.GetName()}},
			}
		}), newCreateOnlyPredicate(func(e event.CreateEvent) bool {
			sts, ok := e.Object.(*appsv1.StatefulSet)
			if ok && isUserNamespaceBfl(sts.Namespace, sts.Name) {
				return true
			}
			return false
		}),
	)
}

func filePathExists(name string) bool {
	_, err := os.Stat(name)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func isUserNamespaceBfl(namespace, name string) bool {
	return name == BflStatefulSetName && strings.HasPrefix(namespace, "user-space")
}

func hasAllAnnotations(annotations map[string]string, keys ...string) bool {
	if annotations == nil {
		return false
	}

	sb := sets.NewByte()

	for _, key := range keys {
		if v, ok := annotations[key]; ok && v != "" {
			sb.Insert('y')
		} else {
			sb.Insert('n')
		}
	}
	return sb.HasAll('y')
}
