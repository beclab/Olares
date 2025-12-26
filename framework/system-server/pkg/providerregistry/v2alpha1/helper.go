package v2alpha1

import (
	"fmt"
	"net/http"
	"strings"

	"bytetrade.io/web3os/system-server/pkg/constants"
)

var (
	headerXForwardedHost = "X-Forwarded-Host"
	headerXProviderProxy = "X-Provider-Proxy"
)

func ProviderServiceAddr(providerRef string) string {
	token := strings.Split(providerRef, "/")
	if len(token) == 1 {
		return token[0]
	}

	return fmt.Sprintf("%s.%s", token[1], token[0])
}

func ProviderRefFromHost(host string) string {
	token := strings.Split(strings.Split(host, ":")[0], ".")
	if len(token) < 2 {
		return fmt.Sprintf("%s.user-space-%s", host, constants.Owner)
	}

	return fmt.Sprintf("%s/%s", token[1], token[0])
}

func ProviderRefName(appName, namespace string) string {
	if len(namespace) == 0 {
		return appName
	}

	return fmt.Sprintf("%s/%s", namespace, appName)
}

// GetXForwardedHost returns the content of the X-Forwarded-URI header, falling back to the start-line request path.
func GetXForwardedHost(req *http.Request) (uri string) {
	for _, uriFunc := range []func() string{
		func() string { return req.Header.Get(headerXProviderProxy) },
		func() string { return req.Header.Get(headerXForwardedHost) },
	} {
		uri = uriFunc()
		if len(uri) > 0 {
			if !strings.HasPrefix(uri, "https://") && !strings.HasPrefix(uri, "http://") {
				uri = "http://" + uri
			}
			return uri
		}
	}

	return req.URL.String() // default
}
