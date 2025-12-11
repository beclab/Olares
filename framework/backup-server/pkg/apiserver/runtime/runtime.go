package runtime

import (
	"fmt"

	"github.com/emicklei/go-restful/v3"
	"olares.com/backup-server/pkg/apiserver/config"
)

type ModuleVersion struct {
	Name    string
	Version string
}

func NewWebService(cfg *config.Config, mv ModuleVersion) *restful.WebService {
	ws := restful.WebService{}

	var routePath string

	if cfg.APIRoutePrefix == "/" {
		routePath = fmt.Sprintf("/%s/%s", mv.Name, mv.Version)
	} else {
		routePath = fmt.Sprintf("%s/%s/%s", cfg.APIRoutePrefix, mv.Name, mv.Version)
	}

	ws.Path(routePath).Produces(restful.MIME_JSON)

	return &ws
}
