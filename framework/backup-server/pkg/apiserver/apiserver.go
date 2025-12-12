package apiserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-openapi/spec"
	"olares.com/backup-server/pkg/apiserver/config"
	"olares.com/backup-server/pkg/apiserver/filters"
	"olares.com/backup-server/pkg/apiserver/response"
	apiruntime "olares.com/backup-server/pkg/apiserver/runtime"
	backupv1 "olares.com/backup-server/pkg/modules/backup/v1"
	"olares.com/backup-server/pkg/util/log"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
)

type APIServer struct {
	cfg *config.Config

	server *http.Server

	container *restful.Container
}

func New(cfg *config.Config) (*APIServer, error) {
	s := APIServer{
		cfg: cfg,
		server: &http.Server{
			Addr: cfg.ListenAddr,
		},
	}
	s.container = restful.NewContainer()

	return &s, nil
}

func (s *APIServer) PrepareRun() error {
	if s.container == nil {
		s.container = restful.NewContainer()
	}
	s.container.Filter(filters.Cors)
	s.container.RecoverHandler(filters.LogStackOnRecover)
	s.container.Filter(filters.LogRequestAndResponse)
	s.container.Filter(func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		defer func() {
			if e := recover(); e != nil {
				response.HandleInternalError(resp, errors.Errorf("internal error: %v", e))
			}
		}()

		chain.ProcessFilter(req, resp)
	})
	s.container.Filter(filters.Authenticate)

	s.container.Router(restful.CurlyRouter{})

	// sub modules
	s.installModuleAPI()
	s.installAPIDocs()

	s.server.Handler = s.container

	log.Info("registered modules:\n")
	for _, ws := range s.container.RegisteredWebServices() {
		fmt.Printf("  - %s\n", ws.RootPath())
	}
	fmt.Printf("\n")

	return nil
}

func (s *APIServer) Run(ctx context.Context) error {
	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		s.server.Shutdown(shutdownCtx)
	}()

	log.Infof("starting apiserver on %q", s.server.Addr)

	err := s.server.ListenAndServe()
	if err != nil {
		return errors.Wrapf(err, "http.server listen")
	}
	return nil
}

func (s *APIServer) installModuleAPI() {
	// add business models
	apiruntime.Must(backupv1.AddContainer(s.cfg, s.container))
}

func (s *APIServer) installAPIDocs() {
	config := restfulspec.Config{
		WebServices:                   s.container.RegisteredWebServices(),
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}

	s.container.Add(restfulspec.NewOpenAPIService(config))
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Backup Server",
			Description: "Go web boilerplate contains all the boilerplate you need to create a Go packages.",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "olares",
					Email: "olares@olares.com",
					URL:   "",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache License 2.0",
					URL:  "http://www.apache.org/licenses/LICENSE-2.0",
				},
			},
			Version: "0.1.1",
		},
	}
}
