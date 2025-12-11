package constants

import (
	"os"
	"strings"

	"github.com/google/uuid"
)

const (
	ProxyServerServiceName    = "system-server"
	ProxyServerListenAddress  = ":28080"
	APIServerListenAddress    = ":80"
	KubeSphereClientAttribute = "ksclient"
	AuthorizationTokenKey     = "X-Authorization"
	BflUserKey                = "X-BFL-USER"
	AuthTokenCookieName       = "auth_token"
	AutheliaNonceKey          = "Authelia-Nonce"
)

var (
	MyNamespace string
	Owner       string
	MyUserspace string
)

var (
	Nonce = uuid.New().String()
)

func init() {
	MyNamespace = os.Getenv("MY_NAMESPACE")
	Owner = os.Getenv("OWNER")
	MyUserspace = strings.Replace(MyNamespace, "user-system-", "user-space-", 1)
}
