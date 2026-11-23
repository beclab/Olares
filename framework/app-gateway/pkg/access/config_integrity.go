package access

import (
	"encoding/json"
	"strings"
)

// Protected runtime ConfigMaps that project security posture (not secrets).
var protectedConfigMaps = map[string]struct{}{
	"app-access-runtime":    {},
	"olares-http-mesh-wave": {},
}

// Exact platform identities allowed to mutate posture ConfigMaps (installer / masters).
var platformUsers = map[string]struct{}{
	"system:masters": {},
	"system:serviceaccount:kube-system:olares-cli": {},
	"system:serviceaccount:kube-system:default":    {},
	"system:serviceaccount:os-platform:olares-app": {},
	"system:serviceaccount:os-app-access:access-admission": {},
}

// AdmitProtectedConfigMap enforces who may change posture ConfigMaps.
func AdmitProtectedConfigMap(username, cmName string) AdmissionResponse {
	if _, ok := protectedConfigMaps[cmName]; !ok {
		return AdmissionResponse{Allowed: true}
	}
	u := strings.TrimSpace(username)
	if _, ok := platformUsers[u]; ok {
		return AdmissionResponse{Allowed: true}
	}
	// Cluster bootstrap often uses the apiserver identity.
	if u == "kubernetes-admin" || strings.HasPrefix(u, "system:node:") {
		return AdmissionResponse{Allowed: true}
	}
	return AdmissionResponse{
		Allowed: false,
		Message: "uniauth posture ConfigMap is release-controlled; mutate via OTA only",
	}
}

type metaOnlyObject struct {
	Kind     string `json:"kind"`
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
}

func configMapNameFromObject(raw json.RawMessage) (kind, name string, ok bool) {
	var o metaOnlyObject
	if err := json.Unmarshal(raw, &o); err != nil {
		return "", "", false
	}
	if o.Metadata.Name == "" {
		return o.Kind, "", false
	}
	if o.Kind == "ConfigMap" || o.Kind == "" {
		return "ConfigMap", o.Metadata.Name, true
	}
	return o.Kind, o.Metadata.Name, true
}
