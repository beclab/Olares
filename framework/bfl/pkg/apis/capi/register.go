package capi

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
)

func AddContainer(c *restful.Container) error {
	handler, err := newHandler()
	if err != nil {
		return err
	}

	ws := new(restful.WebService)
	ws.Path("/capi").Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/app/detail").To(handler.appDetail))
	ws.Route(ws.GET("/app/myapps").To(handler.myApps))
	ws.Route(ws.GET("/system/ifs").To(handler.systemIFS))
	ws.Route(ws.GET("/monitoring.kubesphere.io/v1alpha3/{type}").To(handler.monitoringMetric))
	ws.Route(ws.GET("/namespaces/group").To(handler.namespaceGroup))

	c.Add(ws)
	return nil
}

func writeJSON(resp *restful.Response, status int, value any) {
	resp.WriteHeaderAndEntity(status, value)
}

func writeUpstreamError(resp *restful.Response, err error) {
	writeJSON(resp, http.StatusBadGateway, map[string]string{"message": err.Error()})
}
