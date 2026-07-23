package callerjwt

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-jose/go-jose/v4"
)

const (
	JWKSServiceName      = "caller-jwt-jwks"
	// JWKS Service must live with app-service pods (selector is namespaced).
	JWKSServiceNamespace = "os-framework"
	JWKSPath             = "/.well-known/jwks.json"
	jwksListenEnv        = "CALLER_JWT_JWKS_LISTEN"
	defaultJWKSListen    = ":8444"
	tlsCertEnv           = "WEBHOOK_TLS_CERT"
	tlsKeyEnv            = "WEBHOOK_TLS_KEY"
	defaultCertPath      = "/etc/certs/server.crt"
	defaultKeyPath       = "/etc/certs/server.key"
	jwksServicePort      = int32(443)
	jwksTargetPortName   = "jwks"

	// JWKSIngressNPName is the NetworkPolicy that allows Envoy Gateway pods
	// in os-gateway to fetch JWKS on the app-service container port.
	JWKSIngressNPName           = "allow-app-gateway-caller-jwt-jwks"
	JWKSIngressNPFromNamespace  = "os-gateway"
	JWKSIngressNPComponentValue = "caller-jwt"
	jwksAppServiceSelectorKey   = "tier"
	jwksAppServiceSelectorValue = "app-service"
	managedByComponentLabel     = "app.kubernetes.io/component"
)

// BuildJWKS returns the JSON Web Key Set for the issuer key ring.
func BuildJWKS(ring KeyRing) (jose.JSONWebKeySet, error) {
	if ring.Active.PrivateKey == nil {
		return jose.JSONWebKeySet{}, errors.New("callerjwt: active key required for jwks")
	}
	set := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Key:       ring.Active.PrivateKey.Public(),
				KeyID:     ring.Active.KID,
				Algorithm: string(jose.RS256),
				Use:       "sig",
			},
		},
	}
	if ring.Previous != nil && ring.Previous.PrivateKey != nil {
		set.Keys = append(set.Keys, jose.JSONWebKey{
			Key:       ring.Previous.PrivateKey.Public(),
			KeyID:     ring.Previous.KID,
			Algorithm: string(jose.RS256),
			Use:       "sig",
		})
	}
	return set, nil
}

// JWKSHandler serves the issuer public keys at JWKSPath.
func JWKSHandler(issuer *Issuer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != JWKSPath {
			http.NotFound(w, r)
			return
		}
		if issuer == nil {
			http.Error(w, "issuer unavailable", http.StatusServiceUnavailable)
			return
		}
		set, err := BuildJWKS(issuer.Keys())
		if err != nil {
			http.Error(w, "jwks build failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(set)
	})
}

// JWKSServer serves JWKS over HTTPS for EG remoteJWKS backendRefs.
type JWKSServer struct {
	Issuer     *Issuer
	Reconciler *IssuerReconciler
	Addr       string
}

// Start implements manager.Runnable. Cache is synced before Start, so key
// materialization is safe here (unlike SetupWithManager).
func (s *JWKSServer) Start(ctx context.Context) error {
	if s == nil {
		<-ctx.Done()
		return nil
	}
	issuer := s.Issuer
	if issuer == nil && s.Reconciler != nil {
		if err := s.Reconciler.ensureIssuer(ctx); err != nil {
			return fmt.Errorf("init caller jwt issuer for jwks: %w", err)
		}
		issuer = s.Reconciler.Issuer()
	}
	if issuer == nil {
		<-ctx.Done()
		return nil
	}
	addr := s.Addr
	if addr == "" {
		addr = jwksListenAddress()
	}
	certFile, keyFile := tlsCertPaths()
	server := &http.Server{
		Addr:              addr,
		Handler:           JWKSHandler(issuer),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	if certFile != "" && keyFile != "" {
		if _, err := os.Stat(certFile); err == nil {
			if _, err := os.Stat(keyFile); err == nil {
				cert, err := tls.LoadX509KeyPair(certFile, keyFile)
				if err != nil {
					return fmt.Errorf("load jwks tls cert: %w", err)
				}
				server.TLSConfig = &tls.Config{
					MinVersion:   tls.VersionTLS12,
					Certificates: []tls.Certificate{cert},
				}
				err = server.ListenAndServeTLS("", "")
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return err
			}
		}
	}
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func jwksListenAddress() string {
	if v := os.Getenv(jwksListenEnv); v != "" {
		return v
	}
	return defaultJWKSListen
}

func tlsCertPaths() (string, string) {
	cert := defaultCertPath
	key := defaultKeyPath
	if v := os.Getenv(tlsCertEnv); v != "" {
		cert = v
	}
	if v := os.Getenv(tlsKeyEnv); v != "" {
		key = v
	}
	return cert, key
}

// JWKSListenPort returns the container port name/number for the JWKS server.
func JWKSListenPort() (name string, port int32) {
	return jwksTargetPortName, parseListenPort(jwksListenAddress())
}

func parseListenPort(addr string) int32 {
	_, portStr, ok := splitHostPort(addr)
	if !ok || portStr == "" {
		return 8444
	}
	var port int
	_, _ = fmt.Sscanf(portStr, "%d", &port)
	if port <= 0 {
		return 8444
	}
	return int32(port)
}

func splitHostPort(addr string) (host, port string, ok bool) {
	if addr == "" {
		return "", "", false
	}
	if addr[0] == ':' {
		return "", addr[1:], true
	}
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i], addr[i+1:], true
		}
	}
	return addr, "", false
}

// VerifyJWKSResponse checks that encoded JWKS JSON is parseable.
func VerifyJWKSResponse(data []byte) (jose.JSONWebKeySet, error) {
	var set jose.JSONWebKeySet
	if err := json.Unmarshal(data, &set); err != nil {
		return jose.JSONWebKeySet{}, err
	}
	if len(set.Keys) == 0 {
		return jose.JSONWebKeySet{}, errors.New("callerjwt: empty jwks")
	}
	for _, key := range set.Keys {
		if key.Key == nil {
			return jose.JSONWebKeySet{}, errors.New("callerjwt: jwks key material missing")
		}
	}
	return set, nil
}
