/*
Copyright 2020 KubeSphere Authors

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

package v1alpha2

import (
	"strings"

	"k8s.io/klog"

	"github.com/emicklei/go-restful"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
)

type Member struct {
	Username string `json:"username"`
	RoleRef  string `json:"roleRef"`
}

type GroupMember struct {
	UserName  string `json:"userName"`
	GroupName string `json:"groupName"`
}

type PasswordReset struct {
	CurrentPassword string `json:"currentPassword"`
	Password        string `json:"password"`
}

type iamHandler struct {
	am         am.AccessManagementInterface
	im         im.IdentityManagementInterface
	authorizer authorizer.Authorizer
}

func newIAMHandler(im im.IdentityManagementInterface, am am.AccessManagementInterface, authorizer authorizer.Authorizer) *iamHandler {
	return &iamHandler{
		am:         am,
		im:         im,
		authorizer: authorizer,
	}
}

func (h *iamHandler) DescribeUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")

	user, err := h.im.DescribeUser(username)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(user)
}

func (h *iamHandler) RetrieveMemberRoleTemplates(request *restful.Request, response *restful.Response) {
	if strings.HasSuffix(request.Request.URL.Path, iamv1alpha2.ResourcesPluralGlobalRole) {
		username := request.PathParameter("user")

		globalRole, err := h.am.GetGlobalRoleOfUser(username)
		if err != nil {
			// if role binding not exist return empty list
			if errors.IsNotFound(err) {
				response.WriteEntity([]interface{}{})
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}

		result, err := h.am.ListGlobalRoles(&query.Query{
			Pagination: query.NoPagination,
			SortBy:     "",
			Ascending:  false,
			Filters:    map[query.Field]query.Value{iamv1alpha2.AggregateTo: query.Value(globalRole.Name)},
		})
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}
		response.WriteEntity(result.Items)
		return
	}
	if strings.HasSuffix(request.Request.URL.Path, iamv1alpha2.ResourcesPluralClusterRole) {
		username := request.PathParameter("clustermember")
		clusterRole, err := h.am.GetClusterRoleOfUser(username)
		if err != nil {
			// if role binding not exist return empty list
			if errors.IsNotFound(err) {
				response.WriteEntity([]interface{}{})
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}

		result, err := h.am.ListClusterRoles(&query.Query{
			Pagination: query.NoPagination,
			SortBy:     "",
			Ascending:  false,
			Filters:    map[query.Field]query.Value{iamv1alpha2.AggregateTo: query.Value(clusterRole.Name)},
		})
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}

		response.WriteEntity(result.Items)
		return
	}

	if strings.HasSuffix(request.Request.URL.Path, iamv1alpha2.ResourcesPluralRole) {
		namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
		username := request.PathParameter("member")
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}
		klog.Infof("username....: %s", username)

		user, err := h.im.DescribeUser(username)
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}

		roles, err := h.am.GetNamespaceRoleOfUser(username, user.Spec.Groups, namespace)
		if err != nil {
			// if role binding not exist return empty list
			if errors.IsNotFound(err) {
				response.WriteEntity([]interface{}{})
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}

		templateRoles := make(map[string]*rbacv1.Role)
		for _, role := range roles {
			// merge template Role
			result, err := h.am.ListRoles(namespace, &query.Query{
				Pagination: query.NoPagination,
				SortBy:     "",
				Ascending:  false,
				Filters:    map[query.Field]query.Value{iamv1alpha2.AggregateTo: query.Value(role.Name)},
			})

			if err != nil {
				api.HandleInternalError(response, request, err)
				return
			}

			for _, obj := range result.Items {
				templateRole := obj.(*rbacv1.Role)
				templateRoles[templateRole.Name] = templateRole
			}
		}

		results := make([]*rbacv1.Role, 0, len(templateRoles))
		for _, value := range templateRoles {
			results = append(results, value)
		}

		response.WriteEntity(results)
		return
	}
}

func (h *iamHandler) ListUsers(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	result, err := h.im.ListUsers(queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	for i, item := range result.Items {
		user := item.(*iamv1alpha2.User)
		user = user.DeepCopy()

		result.Items[i] = user
	}
	response.WriteEntity(result)
}

func (h *iamHandler) ListLLdapUsers(request *restful.Request, response *restful.Response) {
	result, err := h.im.ListLLdapUsers(query.New())
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	response.WriteEntity(result)
}

func (h *iamHandler) ListRoles(request *restful.Request, response *restful.Response) {
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	queryParam := query.ParseQueryParameter(request)
	result, err := h.am.ListRoles(namespace, queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	response.WriteEntity(result)
}

func (h *iamHandler) ListClusterRoles(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	result, err := h.am.ListClusterRoles(queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	response.WriteEntity(result)
}

func (h *iamHandler) DescribeClusterRole(request *restful.Request, response *restful.Response) {
	clusterRoleName := request.PathParameter("clusterrole")
	clusterRole, err := h.am.GetClusterRole(clusterRoleName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	response.WriteEntity(clusterRole)
}

func (h *iamHandler) DescribeNamespaceRole(request *restful.Request, response *restful.Response) {
	roleName := request.PathParameter("role")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	role, err := h.am.GetNamespaceRole(namespace, roleName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(role)
}

// resolve the namespace which controlled by the devops project
func (h *iamHandler) resolveNamespace(namespace string, devops string) (string, error) {

	return namespace, nil

	//return h.am.GetDevOpsRelatedNamespace(devops)
}

func (h *iamHandler) ListGlobalRoles(req *restful.Request, resp *restful.Response) {
	queryParam := query.ParseQueryParameter(req)
	result, err := h.am.ListGlobalRoles(queryParam)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *iamHandler) DescribeGlobalRole(request *restful.Request, response *restful.Response) {
	globalRoleName := request.PathParameter("globalrole")
	globalRole, err := h.am.GetGlobalRole(globalRoleName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	response.WriteEntity(globalRole)
}
