package v1

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/api"
	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apis"
	"bytetrade.io/web3os/bfl/pkg/apis/backend/v1/metrics"
	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	monitov1alpha1 "bytetrade.io/web3os/bfl/pkg/apis/monitor/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/app_service/v1"
	"bytetrade.io/web3os/bfl/pkg/constants"

	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
	"k8s.io/utils/pointer"
)

type Handler struct {
	apis.Base
	appService *app_service.Client
}

func New() *Handler {
	as := app_service.NewAppServiceClient()
	return &Handler{
		appService: as,
	}
}

func (h *Handler) handleUserInfo(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	var (
		isEphemeral                           bool
		terminusName, zone, role, createdUser string
		accessLevel                           *int
	)

	userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
	if err != nil {
		response.HandleError(resp, errors.Errorf("user info: new user operator err: %v", err))
		return
	}

	user, err := userOp.GetUser(username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("user info: get user err: %v", err))
		return
	}

	terminusName = user.Name

	isEphemeral, zone, err = userOp.GetUserDomainType(user)
	if err != nil {
		log.Warnf("unable to get user domain type: %v", err)
	}
	role = userOp.GetUserAnnotation(user, constants.UserAnnotationOwnerRole)
	createdUser = userOp.GetUserAnnotation(user, constants.AnnotationUserCreator)

	level := userOp.GetUserAnnotation(user, constants.UserLauncherAccessLevel)
	if level != "" {
		if l, err := strconv.Atoi(level); err == nil {
			accessLevel = pointer.Int(l)
		} else {
			log.Warnf("access level strconv err: %v", err)
		}
	}

	uInfo := UserInfo{
		Name:         username,
		OwnerRole:    role,
		TerminusName: terminusName,
		IsEphemeral:  isEphemeral,
		Zone:         zone,
		CreatedUser:  createdUser,
	}

	if status := userOp.GetTerminusStatus(user); status == string(constants.Completed) {
		uInfo.WizardComplete = true
	}

	if accessLevel != nil {
		uInfo.AccessLevel = accessLevel
	}
	response.Success(resp, uInfo)
}

func (h *Handler) handleTerminusInfo(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}

	userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
	if err != nil {
		response.HandleError(resp, errors.Errorf("olares info: new user operator err: %v", err))
		return
	}

	user, err := userOp.GetUser(username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("olares info: get user err: %v", err))
		return
	}

	tInfo := TerminusInfo{}
	tInfo.TerminusName = user.Name

	status := userOp.GetTerminusStatus(user)
	if status == "" {
		tInfo.WizardStatus = constants.WaitActivateVault
	} else {
		tInfo.WizardStatus = constants.WizardStatus(status)
	}

	selfhosted, terminusd, osVersion, err := userOp.SelfhostedAndOsVersion()
	if err != nil {
		response.HandleError(resp, errors.Errorf("olares info: get olares host type err: %v", err))
		return
	}

	tInfo.Selfhosted = selfhosted
	tInfo.OsVersion = osVersion

	tInfo.Terminusd = "0"
	if terminusd {
		tInfo.Terminusd = "1"
	}

	terminusId, err := userOp.GetTerminusID()
	if err != nil {
		response.HandleError(resp, errors.Errorf("olares info: get olares id err: %v", err))
		return
	}

	tInfo.TerminusID = terminusId

	var denyAllAnno string = userOp.GetDenyAllPolicy(user)
	if denyAllAnno == "" {
		tInfo.TailScaleEnable = false
	} else {
		denyAll, _ := strconv.Atoi(denyAllAnno)
		tInfo.TailScaleEnable = denyAll == 1
	}

	tInfo.LoginBackground, tInfo.Style = userOp.GetLoginBackground(user)
	tInfo.Avatar = userOp.GetAvatar(user)
	tInfo.UserDID = userOp.GetUserDID(user)

	if reverseProxy, err := userOp.GetReverseProxyType(); err == nil {
		tInfo.ReverseProxy = reverseProxy
	}

	response.Success(resp, tInfo)

}

func (h *Handler) handleOlaresInfo(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}

	userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
	if err != nil {
		response.HandleError(resp, errors.Errorf("olares info: new user operator err: %v", err))
		return
	}

	user, err := userOp.GetUser(username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("olares info: get user err: %v", err))
		return
	}

	tInfo := OlaresInfo{}
	tInfo.OlaresID = user.Name

	status := userOp.GetTerminusStatus(user)
	if status == "" {
		tInfo.WizardStatus = constants.WaitActivateVault
	} else {
		tInfo.WizardStatus = constants.WizardStatus(status)
	}

	tInfo.ClusterName, tInfo.VaultURL, err = clusterConnectionInfo(req.Request.Context(), tInfo.WizardStatus)
	if err != nil {
		response.HandleError(resp, errors.Errorf("olares info: get cluster connection info err: %v", err))
		return
	}

	selfhosted, terminusd, osVersion, err := userOp.SelfhostedAndOsVersion()
	if err != nil {
		response.HandleError(resp, errors.Errorf("olares info: get olares host type err: %v", err))
		return
	}

	tInfo.EnableReverseProxy = selfhosted
	tInfo.OsVersion = osVersion

	tInfo.Olaresd = "0"
	if terminusd {
		tInfo.Olaresd = "1"
	}

	terminusId, err := userOp.GetTerminusID()
	if err != nil {
		response.HandleError(resp, errors.Errorf("olares info: get olares id err: %v", err))
		return
	}

	tInfo.ID = terminusId

	var denyAllAnno string = userOp.GetDenyAllPolicy(user)
	if denyAllAnno == "" {
		tInfo.TailScaleEnable = false
	} else {
		denyAll, _ := strconv.Atoi(denyAllAnno)
		tInfo.TailScaleEnable = denyAll == 1
	}

	tInfo.LoginBackground, tInfo.Style = userOp.GetLoginBackground(user)
	tInfo.Avatar = userOp.GetAvatar(user)
	tInfo.UserDID = userOp.GetUserDID(user)

	response.Success(resp, tInfo)

}

func (h *Handler) myapps(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	var opt MyAppsParam
	if err := req.ReadEntity(&opt); err != nil {
		response.HandleError(resp, err)
		return
	}

	req.SetAttribute(constants.MyAppsIsLocalKey, opt.IsLocal)

	list, err := h.Base.GetAppListAndServicePort(req,
		func() (string, []*app_service.AppInfo, error) {
			return h.Base.GetAppForUser(username, h.appService.ForUser(username))
		})
	if err != nil {
		response.HandleInternalError(resp, fmt.Errorf("list apps: %v", err))
		return
	}
	response.Success(resp, api.NewListResult(list))

}

func (h *Handler) getClusterMetric(req *restful.Request, resp *restful.Response) {
	prome, err := metrics.NewPrometheus(metrics.PrometheusEndpoint)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	opts := metrics.QueryOptions{
		Level: metrics.LevelCluster,
	}

	metricsResult := prome.GetNamedMetrics(req.Request.Context(), []string{
		"cluster_cpu_usage",
		"cluster_cpu_total",
		"cluster_disk_size_usage",
		"cluster_disk_size_capacity",
		"cluster_memory_total",
		"cluster_memory_usage_wo_cache",
		"cluster_net_bytes_transmitted",
		"cluster_net_bytes_received",
	}, time.Now(), opts)

	var clusterMetrics monitov1alpha1.ClusterMetrics
	for _, m := range metricsResult {
		switch m.MetricName {
		case "cluster_cpu_usage":
			clusterMetrics.CPU.Usage = metrics.GetValue(&m)
		case "cluster_cpu_total":
			clusterMetrics.CPU.Total = metrics.GetValue(&m)

		case "cluster_disk_size_usage":
			clusterMetrics.Disk.Usage = metrics.GetValue(&m)
		case "cluster_disk_size_capacity":
			clusterMetrics.Disk.Total = metrics.GetValue(&m)

		case "cluster_memory_total":
			clusterMetrics.Memory.Total = metrics.GetValue(&m)
		case "cluster_memory_usage_wo_cache":
			clusterMetrics.Memory.Usage = metrics.GetValue(&m)

		case "cluster_net_bytes_transmitted":
			clusterMetrics.Net.Transmitted = metrics.GetValue(&m)

		case "cluster_net_bytes_received":
			clusterMetrics.Net.Received = metrics.GetValue(&m)
		}
	}

	roundToGB := func(v float64) float64 { return math.Round((v/1000000000.00)*100.00) / 100.00 }
	fmtMetricsValue(&clusterMetrics.CPU, "Cores", func(v float64) float64 { return v })
	fmtMetricsValue(&clusterMetrics.Memory, "GB", roundToGB)
	fmtMetricsValue(&clusterMetrics.Disk, "GB", roundToGB)

	response.Success(resp, clusterMetrics)

}

func fmtMetricsValue(v *monitov1alpha1.MetricV, unit string, unitFunc func(float64) float64) {
	v.Unit = unit

	v.Usage = unitFunc(v.Usage)
	v.Total = unitFunc(v.Total)
	v.Ratio = math.Round((v.Usage / v.Total) * 100)
}
