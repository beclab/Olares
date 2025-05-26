package model

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

type InstallModelReq struct {
	Config struct {
		DomainName      string `json:"terminus_os_domainname"`
		UserName        string `json:"terminus_os_username" validate:"required"`
		KubeType        string `json:"kube_type" validate:"kubeTypeValid"`
		Vendor          string `json:"vendor"`
		GpuEnable       int    `json:"gpu_enable" validate:"oneof=0 1"`
		GpuShare        int    `json:"gpu_share" validate:"required_with=GpuEnable,oneof=0 1"`
		Version         string `json:"version"`
		Proxy           string `json:"proxy"`
		RegistryMirrors string `json:"registry-mirrors"` // need to add to cli
		DebugSaveConfig int    `json:"debug_save_config" validate:"oneof=0 1"`
		DebugDownload   int    `json:"debug_download" validate:"oneof=0 1"`
		DebugInstall    int    `json:"debug_install" validate:"oneof=0 1"`
	} `json:"config"`
}

func KubeTypeValid(fl validator.FieldLevel) bool {
	kubeType := strings.ToLower(fl.Field().String())
	return kubeType == "" || kubeType == "k3s" || kubeType == "k8s"
}
