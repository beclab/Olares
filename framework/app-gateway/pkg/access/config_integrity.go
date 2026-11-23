package access

import (
	"encoding/json"
	"strings"
)

// Protected runtime ConfigMaps that project security posture (not secrets).
// Integrity matters: unauthorized updates can disable mesh/NP/mTLS.
var protectedConfigMaps = map[string]struct{}{
	"app-access-runtime":   {},
	"olares-http-mesh-wave": {},
}

// platformUserPrefixes may mutate protected posture ConfigMaps (installer / control plane).
var platformUserPrefixes = []string{
	"system:serviceaccount:os-app-access:",
	"system:serviceaccount:kube-system:",
	"system:serviceaccount:os-platform:",
	"system:masters",
	"system:serviceaccount:os-framework:",
}

// AdmitProtectedConfigMap enforces who may change end-state posture ConfigMaps.
// Non-protected ConfigMaps are allowed. Fail-closed for protected names.
func AdmitProtectedConfigMap(username, cmName string) AdmissionResponse {
	if _, ok := protectedConfigMaps[cmName]; !ok {
		return AdmissionResponse{Allowed: true}
	}
	if usernameAllowed(username) {
		return AdmissionResponse{Allowed: true}
	}
	return AdmissionResponse{
		Allowed: false,
		Message: "uniauth posture ConfigMap is release-controlled; mutate via OTA/installer only",
	}
}

func usernameAllowed(username string) bool {
	u := strings.TrimSpace(username)
	if u == "" {
		return false
	}
	for _, p := range platformUserPrefixes {
		if p == "system:masters" && u == p {
			return true
		}
		if strings.HasPrefix(u, p) {
			return true
		}
	}
	// kube-apiserver / helm often appear as system:serviceaccount:…:… 
	if strings.Contains(u, "helm") && strings.HasPrefix(u, "system:serviceaccount:") {
		return true
	}
	return false
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
	if o.Kind != "ConfigMap" && o.Kind != "" {
		// admission sometimes omits kind in object; caller may pass resource
		if o.Metadata.Name == "" {
			return o.Kind, "", false
		}
	}
	if o.Kind == "ConfigMap" || o.Kind == "" {
		if o.Metadata.Name != "" {
			return "ConfigMap", o.Metadata.Name, true
		}
	}
	return o.Kind, o.Metadata.Name, o.Metadata.Name != ""
}
