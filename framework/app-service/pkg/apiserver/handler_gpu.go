package apiserver

import (
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"golang.org/x/exp/maps"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func (h *Handler) getGpuTypes(req *restful.Request, resp *restful.Response) {
	var nodes corev1.NodeList
	err := h.ctrlClient.List(req.Request.Context(), &nodes, &client.ListOptions{})
	if err != nil {
		klog.Errorf("list node failed %v", err)
		api.HandleError(resp, req, err)
		return
	}
	gpuTypes, err := utils.GetAllGpuTypesFromNodes(&nodes)
	if err != nil {
		klog.Errorf("get gpu type failed %v", err)
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteAsJson(&map[string]interface{}{
		"gpu_types": maps.Keys(gpuTypes),
	},
	)
}
