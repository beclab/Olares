package controllers

import (
	"encoding/json"

	appv1alpha1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
)

const (
	applicationSettingsPolicyKey        = "policy"
	namespaceFinalizer                  = "finalizers.bytetrade.io/namespaces"
	userFinalizer                       = "finalizers.bytetrade.io/users"
	creator                             = "bytetrade.io/creator"
	deploymentResourceVersionAnnotation = "bytetrade.io/deployment-resource-version"
)

type applicationSettingsSubPolicy struct {
	URI      string `json:"uri"`
	Policy   string `json:"policy"`
	OneTime  bool   `json:"one_time"`
	Duration int32  `json:"valid_duration"`
}

type applicationSettingsPolicy struct {
	DefaultPolicy string                          `json:"default_policy"`
	SubPolicies   []*applicationSettingsSubPolicy `json:"sub_policies"`
	OneTime       bool                            `json:"one_time"`
	Duration      int32                           `json:"valid_duration"`
}

// mergeEntrances merges new entrances with existing ones.
// Preserves authLevel from existing entrances, other fields are updated from new entrances.
func mergeEntrances(existing, incoming []appv1alpha1.Entrance) []appv1alpha1.Entrance {
	if len(existing) == 0 {
		return incoming
	}

	existingByName := make(map[string]*appv1alpha1.Entrance, len(existing))
	for i := range existing {
		existingByName[existing[i].Name] = &existing[i]
	}

	merged := make([]appv1alpha1.Entrance, 0, len(incoming))
	for _, entry := range incoming {
		if old, exists := existingByName[entry.Name]; exists {
			entry.AuthLevel = old.AuthLevel
		}
		merged = append(merged, entry)
	}

	return merged
}

func mergePolicySettings(existingPolicy, incomingPolicy string) string {
	if incomingPolicy == "" {
		return existingPolicy
	}
	if existingPolicy == "" {
		return incomingPolicy
	}

	var existing, incoming map[string]applicationSettingsPolicy
	if err := json.Unmarshal([]byte(existingPolicy), &existing); err != nil {
		return incomingPolicy
	}
	if err := json.Unmarshal([]byte(incomingPolicy), &incoming); err != nil {
		return existingPolicy
	}

	merged := make(map[string]applicationSettingsPolicy, len(incoming))
	for name, incomingEntry := range incoming {
		if existingEntry, exists := existing[name]; exists {
			incomingEntry.DefaultPolicy = existingEntry.DefaultPolicy
			incomingEntry.SubPolicies = existingEntry.SubPolicies
		}
		merged[name] = incomingEntry
	}

	result, err := json.Marshal(merged)
	if err != nil {
		return existingPolicy
	}

	return string(result)
}
