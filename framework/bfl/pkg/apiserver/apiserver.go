package apiserver

import (
	"context"
	"net/http"

	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/api/response"
	backendv1 "bytetrade.io/web3os/bfl/pkg/apis/backend/v1"
	callbackV1alpha1 "bytetrade.io/web3os/bfl/pkg/apis/callback/v1alpha1"

	iamV1alpha1 "bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1"
	monitov1alpha1 "bytetrade.io/web3os/bfl/pkg/apis/monitor/v1alpha1"
	settingsV1alpha1 "bytetrade.io/web3os/bfl/pkg/apis/settings/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	v1alpha1client "bytetrade.io/web3os/bfl/pkg/client/clientset/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
)

type APIServer struct {
	Server *http.Server

	container *restful.Container

	kubeClient v1alpha1client.ClientInterface
}

func New() (*APIServer, error) {
	s := &APIServer{}

	// new kubeclient
	client, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return nil, errors.Errorf("new kubeclient in cluster err: %v", err)
	}
	s.kubeClient = client

	// jwt key
	if err := s.initFetchKsJwtKey(); err != nil {
		return nil, err
	}

	server := &http.Server{
		Addr: constants.APIServerListenAddress,
	}

	s.Server = server
	return s, nil
}

func (s *APIServer) initFetchKsJwtKey() error {
	secret, err := s.kubeClient.Kubernetes().CoreV1().Secrets("os-platform").Get(context.TODO(), "lldap-credentials", metav1.GetOptions{})
	if err != nil {
		return err
	}
	jwtSecretKey := secret.Data["lldap-jwt-secret"]
	constants.KubeSphereJwtKey = jwtSecretKey
	return nil
}

func (s *APIServer) PrepareRun() error {
	s.container = restful.NewContainer()
	s.container.Filter(logRequestAndResponse)
	s.container.Filter(cors)
	s.container.RecoverHandler(logStackOnRecover)
	s.container.Filter(func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		defer func() {
			if e := recover(); e != nil {
				response.HandleInternalError(resp, errors.Errorf("server internal error: %v", e))
			}
		}()

		chain.ProcessFilter(req, resp)
	})
	s.container.Filter(authenticate)

	s.container.Router(restful.CurlyRouter{})

	s.installModuleAPI()
	s.installAPIDocs()

	var modulePaths []string
	for _, ws := range s.container.RegisteredWebServices() {
		modulePaths = append(modulePaths, ws.RootPath())
	}
	log.Infow("registered module", "paths", modulePaths)

	s.Server.Handler = s.container

	return nil
}

func (s *APIServer) Run() error {
	err := s.Server.ListenAndServe()
	if err != nil {
		return errors.Errorf("listen and serve err: %v", err)
	}
	return nil
}

func (s *APIServer) installAPIDocs() {
	config := restfulspec.Config{
		WebServices:                   s.container.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "/bfl/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	s.container.Add(restfulspec.NewOpenAPIService(config))
}

func (s *APIServer) installModuleAPI() {
	urlruntime.Must(callbackV1alpha1.AddToContainer(s.container))
	urlruntime.Must(iamV1alpha1.AddToContainer(s.container, callbackV1alpha1.AddBackupCallbackHandler))
	urlruntime.Must(backendv1.AddContainer(s.container))
	urlruntime.Must(settingsV1alpha1.AddContainer(s.container))
	urlruntime.Must(monitov1alpha1.AddContainer(s.container))
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "BFL",
			Description: "Backend For Launcher",
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
			Version: "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{{TagProps: spec.TagProps{
		Name:        "Launcher",
		Description: "Web 3 OS Launcher"}}}
}
