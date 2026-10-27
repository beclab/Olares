package apiserver

import (
	"context"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
)

const (
	maxComputePreflightBodyBytes = 1 << 20
	maxComputePreflightDemands   = 64
	maxComputePreflightString    = 256
)

type ComputePreflightRequest struct {
	Demands []ComputePreflightDemand `json:"demands"`
}

type ComputePreflightDemand struct {
	ID                string `json:"id"`
	AppID             string `json:"appId,omitempty"`
	Application       string `json:"application"`
	Owner             string `json:"owner"`
	Mode              string `json:"mode"`
	RequiredCPU       string `json:"requiredCPU"`
	RequiredGPU       string `json:"requiredGPU"`
	LimitedGPU        string `json:"limitedGPU"`
	RequiredMemory    string `json:"requiredMemory"`
	LimitedMemory     string `json:"limitedMemory"`
	RequiredDisk      string `json:"requiredDisk"`
	SupportMultiCards bool   `json:"supportMultiCards"`
	SupportMultiNodes bool   `json:"supportMultiNodes"`
}

type ComputePreflightResponse struct {
	api.Response `json:",inline"`
	Data         compute.PreflightReport `json:"data"`
}

type preflightSnapshotCollector interface {
	Collect(ctx context.Context, token string, owners []string) (compute.PreflightSnapshot, error)
}
