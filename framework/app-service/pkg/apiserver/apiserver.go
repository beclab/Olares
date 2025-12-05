package apiserver

import (
	"context"
	"net/http"
	"os"
	"time"

	"bytetrade.io/web3os/app-service/pkg/constants"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultCertPath = "/etc/certs/server.crt"
	defaultKeyPath  = "/etc/certs/server.key"
	tlsCertEnv      = "WEBHOOK_TLS_CERT"
	tlsKeyEnv       = "WEBHOOK_TLS_KEY"
)

var apiHandler *Handler

// APIServer represents an API server for system.
type APIServer struct {
	Server    *http.Server
	SSLServer *http.Server

	// RESTful Server
	container *restful.Container

	serverCtx context.Context
}

// New returns an APIServer.
func New(ctx context.Context) (*APIServer, error) {
	server := &http.Server{
		Addr: constants.APIServerListenAddress,
	}
	sslServer := &http.Server{
		Addr: constants.WebhookServerListenAddress,
	}

	return &APIServer{
		Server:    server,
		SSLServer: sslServer,
		container: restful.NewContainer(),
		serverCtx: ctx,
	}, nil
}

// PrepareRun do prepares for API server.
func (s *APIServer) PrepareRun(ksHost string, kubeConfig *rest.Config, client client.Client, stopCh <-chan struct{}) (err error) {
	s.container.Filter(logRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	// use the server context for goroutine in background
	apiHandlerBuilder := &handlerBuilder{}
	apiHandlerBuilder.WithContext(s.serverCtx).
		WithKubesphereConfig(ksHost).
		WithKubernetesConfig(kubeConfig).
		WithCtrlClient(client).
		WithAppInformer()
	apiHandler, err = apiHandlerBuilder.Build()
	if err != nil {
		return err
	}
	go apiHandler.opController.run()

	err = apiHandler.Run(stopCh)
	if err != nil {
		klog.Infof("wait for cache sync failed %v", err)
		return err
	}
	err = addServiceToContainer(s.container, apiHandler)
	if err != nil {
		return err
	}

	s.installAPIDocs()

	s.Server.Handler = s.container
	s.SSLServer.Handler = s.container

	return nil
}

// Run running a server.
func (s *APIServer) Run() error {
	shutdownCtx, cancel := context.WithTimeout(s.serverCtx, 2*time.Minute)
	defer cancel()

	go func() {
		<-s.serverCtx.Done()
		_ = s.Server.Shutdown(shutdownCtx)
		_ = s.SSLServer.Shutdown(shutdownCtx)
		ctrl.Log.Info("Shutdown apiserver for app-service")
	}()

	go func() {
		tlsCert, tlsKey := defaultCertPath, defaultKeyPath
		if os.Getenv(tlsCertEnv) != "" && os.Getenv(tlsKeyEnv) != "" {
			tlsCert, tlsKey = os.Getenv(tlsCertEnv), os.Getenv(tlsKeyEnv)
		}
		ctrl.Log.Info("Starting webhook server for app-service", "listen", constants.WebhookServerListenAddress)
		if err := s.SSLServer.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
			ctrl.Log.Error(err, "Failed to start webhook server for app-service")
		}
	}()
	ctrl.Log.Info("Starting server for app-service", "listen", constants.APIServerListenAddress)

	return s.Server.ListenAndServe()
}

func (s *APIServer) installAPIDocs() {
	config := restfulspec.Config{
		WebServices:                   s.container.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "/app-service/v1/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	s.container.Add(restfulspec.NewOpenAPIService(config))
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "app-service",
			Description: "application service, running in background",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "bytetrade",
					Email: "dev@bytetrade.io",
					URL:   "http://bytetrade.io",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache License 2.0",
					URL:  "http://www.apache.org/licenses/LICENSE-2.0",
				},
			},
			Version: "0.1.0",
		},
	}
	swo.Tags = []spec.Tag{{TagProps: spec.TagProps{
		Name:        "app-service",
		Description: "Web 3 OS app-service"}}}
}
