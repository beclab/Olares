package containerd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/containerd/containerd/remotes/docker/auth"
	remoteserrors "github.com/containerd/containerd/remotes/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/pelletier/go-toml"
)

var verifyRegistryAuth = validateRegistryAuth

// Check login before saving, using the candidate upstream and bypassing mirrors.
// This does not verify access to any particular image.
func validateRegistryAuth(ctx context.Context, registry string, tree *toml.Tree, credentials *RegistryAuth) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	view := registryConfigView(registry, tree, nil)
	server, err := url.Parse(view.Server)
	expected, _ := url.Parse(canonicalServer(registry))
	if err != nil || server.Host != expected.Host || server.User != nil ||
		(server.Scheme != "https" && server.Scheme != "http") ||
		(server.Path != "" && server.Path != "/") || server.RawQuery != "" || server.Fragment != "" {
		return fiber.NewError(400, "invalid registry server for credential validation")
	}
	tlsConfig, err := registryAuthTLS(registry, tree, view)
	if err != nil {
		return fiber.NewError(400, "cannot load registry TLS configuration for credential validation")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	defer transport.CloseIdleConnections()
	client := &http.Client{
		Transport: transport,
		// Never forward credentials through redirects or accept a login HTML page.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	server.Path = "/v2/"
	probe := func(token string, basic bool) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.String(), nil)
		if err != nil {
			return nil, err
		}
		if basic {
			req.SetBasicAuth(credentials.Username, credentials.Password)
		} else if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return client.Do(req)
	}
	response, err := probe("", false)
	if err != nil {
		return registryAuthRequestError(err)
	}
	response.Body.Close()
	if response.StatusCode == http.StatusOK {
		return fiber.NewError(422, "registry allows anonymous access; supplied credentials could not be verified")
	}
	if response.StatusCode != http.StatusUnauthorized {
		return registryAuthStatusError(response.StatusCode)
	}
	for _, challenge := range auth.ParseAuthHeader(response.Header) {
		var token string
		switch challenge.Scheme {
		case auth.BasicAuth:
		case auth.BearerAuth:
			options, err := auth.GenerateTokenOptions(ctx, server.Host, credentials.Username, credentials.Password, challenge)
			if err != nil {
				return fiber.NewError(502, "registry returned an invalid authentication challenge")
			}
			realm, err := url.Parse(options.Realm)
			if err != nil || !validRegistryAuthRealm(server, realm) {
				return fiber.NewError(422, "registry authentication service URL is invalid or downgrades HTTPS to HTTP")
			}
			result, err := auth.FetchToken(ctx, client, nil, options)
			if err != nil {
				return registryAuthRequestError(err)
			}
			token = result.Token
		default:
			continue
		}
		response, err := probe(token, challenge.Scheme == auth.BasicAuth)
		if err != nil {
			return registryAuthRequestError(err)
		}
		response.Body.Close()
		if response.StatusCode == http.StatusOK {
			return nil
		}
		return registryAuthStatusError(response.StatusCode)
	}
	return fiber.NewError(422, "registry does not provide a supported credential verification method")
}

// The configured registry selects its authentication service through the Bearer
// challenge, including services on a different origin. Check the URL and reject
// transport downgrades without imposing a platform or hostname allowlist.
func validRegistryAuthRealm(server, realm *url.URL) bool {
	if realm == nil || realm.Hostname() == "" || realm.User != nil || realm.Fragment != "" ||
		(realm.Scheme != "https" && realm.Scheme != "http") {
		return false
	}
	return server.Scheme != "https" || realm.Scheme == "https"
}

// Preserve imported upstream CA/client settings, resolving relative certificate
// paths against the registry directory just like containerd.
func registryAuthTLS(registry string, tree *toml.Tree, view *RegistryConfig) (*tls.Config, error) {
	config := &tls.Config{InsecureSkipVerify: view.SkipVerify}
	setting := func(key string) interface{} {
		if host, ok := tree.GetPath([]string{"host", view.Server}).(*toml.Tree); ok && host.Has(key) {
			return host.Get(key)
		}
		return tree.Get(key)
	}
	absolute := func(path string) string {
		if filepath.IsAbs(path) {
			return path
		}
		return filepath.Join(registryDir(registry), path)
	}
	files := func(value interface{}) ([]string, error) {
		switch v := value.(type) {
		case nil:
			return nil, nil
		case string:
			return []string{v}, nil
		case []interface{}:
			result := make([]string, len(v))
			for i, value := range v {
				var ok bool
				if result[i], ok = value.(string); !ok {
					return nil, errors.New("invalid TLS file list")
				}
			}
			return result, nil
		default:
			return nil, errors.New("invalid TLS file configuration")
		}
	}
	caFiles, err := files(setting("ca"))
	if err != nil {
		return nil, err
	}
	if len(caFiles) > 0 {
		config.RootCAs, err = x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		for _, file := range caFiles {
			data, err := os.ReadFile(absolute(file))
			if err != nil {
				return nil, err
			}
			if !config.RootCAs.AppendCertsFromPEM(data) {
				return nil, errors.New("invalid registry CA certificate")
			}
		}
	}
	var clients []interface{}
	switch value := setting("client").(type) {
	case nil:
	case string:
		clients = []interface{}{value}
	case []interface{}:
		clients = value
	default:
		return nil, errors.New("invalid registry client certificate configuration")
	}
	for _, value := range clients {
		pair, err := files(value)
		if err != nil || len(pair) == 0 || len(pair) > 2 {
			return nil, errors.New("invalid registry client certificate pair")
		}
		key := pair[0]
		if len(pair) == 2 && pair[1] != "" {
			key = pair[1]
		}
		certificate, err := tls.LoadX509KeyPair(absolute(pair[0]), absolute(key))
		if err != nil {
			return nil, err
		}
		config.Certificates = append(config.Certificates, certificate)
	}
	return config, nil
}

// Do not expose remote bodies, URLs, or raw transport errors, which may contain
// submitted credentials or the temporary registry token.
func registryAuthRequestError(err error) error {
	var status remoteserrors.ErrUnexpectedStatus
	if errors.As(err, &status) {
		return registryAuthStatusError(status.StatusCode)
	}
	var network net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &network) && network.Timeout()) {
		return fiber.NewError(504, "registry credential validation timed out")
	}
	if errors.Is(err, auth.ErrNoToken) {
		return fiber.NewError(502, "registry authentication service did not return a token")
	}
	return fiber.NewError(502, "registry credential validation failed: connection, TLS, or authentication service error")
}

func registryAuthStatusError(status int) error {
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return fiber.NewError(400, "registry rejected the supplied credentials or denied registry access")
	}
	return fiber.NewError(502, "registry credential validation received an unexpected response")
}
