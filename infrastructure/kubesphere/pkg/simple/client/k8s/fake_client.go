/*
Copyright 2020 KubeSphere Authors

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

package k8s

import (
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

type FakeClient struct {
	// kubernetes client interface
	K8sClient kubernetes.Interface

	// discovery client
	DiscoveryClient *discovery.DiscoveryClient

	// generated clientset
	KubeSphereClient kubesphere.Interface

	IstioClient istioclient.Interface

	SnapshotClient snapshotclient.Interface

	ApiExtensionClient apiextensionsclient.Interface

	prometheusClient promresourcesclient.Interface

	MasterURL string

	KubeConfig *rest.Config
}
