package gateway

import (
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// ComputeCallerInClusterPatch decides whether an Application with clusterAppRef
// should receive gateway.olares.io/in-cluster=gateway. Explicit in-cluster
// annotations (gateway or direct) are never overwritten.
func ComputeCallerInClusterPatch(app *appv1alpha1.Application) (needsPatch bool, value string) {
	if app == nil || app.Spec.Settings == nil {
		return false, ""
	}
	if strings.TrimSpace(app.Spec.Settings["clusterAppRef"]) == "" {
		return false, ""
	}
	if _, ok := explicitInClusterAnnotation(app); ok {
		return false, ""
	}
	if s := settingsInClusterMode(app); s == InClusterGateway || s == InClusterDirect {
		if cur := strings.ToLower(strings.TrimSpace(app.Annotations[AnnotationInCluster])); cur == s {
			return false, ""
		}
		return true, s
	}
	cur := strings.ToLower(strings.TrimSpace(app.Annotations[AnnotationInCluster]))
	if cur == InClusterGateway {
		return false, ""
	}
	return true, InClusterGateway
}

// ApplyCallerInClusterAnnotation sets in-cluster=gateway on app when needed.
func ApplyCallerInClusterAnnotation(app *appv1alpha1.Application) {
	need, v := ComputeCallerInClusterPatch(app)
	if !need {
		return
	}
	if app.Annotations == nil {
		app.Annotations = map[string]string{}
	}
	app.Annotations[AnnotationInCluster] = v
}

func explicitInClusterAnnotation(app *appv1alpha1.Application) (string, bool) {
	if app == nil || app.Annotations == nil {
		return "", false
	}
	v := strings.ToLower(strings.TrimSpace(app.Annotations[AnnotationInCluster]))
	if v != InClusterGateway && v != InClusterDirect {
		return "", false
	}
	return v, true
}

func settingsInClusterMode(app *appv1alpha1.Application) string {
	if app == nil || app.Spec.Settings == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(app.Spec.Settings[SettingInClusterMode]))
}
