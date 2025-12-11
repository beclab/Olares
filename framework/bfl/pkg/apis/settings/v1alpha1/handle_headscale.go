package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/api/response"
	v1alpha1client "bytetrade.io/web3os/bfl/pkg/client/clientset/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/client/dynamic_client"
	"bytetrade.io/web3os/bfl/pkg/client/dynamic_client/apps"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

type AclState string

const (
	AclStateApplying AclState = "applying"
	AclStateApplied  AclState = "applied"

	ENV_HEADSCALE_ACL_SSH = "HEADSCALE_ACL_SSH"
	systemAppName         = "olares-app"
)

type SshAcl struct {
	AllowSSH bool     `json:"allow_ssh"`
	State    AclState `json:"state"`
}

type Acl struct {
	Proto string   `json:"proto"`
	Dst   []string `json:"dst"`
}

// settings' acl

func (h *Handler) handleGetHeadscaleSshAcl(req *restful.Request, resp *restful.Response) {
	app, err := h.findApp(req.Request.Context(), systemAppName)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	// get headscale acl env
	allowSSH := false
	for _, acl := range app.Spec.TailScale.ACLs {
		if utils.ListContains(acl.Dst, "*:22") && strings.ToLower(acl.Proto) == "tcp" {
			allowSSH = true
			break
		}
	}

	for _, acl := range app.Spec.TailScaleACLs {
		if utils.ListContains(acl.Dst, "*:22") && strings.ToLower(acl.Proto) == "tcp" {
			allowSSH = true
			break
		}
	}

	response.Success(resp, SshAcl{AllowSSH: allowSSH, State: AclStateApplied})
}

func (h *Handler) handleDisableHeadscaleSshAcl(req *restful.Request, resp *restful.Response) {
	app, err := h.findApp(req.Request.Context(), systemAppName)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	acls := make([]apps.ACL, 0)
	for _, acl := range app.Spec.TailScale.ACLs {
		if acl.Dst[0] == "*:22" {
			continue
		}
		acls = append(acls, acl)
	}

	h.setHeadscaleAcl(req, resp, systemAppName, acls)
}

func (h *Handler) handleEnableHeadscaleSshAcl(req *restful.Request, resp *restful.Response) {
	app, err := h.findApp(req.Request.Context(), systemAppName)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	acls := app.Spec.TailScale.ACLs
	acls = append(acls, apps.ACL{Proto: "tcp", Dst: []string{"*:22"}})
	h.setHeadscaleAcl(req, resp, systemAppName, acls)
}

// app's acl

func (h *Handler) handleGetHeadscaleAppAcl(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)

	app, err := h.findApp(req.Request.Context(), appName)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	acls := make([]Acl, 0)
	for _, acl := range app.Spec.TailScale.ACLs {
		acls = append(acls, Acl{
			Proto: acl.Proto,
			Dst:   acl.Dst,
		})
	}
	// just to maintain compatibility with existing application
	for _, acl := range app.Spec.TailScaleACLs {
		acls = append(acls, Acl{
			Proto: acl.Proto,
			Dst:   acl.Dst,
		})
	}

	response.Success(resp, acls)
}

func (h *Handler) handleUpdateHeadscaleAppAcl(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	acls, err := h.parseAcl(req)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	klog.Infof("appsacl: %v", acls)
	if isPortDuplicate(acls) {
		response.HandleBadRequest(resp, errors.New("ports duplicated"))
		return
	}

	h.setHeadscaleAcl(req, resp, appName, acls)
}

type wrapACL struct {
	AppName  string `json:"appName"`
	AppOwner string `json:"appOwner"`
	Acl      `json:",inline"`
}

func (h *Handler) handleHeadscaleACLList(req *restful.Request, resp *restful.Response) {
	apps, err := h.appList(req.Request.Context())
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	acls := make([]wrapACL, 0)
	for _, app := range apps {
		for _, acl := range app.Spec.TailScale.ACLs {
			acls = append(acls, wrapACL{
				AppName:  app.Spec.Name,
				AppOwner: app.Spec.Owner,
				Acl: Acl{
					Proto: acl.Proto,
					Dst:   acl.Dst,
				},
			})
		}
		// just to maintain compatibility with existing application
		for _, acl := range app.Spec.TailScaleACLs {
			acls = append(acls, wrapACL{
				AppName:  app.Spec.Name,
				AppOwner: app.Spec.Owner,
				Acl: Acl{
					Proto: acl.Proto,
					Dst:   acl.Dst,
				},
			})
		}
	}
	response.Success(resp, acls)
}

func (h *Handler) setHeadscaleAcl(req *restful.Request, resp *restful.Response, appName string, acls []apps.ACL) {

	app, err := h.findApp(req.Request.Context(), appName)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {

		client, err := dynamic_client.NewResourceClient[apps.Application](apps.ApplicationGvr)
		if err != nil {
			klog.Error("failed to get client: ", err)
			return err
		}

		updateApp, err := client.Get(req.Request.Context(), app.Name, metav1.GetOptions{})
		if err != nil {
			klog.Error("failed to get app: ", err, ", ", app.Name)
			return err
		}

		updateApp.Spec.TailScale.ACLs = acls
		updateApp.Spec.TailScaleACLs = []apps.ACL{}
		_, err = client.Update(req.Request.Context(), updateApp, metav1.UpdateOptions{})

		return err
	})

	if err != nil {
		klog.Error("Failed to update headscale acl: ", err)
		response.HandleError(resp, err)
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) findApp(ctx context.Context, appName string) (*apps.Application, error) {
	client, err := dynamic_client.NewResourceClient[apps.Application](apps.ApplicationGvr)
	if err != nil {
		klog.Error("failed to get client: ", err)
		return nil, err
	}

	apps, err := client.List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list app error: ", err)
		return nil, err
	}

	for _, a := range apps {
		if a.Spec.Name == appName && a.Spec.Owner == constants.Username {
			return a, nil
		}
	}

	return nil, errors.New("app not found")
}

func (h *Handler) parseAcl(req *restful.Request) ([]apps.ACL, error) {
	var acls []apps.ACL
	err := req.ReadEntity(&acls)
	if err != nil {
		klog.Error("parse request acl body error, ", err)
		return nil, err
	}

	err = apps.CheckTailScaleACLs(acls)
	if err != nil {
		klog.Error("check acl error, ", err)
		return nil, err
	}

	return acls, nil
}

func calTailScaleSubnet() (subnets []string, err error) {
	kubeClient := v1alpha1client.KubeClient.Kubernetes()
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		return subnets, errors.New("get node name from env failed")
	}
	node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return subnets, err
	}
	ipAddress := node.Annotations["projectcalico.org/IPv4Address"]
	//_, ipnet, err := net.ParseCIDR(ipAddress)
	//if err != nil {
	//	return subnets, err
	//}
	ipnet := subtractOneMask(ipAddress)
	subnets = append(subnets, ipnet)
	return subnets, nil
}

func (h *Handler) handleGetTailScaleSubnet(req *restful.Request, resp *restful.Response) {
	app, err := h.findApp(req.Request.Context(), systemAppName)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	subRoutes := make([]string, 0)
	subRoutes = append(subRoutes, app.Spec.TailScale.SubRoutes...)
	response.Success(resp, subRoutes)
}

func (h *Handler) handleEnableTailScaleSubnet(req *restful.Request, resp *restful.Response) {
	tailScaleSubRoutes, err := calTailScaleSubnet()
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	h.setTailScaleSubRoutes(req, resp, systemAppName, tailScaleSubRoutes)
}

func (h *Handler) handleDisableTailScaleSubnet(req *restful.Request, resp *restful.Response) {
	var tailScaleSubRoutes []string
	h.setTailScaleSubRoutes(req, resp, systemAppName, tailScaleSubRoutes)
}

func (h *Handler) setTailScaleSubRoutes(req *restful.Request, resp *restful.Response, appName string, subRoutes []string) {
	app, err := h.findApp(req.Request.Context(), appName)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		client, err := dynamic_client.NewResourceClient[apps.Application](apps.ApplicationGvr)
		if err != nil {
			klog.Error("failed to get client: ", err)
			return err
		}
		updateApp, err := client.Get(req.Request.Context(), app.Name, metav1.GetOptions{})
		if err != nil {
			klog.Error("failed to get app: ", err, ", ", app.Name)
			return err
		}
		updateApp.Spec.TailScale.SubRoutes = subRoutes
		_, err = client.Update(req.Request.Context(), updateApp, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		klog.Errorf("Failed to update tailScale subRoutes %v", err)
		response.HandleError(resp, err)
		return
	}
	response.SuccessNoData(resp)
}

func (h *Handler) appList(ctx context.Context) ([]*apps.Application, error) {
	client, err := dynamic_client.NewResourceClient[apps.Application](apps.ApplicationGvr)
	if err != nil {
		klog.Error("failed to get client: ", err)
		return nil, err
	}

	appList, err := client.List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list app error: ", err)
		return nil, err
	}
	filteredApps := make([]*apps.Application, 0)
	for _, a := range appList {
		if a.Spec.Owner != constants.Username {
			continue
		}
		filteredApps = append(filteredApps, a)
	}

	return filteredApps, nil
}

func isPortDuplicate(acls []apps.ACL) bool {
	portMap := make(map[string]struct{})

	for _, acl := range acls {
		for _, dst := range acl.Dst {
			if acl.Proto == "" {
				tcpKey := fmt.Sprintf("tcp:%s", dst)
				udpKey := fmt.Sprintf("udp:%s", dst)
				if _, ok := portMap[tcpKey]; ok {
					return true
				}
				if _, ok := portMap[udpKey]; ok {
					return true
				}

				portMap[tcpKey] = struct{}{}
				portMap[udpKey] = struct{}{}
			} else {
				key := fmt.Sprintf("%s:%s", acl.Proto, dst)
				if _, exists := portMap[key]; exists {
					return true
				}
				portMap[key] = struct{}{}
			}
		}
	}

	return false
}

func subtractOneMask(subnet string) string {
	_, network, err := net.ParseCIDR(subnet)
	if err != nil {
		klog.Errorf("parseCIDR failed: %v", err)
		return subnet
	}
	ones, bits := network.Mask.Size()
	if ones <= 0 {
		klog.Infof("network mask ones <=0 ")
		return subnet
	}
	newMask := net.CIDRMask(ones-1, bits)
	ip := network.IP.Mask(newMask)
	newCIDR := net.IPNet{
		IP:   ip,
		Mask: newMask,
	}
	return newCIDR.String()
}
