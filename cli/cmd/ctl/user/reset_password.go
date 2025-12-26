package user

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"
)

const (
	resetNamespace       = "os-framework"
	resetServiceAccount  = "olares-cli-sa"
	resetServiceName     = "auth-provider-svc"
	resetServicePortName = "server"
	defaultServicePort   = 28080
	passwordSaltSuffix   = "@Olares2025"
	authHeaderBearer     = "Bearer "
	cliAuthHeader        = "Olares-CLI-Authorization"
	resetRequestPathTmpl = "http://%s:%d/cli/api/reset/%s/password"
)

type resetPasswordOptions struct {
	username   string
	password   string
	kubeConfig string
}

func NewCmdResetPassword() *cobra.Command {
	o := &resetPasswordOptions{}
	cmd := &cobra.Command{
		Use:   "reset-password {username}",
		Short: "forcefully reset a user's password via auth-provider",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			o.username = args[0]
			if err := o.Validate(); err != nil {
				log.Fatal(err)
			}
			if err := o.Run(cmd.Context()); err != nil {
				log.Fatal(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func (o *resetPasswordOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.password, "password", "p", "", "new password to set")
	cmd.Flags().StringVar(&o.kubeConfig, "kubeconfig", "", "path to kubeconfig file (optional)")
}

func (o *resetPasswordOptions) Validate() error {
	if o.username == "" {
		return fmt.Errorf("username is required")
	}
	if o.password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (o *resetPasswordOptions) Run(ctx context.Context) error {
	cfgPath := o.kubeConfig
	if cfgPath == "" {
		cfgPath = os.Getenv("KUBECONFIG")
		if cfgPath == "" {
			cfgPath = clientcmd.RecommendedHomeFile
		}
	}
	restCfg, err := clientcmd.BuildConfigFromFlags("", cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	k8s, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	expires := int64(3600)
	tr := &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			ExpirationSeconds: &expires,
		},
	}
	tokenResp, err := k8s.CoreV1().ServiceAccounts(resetNamespace).CreateToken(ctx, resetServiceAccount, tr, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service account token: %w", err)
	}
	saToken := tokenResp.Status.Token
	if saToken == "" {
		return fmt.Errorf("received empty token for service account %s/%s", resetNamespace, resetServiceAccount)
	}

	svc, err := k8s.CoreV1().Services(resetNamespace).Get(ctx, resetServiceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get service %s/%s: %w", resetNamespace, resetServiceName, err)
	}
	clusterIP := svc.Spec.ClusterIP
	port := int32(defaultServicePort)
	if len(svc.Spec.Ports) > 0 {
		chosen := svc.Spec.Ports[0].Port
		for _, p := range svc.Spec.Ports {
			if p.Name == resetServicePortName {
				chosen = p.Port
				break
			}
		}
		port = chosen
	}
	if clusterIP == "" {
		return fmt.Errorf("service %s/%s has empty ClusterIP", resetNamespace, resetServiceName)
	}

	url := fmt.Sprintf(resetRequestPathTmpl, clusterIP, port, o.username)
	bodyMap := map[string]string{
		"password": saltedMD5(o.password),
	}
	payload, err := json.Marshal(bodyMap)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeaderBearer+saToken)
	req.Header.Set(cliAuthHeader, authHeaderBearer+saToken)
	req.Header.Set("X-Forwarded-Host", fmt.Sprintf("%s.%s:%d", resetServiceName, resetNamespace, port))

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("reset password request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		codeText := http.StatusText(resp.StatusCode)
		if len(body) > 0 {
			return fmt.Errorf("reset password failed: %d(%s), %s", resp.StatusCode, codeText, string(body))
		}
		return fmt.Errorf("reset password failed: %d(%s)", resp.StatusCode, codeText)
	}

	fmt.Printf("Password for user '%s' reset successfully\n", o.username)
	return nil
}

func saltedMD5(s string) string {
	sum := md5.Sum([]byte(s + passwordSaltSuffix))
	return hex.EncodeToString(sum[:])
}
