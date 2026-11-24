package network_activation

import "testing"

func TestCertificateRefreshCronJobUsesNativeSidecar(t *testing.T) {
	job := renewCertCronJob("user-space-example", "0 0 1 1 ?", true)
	pod := job.Spec.JobTemplate.Spec.Template
	if pod.Spec.ServiceAccountName == nil || *pod.Spec.ServiceAccountName != "bfl-cert-refresh" ||
		pod.Labels["app.kubernetes.io/name"] != "bfl-cert-refresh" ||
		pod.Annotations["linkerd.io/inject"] != "enabled" ||
		pod.Annotations["config.linkerd.io/proxy-enable-native-sidecar"] != "true" {
		t.Fatalf("unexpected certificate refresh CronJob identity: %+v", pod)
	}
	legacy := renewCertCronJob("user-space-legacy", "0 0 1 1 ?", false).Spec.JobTemplate.Spec.Template
	if legacy.Spec.ServiceAccountName != nil || legacy.ObjectMetaApplyConfiguration != nil {
		t.Fatal("a tenant without the dedicated ServiceAccount must retain its existing CronJob template")
	}
}
