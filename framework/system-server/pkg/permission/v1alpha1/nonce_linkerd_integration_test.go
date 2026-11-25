//go:build integration

package permission

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Read-only acceptance against an explicitly selected, predeployed candidate.
// Fixtures are prepared by the operator; this test never creates resources or
// reads ServiceAccount tokens. Run on a host that can reach cluster Pod IPs.
func TestNonceLinkerdAcceptance(t *testing.T) {
	required := func(key string) string {
		t.Helper()
		v := os.Getenv(key)
		if v == "" {
			t.Fatalf("required acceptance input missing: %s", key)
		}
		return v
	}
	name := func(key string) string {
		t.Helper()
		v := required(key)
		if len(validation.IsDNS1123Subdomain(v)) != 0 {
			t.Fatalf("invalid Kubernetes name input: %s", key)
		}
		return v
	}
	contextName := required("NONCE_TEST_CONTEXT")
	targetNS := name("NONCE_TEST_TARGET_NAMESPACE")
	if !strings.HasPrefix(targetNS, "user-system-") {
		t.Fatal("target must be an isolated user-system candidate namespace")
	}
	expectedDigest := required("NONCE_TEST_SYSTEMSERVER_DIGEST")
	if !regexp.MustCompile(`^sha256:[a-f0-9]{64}$`).MatchString(expectedDigest) {
		t.Fatal("candidate must be selected by immutable digest")
	}
	run := func(args ...string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		defer cancel()
		return exec.CommandContext(ctx, "kubectl", append([]string{"--context", contextName}, args...)...).CombinedOutput()
	}
	namespaceRaw, err := run("get", "namespace", targetNS, "-o", "json")
	if err != nil {
		t.Fatal("cannot read acceptance namespace")
	}
	var targetNamespace corev1.Namespace
	if err := json.Unmarshal(namespaceRaw, &targetNamespace); err != nil || targetNamespace.Labels["traffic.olares.io/nonce-acceptance"] != "true" {
		t.Fatal("target namespace is not explicitly enrolled for isolated nonce acceptance")
	}
	getPod := func(ns, name string) corev1.Pod {
		t.Helper()
		raw, err := run("-n", ns, "get", "pod", name, "-o", "json")
		if err != nil {
			t.Fatal("cannot read acceptance Pod")
		}
		var pod corev1.Pod
		if err := json.Unmarshal(raw, &pod); err != nil {
			t.Fatal("invalid Pod response")
		}
		return pod
	}
	raw, err := run("-n", targetNS, "get", "pods", "-l", "app=systemserver", "-o", "json")
	if err != nil {
		t.Fatal("cannot read target Pods")
	}
	var pods corev1.PodList
	if err := json.Unmarshal(raw, &pods); err != nil || len(pods.Items) != 1 {
		t.Fatal("acceptance requires exactly one candidate system-server Pod")
	}
	target := pods.Items[0]
	if target.Annotations["config.linkerd.io/default-inbound-policy"] != "deny" {
		t.Fatal("candidate default policy must deny")
	}
	matched := false
	for _, status := range target.Status.ContainerStatuses {
		if status.Name == "system-server" && status.Ready && strings.HasSuffix(status.ImageID, expectedDigest) {
			matched = true
		}
	}
	if !matched {
		t.Fatal("running candidate does not match the required digest")
	}
	authPod := name("NONCE_TEST_AUTH_POD")
	authContainer := name("NONCE_TEST_AUTH_CONTAINER")
	source := getPod("os-framework", authPod)
	if source.Spec.ServiceAccountName != "authelia-backend" {
		t.Fatal("positive source is not the dedicated identity")
	}
	deniedNS := name("NONCE_TEST_DENIED_NAMESPACE")
	deniedPod := name("NONCE_TEST_DENIED_POD")
	deniedContainer := name("NONCE_TEST_DENIED_CONTAINER")
	denied := getPod(deniedNS, deniedPod)
	if denied.Spec.ServiceAccountName != "authelia-backend" {
		t.Fatal("cross-tenant fixture must use the same account name in a different namespace")
	}
	if deniedNS == "os-framework" {
		t.Fatal("cross-tenant negative source must be outside os-framework")
	}
	if denied.Status.Phase != corev1.PodRunning {
		t.Fatal("negative source is not running")
	}
	platformPod := name("NONCE_TEST_PLATFORM_DENIED_POD")
	platformContainer := name("NONCE_TEST_PLATFORM_DENIED_CONTAINER")
	platform := getPod("os-framework", platformPod)
	if platform.Spec.ServiceAccountName == "authelia-backend" {
		t.Fatal("platform negative source has the allowed identity")
	}
	url := "http://system-server." + targetNS + noncePath
	statusPattern := regexp.MustCompile(`HTTP/[0-9.]+ ([0-9]{3})`)
	request := func(ns, pod, container string, forged bool) ([]byte, []byte) {
		args := []string{"-n", ns, "exec", pod, "-c", container, "--", "wget", "-T", "5", "-S", "-O", "-"}
		if forged {
			args = append(args, "--header=Remote-User: admin", "--header=X-BFL-USER: other-tenant", "--header=X-Forwarded-For: "+source.Status.PodIP, "--header=l5d-client-id: authelia-backend.os-framework.serviceaccount.identity.linkerd.cluster.local", "--header=Authorization: Bearer forged")
		}
		out, _ := run(append(args, url)...)
		statuses := statusPattern.FindAllSubmatch(out, -1)
		if len(statuses) == 0 {
			t.Fatal("request produced no HTTP status; tool or network failures are not accepted as denials")
		}
		return statuses[len(statuses)-1][1], out
	}
	for _, forged := range []bool{false, true} {
		status, body := request("os-framework", authPod, authContainer, forged)
		if string(status) != "200" {
			t.Fatal("authorized mesh caller could not fetch nonce")
		}
		found := false
		for _, word := range strings.Fields(string(body)) {
			if value, err := uuid.Parse(word); err == nil && value.String() == word && value != uuid.Nil {
				found = true
			}
		}
		if !found {
			t.Fatal("authorized response did not contain the backend nonce")
		}
		status, _ = request(deniedNS, deniedPod, deniedContainer, forged)
		if string(status) != "403" {
			t.Fatal("cross-tenant workload was not denied")
		}
		status, _ = request("os-framework", platformPod, platformContainer, forged)
		if string(status) != "403" {
			t.Fatal("shared platform identity was not denied")
		}
	}
	serviceRaw, err := run("-n", targetNS, "get", "service", "system-server", "-o", "json")
	if err != nil {
		t.Fatal("cannot read candidate Service")
	}
	var service corev1.Service
	if err := json.Unmarshal(serviceRaw, &service); err != nil {
		t.Fatal("invalid Service response")
	}
	client := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	for _, ip := range []string{service.Spec.ClusterIP, target.Status.PodIP} {
		for _, forged := range []bool{false, true} {
			req, err := http.NewRequestWithContext(t.Context(), "GET", "http://"+ip+noncePath, nil)
			if err != nil {
				t.Fatal("invalid candidate address")
			}
			if forged {
				req.Header.Set("l5d-client-id", "authelia-backend.os-framework.serviceaccount.identity.linkerd.cluster.local")
				req.Header.Set("X-Forwarded-For", source.Status.PodIP)
				req.Header.Set("X-BFL-USER", "admin")
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal("plain HTTP produced a network error; explicit policy denial required")
			}
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			if resp.StatusCode != 403 {
				t.Fatal("Service/Pod IP bypass was not rejected by the mesh")
			}
		}
	}
}
