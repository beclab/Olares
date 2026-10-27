package v2alpha1

import (
	"crypto/x509"
	"net/url"

	apiexternalversions "github.com/beclab/api/pkg/generated/informers/externalversions"
	applisters "github.com/beclab/api/pkg/generated/listers/app.bytetrade.io/v1alpha1"
	"github.com/brancz/kube-rbac-proxy/cmd/kube-rbac-proxy/app/options"
	"github.com/brancz/kube-rbac-proxy/pkg/proxy"
	"golang.org/x/net/http2"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type completedProxyRunOptions struct {
	insecureListenAddress string // DEPRECATED
	secureListenAddress   string
	proxyEndpointsPort    int

	upstreamURL      *url.URL
	upstreamForceH2C bool
	upstreamCABundle *x509.CertPool

	http2Disable bool
	http2Options *http2.Server

	auth *proxy.Config
	tls  *options.TLSConfig

	kubeClient *kubernetes.Clientset

	allowPaths  []string
	ignorePaths []string
	lldapServer string
	lldapPort   int

	informerFactory    informers.SharedInformerFactory
	appInformerFactory apiexternalversions.SharedInformerFactory
	appLister          applisters.ApplicationLister
}
