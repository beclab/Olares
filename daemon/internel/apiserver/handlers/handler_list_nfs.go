package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type ListNfsReq struct {
	Server string `json:"server"`
}

type nfsInfo struct {
	Path string `json:"path"`
	Acl  string `json:"acl"`
}

func (h *Handlers) PostListNfs(ctx *fiber.Ctx) error {
	var req ListNfsReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	// validate the server field is not empty or not a valid IP address or domain name
	if req.Server == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "server is empty")
	}

	if !utils.IsValidIP(req.Server) && !utils.IsValidDomain(req.Server) {
		return h.ErrJSON(ctx, http.StatusBadRequest, "server is not a valid IP address or domain name")
	}

	nfsList, err := utils.ListNfsDriver(ctx.Context(), req.Server)
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	infoRes := make([]*nfsInfo, 0, len(nfsList))
	for _, n := range nfsList {
		infoRes = append(infoRes, &nfsInfo{
			Path: n.Path,
			Acl:  n.Acl,
		})
	}

	return h.OkJSON(ctx, "success to list nfs server", infoRes)
}
