package apiserver

import (
	"net/http"
	"os"

	"bytetrade.io/web3os/terminusd/pkg/nets"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

func (h *handlers) GetHostsfile(ctx *fiber.Ctx) error {
	items, err := nets.GetHostsFile()
	if err != nil {
		return h.ErrJSON(ctx, http.StatusServiceUnavailable, err.Error())
	}

	return h.OkJSON(ctx, "", items)
}

type writeHostsfileReq struct {
	Items []*nets.HostsItem `json:"items"`
}

func (h *handlers) PostHostsfile(ctx *fiber.Ctx) error {
	var req writeHostsfileReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	for _, i := range req.Items {
		for _, b := range blackList {
			if b == i.Host {
				return h.ErrJSON(ctx, http.StatusBadRequest, "cannot modify the host "+i.Host)
			}
		}
	}

	if err := nets.ForceUpdateHostsFile(req.Items); err != nil {
		klog.Error("write hosts error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	return h.OkJSON(ctx, "success to write hosts file")
}

var (
	blackList []string = []string{
		"localhost",
		"lb.kubesphere.local",
	}
)

func init() {
	if hosts, err := os.Hostname(); err == nil {
		blackList = append(blackList, hosts)
	}
}
