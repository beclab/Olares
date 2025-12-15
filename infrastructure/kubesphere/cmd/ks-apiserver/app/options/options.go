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

package options

import (
	"crypto/tls"
	"flag"
	"fmt"

	"k8s.io/client-go/kubernetes/scheme"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"kubesphere.io/kubesphere/pkg/apis"
	"kubesphere.io/kubesphere/pkg/apiserver"
	apiserverconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/informers"
	genericoptions "kubesphere.io/kubesphere/pkg/server/options"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"

	"net/http"
	"strings"

	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/metricsserver"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/prometheus"
)

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *genericoptions.ServerRunOptions
	*apiserverconfig.Config

	//
	DebugMode bool
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		Config:                  apiserverconfig.New(),
	}

	return s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "Don't enable this if you don't know what it means.")
	s.GenericServerRunOptions.AddFlags(fs, s.GenericServerRunOptions)
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.AuthorizationOptions.AddFlags(fss.FlagSet("authorization"), s.AuthorizationOptions)
	s.MonitoringOptions.AddFlags(fss.FlagSet("monitoring"), s.MonitoringOptions)
	s.LoggingOptions.AddFlags(fss.FlagSet("logging"), s.LoggingOptions)
	s.EventsOptions.AddFlags(fss.FlagSet("events"), s.EventsOptions)
	s.AlertingOptions.AddFlags(fss.FlagSet("alerting"), s.AlertingOptions)

	fs = fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return fss
}

const fakeInterface string = "FAKE"

// NewAPIServer creates an APIServer instance using given options
func (s *ServerRunOptions) NewAPIServer(stopCh <-chan struct{}) (*apiserver.APIServer, error) {
	apiServer := &apiserver.APIServer{
		Config: s.Config,
	}

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		return nil, err
	}
	apiServer.KubernetesClient = kubernetesClient

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes(), kubernetesClient.KubeSphere(),
		kubernetesClient.Istio(), kubernetesClient.Snapshot(), kubernetesClient.ApiExtensions(), kubernetesClient.Prometheus())
	apiServer.InformerFactory = informerFactory

	if s.MonitoringOptions == nil || len(s.MonitoringOptions.Endpoint) == 0 {
		return nil, fmt.Errorf("moinitoring service address in configuration MUST not be empty, please check configmap/kubesphere-config in kubesphere-system namespace")
	} else {
		monitoringClient, err := prometheus.NewPrometheus(s.MonitoringOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to prometheus, please check prometheus status, error: %v", err)
		}
		apiServer.MonitoringClient = monitoringClient
	}

	apiServer.MetricsClient = metricsserver.NewMetricsClient(kubernetesClient.Kubernetes(), s.KubernetesOptions)

	apiServer.CacheClient = cache.NewSimpleCache()

	if s.AlertingOptions != nil && (s.AlertingOptions.PrometheusEndpoint != "" || s.AlertingOptions.ThanosRulerEndpoint != "") {
		alertingClient, err := alerting.NewRuleClient(s.AlertingOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to init alerting client: %v", err)
		}
		apiServer.AlertingClient = alertingClient
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}

	if s.GenericServerRunOptions.SecurePort != 0 {
		certificate, err := tls.LoadX509KeyPair(s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey)
		if err != nil {
			return nil, err
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
		server.Addr = fmt.Sprintf(":%d", s.GenericServerRunOptions.SecurePort)
	}

	sch := scheme.Scheme
	if err := apis.AddToScheme(sch); err != nil {
		klog.Fatalf("unable add APIs to scheme: %v", err)
	}
	if err := extv1.AddToScheme(sch); err != nil {
		klog.Fatalf("unable add apiextensions v1 APIs to scheme: %v", err)
	}

	apiServer.RuntimeCache, err = runtimecache.New(apiServer.KubernetesClient.Config(), runtimecache.Options{Scheme: sch})
	if err != nil {
		klog.Fatalf("unable to create controller runtime cache: %v", err)
	}

	apiServer.RuntimeClient, err = runtimeclient.New(apiServer.KubernetesClient.Config(), runtimeclient.Options{Scheme: sch})
	if err != nil {
		klog.Fatalf("unable to create controller runtime client: %v", err)
	}

	apiServer.Server = server

	return apiServer, nil
}
