package permission

import (
	"net/http"

	"bytetrade.io/web3os/system-server/pkg/constants"
	"github.com/emicklei/go-restful/v3"
)

const noncePath = constants.AutheliaNoncePath

// The nonce route is authorized by Linkerd, not by a forwarded identity or the
// last TCP hop. The deployment must enroll the API port in default-deny mesh
// policy before enabling this handler. Other API routes have disjoint matches.
func (h *Handler) nonce(req *restful.Request, resp *restful.Response) {
	resp.Header().Set("Cache-Control", "no-store")
	if !h.nonceMeshRequired {
		http.Error(resp, "nonce mesh enrollment required", http.StatusServiceUnavailable)
		return
	}
	// Do not let a route authorized under another prefix reach the nonce via
	// backend path normalization, percent decoding, an absolute URI or query.
	if req.Request.Method != http.MethodGet || req.Request.RequestURI != noncePath || req.Request.URL.Path != noncePath {
		http.Error(resp, "noncanonical nonce request", http.StatusNotFound)
		return
	}
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.Write([]byte(constants.Nonce))
}
