package apiserver

import (
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
)

// ResumeRequest is the request body for POST /apps/{app}/resume and
// related resume paths. Kept in this file so both handler_suspend_v2.go
// and handler_compute_resources.go can decode it without introducing a
// dedicated types file.
type ResumeRequest struct {
	ComputeBinding []compute.BindingSelection `json:"computeBinding,omitempty"`
	// ComputeBindings picks cards per GPU consumer for apps declaring
	// workload accelerators; consumers left out are placed automatically.
	ComputeBindings map[string][]compute.BindingSelection `json:"computeBindings,omitempty"`
}
