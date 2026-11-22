package access

import (
	"encoding/json"
	"fmt"
)

// AdmissionHTTPRulesRequest is the subset of an AppCapabilitySet / policy
// payload that the validating admission path must check before persist.
type AdmissionHTTPRulesRequest struct {
	Rules []HTTPRule `json:"rules"`
}

// AdmissionResponse is a fail-closed decision suitable for ValidatingAdmissionWebhook.
type AdmissionResponse struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message,omitempty"`
}

// AdmitHTTPRules validates capability HTTP rules. Empty or invalid input is denied.
func AdmitHTTPRules(req AdmissionHTTPRulesRequest) AdmissionResponse {
	if err := ValidateHTTPRules(req.Rules); err != nil {
		return AdmissionResponse{Allowed: false, Message: err.Error()}
	}
	return AdmissionResponse{Allowed: true}
}

// AdmitHTTPRulesJSON decodes a JSON body and runs AdmitHTTPRules.
func AdmitHTTPRulesJSON(raw []byte) (AdmissionResponse, error) {
	if len(raw) == 0 {
		return AdmissionResponse{Allowed: false, Message: "empty admission payload"}, fmt.Errorf("empty admission payload")
	}
	var req AdmissionHTTPRulesRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return AdmissionResponse{Allowed: false, Message: "invalid json"}, err
	}
	return AdmitHTTPRules(req), nil
}
