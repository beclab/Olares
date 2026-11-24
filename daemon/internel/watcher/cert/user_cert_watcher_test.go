package cert

import "testing"

func TestCertificateRefreshJobUsesDedicatedMeshIdentity(t *testing.T) {
	job := newDownloadUserCertJob("user-space-example", true)
	pod := job.Spec.Template
	if job.Namespace != "user-space-example" || pod.Spec.ServiceAccountName != "bfl-cert-refresh" ||
		pod.Labels["app.kubernetes.io/name"] != "bfl-cert-refresh" ||
		pod.Annotations["linkerd.io/inject"] != "enabled" ||
		pod.Annotations["config.linkerd.io/proxy-enable-native-sidecar"] != "true" {
		t.Fatalf("unexpected certificate refresh Job identity: %+v", pod)
	}
	old := newDownloadUserCertJob("user-space-legacy", false).Spec.Template
	if old.Spec.ServiceAccountName != "" || len(old.Labels) != 0 || len(old.Annotations) != 0 {
		t.Fatal("a tenant without the dedicated ServiceAccount must retain its existing Job template")
	}
}
