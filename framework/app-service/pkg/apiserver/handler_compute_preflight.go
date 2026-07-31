package apiserver

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/resource"
)

func (h *Handler) computeResourcesPreflight(req *restful.Request, resp *restful.Response) {
	if req.Request.ContentLength > maxComputePreflightBodyBytes {
		handleComputePreflightBodyTooLarge(req, resp)
		return
	}
	req.Request.Body = http.MaxBytesReader(resp.ResponseWriter, req.Request.Body, maxComputePreflightBodyBytes)
	var request ComputePreflightRequest
	if err := req.ReadEntity(&request); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			handleComputePreflightBodyTooLarge(req, resp)
			return
		}
		api.HandleBadRequest(resp, req, err)
		return
	}
	if len(request.Demands) == 0 {
		api.HandleBadRequest(resp, req, fmt.Errorf("at least one demand is required"))
		return
	}
	if len(request.Demands) > maxComputePreflightDemands {
		api.HandleBadRequest(resp, req, fmt.Errorf("at most %d demands are allowed", maxComputePreflightDemands))
		return
	}

	user, ok := req.Attribute(constants.UserContextAttribute).(string)
	if !ok || user == "" {
		api.HandleUnauthorized(resp, req, fmt.Errorf("authenticated user is required"))
		return
	}
	if h.preflightIsAdmin == nil {
		api.HandleInternalError(resp, req, fmt.Errorf("compute preflight admin checker is not configured"))
		return
	}
	isAdmin, err := h.preflightIsAdmin(req.Request.Context(), user)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	demands := make([]compute.PreflightDemand, 0, len(request.Demands))
	owners := make([]string, 0, len(request.Demands))
	seenOwners := make(map[string]struct{}, len(request.Demands))
	seenIDs := make(map[string]struct{}, len(request.Demands))
	for i, wire := range request.Demands {
		if !isAdmin && wire.Owner != user {
			api.HandleForbidden(resp, req, fmt.Errorf("demand %d owner %q is not authorized", i, wire.Owner))
			return
		}
		demand, err := wire.toCompute()
		if err != nil {
			api.HandleBadRequest(resp, req, fmt.Errorf("demand %d: %w", i, err))
			return
		}
		if _, exists := seenIDs[demand.ID]; exists {
			api.HandleBadRequest(resp, req, fmt.Errorf("demand %d: duplicate id %q", i, demand.ID))
			return
		}
		seenIDs[demand.ID] = struct{}{}
		demands = append(demands, demand)
		if _, exists := seenOwners[demand.Owner]; !exists {
			seenOwners[demand.Owner] = struct{}{}
			owners = append(owners, demand.Owner)
		}
	}

	if h.preflightCollector == nil {
		api.HandleInternalError(resp, req, fmt.Errorf("compute preflight snapshot collector is not configured"))
		return
	}
	token := ""
	if h.kubeConfig != nil {
		token = h.kubeConfig.BearerToken
	}
	snapshot, err := h.preflightCollector.Collect(req.Request.Context(), token, owners)
	if err != nil {
		api.HandleInternalError(resp, req, fmt.Errorf("collect compute preflight snapshot: %w", err))
		return
	}
	report, err := compute.SimulatePreflight(demands, snapshot)
	if err != nil {
		api.HandleInternalError(resp, req, fmt.Errorf("simulate compute preflight: %w", err))
		return
	}
	resp.WriteAsJson(ComputePreflightResponse{
		Response: api.Response{Code: api.CodeSuccess},
		Data:     report,
	})
}

func (d ComputePreflightDemand) toCompute() (compute.PreflightDemand, error) {
	if d.ID == "" || d.Application == "" || d.Owner == "" || d.Mode == "" {
		return compute.PreflightDemand{}, fmt.Errorf("id, application, owner, and mode are required")
	}
	if !validComputePreflightMode(d.Mode) {
		return compute.PreflightDemand{}, fmt.Errorf("unsupported mode %q", d.Mode)
	}
	for _, item := range []struct{ field, value string }{
		{field: "id", value: d.ID},
		{field: "appId", value: d.AppID},
		{field: "application", value: d.Application},
		{field: "owner", value: d.Owner},
		{field: "mode", value: d.Mode},
		{field: "requiredCPU", value: d.RequiredCPU},
		{field: "requiredGPU", value: d.RequiredGPU},
		{field: "limitedGPU", value: d.LimitedGPU},
		{field: "requiredMemory", value: d.RequiredMemory},
		{field: "limitedMemory", value: d.LimitedMemory},
		{field: "requiredDisk", value: d.RequiredDisk},
	} {
		if len(item.value) > maxComputePreflightString {
			return compute.PreflightDemand{}, fmt.Errorf("%s exceeds %d bytes", item.field, maxComputePreflightString)
		}
	}
	var requiredCPU, requiredGPU, limitedGPU int64
	var requiredMemory, limitedMemory, requiredDisk int64
	quantities := [...]struct {
		field, value string
		milli        bool
		out          *int64
	}{
		{"requiredCPU", d.RequiredCPU, true, &requiredCPU},
		{"requiredGPU", d.RequiredGPU, false, &requiredGPU},
		{"limitedGPU", d.LimitedGPU, false, &limitedGPU},
		{"requiredMemory", d.RequiredMemory, false, &requiredMemory},
		{"limitedMemory", d.LimitedMemory, false, &limitedMemory},
		{"requiredDisk", d.RequiredDisk, false, &requiredDisk},
	}
	for _, item := range quantities {
		value, err := parseWireQuantity(item.field, item.value, item.milli)
		if err != nil {
			return compute.PreflightDemand{}, err
		}
		*item.out = value
	}
	if limitedGPU != 0 && limitedGPU < requiredGPU {
		return compute.PreflightDemand{}, fmt.Errorf("limitedGPU must be zero or at least requiredGPU")
	}
	if limitedMemory != 0 && limitedMemory < requiredMemory {
		return compute.PreflightDemand{}, fmt.Errorf("limitedMemory must be zero or at least requiredMemory")
	}
	supportMultiNodes := d.Mode == utils.NvidiaCardType && d.SupportMultiNodes
	supportMultiCards := d.Mode == utils.NvidiaCardType && (d.SupportMultiCards || supportMultiNodes)
	return compute.PreflightDemand{
		ID:          d.ID,
		AppID:       d.AppID,
		Application: d.Application,
		Owner:       d.Owner,
		Requirement: compute.Requirement{
			Mode:              d.Mode,
			RequiredCPU:       requiredCPU,
			RequiredGPU:       requiredGPU,
			LimitedGPU:        limitedGPU,
			RequiredMemory:    requiredMemory,
			LimitedMemory:     limitedMemory,
			RequiredDisk:      requiredDisk,
			SupportMultiCards: supportMultiCards,
			SupportMultiNodes: supportMultiNodes,
		},
	}, nil
}

func validComputePreflightMode(mode string) bool {
	switch mode {
	case utils.CPUType, utils.NvidiaCardType, utils.GB10ChipType, utils.AppleMChipType,
		utils.IntelType, utils.AMDType, utils.IntelGPUType, utils.AMDGPUType, utils.MooreSocType:
		return true
	default:
		return false
	}
}

func parseWireQuantity(field, value string, milli bool) (int64, error) {
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("%s is required", field)
	}
	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", field, err)
	}
	if quantity.Sign() < 0 {
		return 0, fmt.Errorf("%s must not be negative", field)
	}
	var max *resource.Quantity
	if milli {
		max = resource.NewMilliQuantity(math.MaxInt64, resource.DecimalSI)
	} else {
		max = resource.NewQuantity(math.MaxInt64, resource.DecimalSI)
	}
	if quantity.Cmp(*max) > 0 {
		return 0, fmt.Errorf("%s overflows int64", field)
	}
	if milli {
		return quantity.MilliValue(), nil
	}
	return quantity.Value(), nil
}

func handleComputePreflightBodyTooLarge(req *restful.Request, resp *restful.Response) {
	api.HandleError(resp, req, restful.ServiceError{
		Code:    http.StatusRequestEntityTooLarge,
		Message: fmt.Sprintf("request body exceeds %d bytes", maxComputePreflightBodyBytes),
	})
}
