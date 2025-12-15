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
	"fmt"
	"k8s.io/klog"
	"strings"

	"github.com/emicklei/go-restful"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/request"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
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

func (h *iamHandler) ListLLdapGroups(request *restful.Request, response *restful.Response) {
	result, err := h.im.ListLLdapGroups(query.New())
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	response.WriteEntity(result)
}

func appendGlobalRoleAnnotation(user *iamv1alpha2.User, globalRole string) *iamv1alpha2.User {
	if user.Annotations == nil {
		user.Annotations = make(map[string]string, 0)
	}
	user.Annotations[iamv1alpha2.GlobalRoleAnnotation] = globalRole
	return user
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

func (h *iamHandler) CreateUser(req *restful.Request, resp *restful.Response) {
	var user iamv1alpha2.User
	err := req.ReadEntity(&user)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	operator, ok := request.UserFrom(req.Request.Context())
	if ok && operator.GetName() == iamv1alpha2.PreRegistrationUser {
		extra := operator.GetExtra()
		// The token used for registration must contain additional information
		if len(extra[iamv1alpha2.ExtraIdentityProvider]) != 1 || len(extra[iamv1alpha2.ExtraUID]) != 1 {
			err = errors.NewBadRequest("invalid registration token")
			api.HandleBadRequest(resp, req, err)
			return
		}
		if user.Labels == nil {
			user.Labels = make(map[string]string)
		}
		user.Labels[iamv1alpha2.IdentifyProviderLabel] = extra[iamv1alpha2.ExtraIdentityProvider][0]
		user.Labels[iamv1alpha2.OriginUIDLabel] = extra[iamv1alpha2.ExtraUID][0]
		// default role
		delete(user.Annotations, iamv1alpha2.GlobalRoleAnnotation)
	}

	created, err := h.im.CreateUser(&user)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(created)
}

func (h *iamHandler) UpdateUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")

	var user iamv1alpha2.User

	err := request.ReadEntity(&user)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if username != user.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", user.Name, username)
		api.HandleBadRequest(response, request, err)
		return
	}

	//globalRole := user.Annotations[iamv1alpha2.GlobalRoleAnnotation]
	//delete(user.Annotations, iamv1alpha2.GlobalRoleAnnotation)

	updated, err := h.im.UpdateUser(&user)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	//operator, ok := apirequest.UserFrom(request.Request.Context())
	//if globalRole != "" && ok {
	//	err = h.updateGlobalRoleBinding(operator, updated, globalRole)
	//	if err != nil {
	//		api.HandleError(response, request, err)
	//		return
	//	}
	//	updated = appendGlobalRoleAnnotation(updated, globalRole)
	//}

	response.WriteEntity(updated)
}

func (h *iamHandler) DeleteUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")

	err := h.im.DeleteUser(username)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) CreateClusterRole(request *restful.Request, response *restful.Response) {
	var clusterRole rbacv1.ClusterRole
	err := request.ReadEntity(&clusterRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateClusterRole(&clusterRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteClusterRole(request *restful.Request, response *restful.Response) {
	clusterrole := request.PathParameter("clusterrole")

	err := h.am.DeleteClusterRole(clusterrole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateClusterRole(request *restful.Request, response *restful.Response) {
	clusterRoleName := request.PathParameter("clusterrole")

	var clusterRole rbacv1.ClusterRole

	err := request.ReadEntity(&clusterRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if clusterRoleName != clusterRole.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", clusterRole.Name, clusterRoleName)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateClusterRole(&clusterRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
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

func (h *iamHandler) CreateNamespaceRole(request *restful.Request, response *restful.Response) {

	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	var role rbacv1.Role
	err = request.ReadEntity(&role)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateNamespaceRole(namespace, &role)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteNamespaceRole(request *restful.Request, response *restful.Response) {
	role := request.PathParameter("role")

	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	err = h.am.DeleteNamespaceRole(namespace, role)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateNamespaceRole(request *restful.Request, response *restful.Response) {
	roleName := request.PathParameter("role")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	var role rbacv1.Role
	err = request.ReadEntity(&role)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if roleName != role.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", role.Name, roleName)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateNamespaceRole(namespace, &role)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
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

func (h *iamHandler) PatchNamespaceRole(request *restful.Request, response *restful.Response) {
	roleName := request.PathParameter("role")
	namespaceName, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	var role rbacv1.Role
	err = request.ReadEntity(&role)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	role.Name = roleName
	patched, err := h.am.PatchNamespaceRole(namespaceName, &role)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *iamHandler) PatchClusterRole(request *restful.Request, response *restful.Response) {
	clusterRoleName := request.PathParameter("clusterrole")

	var clusterRole rbacv1.ClusterRole
	err := request.ReadEntity(&clusterRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	clusterRole.Name = clusterRoleName
	patched, err := h.am.PatchClusterRole(&clusterRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *iamHandler) CreateRoleBinding(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	var roleBindings []rbacv1.RoleBinding
	err := request.ReadEntity(&roleBindings)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	var results []rbacv1.RoleBinding
	for _, item := range roleBindings {
		r, err := h.am.CreateRoleBinding(namespace, &item)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
		results = append(results, *r)
	}

	response.WriteEntity(results)
}

func (h *iamHandler) DeleteRoleBinding(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("rolebinding")
	namespace := request.PathParameter("namespace")

	err := h.am.DeleteRoleBinding(namespace, name)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
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

func (h *iamHandler) CreateGlobalRole(request *restful.Request, response *restful.Response) {

	var globalRole iamv1alpha2.GlobalRole
	err := request.ReadEntity(&globalRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateGlobalRole(&globalRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteGlobalRole(request *restful.Request, response *restful.Response) {
	globalRole := request.PathParameter("globalrole")
	err := h.am.DeleteGlobalRole(globalRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateGlobalRole(request *restful.Request, response *restful.Response) {
	globalRoleName := request.PathParameter("globalrole")

	var globalRole iamv1alpha2.GlobalRole
	err := request.ReadEntity(&globalRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if globalRoleName != globalRole.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", globalRole.Name, globalRoleName)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateGlobalRole(&globalRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
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

func (h *iamHandler) PatchGlobalRole(request *restful.Request, response *restful.Response) {
	globalRoleName := request.PathParameter("globalrole")

	var globalRole iamv1alpha2.GlobalRole
	err := request.ReadEntity(&globalRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	globalRole.Name = globalRoleName
	patched, err := h.am.PatchGlobalRole(&globalRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(patched)
}
