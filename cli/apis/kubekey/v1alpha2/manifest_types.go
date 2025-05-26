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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Iso struct {
	LocalPath string `yaml:"localPath" json:"localPath"`
	Url       string `yaml:"url" json:"url"`
}

type Repository struct {
	Iso Iso `yaml:"iso" json:"iso"`
}

type OperatingSystem struct {
	Arch       string     `yaml:"arch" json:"arch"`
	Type       string     `yaml:"type" json:"type,omitempty"`
	Id         string     `yaml:"id" json:"id"`
	Version    string     `yaml:"version" json:"version"`
	OsImage    string     `yaml:"osImage" json:"osImage"`
	Repository Repository `yaml:"repository" json:"repository"`
}

type KubernetesDistribution struct {
	Type    string `yaml:"type" json:"type"`
	Version string `yaml:"version" json:"version"`
}

type Helm struct {
	Version string `yaml:"version" json:"version"`
}

type CNI struct {
	Version string `yaml:"version" json:"version"`
}

type ETCD struct {
	Version string `yaml:"version" json:"version"`
}

type DockerManifest struct {
	Version string `yaml:"version" json:"version"`
}

type Crictl struct {
	Version string `yaml:"version" json:"version"`
}

type ContainerRuntime struct {
	Type    string `yaml:"type" json:"type"`
	Version string `yaml:"version" json:"version"`
}

type DockerRegistry struct {
	Version string `yaml:"version" json:"version"`
}

type Harbor struct {
	Version string `yaml:"version" json:"version"`
}

type DockerCompose struct {
	Version string `yaml:"version" json:"version"`
}

type Components struct {
	Helm              Helm               `yaml:"helm" json:"helm"`
	CNI               CNI                `yaml:"cni" json:"cni"`
	ETCD              ETCD               `yaml:"etcd" json:"etcd"`
	ContainerRuntimes []ContainerRuntime `yaml:"containerRuntimes" json:"containerRuntimes"`
	Crictl            Crictl             `yaml:"crictl" json:"crictl,omitempty"`
	DockerRegistry    DockerRegistry     `yaml:"docker-registry" json:"docker-registry"`
	Harbor            Harbor             `yaml:"harbor" json:"harbor"`
	DockerCompose     DockerCompose      `yaml:"docker-compose" json:"docker-compose"`
}

type ManifestRegistry struct {
	Auths runtime.RawExtension `yaml:"auths" json:"auths,omitempty"`
}

// ManifestSpec defines the desired state of Manifest
type ManifestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Arches                  []string                 `yaml:"arches" json:"arches"`
	OperatingSystems        []OperatingSystem        `yaml:"operatingSystems" json:"operatingSystems"`
	KubernetesDistributions []KubernetesDistribution `yaml:"kubernetesDistributions" json:"kubernetesDistributions"`
	Components              Components               `yaml:"components" json:"components"`
	Images                  []string                 `yaml:"images" json:"images"`
	ManifestRegistry        ManifestRegistry         `yaml:"registry" json:"registry"`
}

// ManifestStatus defines the observed state of Manifest
type ManifestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Manifest is the Schema for the manifests API
type Manifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManifestSpec   `json:"spec,omitempty"`
	Status ManifestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ManifestList contains a list of Manifest
type ManifestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Manifest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Manifest{}, &ManifestList{})
}
