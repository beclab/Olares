package webhook

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const multusNetworksAnnotation = "k8s.v1.cni.cncf.io/networks"

// ValidateMacvlanAnnotation rejects a direct underlay-macvlan selection unless
// the platform-owned label is present. The label is the hand-off contract
// between this validating guard and the mutating allocator.
func ValidateMacvlanAnnotation(pod *corev1.Pod) error {
	if pod == nil || pod.Annotations == nil {
		return nil
	}
	raw := strings.TrimSpace(pod.Annotations[multusNetworksAnnotation])
	if raw == "" {
		return nil
	}
	if raw == "kube-system/underlay-macvlan" {
		if pod.Labels[constants.ApplicationMacvlanInitLabel] != "true" {
			return fmt.Errorf("direct underlay-macvlan annotation requires platform-owned label %s=true", constants.ApplicationMacvlanInitLabel)
		}
		return nil
	}

	var selections []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &selections); err != nil {
		return fmt.Errorf("unsupported Multus networks annotation: %w", err)
	}
	for _, selection := range selections {
		name, _, _ := unstructured.NestedString(selection, "name")
		namespace, _, _ := unstructured.NestedString(selection, "namespace")
		if name == "underlay-macvlan" && (namespace == "" || namespace == "kube-system") &&
			pod.Labels[constants.ApplicationMacvlanInitLabel] != "true" {
			return fmt.Errorf("direct underlay-macvlan annotation requires platform-owned label %s=true", constants.ApplicationMacvlanInitLabel)
		}
	}
	return nil
}
