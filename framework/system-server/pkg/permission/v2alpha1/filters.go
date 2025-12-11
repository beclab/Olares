package v2alpha1

import (
	"fmt"
	"net/http"
	"net/url"

	apiv1alpha1 "bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/constants"
	providerv2alpha1 "bytetrade.io/web3os/system-server/pkg/providerregistry/v2alpha1"
	"bytetrade.io/web3os/system-server/pkg/utils"
	"github.com/brancz/kube-rbac-proxy/pkg/authz"
	"github.com/brancz/kube-rbac-proxy/pkg/proxy"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/request"
	corev1 "k8s.io/client-go/listers/core/v1"

	"k8s.io/klog/v2"
)

func WithAuthentication(
	authReq authenticator.Request,
	audiences []string,
	handler http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if len(audiences) > 0 {
			ctx = authenticator.WithAudiences(ctx, audiences)
			req = req.WithContext(ctx)
		}

		res, ok, err := authReq.AuthenticateRequest(req)
		if err != nil {
			klog.Errorf("Unable to authenticate the request due to an error: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		req = req.WithContext(request.WithUser(req.Context(), res.User))
		handler.ServeHTTP(w, req)
	}
}

func WithAuthorization(
	authz Authorizer,
	cfg *authz.Config,
	handler http.HandlerFunc,
) http.HandlerFunc {
	getRequestAttributes := func(u user.Info, r *http.Request) []authorizer.Attributes {
		allAttrs := proxy.
			NewKubeRBACProxyAuthorizerAttributesGetter(cfg).
			GetRequestAttributes(u, r)

		for i, attrs := range allAttrs {
			if attrs.GetPath() != "" && !attrs.IsResourceRequest() {
				// for non-resource requests, setup the provider reference
				uri := providerv2alpha1.GetXForwardedHost(r)
				requestUrl, err := url.Parse(uri)
				if err != nil {
					klog.Errorf("failed to parse X-Forwarded-Host: %v", err)
					return nil
				}
				hostStr := requestUrl.Host
				if hostStr == "" {
					hostStr = r.Host
				}
				klog.V(5).Infof("RBAC: using provider host %q, url: %q", hostStr, uri)
				ref := hostStr

				path := attrs.GetPath()
				queryString := r.URL.RawQuery
				if queryString != "" {
					path = fmt.Sprintf("%s?%s", path, queryString)
				}
				a := authorizer.AttributesRecord{
					User:            attrs.GetUser(),
					Verb:            attrs.GetVerb(),
					Namespace:       attrs.GetNamespace(),
					APIGroup:        attrs.GetAPIGroup(),
					APIVersion:      attrs.GetAPIVersion(),
					Resource:        ref,
					Subresource:     attrs.GetSubresource(),
					Name:            attrs.GetName(),
					ResourceRequest: attrs.IsResourceRequest(),
					Path:            path,
				}
				allAttrs[i] = a
			}
		}

		return allAttrs
	}

	return func(w http.ResponseWriter, req *http.Request) {
		u, ok := request.UserFrom(req.Context())
		if !ok {
			http.Error(w, "user not in context", http.StatusBadRequest)
			return
		}

		// Get authorization attributes
		allAttrs := getRequestAttributes(u, req)
		if len(allAttrs) == 0 {
			msg := "Bad Request. The request or configuration is malformed."
			klog.V(2).Info(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		var service string
		for _, attrs := range allAttrs {
			// Authorize
			s, authorized, reason, err := authz.Authorize(req.Context(), attrs)
			if err != nil {
				msg := fmt.Sprintf("Authorization error (user=%s, verb=%s, resource=%s, subresource=%s)", u.GetName(), attrs.GetVerb(), attrs.GetResource(), attrs.GetSubresource())
				klog.Errorf("%s: %s", msg, err)
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}
			klog.V(5).Infof("Authorization result, %d, reason=%s", authorized, reason)
			if authorized != authorizer.DecisionAllow {
				msg := fmt.Sprintf("Forbidden (user=%s, verb=%s, resource=%s, subresource=%s)", u.GetName(), attrs.GetVerb(), attrs.GetResource(), attrs.GetSubresource())
				klog.V(2).Infof("%s. Reason: %q.", msg, reason)
				http.Error(w, msg, http.StatusForbidden)
				return
			}

			if s != "" {
				service = s
			}
		}

		if service != "" {
			req = req.WithContext(WithProviderService(req.Context(), service))
		}

		handler.ServeHTTP(w, req)
	}
}

func WithUserHeader(
	convert func(account string) string,
	handler http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		u, ok := request.UserFrom(req.Context())
		if ok {
			user := u.GetName()
			if convert != nil {
				user = convert(user)
			}
			req.Header.Set(constants.BflUserKey, user)
			req.Header.Set(apiv1alpha1.BackendTokenHeader, constants.Nonce)
		}

		handler.ServeHTTP(w, req)
	}
}

func MustHaveProviderService(
	handler http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		_, ok := ProviderServiceFrom(req.Context())
		if !ok {
			http.Error(w, "provider service not found", http.StatusBadRequest)
			return
		}

		handler.ServeHTTP(w, req)
	}
}

func RecoverHeader(
	handler http.HandlerFunc,
) http.HandlerFunc {
	recoverHeaders := map[string]string{
		"Temp-Authorization": "Authorization",
	}
	return func(w http.ResponseWriter, req *http.Request) {
		for tempHeader, originalHeader := range recoverHeaders {
			if val := req.Header.Get(tempHeader); val != "" {
				req.Header.Set(originalHeader, val)
				req.Header.Del(tempHeader)
			}
		}

		handler.ServeHTTP(w, req)
	}
}

func UserFromServiceAccount(namespaceLister corev1.NamespaceLister) func(string) string {
	return func(sa string) string {
		if isSA, saNamespace, _ := utils.IsServiceAccount(sa); isSA {
			if isUserNamespace, username := utils.IsUserNamespace(saNamespace); isUserNamespace {
				return username
			} // end of service account check

			// if not a user namespace, we should get owner from namespace
			ns, err := namespaceLister.Get(saNamespace)
			if err != nil {
				klog.Errorf("failed to get namespace %s: %v", saNamespace, err)
				return sa // return original service account if namespace not found
			}

			if owner, ok := ns.Labels["bytetrade.io/ns-owner"]; ok {
				return owner
			}
		}

		return sa
	}
}
