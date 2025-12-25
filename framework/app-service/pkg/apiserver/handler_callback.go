package apiserver

import (
	"sync"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appwatchers"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
)

var singleTask *sync.Once = &sync.Once{}

func (h *Handler) highload(req *restful.Request, resp *restful.Response) {
	var load struct {
		CPU    float64 `json:"cpu"`
		Memory float64 `json:"memory"`
	}

	err := req.ReadEntity(&load)
	if err != nil {
		klog.Errorf("Failed to read request err=%v", err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	klog.Infof("System resources high load cpu=%v memory=%v", load.CPU, load.Memory)

	// start application suspending task
	singleTask.Do(func() {
		go func() {
			err := appwatchers.SuspendTopApp(h.serviceCtx, h.ctrlClient)
			if err != nil {
				klog.Errorf("Failed to suspend applications err=%v", err)
			}
			singleTask = &sync.Once{}
		}()
	})

	resp.WriteAsJson(map[string]int{"code": 0})

}

var userSingleTask = &sync.Once{}

func (h *Handler) userHighLoad(req *restful.Request, resp *restful.Response) {
	var load struct {
		CPU    float64 `json:"cpu"`
		Memory float64 `json:"memory"`
		User   string  `json:"user"`
	}

	err := req.ReadEntity(&load)
	if err != nil {
		klog.Errorf("Failed to read request err=%v", err)
		api.HandleBadRequest(resp, req, err)
		return
	}
	klog.Infof("User: %s resources high load, cpu %.2f, mem %.2f", load.User, load.CPU, load.Memory)

	userSingleTask.Do(func() {
		go func() {
			err := appwatchers.SuspendUserTopApp(h.serviceCtx, h.ctrlClient, load.User)
			if err != nil {
				klog.Errorf("Failed to suspend application user=%s err=%v", load.User, err)
			}
			userSingleTask = &sync.Once{}
		}()
	})
	resp.WriteAsJson(map[string]int{"code": 0})
}
