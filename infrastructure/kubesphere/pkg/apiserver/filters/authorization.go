/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package filters

import (
	"context"
	"errors"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"net/http"
)

// WithAuthorization passes all authorized requests on to handler, and returns forbidden error otherwise.
func WithAuthorization(handler http.Handler, authorizers authorizer.Authorizer) http.Handler {
	if authorizers == nil {
		return handler
	}

	defaultSerializer := serializer.NewCodecFactory(runtime.NewScheme()).WithoutConversion()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		attributes, err := getAuthorizerAttributes(ctx)
		if err != nil {
			responsewriters.InternalError(w, req, err)
		}

		authorized, reason, err := authorizers.Authorize(attributes)
		if authorized == authorizer.DecisionAllow {
			handler.ServeHTTP(w, req)
			return
		}

		if err != nil {
			responsewriters.InternalError(w, req, err)
			return
		}

		responsewriters.Forbidden(ctx, attributes, w, req, reason, defaultSerializer)
	})
}

func getAuthorizerAttributes(ctx context.Context) (authorizer.Attributes, error) {
	attribs := authorizer.AttributesRecord{}

	user, ok := request.UserFrom(ctx)
	if ok {
		attribs.User = user
	}

	requestInfo, found := request.RequestInfoFrom(ctx)
	if !found {
		return nil, errors.New("no RequestInfo found in the context")
	}

	// Start with common attributes that apply to resource and non-resource requests
	attribs.ResourceScope = requestInfo.ResourceScope
	attribs.ResourceRequest = requestInfo.IsResourceRequest
	attribs.Path = requestInfo.Path
	attribs.Verb = requestInfo.Verb
	attribs.KubernetesRequest = requestInfo.IsKubernetesRequest
	attribs.APIGroup = requestInfo.APIGroup
	attribs.APIVersion = requestInfo.APIVersion
	attribs.Resource = requestInfo.Resource
	attribs.Subresource = requestInfo.Subresource
	attribs.Namespace = requestInfo.Namespace
	attribs.Name = requestInfo.Name

	return &attribs, nil
}

// if using basic auth. But only treats request with requestURI `/oauth/authorize` as login attempt
func WithAuthentication(handler http.Handler, authRequest authenticator.Request) http.Handler {
	if authRequest == nil {
		klog.Warningf("Authentication is disabled")
		return handler
	}
	s := serializer.NewCodecFactory(runtime.NewScheme()).WithoutConversion()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		resp, ok, err := authRequest.AuthenticateRequest(req)
		_, _, usingBasicAuth := req.BasicAuth()

		defer func() {
			// if we authenticated successfully, go ahead and remove the bearer token so that no one
			// is ever tempted to use it inside of the API server
			if usingBasicAuth && ok {
				req.Header.Del("Authorization")
			}
		}()

		if err != nil || !ok {
			ctx := req.Context()
			requestInfo, found := request.RequestInfoFrom(ctx)
			if !found {
				responsewriters.InternalError(w, req, errors.New("no RequestInfo found in the context"))
				return
			}
			gv := schema.GroupVersion{Group: requestInfo.APIGroup, Version: requestInfo.APIVersion}
			responsewriters.ErrorNegotiated(apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)), s, gv, w, req)
			return
		}

		req = req.WithContext(request.WithUser(req.Context(), resp.User))
		handler.ServeHTTP(w, req)
	})
}
