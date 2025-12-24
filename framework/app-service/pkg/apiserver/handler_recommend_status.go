package apiserver

import (
	"context"
	"errors"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/argo"
	"github.com/beclab/Olares/framework/app-service/pkg/client/clientset"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	installerv1 "github.com/beclab/Olares/framework/app-service/pkg/workflowinstaller/v1"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type statusResp struct {
	api.Response
	Data *statusData `json:"data"`
}

type statusListResp struct {
	api.Response
	Data []*statusData `json:"data"`
}

type statusData struct {
	UUID           string                   `json:"uuid"`
	Namespace      string                   `json:"namespace"`
	User           string                   `json:"user"`
	ResourceStatus string                   `json:"resourceStatus"`
	ResourceType   string                   `json:"resourceType"`
	CreateTime     metav1.Time              `json:"createTime"`
	UpdataTime     metav1.Time              `json:"updateTime"`
	Metadata       metadata                 `json:"metadata"`
	Version        string                   `json:"version"`
	Title          string                   `json:"title"`
	SyncProvider   []map[string]interface{} `json:"syncProvider,omitempty"`
}

type metadata struct {
	Name string `json:"name"`
}

type status bool

func (s status) String() string {
	if s {
		return "running"
	}

	return "notfound"
}

func (h *Handler) statusRecommend(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamWorkflowName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	data, err := getWorkflowStatus(req.Request.Context(), h.kubeConfig, app, owner)

	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	ns, err := client.KubeClient.Kubernetes().CoreV1().Namespaces().Get(req.Request.Context(), data.Namespace, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		api.HandleError(resp, req, err)
		return
	}

	if title, ok := ns.Annotations[constants.WorkflowTitleAnnotation]; ok {
		data.Title = title
	}

	resp.WriteEntity(statusResp{
		Response: api.Response{Code: 200},
		Data:     data,
	})

}

func (h *Handler) statusRecommendList(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	namespaces, err := client.KubeClient.Kubernetes().CoreV1().Namespaces().List(req.Request.Context(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list namespace err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	var statusList []*statusData
	for _, ns := range namespaces.Items {
		release, ok := ns.Labels[constants.WorkflowNameLabel]
		if !ok {
			continue
		}

		user, ok := ns.Labels["bytetrade.io/ns-owner"]
		if !ok || user != owner {
			continue
		}

		data, err := getWorkflowStatus(req.Request.Context(), h.kubeConfig, release, owner)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}

		if title, ok := ns.Annotations[constants.WorkflowTitleAnnotation]; ok {
			data.Title = title
		}

		statusList = append(statusList, data)
	}

	resp.WriteEntity(statusListResp{
		Response: api.Response{Code: 200},
		Data:     statusList,
	})
}

func getWorkflowStatus(ctx context.Context, kubeConfig *rest.Config, app, owner string) (*statusData, error) {
	namespace, err := utils.AppNamespace(app, owner, "")
	if err != nil {
		return nil, err
	}

	helmClient, err := installerv1.NewHelmClient(ctx, kubeConfig, namespace)
	if err != nil {
		klog.Errorf("Failed to build helm client err=%v", err)
		return nil, err
	}

	installed, release, err := helmClient.Status(app)
	if err != nil {
		klog.Errorf("Failed to get install history app=%s err=%v", app, err)
		return nil, err
	}
	res := statusData{
		UUID:           "",
		Namespace:      namespace,
		User:           owner,
		ResourceStatus: status(installed).String(),
		ResourceType:   v1alpha1.Recommend.String(),
		Metadata:       metadata{Name: app},
	}
	client, _ := utils.GetClient()

	name, _ := apputils.FmtAppMgrName(app, owner, namespace)
	recommendMgr, err := client.AppV1alpha1().ApplicationManagers().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if recommendMgr.Status.OpType == v1alpha1.UninstallOp && recommendMgr.Status.State == v1alpha1.Uninstalling {
		res.ResourceStatus = v1alpha1.Uninstalling.String()
	}

	if release != nil {
		res.CreateTime = metav1.Time(release.Info.FirstDeployed)
		res.UpdataTime = metav1.Time(release.Info.LastDeployed)
	}

	if installed {
		version, err := helmClient.Version(app)
		if err != nil {
			klog.Errorf("Failed to get deployed chart version app=%s err=%v", app, err)
			return nil, err
		}

		res.Version = version
	}

	return &res, nil
}

func (h *Handler) statusListDev(req *restful.Request, resp *restful.Response) {
	owner := req.PathParameter(UserName)
	if owner == "" {
		klog.Error("owner is nil")
		api.HandleError(resp, req, errors.New("owner is nil"))
	}

	kubeClient, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		klog.Errorf("Failed to build k8s client err=%v", err)
		api.HandleError(resp, req, err)
	}

	namespaces, err := kubeClient.CoreV1().Namespaces().List(req.Request.Context(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list namespace err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	var statusList []*statusData
	for _, ns := range namespaces.Items {
		release, ok := ns.Labels[constants.WorkflowNameLabel]
		if !ok {
			continue
		}

		user, ok := ns.Labels["bytetrade.io/ns-owner"]
		if !ok || user != owner {
			continue
		}

		data, err := getWorkflowStatus(req.Request.Context(), h.kubeConfig, release, owner)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}

		if title, ok := ns.Annotations[constants.WorkflowTitleAnnotation]; ok {
			data.Title = title
		}

		data.SyncProvider, err = argo.GetProviderData(kubeClient, ns.Name)
		if err != nil {
			klog.Warningf("%s GetProviderData err=%v", ns.Name, err)
		}

		statusList = append(statusList, data)
	}

	resp.WriteEntity(statusListResp{
		Response: api.Response{Code: 200},
		Data:     statusList,
	})
}
