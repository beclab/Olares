package onboarding

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ResetWithAuthelia uses the same authenticated backend as olares-cli ctl user
// reset-password. Callers must first authenticate the reserved onboarding user.
func ResetWithAuthelia(kube kubernetes.Interface) func(context.Context, string, string) error {
	client := &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return func(ctx context.Context, username, password string) error {
		expires := int64(600)
		token, err := kube.CoreV1().ServiceAccounts(Namespace).CreateToken(ctx, "olares-cli-sa", &authv1.TokenRequest{
			Spec: authv1.TokenRequestSpec{ExpirationSeconds: &expires},
		}, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		if token.Status.Token == "" {
			return fmt.Errorf("empty auth-provider credential")
		}
		svc, err := kube.CoreV1().Services(Namespace).Get(ctx, "auth-provider-svc", metav1.GetOptions{})
		if err != nil {
			return err
		}
		if net.ParseIP(svc.Spec.ClusterIP) == nil {
			return fmt.Errorf("auth-provider has no cluster IP")
		}
		port := int32(0)
		for _, p := range svc.Spec.Ports {
			if p.Name == "server" {
				port = p.Port
				break
			}
		}
		if port == 0 {
			return fmt.Errorf("auth-provider server port is missing")
		}
		// This is the existing auth-provider password wire format.
		payload, _ := json.Marshal(map[string]string{"password": passwordWire(password)})
		url := "http://" + net.JoinHostPort(svc.Spec.ClusterIP, strconv.Itoa(int(port))) + "/cli/api/reset/" + username + "/password"
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token.Status.Token)
		req.Header.Set("Olares-CLI-Authorization", "Bearer "+token.Status.Token)
		req.Header.Set("X-Forwarded-Host", "auth-provider-svc.os-framework:"+strconv.Itoa(int(port)))
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("auth-provider returned HTTP %d", resp.StatusCode)
		}
		return nil
	}
}

func passwordWire(password string) string {
	sum := md5.Sum([]byte(password + "@Olares2025"))
	return hex.EncodeToString(sum[:])
}
