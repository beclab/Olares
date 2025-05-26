/*
Copyright 2021.

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

package v1alpha1

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Cluster. Edit Cluster_types.go to remove/update
	Hosts                []HostCfg            `yaml:"hosts" json:"hosts,omitempty"`
	RoleGroups           RoleGroups           `yaml:"roleGroups" json:"roleGroups,omitempty"`
	ControlPlaneEndpoint ControlPlaneEndpoint `yaml:"controlPlaneEndpoint" json:"controlPlaneEndpoint,omitempty"`
	Kubernetes           Kubernetes           `yaml:"kubernetes" json:"kubernetes,omitempty"`
	Network              NetworkConfig        `yaml:"network" json:"network,omitempty"`
	Registry             RegistryConfig       `yaml:"registry" json:"registry,omitempty"`
	Addons               []Addon              `yaml:"addons" json:"addons,omitempty"`
	KubeSphere           KubeSphere           `json:"kubesphere,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	JobInfo       JobInfo      `json:"jobInfo,omitempty"`
	Version       string       `json:"version,omitempty"`
	NetworkPlugin string       `json:"networkPlugin,omitempty"`
	NodesCount    int          `json:"nodesCount,omitempty"`
	EtcdCount     int          `json:"etcdCount,omitempty"`
	MasterCount   int          `json:"masterCount,omitempty"`
	WorkerCount   int          `json:"workerCount,omitempty"`
	Nodes         []NodeStatus `json:"nodes,omitempty"`
	Conditions    []Condition  `json:"Conditions,omitempty"`
}

// JobInfo defines the job information to be used to create a cluster or add a node.
type JobInfo struct {
	Namespace string    `json:"namespace,omitempty"`
	Name      string    `json:"name,omitempty"`
	Pods      []PodInfo `json:"pods,omitempty"`
}

// PodInfo defines the pod information to be used to create a cluster or add a node.
type PodInfo struct {
	Name       string          `json:"name,omitempty"`
	Containers []ContainerInfo `json:"containers,omitempty"`
}

// ContainerInfo defines the container information to be used to create a cluster or add a node.
type ContainerInfo struct {
	Name string `json:"name,omitempty"`
}

// NodeStatus defines the status information of the nodes in the cluster.
type NodeStatus struct {
	InternalIP string          `json:"internalIP,omitempty"`
	Hostname   string          `json:"hostname,omitempty"`
	Roles      map[string]bool `json:"roles,omitempty"`
}

// Condition defines the process information.
type Condition struct {
	Step      string      `json:"step,omitempty"`
	StartTime metav1.Time `json:"startTime,omitempty"`
	EndTime   metav1.Time `json:"endTime,omitempty"`
	Status    bool        `json:"status,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
// +kubebuilder:resource:path=clusters,scope=Cluster
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}

// HostCfg defines host information for cluster.
type HostCfg struct {
	Name            string `yaml:"name,omitempty" json:"name,omitempty"`
	Address         string `yaml:"address,omitempty" json:"address,omitempty"`
	InternalAddress string `yaml:"internalAddress,omitempty" json:"internalAddress,omitempty"`
	Port            int    `yaml:"port,omitempty" json:"port,omitempty"`
	User            string `yaml:"user,omitempty" json:"user,omitempty"`
	Password        string `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKey      string `yaml:"privateKey,omitempty" json:"privateKey,omitempty"`
	PrivateKeyPath  string `yaml:"privateKeyPath,omitempty" json:"privateKeyPath,omitempty"`
	Arch            string `yaml:"arch,omitempty" json:"arch,omitempty"`

	Labels   map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	ID       string            `yaml:"id,omitempty" json:"id,omitempty"`
	Index    int               `json:"-"`
	IsEtcd   bool              `json:"-"`
	IsMaster bool              `json:"-"`
	IsWorker bool              `json:"-"`

	EtcdExist bool   `json:"-"`
	EtcdName  string `json:"-"`
}

// RoleGroups defines the grouping of role for hosts (etcd / master / worker).
type RoleGroups struct {
	Etcd   []string `yaml:"etcd" json:"etcd,omitempty"`
	Master []string `yaml:"master" json:"master,omitempty"`
	Worker []string `yaml:"worker" json:"worker,omitempty"`
}

// HostGroups defines the grouping of hosts for cluster (all / etcd / master / worker / k8s).
type HostGroups struct {
	All    []HostCfg
	Etcd   []HostCfg
	Master []HostCfg
	Worker []HostCfg
	K8s    []HostCfg
}

// ControlPlaneEndpoint defines the control plane endpoint information for cluster.
type ControlPlaneEndpoint struct {
	InternalLoadbalancer string `yaml:"internalLoadbalancer" json:"internalLoadbalancer,omitempty"`
	Domain               string `yaml:"domain" json:"domain,omitempty"`
	Address              string `yaml:"address" json:"address,omitempty"`
	Port                 int    `yaml:"port" json:"port,omitempty"`
}

// RegistryConfig defines the configuration information of the image's repository.
type RegistryConfig struct {
	RegistryMirrors    []string `yaml:"registryMirrors" json:"registryMirrors,omitempty"`
	InsecureRegistries []string `yaml:"insecureRegistries" json:"insecureRegistries,omitempty"`
	PrivateRegistry    string   `yaml:"privateRegistry" json:"privateRegistry,omitempty"`
}

// KubeSphere defines the configuration information of the KubeSphere.
type KubeSphere struct {
	Enabled        bool   `json:"enabled,omitempty"`
	Version        string `json:"version,omitempty"`
	Configurations string `json:"configurations,omitempty"`
}

// ExternalEtcd defines configuration information of external etcd.
type ExternalEtcd struct {
	Endpoints []string
	CaFile    string
	CertFile  string
	KeyFile   string
}

// Copy is used to create a copy for Runtime.
func (h *HostCfg) Copy() *HostCfg {
	host := *h
	return &host
}

// GenerateCertSANs is used to generate cert sans for cluster.
func (cfg *ClusterSpec) GenerateCertSANs() []string {
	clusterSvc := fmt.Sprintf("kubernetes.default.svc.%s", cfg.Kubernetes.ClusterName)
	defaultCertSANs := []string{"kubernetes", "kubernetes.default", "kubernetes.default.svc", clusterSvc, "localhost", "127.0.0.1"}
	extraCertSANs := []string{}

	extraCertSANs = append(extraCertSANs, cfg.ControlPlaneEndpoint.Domain)
	extraCertSANs = append(extraCertSANs, cfg.ControlPlaneEndpoint.Address)

	for _, host := range cfg.Hosts {
		extraCertSANs = append(extraCertSANs, host.Name)
		extraCertSANs = append(extraCertSANs, fmt.Sprintf("%s.%s", host.Name, cfg.Kubernetes.ClusterName))
		if host.Address != cfg.ControlPlaneEndpoint.Address {
			extraCertSANs = append(extraCertSANs, host.Address)
		}
		if host.InternalAddress != host.Address && host.InternalAddress != cfg.ControlPlaneEndpoint.Address {
			extraCertSANs = append(extraCertSANs, host.InternalAddress)
		}
	}

	extraCertSANs = append(extraCertSANs, util.ParseIp(cfg.Network.KubeServiceCIDR)[0])

	defaultCertSANs = append(defaultCertSANs, extraCertSANs...)

	if cfg.Kubernetes.ApiserverCertExtraSans != nil {
		defaultCertSANs = append(defaultCertSANs, cfg.Kubernetes.ApiserverCertExtraSans...)
	}

	return defaultCertSANs
}

// GroupHosts is used to group hosts according to the configuration file.s
func (cfg *ClusterSpec) GroupHosts() (*HostGroups, error) {
	clusterHostsGroups := HostGroups{}

	hostList := map[string]string{}
	for _, host := range cfg.Hosts {
		hostList[host.Name] = host.Name
	}

	etcdGroup, masterGroup, workerGroup, err := cfg.ParseRolesList(hostList)
	if err != nil {
		return nil, err
	}
	for index, host := range cfg.Hosts {
		host.Index = index
		if len(etcdGroup) > 0 {
			for _, hostName := range etcdGroup {
				if host.Name == hostName {
					host.IsEtcd = true
					break
				}
			}
		}

		if len(masterGroup) > 0 {
			for _, hostName := range masterGroup {
				if host.Name == hostName {
					host.IsMaster = true
					break
				}
			}
		}

		if len(workerGroup) > 0 {
			for _, hostName := range workerGroup {
				if hostName != "" && host.Name == hostName {
					host.IsWorker = true
					break
				}
			}
		}

		if host.IsEtcd {
			clusterHostsGroups.Etcd = append(clusterHostsGroups.Etcd, host)
		}
		if host.IsMaster {
			clusterHostsGroups.Master = append(clusterHostsGroups.Master, host)
		}
		if host.IsWorker {
			clusterHostsGroups.Worker = append(clusterHostsGroups.Worker, host)
		}
		if host.IsMaster || host.IsWorker {
			clusterHostsGroups.K8s = append(clusterHostsGroups.K8s, host)
		}
		clusterHostsGroups.All = append(clusterHostsGroups.All, host)
	}

	//Check that the parameters under roleGroups are incorrect
	if len(masterGroup) == 0 {
		logger.Fatal(errors.New("The number of master cannot be 0"))
	}
	if len(etcdGroup) == 0 {
		logger.Fatal(errors.New("The number of etcd cannot be 0"))
	}

	if len(masterGroup) != len(clusterHostsGroups.Master) {
		return nil, errors.New("Incorrect nodeName under roleGroups/master in the configuration file")
	}
	if len(etcdGroup) != len(clusterHostsGroups.Etcd) {
		return nil, errors.New("Incorrect nodeName under roleGroups/etcd in the configuration file")
	}
	if len(workerGroup) != len(clusterHostsGroups.Worker) {
		return nil, errors.New("Incorrect nodeName under roleGroups/work in the configuration file")
	}

	return &clusterHostsGroups, nil
}

// ClusterIP is used to get the kube-apiserver service address inside the cluster.
func (cfg *ClusterSpec) ClusterIP() string {
	return util.ParseIp(cfg.Network.KubeServiceCIDR)[0]
}

// CorednsClusterIP is used to get the coredns service address inside the cluster.
func (cfg *ClusterSpec) CorednsClusterIP() string {
	return util.ParseIp(cfg.Network.KubeServiceCIDR)[2]
}

// ClusterDNS is used to get the dns server address inside the cluster.
func (cfg *ClusterSpec) ClusterDNS() string {
	if cfg.Kubernetes.EnableNodelocaldns() {
		return "169.254.25.10"
	}
	return cfg.CorednsClusterIP()
}

// ParseRolesList is used to parse the host grouping list.
func (cfg *ClusterSpec) ParseRolesList(hostList map[string]string) ([]string, []string, []string, error) {
	etcdGroupList := []string{}
	masterGroupList := []string{}
	workerGroupList := []string{}

	for _, host := range cfg.RoleGroups.Etcd {
		if strings.Contains(host, "[") && strings.Contains(host, "]") && strings.Contains(host, ":") {
			etcdGroupList = append(etcdGroupList, getHostsRange(host, hostList, "etcd")...)
		} else {
			if err := hostVerify(hostList, host, "etcd"); err != nil {
				logger.Fatal(err)
			}
			etcdGroupList = append(etcdGroupList, host)
		}
	}

	for _, host := range cfg.RoleGroups.Master {
		if strings.Contains(host, "[") && strings.Contains(host, "]") && strings.Contains(host, ":") {
			masterGroupList = append(masterGroupList, getHostsRange(host, hostList, "master")...)
		} else {
			if err := hostVerify(hostList, host, "master"); err != nil {
				logger.Fatal(err)
			}
			masterGroupList = append(masterGroupList, host)
		}
	}

	for _, host := range cfg.RoleGroups.Worker {
		if strings.Contains(host, "[") && strings.Contains(host, "]") && strings.Contains(host, ":") {
			workerGroupList = append(workerGroupList, getHostsRange(host, hostList, "worker")...)
		} else {
			if err := hostVerify(hostList, host, "worker"); err != nil {
				logger.Fatal(err)
			}
			workerGroupList = append(workerGroupList, host)
		}
	}
	return etcdGroupList, masterGroupList, workerGroupList, nil
}

func getHostsRange(rangeStr string, hostList map[string]string, group string) []string {
	hostRangeList := []string{}
	r := regexp.MustCompile(`\[(\d+)\:(\d+)\]`)
	nameSuffix := r.FindStringSubmatch(rangeStr)
	namePrefix := strings.Split(rangeStr, nameSuffix[0])[0]
	nameSuffixStart, _ := strconv.Atoi(nameSuffix[1])
	nameSuffixEnd, _ := strconv.Atoi(nameSuffix[2])
	for i := nameSuffixStart; i <= nameSuffixEnd; i++ {
		if err := hostVerify(hostList, fmt.Sprintf("%s%d", namePrefix, i), group); err != nil {
			logger.Fatal(err)
		}
		hostRangeList = append(hostRangeList, fmt.Sprintf("%s%d", namePrefix, i))
	}
	return hostRangeList
}

func hostVerify(hostList map[string]string, hostName string, group string) error {
	if _, ok := hostList[hostName]; !ok {
		return fmt.Errorf("[%s] is in [%s] group, but not in hosts list", hostName, group)
	}
	return nil
}

const (
	Haproxy = "haproxy"
)

func (c ControlPlaneEndpoint) IsInternalLBEnabled() bool {
	if c.InternalLoadbalancer == Haproxy {
		return true
	}
	return false
}
