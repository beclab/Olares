/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubekey

import (
	"bytes"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"sigs.k8s.io/yaml"

	kubekeyv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	yamlV2 "gopkg.in/yaml.v2"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	CreateCluster = "create cluster"
	AddNodes      = "add nodes"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=storage.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=*
// +kubebuilder:rbac:groups=installer.kubesphere.io,resources=clusterconfigurations,verbs=*
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=apiregistration.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=auditing.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=autoscaling,resources=*,verbs=*
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=config.istio.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=core.kubefed.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=devops.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=events.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=batch,resources=*,verbs=*
// +kubebuilder:rbac:groups=extensions,resources=*,verbs=*
// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=jaegertracing.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=logging.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=*,verbs=*
// +kubebuilder:rbac:groups=networking.istio.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=notification.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=policy,resources=*,verbs=*
// +kubebuilder:rbac:groups=storage.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=cluster.kubesphere.io,resources=*,verbs=*

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	cluster := &kubekeyv1alpha2.Cluster{}
	cmFound := &corev1.ConfigMap{}
	jobFound := &batchv1.Job{}
	var (
		clusterAlreadyExist   bool
		addHosts, removeHosts []kubekeyv1alpha2.HostCfg
	)
	// Fetch the Cluster object
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if kubeErr.IsNotFound(err) {
			logger.Info("Cluster resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Cluster")
		return ctrl.Result{}, err
	}

	// Check if the configMap already exists
	if err := r.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: req.Namespace}, cmFound); err == nil {
		clusterAlreadyExist = true
	}

	// create a new cluster
	if cluster.Status.NodesCount == 0 {
		// Ensure that the current pipeline execution is complete
		if cluster.Status.PiplineInfo.Status != "" {
			return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
		}

		if !clusterAlreadyExist {
			// create kubesphere cluster
			if err := newKubeSphereCluster(r, cluster); err != nil {
				return ctrl.Result{RequeueAfter: 2 * time.Second}, err
			}

			if err := updateClusterConfigMap(r, ctx, cluster, cmFound); err != nil {
				return ctrl.Result{RequeueAfter: 2 * time.Second}, err
			}
			return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
		}

		// If the CR cluster define current cluster
		nodes, err := currentClusterDiff(r, ctx, cluster)
		if err != nil {
			return ctrl.Result{RequeueAfter: 2 * time.Second}, err
		}
		if len(nodes) != 0 {
			logger.Info("Cluster resource defines current cluster")
			if err := adaptExistedCluster(nodes, cluster); err != nil {
				return ctrl.Result{RequeueAfter: 2 * time.Second}, err
			}
			if err := r.Status().Update(context.TODO(), cluster); err != nil {
				return ctrl.Result{RequeueAfter: 2 * time.Second}, err
			}

			return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
		}

		// If the CR cluster define other existed cluster
		otherClusterNodes, err := otherClusterDiff(r, ctx, cluster)
		if err != nil {
			return ctrl.Result{RequeueAfter: 2 * time.Second}, err
		}
		if len(otherClusterNodes) != 0 {
			logger.Info("Cluster resource defines other existed cluster")
			if err := adaptExistedCluster(otherClusterNodes, cluster); err != nil {
				return ctrl.Result{RequeueAfter: 2 * time.Second}, err
			}
			if err := r.Status().Update(context.TODO(), cluster); err != nil {
				return ctrl.Result{RequeueAfter: 2 * time.Second}, err
			}

			return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
		}

		if err := updateRunJob(r, req, ctx, cluster, jobFound, CreateCluster); err != nil {
			return ctrl.Result{RequeueAfter: 2 * time.Second}, err
		}

		addHosts = cluster.Spec.Hosts
		sendHostsAction(1, addHosts)

		// Ensure that the cluster has been created successfully, otherwise re-enter Reconcile.
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	}

	// add nodes to cluster
	if cluster.Status.NodesCount != 0 && len(cluster.Spec.Hosts) > cluster.Status.NodesCount {
		// Ensure that the current pipeline execution is complete
		if cluster.Status.PiplineInfo.Status != "" {
			return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
		}

		if err := updateClusterConfigMap(r, ctx, cluster, cmFound); err != nil {
			return ctrl.Result{}, err
		}
		if err := updateRunJob(r, req, ctx, cluster, jobFound, AddNodes); err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		currentNodes := map[string]string{}
		for _, node := range cluster.Status.Nodes {
			currentNodes[node.Hostname] = node.Hostname
		}

		for _, host := range cluster.Spec.Hosts {
			if _, ok := currentNodes[host.Name]; !ok {
				addHosts = append(addHosts, host)
			}
		}
		sendHostsAction(1, addHosts)

		// Ensure that the nodes has been added successfully, otherwise re-enter Reconcile.
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	}

	// Completion of pipeline execution, clearing pipeline status information
	cluster.Status.PiplineInfo.Status = ""
	if err := r.Status().Update(context.TODO(), cluster); err != nil {
		return ctrl.Result{RequeueAfter: 3 * time.Second}, err
	}

	// Synchronizing Node Information
	kubeConfigCm := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-kubeconfig", cluster.Name), Namespace: req.Namespace}, kubeConfigCm); err == nil && len(cluster.Status.Nodes) != 0 {
		// fixme: this code will delete the kube config configmap when user delete the CR cluster.
		// And if user apply this deleted CR cluster again, kk will no longer be able to find the kube config.

		//kubeConfigCm.OwnerReferences = []metav1.OwnerReference{{
		//	APIVersion: cluster.APIVersion,
		//	Kind:       cluster.Kind,
		//	Name:       cluster.Name,
		//	UID:        cluster.UID,
		//}}
		//if err := r.Update(ctx, kubeConfigCm); err != nil {
		//	return ctrl.Result{Requeue: true}, err
		//}
		kubeconfig, err := base64.StdEncoding.DecodeString(kubeConfigCm.Data["kubeconfig"])
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		currentNodes := map[string]string{}
		for _, node := range nodeList.Items {
			currentNodes[node.Name] = node.Name
		}
		for _, etcd := range cluster.Spec.RoleGroups["etcd"] {
			if _, ok := currentNodes[etcd]; !ok {
				currentNodes[etcd] = etcd
			}
		}

		nodes := cluster.Status.Nodes
		var newNodes []kubekeyv1alpha2.NodeStatus

		for _, node := range nodes {
			if _, ok := currentNodes[node.Hostname]; ok {
				newNodes = append(newNodes, node)
			}
		}

		hosts := cluster.Spec.Hosts
		var newHosts []kubekeyv1alpha2.HostCfg
		for _, host := range hosts {
			if _, ok := currentNodes[host.Name]; ok {
				newHosts = append(newHosts, host)
			} else {
				removeHosts = append(removeHosts, host)
			}
		}

		sendHostsAction(0, removeHosts)

		var newEtcd, newMaster, newWorker []string
		for _, node := range newNodes {
			if node.Roles["etcd"] {
				newEtcd = append(newEtcd, node.Hostname)
			}
			if node.Roles["master"] {
				newMaster = append(newMaster, node.Hostname)
			}
			if node.Roles["worker"] {
				newWorker = append(newWorker, node.Hostname)
			}
		}

		cluster.Spec.Hosts = newHosts
		cluster.Spec.RoleGroups = map[string][]string{
			"master": newMaster,
			"etcd":   newEtcd,
			"worker": newWorker,
		}

		if err := r.Update(ctx, cluster); err != nil {
			return ctrl.Result{Requeue: true}, nil
		}

		// Fetch the Cluster object
		if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
			if kubeErr.IsNotFound(err) {
				logger.Info("Cluster resource not found. Ignoring since object must be deleted")
				return ctrl.Result{}, nil
			}
			logger.Error(err, "Failed to get Cluster")
			return ctrl.Result{}, err
		}

		cluster.Status.Nodes = newNodes
		cluster.Status.NodesCount = len(newNodes)
		cluster.Status.MasterCount = len(newMaster)
		cluster.Status.WorkerCount = len(newWorker)
		if err := r.Status().Update(ctx, cluster); err != nil {
			return ctrl.Result{Requeue: true}, nil
		}
	}
	return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubekeyv1alpha2.Cluster{}).
		WithEventFilter(ignoreDeletionPredicate()).
		Complete(r)
}

func ignoreDeletionPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

func (r *ClusterReconciler) configMapForCluster(c *kubekeyv1alpha2.Cluster) *corev1.ConfigMap {
	type Metadata struct {
		Name string `yaml:"name" json:"name,omitempty"`
	}
	clusterConfiguration := struct {
		ApiVersion string                      `yaml:"apiVersion" json:"apiVersion,omitempty"`
		Kind       string                      `yaml:"kind" json:"kind,omitempty"`
		Metadata   Metadata                    `yaml:"metadata" json:"metadata,omitempty"`
		Spec       kubekeyv1alpha2.ClusterSpec `yaml:"spec" json:"spec,omitempty"`
	}{ApiVersion: c.APIVersion, Kind: c.Kind, Metadata: Metadata{Name: c.Name}, Spec: c.Spec}

	clusterStr, _ := yaml.Marshal(clusterConfiguration)

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
			Labels:    map[string]string{"kubekey.kubesphere.io/name": c.Name},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: c.APIVersion,
				Kind:       c.Kind,
				Name:       c.Name,
				UID:        c.UID,
			}},
		},
		Data: map[string]string{"cluster.yaml": string(clusterStr)},
	}
	return cm
}

func (r *ClusterReconciler) jobForCluster(c *kubekeyv1alpha2.Cluster, action string) *batchv1.Job {
	var (
		backoffLimit int32 = 0
		name         string
		args         []string
	)
	if action == CreateCluster {
		name = fmt.Sprintf("%s-create-cluster", c.Name)
		args = []string{"create", "cluster", "-f", "/home/kubekey/config/cluster.yaml", "-y", "--in-cluster", "true", "--namespace", c.Namespace}
	} else if action == AddNodes {
		name = fmt.Sprintf("%s-add-nodes", c.Name)
		args = []string{"add", "nodes", "-f", "/home/kubekey/config/cluster.yaml", "-y", "--in-cluster", "true", "--ignore-err", "true", "--namespace", c.Namespace}
	}

	podlist := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(c.Namespace),
		client.MatchingLabels{"control-plane": "controller-manager"},
	}
	err := r.List(context.TODO(), podlist, listOpts...)
	if err != nil {
		logger.Error(err, "Failed to list kubekey controller-manager pod")
	}
	nodeName := podlist.Items[0].Spec.NodeName
	var image string
	for _, container := range podlist.Items[0].Spec.Containers {
		if container.Name == "manager" {
			image = container.Image
		}
	}

	imageRepoList := strings.Split(image, "/")
	var kubekey int64
	kubekey = 1000

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
			Labels:    map[string]string{"kubekey.kubesphere.io/name": c.Name},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: c.APIVersion,
				Kind:       c.Kind,
				Name:       c.Name,
				UID:        c.UID,
			}},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: c.Name,
									},
									Items: []corev1.KeyToPath{{
										Key:  "cluster.yaml",
										Path: "cluster.yaml",
									}},
								},
							},
						},
						{
							Name: "kube-binaries",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &kubekey,
						FSGroup:   &kubekey,
					},
					InitContainers: []corev1.Container{
						{
							Name:  "kube-binaries",
							Image: fmt.Sprintf("%s/kube-binaries:%s", strings.Join(imageRepoList[:len(imageRepoList)-1], "/"), c.Spec.Kubernetes.Version),
							Command: []string{
								"sh",
								"-c",
								"cp -r -f /kubekey/* /home/kubekey/kubekey/",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "kube-binaries",
									MountPath: "/home/kubekey/kubekey",
								},
							},
						},
					},
					Containers: []corev1.Container{{
						Name:            "runner",
						Image:           image,
						ImagePullPolicy: "IfNotPresent",
						Command:         []string{"/home/kubekey/kk"},
						Args:            args,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "config",
								MountPath: "/home/kubekey/config",
							},
							{
								Name:      "kube-binaries",
								MountPath: "/home/kubekey/kubekey",
							},
						},
					}},
					NodeName:           nodeName,
					ServiceAccountName: "kubekey-controller-manager",
					RestartPolicy:      "Never",
				},
			},
		},
	}
	return job
}

func updateStatusRunner(r *ClusterReconciler, req ctrl.Request, cluster *kubekeyv1alpha2.Cluster, action string) error {
	var (
		name string
	)
	if action == CreateCluster {
		name = fmt.Sprintf("%s-create-cluster", cluster.Name)
	} else if action == AddNodes {
		name = fmt.Sprintf("%s-add-nodes", cluster.Name)
	}

	podlist := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(req.Namespace),
		client.MatchingLabels{"job-name": name},
	}
	for i := 0; i < 100; i++ {
		_ = r.List(context.TODO(), podlist, listOpts...)
		if len(podlist.Items) == 1 {
			// Fetch the Cluster object
			if err := r.Get(context.TODO(), req.NamespacedName, cluster); err != nil {
				if kubeErr.IsNotFound(err) {
					return nil
				}
				return err
			}

			if len(podlist.Items[0].ObjectMeta.GetName()) != 0 && len(podlist.Items[0].Status.ContainerStatuses[0].Name) != 0 && podlist.Items[0].Status.Phase != "Pending" {
				cluster.Status.JobInfo = kubekeyv1alpha2.JobInfo{
					Namespace: req.Namespace,
					Name:      name,
					Pods: []kubekeyv1alpha2.PodInfo{{
						Name:       podlist.Items[0].ObjectMeta.GetName(),
						Containers: []kubekeyv1alpha2.ContainerInfo{{Name: podlist.Items[0].Status.ContainerStatuses[0].Name}},
					}},
				}

				if err := r.Status().Update(context.TODO(), cluster); err != nil {
					return err
				}

				break
			}
		}
		time.Sleep(6 * time.Second)
	}

	return nil
}

func updateClusterConfigMap(r *ClusterReconciler, ctx context.Context, cluster *kubekeyv1alpha2.Cluster, cmFound *corev1.ConfigMap) error {
	// Check if the configmap already exists, if not create a new one
	if err := r.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, cmFound); err != nil && !kubeErr.IsNotFound(err) {
		logger.Error(err, "Failed to get ConfigMap", "ConfigMap.Namespace", cmFound.Namespace, "ConfigMap.Name", cmFound.Name)
		return err
	} else if err == nil {
		if err := r.Delete(ctx, cmFound); err != nil {
			logger.Error(err, "Failed to delete old ConfigMap", "ConfigMap.Namespace", cmFound.Namespace, "ConfigMap.Name", cmFound.Name)
			return err
		}
	}

	// Define a new configmap
	cmCluster := r.configMapForCluster(cluster)
	logger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cmCluster.Namespace, "ConfigMap.Name", cmCluster.Name)
	if err := r.Create(ctx, cmCluster); err != nil {
		logger.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", cmCluster.Namespace, "ConfigMap.Name", cmCluster.Name)
		return err
	}
	return nil
}

func updateRunJob(r *ClusterReconciler, req ctrl.Request, ctx context.Context, cluster *kubekeyv1alpha2.Cluster, jobFound *batchv1.Job, action string) error {
	var (
		name string
	)
	if action == CreateCluster {
		name = fmt.Sprintf("%s-create-cluster", cluster.Name)
	} else if action == AddNodes {
		name = fmt.Sprintf("%s-add-nodes", cluster.Name)
	}

	// Check if the job already exists, if not create a new one
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: req.Namespace}, jobFound); err != nil && !kubeErr.IsNotFound(err) {
		return err
	} else if err == nil && (jobFound.Status.Failed != 0 || jobFound.Status.Succeeded != 0) {
		// delete old pods
		podlist := &corev1.PodList{}
		listOpts := []client.ListOption{
			client.InNamespace(req.Namespace),
			client.MatchingLabels{"job-name": name},
		}
		if err := r.List(context.TODO(), podlist, listOpts...); err == nil && len(podlist.Items) != 0 {
			for _, pod := range podlist.Items {
				_ = r.Delete(ctx, &pod)
			}
		}
		logger.Info("Prepare to delete old job", "Job.Namespace", jobFound.Namespace, "Job.Name", jobFound.Name)
		if err := r.Delete(ctx, jobFound); err != nil {
			logger.Error(err, "Failed to delete old Job", "Job.Namespace", jobFound.Namespace, "Job.Name", jobFound.Name)
			return err
		}
		logger.Info("Deleting old job success", "Job.Namespace", jobFound.Namespace, "Job.Name", jobFound.Name)

		err := wait.PollInfinite(1*time.Second, func() (bool, error) {
			logger.Info("Checking old job is deleted", "Job.Namespace", jobFound.Namespace, "Job.Name", jobFound.Name)
			if e := r.Get(ctx, types.NamespacedName{Name: name, Namespace: req.Namespace}, jobFound); e != nil {
				if kubeErr.IsNotFound(e) {
					return true, nil
				}
				return false, e
			}
			return false, nil
		})
		if err != nil {
			logger.Error(err, "Failed to loop check old job is deleted", "Job.Namespace", jobFound.Namespace, "Job.Name", jobFound.Name)
			return err
		}

		jobCluster := r.jobForCluster(cluster, action)
		logger.Info("Creating a new Job to scale cluster", "Job.Namespace", jobCluster.Namespace, "Job.Name", jobCluster.Name)
		if err := r.Create(ctx, jobCluster); err != nil {
			logger.Error(err, "Failed to create new Job", "Job.Namespace", jobCluster.Namespace, "Job.Name", jobCluster.Name)
			return err
		}
	} else if kubeErr.IsNotFound(err) {
		jobCluster := r.jobForCluster(cluster, action)
		logger.Info("Creating a new Job to create cluster", "Job.Namespace", jobCluster.Namespace, "Job.Name", jobCluster.Name)
		if err := r.Create(ctx, jobCluster); err != nil {
			logger.Error(err, "Failed to create new Job", "Job.Namespace", jobCluster.Namespace, "Job.Name", jobCluster.Name)
			return err
		}
	}
	if err := updateStatusRunner(r, req, cluster, action); err != nil {
		return err
	}
	return nil
}

func sendHostsAction(action int, hosts []kubekeyv1alpha2.HostCfg) {
	if os.Getenv("HOSTS_MANAGER") == "true" {
		type HostsAction struct {
			Hosts  []kubekeyv1alpha2.HostCfg `json:"hosts,omitempty"`
			Action int                       `json:"action,omitempty"`
		}
		newHostsAction := HostsAction{
			Hosts:  hosts,
			Action: action,
		}

		fmt.Println(newHostsAction)
		hostsInfoBytes, err := json.Marshal(newHostsAction)
		if err != nil {
			logger.Error(err, "Failed to marshal hosts info")
		}

		fmt.Println(string(hostsInfoBytes))
		req, err := http.NewRequest("POST", "http://localhost:8090/api/v1alpha2/hosts", bytes.NewReader(hostsInfoBytes))
		if err != nil {
			logger.Error(err, "Failed to create request")
		}

		req.Header.Add("Content-Type", "application/json")

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			logger.Error(err, "Failed to  send hosts info")
		}
	}
}

type NodeInfo struct {
	Name    string
	Address string
	Master  bool
	Worker  bool
	Etcd    bool
}

type Configuration struct {
	Etcd Etcd `yaml:"etcd"`
}

type Etcd struct {
	External External `yaml:"external"`
}

type External struct {
	Endpoints []string `yaml:"endpoints"`
}

func currentClusterDiff(r *ClusterReconciler, ctx context.Context, c *kubekeyv1alpha2.Cluster) ([]kubekeyv1alpha2.NodeStatus, error) {
	newNodes := make([]kubekeyv1alpha2.NodeStatus, 0)

	allSpecHostsMap := make(map[string]NodeInfo)
	for _, host := range c.Spec.Hosts {
		allSpecHostsMap[host.InternalAddress] = NodeInfo{
			Name:    host.Name,
			Address: host.InternalAddress,
		}
	}

	nodeList := &corev1.NodeList{}
	if err := r.List(ctx, nodeList, &client.ListOptions{}); err != nil {
		return newNodes, err
	}

	// get cluster nodes which in spec hosts
	currentHostsMap := make(map[string]NodeInfo)
	for _, node := range nodeList.Items {
		isMaster := false
		isWorker := false

		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			isMaster = true
		}
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			isMaster = true
		}
		if _, ok := node.Labels["node-role.kubernetes.io/worker"]; ok {
			isWorker = true
		}

		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeInternalIP {
				if _, ok := allSpecHostsMap[address.Address]; ok {
					currentHostsMap[address.Address] = NodeInfo{
						Address: address.Address,
						Name:    node.Name,
						Master:  isMaster,
						Worker:  isWorker,
					}
				}
			}
		}
	}

	// spec cluster is different from current cluster
	if len(currentHostsMap) == 0 {
		return newNodes, nil
	}

	if len(currentHostsMap) < len(nodeList.Items) {
		return newNodes, errors.New("the number of spec hosts are not enough compared with the cluster exist hosts")
	}

	// mark which ip or node name is etcd role
	etcdSpecMap := make(map[string]bool)
	kubeadmConfig := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: "kubeadm-config", Namespace: "kube-system"}, kubeadmConfig)
	if err == nil {
		configuration, ok := kubeadmConfig.Data["ClusterConfiguration"]
		if !ok {
			return newNodes, errors.Errorf("Failed to get ClusterConfiguration key in kubeadm-config configmap")
		}

		conf := &Configuration{}
		if err := yamlV2.Unmarshal([]byte(configuration), conf); err != nil {
			return newNodes, err
		}

		for _, endpoint := range conf.Etcd.External.Endpoints {
			ip, err := findIpAddress(endpoint)
			if err != nil {
				return newNodes, errors.Wrap(err, "Failed to find etcd external endpoints")
			}
			etcdSpecMap[ip] = true
		}
	} else {
		if kubeErr.IsNotFound(err) {
			for _, etcdHostName := range c.Spec.RoleGroups["etcd"] {
				etcdSpecMap[etcdHostName] = true
			}
		} else {
			return newNodes, errors.Wrap(errors.WithStack(err), "Failed to get kubeadm-config")
		}
	}

	// add etcd node info to current hosts map
	for _, host := range allSpecHostsMap {
		_, nameOk := etcdSpecMap[host.Name]
		_, ipOk := etcdSpecMap[host.Address]
		if !nameOk && !ipOk {
			continue
		}
		if exist, ok := currentHostsMap[host.Address]; ok {
			exist.Etcd = true
			currentHostsMap[host.Address] = exist
		} else {
			currentHostsMap[host.Address] = NodeInfo{
				Address: host.Address,
				Name:    host.Name,
				Etcd:    true,
			}
		}
	}

	for _, host := range currentHostsMap {
		newNodes = append(newNodes, kubekeyv1alpha2.NodeStatus{
			InternalIP: host.Address,
			Hostname:   host.Name,
			Roles:      map[string]bool{"master": host.Master, "worker": host.Worker, "etcd": host.Etcd},
		})
	}
	return newNodes, nil
}

func findIpAddress(endpoint string) (string, error) {
	ipv4Regexp, err := regexp.Compile(`[\d]+\.[\d]+\.[\d]+\.[\d]+`)
	if err != nil {
		return "", err
	}
	ip := ipv4Regexp.FindStringSubmatch(endpoint)

	if len(ip) != 0 {
		return ip[0], nil
	}
	return "", errors.Errorf("Failed to find ip address in %s", endpoint)
}

func adaptExistedCluster(newNodes []kubekeyv1alpha2.NodeStatus, c *kubekeyv1alpha2.Cluster) error {
	var newMaster, newWorker, newEtcd []string
	for _, node := range newNodes {
		if node.Roles["etcd"] {
			newEtcd = append(newEtcd, node.Hostname)
		}
		if node.Roles["master"] {
			newMaster = append(newMaster, node.Hostname)
		}
		if node.Roles["worker"] {
			newWorker = append(newWorker, node.Hostname)
		}
	}

	c.Status.NodesCount = len(newNodes)
	c.Status.MasterCount = len(newMaster)
	c.Status.EtcdCount = len(newEtcd)
	c.Status.WorkerCount = len(newWorker)
	c.Status.Nodes = newNodes
	c.Status.Version = c.Spec.Kubernetes.Version
	c.Status.NetworkPlugin = c.Spec.Network.Plugin

	return nil
}

func otherClusterDiff(r *ClusterReconciler, ctx context.Context, c *kubekeyv1alpha2.Cluster) ([]kubekeyv1alpha2.NodeStatus, error) {
	newNodes := make([]kubekeyv1alpha2.NodeStatus, 0)

	allSpecHostsMap := make(map[string]NodeInfo)
	for _, host := range c.Spec.Hosts {
		allSpecHostsMap[host.InternalAddress] = NodeInfo{
			Name:    host.Name,
			Address: host.InternalAddress,
		}
	}

	kubeConfigFound := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-kubeconfig", c.Name), Namespace: c.Namespace}, kubeConfigFound); err != nil {
		if kubeErr.IsNotFound(err) {
			return newNodes, nil
		}
		return newNodes, err
	}

	kubeconfig, err := base64.StdEncoding.DecodeString(kubeConfigFound.Data["kubeconfig"])
	if err != nil {
		return newNodes, err
	}

	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return newNodes, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return newNodes, err
	}

	nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return newNodes, err
	}

	// get cluster nodes which in spec hosts
	currentHostsMap := make(map[string]NodeInfo)
	for _, node := range nodeList.Items {
		isMaster := false
		isWorker := false

		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			isMaster = true
		}
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			isMaster = true
		}
		if _, ok := node.Labels["node-role.kubernetes.io/worker"]; ok {
			isWorker = true
		}

		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeInternalIP {
				if _, ok := allSpecHostsMap[address.Address]; ok {
					currentHostsMap[address.Address] = NodeInfo{
						Address: address.Address,
						Name:    node.Name,
						Master:  isMaster,
						Worker:  isWorker,
					}
				}
			}
		}
	}

	// spec cluster is different from current cluster
	if len(currentHostsMap) == 0 {
		return newNodes, nil
	}

	if len(currentHostsMap) < len(nodeList.Items) {
		return newNodes, errors.New("the number of spec hosts are not enough compared with the cluster exist hosts")
	}

	// mark which ip or node name is etcd role
	etcdSpecMap := make(map[string]bool)
	kubeadmConfig, err := clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "kubeadm-config", metav1.GetOptions{})
	if err == nil {
		configuration, ok := kubeadmConfig.Data["ClusterConfiguration"]
		if !ok {
			return newNodes, errors.Errorf("Failed to get ClusterConfiguration key in kubeadm-config configmap")
		}

		conf := &Configuration{}
		if err := yamlV2.Unmarshal([]byte(configuration), conf); err != nil {
			return newNodes, err
		}

		for _, endpoint := range conf.Etcd.External.Endpoints {
			ip, err := findIpAddress(endpoint)
			if err != nil {
				return newNodes, errors.Wrap(err, "Failed to find etcd external endpoints")
			}
			etcdSpecMap[ip] = true
		}
	} else {
		if kubeErr.IsNotFound(err) {
			for _, etcdHostName := range c.Spec.RoleGroups["etcd"] {
				etcdSpecMap[etcdHostName] = true
			}
		} else {
			return newNodes, errors.Wrap(errors.WithStack(err), "Failed to get kubeadm-config")
		}
	}

	// add etcd node info to current hosts map
	for _, host := range allSpecHostsMap {

		_, nameOk := etcdSpecMap[host.Name]
		_, ipOk := etcdSpecMap[host.Address]
		if !nameOk && !ipOk {
			continue
		}
		if exist, ok := currentHostsMap[host.Address]; ok {
			exist.Etcd = true
			currentHostsMap[host.Address] = exist
		} else {
			currentHostsMap[host.Address] = NodeInfo{
				Address: host.Address,
				Name:    host.Name,
				Etcd:    true,
			}
		}
	}

	for _, host := range currentHostsMap {
		newNodes = append(newNodes, kubekeyv1alpha2.NodeStatus{
			InternalIP: host.Address,
			Hostname:   host.Name,
			Roles:      map[string]bool{"master": host.Master, "worker": host.Worker, "etcd": host.Etcd},
		})
	}
	return newNodes, nil
}
