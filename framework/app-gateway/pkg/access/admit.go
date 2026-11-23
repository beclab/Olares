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

// admissionReview is the Kubernetes admission.k8s.io/v1 subset we need.
type admissionReview struct {
	APIVersion string             `json:"apiVersion"`
	Kind       string             `json:"kind"`
	Request    *admissionRequest  `json:"request,omitempty"`
	Response   *admissionResponse  `json:"response,omitempty"`
}

type admissionRequest struct {
	UID       string          `json:"uid"`
	Operation string          `json:"operation"`
	Object    json.RawMessage `json:"object"`
}

type admissionResponse struct {
	UID     string `json:"uid"`
	Allowed bool   `json:"allowed"`
	Status  *struct {
		Message string `json:"message,omitempty"`
	} `json:"status,omitempty"`
}

type capabilityObject struct {
	Spec struct {
		Capabilities []struct {
			HTTP []HTTPRule `json:"http"`
		} `json:"capabilities"`
		Rules []HTTPRule `json:"rules"`
	} `json:"spec"`
}

// AdmitAdmissionReview validates an AdmissionReview body for access.olares.io objects.
// Fail-closed: empty/malformed reviews are denied.
func AdmitAdmissionReview(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return encodeReviewResponse("", false, "empty admission review")
	}
	var review admissionReview
	if err := json.Unmarshal(raw, &review); err != nil {
		return encodeReviewResponse("", false, "invalid admission review json")
	}
	if review.Request == nil {
		return encodeReviewResponse("", false, "missing admission request")
	}
	uid := review.Request.UID
	if len(review.Request.Object) == 0 || string(review.Request.Object) == "null" {
		return encodeReviewResponse(uid, false, "missing object")
	}
	var obj capabilityObject
	if err := json.Unmarshal(review.Request.Object, &obj); err != nil {
		return encodeReviewResponse(uid, false, "invalid object json")
	}
	var rules []HTTPRule
	if len(obj.Spec.Rules) > 0 {
		rules = obj.Spec.Rules
	} else {
		for _, cap := range obj.Spec.Capabilities {
			rules = append(rules, cap.HTTP...)
		}
	}
	if len(rules) == 0 {
		// Objects without HTTP rules (e.g. non-capability types) are allowed.
		return encodeReviewResponse(uid, true, "")
	}
	dec := AdmitHTTPRules(AdmissionHTTPRulesRequest{Rules: rules})
	return encodeReviewResponse(uid, dec.Allowed, dec.Message)
}

func encodeReviewResponse(uid string, allowed bool, msg string) ([]byte, error) {
	resp := admissionReview{
		APIVersion: "admission.k8s.io/v1",
		Kind:       "AdmissionReview",
		Response: &admissionResponse{
			UID:     uid,
			Allowed: allowed,
		},
	}
	if msg != "" {
		resp.Response.Status = &struct {
			Message string `json:"message,omitempty"`
		}{Message: msg}
	}
	return json.Marshal(resp)
}
