/*
 Copyright 2021 The KubeSphere Authors.

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

package v1alpha2

const (
	KubeKey  = "kubekey"
	Kubeadm  = "kubeadm"
	External = "external"
	MiniKube = "minikube"
)

type EtcdCluster struct {
	// Type of etcd cluster, can be set to 'kubekey' 'kubeadm' 'external'
	Type string `yaml:"type" json:"type,omitempty"`
	// ExternalEtcd describes how to connect to an external etcd cluster when type is set to external
	External         ExternalEtcd `yaml:"external" json:"external,omitempty"`
	BackupDir        string       `yaml:"backupDir" json:"backupDir,omitempty"`
	BackupPeriod     int          `yaml:"backupPeriod" json:"backupPeriod,omitempty"`
	KeepBackupNumber int          `yaml:"keepBackupNumber" json:"keepBackupNumber,omitempty"`
	BackupScriptDir  string       `yaml:"backupScript" json:"backupScript,omitempty"`
}

// ExternalEtcd describes how to connect to an external etcd cluster
// KubeKey, Kubeadm and External are mutually exclusive
type ExternalEtcd struct {
	// Endpoints of etcd members. Useful for using external etcd.
	// If not provided, kubeadm will run etcd in a static pod.
	Endpoints []string `yaml:"endpoints" json:"endpoints,omitempty"`
	// CAFile is an SSL Certificate Authority file used to secure etcd communication.
	CAFile string `yaml:"caFile" json:"caFile,omitempty"`
	// CertFile is an SSL certification file used to secure etcd communication.
	CertFile string `yaml:"certFile" json:"certFile,omitempty"`
	// KeyFile is an SSL key file used to secure etcd communication.
	KeyFile string `yaml:"keyFile" json:"keyFile,omitempty"`
}
