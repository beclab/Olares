package v1alpha1

import (
	"bytetrade.io/web3os/bfl/pkg/api/response"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
)

const BackupCancelCode = 493

type handler struct {
}

func newHandler() *handler {
	return &handler{}
}

func (h *handler) backupNew(req *restful.Request, resp *restful.Response) {
	klog.Info("backup start callback")
	for _, cb := range callbackHandlers {
		err := cb.BackupNew()
		if err != nil {
			klog.Error("backup start response error, ", err)
			resp.WriteError(BackupCancelCode, err)
			return
		}
	}

	response.SuccessNoData(resp)
}

func (h *handler) backupFinish(req *restful.Request, resp *restful.Response) {
	klog.Info("backup finished callback")
	for _, cb := range callbackHandlers {
		err := cb.BackupFinish()
		if err != nil {
			klog.Warning("backup finished callback error, ", err)
		}
	}

	response.SuccessNoData(resp)
}
