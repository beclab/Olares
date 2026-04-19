package utils

import (
	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	UserSchemeGroupVersion = schema.GroupVersion{Group: "iam.kubesphere.io", Version: "v1alpha2"}

	UserGVR = schema.GroupVersionResource{
		Group:    UserSchemeGroupVersion.Group,
		Version:  UserSchemeGroupVersion.Version,
		Resource: "users",
	}
)

// NodePressure is a daemon-local alias of the canonical wire type
// shared with the olares-cli module. The Type field is plain string
// rather than corev1.NodeConditionType because the JSON wire format
// uses a string anyway and the cli module avoids depending on
// k8s.io/api/core/v1.
type NodePressure = clistate.NodePressure
