package permission

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"

	"bytetrade.io/web3os/system-server/pkg/constants"
	"github.com/emicklei/go-restful/v3"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nonceAudience = "olares-system-server-nonce"
	nonceCaller   = "system:serviceaccount:os-framework:authelia-backend"
)

// nonce is also protected by the inbound mesh route. Verify a bound, audience-
// scoped credential here so missing mesh policy or a direct connection cannot
// turn the endpoint into an unauthenticated credential dispenser.
func (h *Handler) nonce(req *restful.Request, resp *restful.Response) {
	resp.Header().Set("Cache-Control", "no-store")
	headers := req.Request.Header.Values("Authorization")
	if len(headers) != 1 {
		http.Error(resp, "nonce access denied", http.StatusForbidden)
		return
	}
	fields := strings.Fields(headers[0])
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") || len(fields[1]) > 16384 {
		http.Error(resp, "nonce access denied", http.StatusForbidden)
		return
	}

	ctx, cancel := context.WithTimeout(req.Request.Context(), time.Second)
	defer cancel()
	review, err := h.kubeClientSet.AuthenticationV1().TokenReviews().Create(ctx,
		&authenticationv1.TokenReview{Spec: authenticationv1.TokenReviewSpec{
			Token: fields[1], Audiences: []string{nonceAudience},
		}}, metav1.CreateOptions{})
	if err != nil {
		// Do not log the review object, token, or API error response body.
		http.Error(resp, "nonce authentication unavailable", http.StatusServiceUnavailable)
		return
	}
	if review == nil || !review.Status.Authenticated || review.Status.Error != "" ||
		review.Status.User.Username != nonceCaller || !slices.Contains(review.Status.Audiences, nonceAudience) {
		http.Error(resp, "nonce access denied", http.StatusForbidden)
		return
	}
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.Write([]byte(constants.Nonce))
}
