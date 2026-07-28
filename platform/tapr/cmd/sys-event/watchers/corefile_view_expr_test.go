package watchers

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestBuildViewExprs_TC_EXPR_01_inclusterPodCIDR(t *testing.T) {
	incluster, _ := buildViewExprs("10.233.64.0/18", "10.233.0.0/16", "192.168.50.230", "")
	want := "incidr(client_ip(), '10.233.64.0/18')"
	if incluster != want {
		t.Fatalf("TC-EXPR-01 incluster=%q want %q", incluster, want)
	}
}

func TestBuildViewExprs_TC_EXPR_02_vpnClusterCIDR(t *testing.T) {
	_, vpn := buildViewExprs("10.233.64.0/18", "10.233.0.0/16", "192.168.50.230", "")
	wantParts := []string{
		"incidr(client_ip(), '100.64.0.0/16')",
		"client_ip() == '192.168.50.230'",
		"incidr(client_ip(), '10.233.0.0/16')",
	}
	for _, part := range wantParts {
		if !strings.Contains(vpn, part) {
			t.Fatalf("TC-EXPR-02 vpn=%q missing %q", vpn, part)
		}
	}
}

func TestBuildViewExprs_TC_EXPR_03_adguardExclusion(t *testing.T) {
	incluster, _ := buildViewExprs("10.233.64.0/18", "10.233.0.0/16", "192.168.50.230", "10.233.3.99")
	want := "( incidr(client_ip(), '10.233.64.0/18') && client_ip() != '10.233.3.99' )"
	if incluster != want {
		t.Fatalf("TC-EXPR-03 incluster=%q want %q", incluster, want)
	}
}

func TestClusterCIDRFromPod_TC_EXPR_04(t *testing.T) {
	if got := clusterCIDRFromPod("10.233.64.0/18"); got != "10.233.0.0/16" {
		t.Fatalf("TC-EXPR-04 clusterCIDRFromPod=%q want 10.233.0.0/16", got)
	}
	if got := clusterCIDRFromPod("172.20.0.0/16"); got != "172.20.0.0/16" {
		t.Fatalf("TC-EXPR-04 /16 pod CIDR clusterCIDR=%q want 172.20.0.0/16", got)
	}
}

func TestDetectPodCIDR_TC_EXPR_05_fallback(t *testing.T) {
	origProbe := podCIDRFromIptablesProbe
	podCIDRFromIptablesProbe = func() (string, bool) { return "", false }
	t.Cleanup(func() { podCIDRFromIptablesProbe = origProbe })

	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		calicoIPPoolGVR: "IPPoolList",
	})
	got := detectPodCIDR(context.Background(), kubefake.NewSimpleClientset(), dynamicClient)
	if got != defaultPodCIDR {
		t.Fatalf("TC-EXPR-05 detectPodCIDR=%q want fallback %q", got, defaultPodCIDR)
	}
}

func TestDetectPodCIDR_TC_EXPR_06_sources(t *testing.T) {
	t.Run("iptables KUBE-SERVICES", func(t *testing.T) {
		output := "-A KUBE-SERVICES ! -s 10.233.64.0/18 -m comment --comment \"kubernetes service portals\" -j KUBE-MARK-MASQ\n"
		got, ok := parsePodCIDRFromKubeServicesIPTables(output)
		if !ok || got != "10.233.64.0/18" {
			t.Fatalf("TC-EXPR-06 iptables parse=%q ok=%v", got, ok)
		}
	})

	t.Run("Calico IPPool", func(t *testing.T) {
		scheme := runtime.NewScheme()
		dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
			calicoIPPoolGVR: "IPPoolList",
		}, &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "crd.projectcalico.org/v1",
				"kind":       "IPPool",
				"metadata": map[string]interface{}{
					"name": "default-ipv4-ippool",
				},
				"spec": map[string]interface{}{
					"cidr": "10.233.64.0/18",
				},
			},
		})
		got, ok := podCIDRFromCalicoIPPools(context.Background(), dynamicClient)
		if !ok || got != "10.233.64.0/18" {
			t.Fatalf("TC-EXPR-06 Calico IPPool=%q ok=%v", got, ok)
		}
	})

	t.Run("kube-apiserver cluster-cidr arg", func(t *testing.T) {
		kubeClient := kubefake.NewSimpleClientset(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-apiserver-node1",
				Namespace: "kube-system",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "kube-apiserver",
					Command: []string{
						"kube-apiserver",
						"--cluster-cidr=10.233.64.0/18",
					},
				}},
			},
		})
		got, ok := podCIDRFromKubeClusterCIDRArg(context.Background(), kubeClient)
		if !ok || got != "10.233.64.0/18" {
			t.Fatalf("TC-EXPR-06 kube-apiserver arg=%q ok=%v", got, ok)
		}
	})
}
