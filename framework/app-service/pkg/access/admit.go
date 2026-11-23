package access

import (
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PolicyNamespace    = "os-app-access"
	CatalogWebhookPath = "/app-service/v1/access/catalog/validate"
)

type AdmissionHTTPRulesRequest struct {
	Rules []HTTPRule `json:"rules"`
}

type AdmissionResponse struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message,omitempty"`
}

func AdmitHTTPRules(req AdmissionHTTPRulesRequest) AdmissionResponse {
	if err := ValidateHTTPRules(req.Rules); err != nil {
		return AdmissionResponse{Message: err.Error()}
	}
	return AdmissionResponse{Allowed: true}
}

func AdmitHTTPRulesJSON(raw []byte) (AdmissionResponse, error) {
	var req AdmissionHTTPRulesRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return AdmissionResponse{Message: "invalid JSON"}, err
	}
	return AdmitHTTPRules(req), nil
}

// CatalogAdmissionHandler serves the non-declarative capability checks on the
// existing TLS listener. Grant revocation does not depend on this endpoint.
func CatalogAdmissionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "admission requires POST", http.StatusMethodNotAllowed)
		return
	}
	var review admissionv1.AdmissionReview
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&review); err != nil || review.Request == nil || review.Request.UID == "" ||
		review.APIVersion != admissionv1.SchemeGroupVersion.String() || review.Kind != "AdmissionReview" {
		http.Error(w, "invalid admission review", http.StatusBadRequest)
		return
	}
	req := review.Request
	resp := admissionv1.AdmissionResponse{UID: req.UID}
	if err := validateCatalogRequest(req); err != nil {
		resp.Result = &metav1.Status{Message: err.Error()}
	} else {
		resp.Allowed = true
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(admissionv1.AdmissionReview{TypeMeta: review.TypeMeta, Response: &resp})
}

func validateCatalogRequest(req *admissionv1.AdmissionRequest) error {
	if req.Resource.Group != "access.olares.io" || req.Resource.Version != "v1alpha1" ||
		req.Resource.Resource != "appcapabilitysets" || req.SubResource != "" ||
		req.Namespace != PolicyNamespace || (req.Operation != admissionv1.Create && req.Operation != admissionv1.Update) {
		return fmt.Errorf("expected a capability catalog write in %s", PolicyNamespace)
	}
	var catalog struct {
		Spec struct {
			Capabilities []struct {
				Rules []HTTPRule `json:"rules"`
			} `json:"capabilities"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(req.Object.Raw, &catalog); err != nil {
		return fmt.Errorf("invalid capability catalog")
	}
	if len(catalog.Spec.Capabilities) == 0 {
		return fmt.Errorf("at least one capability is required")
	}
	for i, cap := range catalog.Spec.Capabilities {
		if err := ValidateHTTPRules(cap.Rules); err != nil {
			return fmt.Errorf("capability[%d]: %w", i, err)
		}
	}
	return nil
}
