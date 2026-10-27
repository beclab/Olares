package controllers

import (
	"context"
	"encoding/json"
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
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
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

				// klog.Info("refresh juicefs config with ", args)
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

	nodeName := ""
	for _, node := range nodeList.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeInternalIP && address.Address == NodeIP {
				nodeName = node.Name
			}
		}
	}
	if nodeName == "" {
		return ctrl.Result{}, fmt.Errorf("cannot identify current node")
	}
	var claims corev1.PersistentVolumeClaimList
	if err = r.List(ctx, &claims, client.HasLabels{UserStorageRoleLabel}); err != nil {
		return ctrl.Result{}, err
	}
	for _, claim := range claims.Items {
		if claim.DeletionTimestamp != nil || !strings.HasPrefix(claim.Namespace, "user-space-") {
			continue
		}
		if err := createDataDirs(&claim); err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			var current corev1.PersistentVolumeClaim
			if err := r.Get(ctx, client.ObjectKeyFromObject(&claim), &current); err != nil {
				return client.IgnoreNotFound(err)
			}
			var initialized map[string]bool
			raw := current.Annotations[InitializedNodesAnnotation]
			if raw != "" {
				if err := json.Unmarshal([]byte(raw), &initialized); err != nil {
					return err
				}
			}
			if initialized[nodeName] {
				return nil
			}
			if initialized == nil {
				initialized = map[string]bool{}
			}
			initialized[nodeName] = true
			data, err := json.Marshal(initialized)
			if err != nil {
				return err
			}
			if current.Annotations == nil {
				current.Annotations = map[string]string{}
			}
			current.Annotations[InitializedNodesAnnotation] = string(data)
			return r.Update(ctx, &current)
		}); err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}
	}

	return ctrl.Result{}, nil
}

func createDataDirs(claim *corev1.PersistentVolumeClaim) error {
	role := claim.Labels[UserStorageRoleLabel]
	directories := map[string]string{"userspace": "rootfs/userspace", "appcache": "userdata/Cache", "dbdata": "userdata/dbdata"}
	directory, ok := directories[role]
	if !ok {
		return fmt.Errorf("unknown personal storage role %q", role)
	}
	relative := claim.Annotations[StorageRelativePathAnnotation]
	if relative != filepath.Join(directory, claim.Name) {
		return fmt.Errorf("invalid relative path on PVC %s/%s", claim.Namespace, claim.Name)
	}
	root := filepath.Join("/olares", relative)
	type ownedDir struct {
		path     string
		uid, gid int
	}
	dirs := []ownedDir{{root, 1000, 1000}}
	switch role {
	case "userspace":
		for _, subdir := range []string{"Home", "Home/Documents", "Home/Downloads", "Home/Pictures", "Home/Movies", "Home/Music", "Home/Code", "Data"} {
			dirs = append(dirs, ownedDir{filepath.Join(root, subdir), 1000, 1000})
		}
	case "appcache":
		for name, perm := range AppSubDirs {
			dirs = append(dirs, ownedDir{filepath.Join(root, name), perm[0], perm[1]})
		}
	case "dbdata":
		dirs[0].uid, dirs[0].gid = 0, 0
		for name, perm := range DbDataSubDirs {
			dirs = append(dirs, ownedDir{filepath.Join(root, name), perm[0], perm[1]})
		}
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir.path, 0755); err != nil {
			return err
		}
		info, err := os.Lstat(dir.path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		stat := info.Sys().(*syscall.Stat_t)
		if stat.Uid != uint32(dir.uid) || stat.Gid != uint32(dir.gid) {
			if err := os.Chown(dir.path, dir.uid, dir.gid); err != nil {
				return err
			}
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
	labels := client.HasLabels{"node-role.kubernetes.io/control-plane"}
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

	return c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}},
		handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
			return []reconcile.Request{{NamespacedName: types.NamespacedName{
				Namespace: o.GetNamespace(),
				Name:      o.GetName()}},
			}
		}), newCreateOnlyPredicate(func(e event.CreateEvent) bool {
			claim, ok := e.Object.(*corev1.PersistentVolumeClaim)
			if ok && claim.Labels[UserStorageRoleLabel] != "" && strings.HasPrefix(claim.Namespace, "user-space-") {
				return true
			}
			return false
		}),
	)
}
