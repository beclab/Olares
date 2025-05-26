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
	"context"
	"encoding/base64"
	"fmt"
	"text/template"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/addons"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/ending"
	"github.com/pkg/errors"

	kubekeyclientset "bytetrade.io/web3os/installer/clients/clientset/versioned"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/lithammer/dedent"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Pending representative the cluster is being created.
	Pending = "pending"
	// Success representative the cluster cluster has been created successfully.
	Success = "success"
	// Failed creation failure.
	Failed = "failed"
)

var (
	newNodes          []string
	clusterKubeSphere = template.Must(template.New("cluster.kubesphere.io").Parse(
		dedent.Dedent(`apiVersion: cluster.kubesphere.io/v1alpha1
kind: Cluster
metadata:
  finalizers:
  - finalizer.cluster.kubesphere.io
  labels:
    cluster-role.kubesphere.io/member: ""
    kubesphere.io/managed: "true"
    kubekey.kubesphere.io/name: {{ .Name }}
  name: {{ .Name }} 
spec:
  connection:
    kubeconfig: {{ .Kubeconfig }}
    type: direct
  enable: {{ .Enable }}
  joinFederation: {{ .JoinFedration }}
  provider: kubesphere
    `)))
)

func generateClusterKubeSphere(name string, kubeconfig string, enable, joinFedration bool) (string, error) {
	return util.Render(clusterKubeSphere, util.Data{
		"Name":          name,
		"Kubeconfig":    kubeconfig,
		"Enable":        enable,
		"JoinFedration": joinFedration,
	})
}

// CheckClusterRole is used to check the cluster's role (host or member).
func CheckClusterRole() (bool, *rest.Config, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return false, nil, err
	}
	// creates the clientset
	clientset, err := kube.NewForConfig(config)
	if err != nil {
		return false, nil, err
	}
	var hostClusterFlag bool
	if err := clientset.RESTClient().Get().
		AbsPath("/apis/cluster.kubesphere.io/v1alpha1/clusters").
		Name("host").
		Do(context.TODO()).Error(); err == nil {
		hostClusterFlag = true
	}
	return hostClusterFlag, config, nil
}

func newKubeSphereCluster(r *ClusterReconciler, c *kubekeyapiv1alpha2.Cluster) error {
	if hostClusterFlag, config, err := CheckClusterRole(); err != nil {
		return err
	} else if hostClusterFlag {
		obj, err := generateClusterKubeSphere(c.Name, "Cg==", false, false)
		if err != nil {
			return err
		}
		if err := addons.DoServerSideApply(context.TODO(), config, []byte(obj)); err != nil {
			err = r.Delete(context.TODO(), c)
			if err != nil {
				return err
			}
			return err
		}

		kscluster, err := addons.GetCluster(c.Name)
		if err != nil {
			return err
		}
		ownerReferencePatch := fmt.Sprintf(`{"metadata": {"ownerReferences": [{"apiVersion": "%s", "kind": "%s", "name": "%s", "uid": "%s"}]}}`, kscluster.GetAPIVersion(), kscluster.GetKind(), kscluster.GetName(), kscluster.GetUID())
		if err := r.Patch(context.TODO(), c, client.RawPatch(types.MergePatchType, []byte(ownerReferencePatch))); err != nil {
			return err
		}
	}
	return nil
}

// UpdateKubeSphereCluster is used to update the cluster object of KubeSphere's multicluster.
func UpdateKubeSphereCluster(runtime *common.KubeRuntime) error {
	if hostClusterFlag, config, err := CheckClusterRole(); err != nil {
		return err
	} else if hostClusterFlag {
		obj, err := generateClusterKubeSphere(runtime.ClusterName, runtime.Kubeconfig, true, true)
		if err != nil {
			return err
		}
		if err := addons.DoServerSideApply(context.TODO(), config, []byte(obj)); err != nil {
			return err
		}
	}
	return nil
}

// NewKubekeyClient is used to create a kubekey cluster client.
func NewKubekeyClient() (*kubekeyclientset.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset := kubekeyclientset.NewForConfigOrDie(config)

	return clientset, nil
}

func getCluster(name string) (*kubekeyapiv1alpha2.Cluster, error) {
	clientset, err := NewKubekeyClient()
	if err != nil {
		return nil, err
	}
	clusterObj, err := clientset.KubekeyV1alpha2().Clusters().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return clusterObj, nil
}

// UpdateClusterConditions is used for updating cluster installation process information or adding nodes.
func UpdateClusterConditions(runtime *common.KubeRuntime, step string, result *ending.ModuleResult) error {
	m := make(map[string]kubekeyapiv1alpha2.Event)
	allStatus := true
	for k, v := range result.HostResults {
		if v.GetStatus() == ending.FAILED {
			allStatus = false
		}
		e := kubekeyapiv1alpha2.Event{
			Step:   step,
			Status: v.GetStatus().String(),
		}
		if v.GetErr() != nil {
			e.Message = v.GetErr().Error()
		}
		m[k] = e
	}
	condition := kubekeyapiv1alpha2.Condition{
		Step:      step,
		StartTime: metav1.Time{Time: result.StartTime},
		EndTime:   metav1.Time{Time: result.EndTime},
		Status:    allStatus,
		Events:    m,
	}

	cluster, err := getCluster(runtime.ClusterName)
	if err != nil {
		return err
	}

	length := len(cluster.Status.Conditions)
	if length <= 0 {
		cluster.Status.Conditions = append(cluster.Status.Conditions, condition)
	} else if cluster.Status.Conditions[length-1].Step == condition.Step {
		cluster.Status.Conditions[length-1] = condition
	} else {
		cluster.Status.Conditions = append(cluster.Status.Conditions, condition)
	}

	cluster.Status.PiplineInfo.Status = "Running"
	if _, err := runtime.ClientSet.KubekeyV1alpha2().Clusters().UpdateStatus(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

// UpdateStatus is used to update status for a new or expanded cluster.
func UpdateStatus(runtime *common.KubeRuntime) error {
	cluster, err := getCluster(runtime.ClusterName)
	if err != nil {
		return err
	}
	cluster.Status.Version = runtime.Cluster.Kubernetes.Version
	cluster.Status.NodesCount = len(runtime.GetAllHosts())
	cluster.Status.MasterCount = len(runtime.GetHostsByRole(common.Master))
	cluster.Status.WorkerCount = len(runtime.GetHostsByRole(common.Worker))
	cluster.Status.EtcdCount = len(runtime.GetHostsByRole(common.ETCD))
	cluster.Status.NetworkPlugin = runtime.Cluster.Network.Plugin
	cluster.Status.Nodes = []kubekeyapiv1alpha2.NodeStatus{}

	for _, node := range runtime.GetAllHosts() {
		cluster.Status.Nodes = append(cluster.Status.Nodes, kubekeyapiv1alpha2.NodeStatus{
			InternalIP: node.GetInternalAddress(),
			Hostname:   node.GetName(),
			Roles:      map[string]bool{"etcd": node.IsRole(common.ETCD), "master": node.IsRole(common.Master), "worker": node.IsRole(common.Worker)},
		})
	}

	cluster.Status.PiplineInfo.Status = "Terminated"
	if _, err := runtime.ClientSet.KubekeyV1alpha2().Clusters().UpdateStatus(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func getClusterClientSet(runtime *common.KubeRuntime) (*kube.Clientset, error) {
	// creates the in-cluster config
	inClusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kube.NewForConfig(inClusterConfig)
	if err != nil {
		return nil, err
	}

	cm, err := clientset.
		CoreV1().
		ConfigMaps(runtime.Arg.Namespace).
		Get(context.TODO(), fmt.Sprintf("%s-kubeconfig", runtime.ClusterName), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	kubeConfigBase64, ok := cm.Data["kubeconfig"]
	if !ok {
		return nil, errors.Errorf("get kubeconfig from %s configmap failed", runtime.ClusterName)
	}

	kubeConfigStr, err := base64.StdEncoding.DecodeString(kubeConfigBase64)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.NewClientConfigFromBytes(kubeConfigStr)
	if err != nil {
		return nil, err
	}
	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}
	clientsetForCluster, err := kube.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return clientsetForCluster, nil
}

func nodeForCluster(name string, labels map[string]string) *corev1.Node {
	node := &corev1.Node{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec:   corev1.NodeSpec{},
		Status: corev1.NodeStatus{},
	}

	return node
}

// CreateNodeForCluster is used to create new nodes for the cluster to be add nodes.
func CreateNodeForCluster(runtime *common.KubeRuntime) error {
	clientsetForCluster, err := getClusterClientSet(runtime)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			return nil
		}
		return err
	}

	nodeInfo := make(map[string]string)
	nodeList, err := clientsetForCluster.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, node := range nodeList.Items {
		nodeInfo[node.Name] = node.Status.NodeInfo.KubeletVersion
	}

	for _, host := range runtime.GetHostsByRole(common.K8s) {
		if _, ok := nodeInfo[host.GetName()]; !ok {
			labels := map[string]string{"kubekey.kubesphere.io/import-status": Pending}
			if host.IsRole(common.Master) {
				labels["node-role.kubernetes.io/master"] = ""
			}
			if host.IsRole(common.Worker) {
				labels["node-role.kubernetes.io/worker"] = ""
			}
			node := nodeForCluster(host.GetName(), labels)
			if _, err = clientsetForCluster.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{}); err != nil {
				return err
			}
			newNodes = append(newNodes, host.GetName())
		}
	}

	return nil
}

// PatchNodeImportStatus is used to update new node's status.
func PatchNodeImportStatus(runtime *common.KubeRuntime, status string) error {
	clientsetForCluster, err := getClusterClientSet(runtime)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			return nil
		}
		return err
	}

	patchStr := fmt.Sprintf(`{"metadata": {"labels": {"kubekey.kubesphere.io/import-status": "%s"}}}`, status)
	nodeList, err := clientsetForCluster.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, node := range nodeList.Items {
		for k, v := range node.Labels {
			if k == "kubekey.kubesphere.io/import-status" && v != Success {
				_, err = clientsetForCluster.CoreV1().Nodes().Patch(context.TODO(), node.Name, types.StrategicMergePatchType, []byte(patchStr), metav1.PatchOptions{})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// SaveKubeConfig is used to save the kubeconfig for the new cluster.
func SaveKubeConfig(runtime *common.KubeRuntime) error {
	// creates the in-cluster config
	inClusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	// creates the clientset
	clientset, err := kube.NewForConfig(inClusterConfig)
	if err != nil {
		return err
	}
	cmClientset := clientset.CoreV1().ConfigMaps(runtime.Arg.Namespace)

	if _, err := cmClientset.Get(context.TODO(), fmt.Sprintf("%s-kubeconfig", runtime.ClusterName), metav1.GetOptions{}); err != nil {
		if kubeErr.IsNotFound(err) {
			_, err = cmClientset.Create(context.TODO(), configMapForKubeconfig(runtime), metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		kubeconfigStr := fmt.Sprintf(`{"kubeconfig": "%s"}`, runtime.Kubeconfig)
		_, err = cmClientset.Patch(context.TODO(), runtime.ClusterName, types.ApplyPatchType, []byte(kubeconfigStr), metav1.PatchOptions{})
		if err != nil {
			return err
		}
	}
	// clientset.CoreV1().ConfigMaps("kubekey-system").Create(context.TODO(), kubeconfigConfigMap, metav1.CreateOptions{}
	return nil
}

// configMapForKubeconfig is used to generate configmap scheme for cluster's kubeconfig.
func configMapForKubeconfig(runtime *common.KubeRuntime) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-kubeconfig", runtime.ClusterName),
		},
		Data: map[string]string{
			"kubeconfig": runtime.Kubeconfig,
		},
	}

	return cm
}

func ClearConditions(runtime *common.KubeRuntime) error {
	cluster, err := getCluster(runtime.ClusterName)
	if err != nil {
		return err
	}
	cluster.Status.Conditions = make([]kubekeyapiv1alpha2.Condition, 0)

	if _, err := runtime.ClientSet.KubekeyV1alpha2().Clusters().UpdateStatus(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}
