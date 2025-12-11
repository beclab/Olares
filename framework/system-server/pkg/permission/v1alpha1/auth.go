package permission

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/constants"
	"bytetrade.io/web3os/system-server/pkg/utils"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
)

func (h *Handler) nonce(req *restful.Request, resp *restful.Response) {
	h.limitIP(req, resp, func(req *restful.Request, resp *restful.Response) {
		resp.Write([]byte(constants.Nonce))
	})
}

func (h *Handler) limitIP(req *restful.Request, resp *restful.Response, next func(req *restful.Request, resp *restful.Response)) {
	authIps, err := h.getAutheliaIP(req.Request.Context())
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	ip, err := getIPFromReq(req.Request)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	if !utils.ListContains(authIps, ip.String()) {
		api.HandleForbidden(resp, req, fmt.Errorf("request from %s is forbidden", ip.String()))
		return
	}

	next(req, resp)
}

func (h *Handler) getAutheliaIP(ctx context.Context) ([]string, error) {
	pods, err := h.kubeClientSet.CoreV1().Pods("os-framework").List(ctx, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			labels.Set{
				"app": "authelia-backend",
			},
		).String(),
	})

	if err != nil {
		klog.Error("get authelia pods error, ", err)
		return nil, err
	}

	var ips []string
	for _, p := range pods.Items {
		ips = append(ips, p.Status.PodIP)
	}

	return ips, nil
}

func getIPFromReq(req *http.Request) (net.IP, error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}

	return userIP, nil
}
