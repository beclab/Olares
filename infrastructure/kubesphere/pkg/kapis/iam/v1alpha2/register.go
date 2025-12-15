/*
Copyright 2019 The KubeSphere Authors.

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
	"net/http"

	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

const (
	GroupName = "iam.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(container *restful.Container, im im.IdentityManagementInterface, am am.AccessManagementInterface, authorizer authorizer.Authorizer) error {
	ws := runtime.NewWebService(GroupVersion)
	handler := newIAMHandler(im, am, authorizer)

	// users
	ws.Route(ws.POST("/users").
		To(handler.CreateUser).
		Doc("Create a global user account.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Reads(iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))

	ws.Route(ws.DELETE("/users/{user}").
		To(handler.DeleteUser).
		Doc("Delete the specified user.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))
	ws.Route(ws.PUT("/users/{user}").
		To(handler.UpdateUser).
		Doc("Update user profile.").
		Reads(iamv1alpha2.User{}).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))

	ws.Route(ws.GET("/users/{user}").
		To(handler.DescribeUser).
		Doc("Retrieve user details.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))
	ws.Route(ws.GET("/users").
		To(handler.ListUsers).
		Doc("List all users.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))

	ws.Route(ws.GET("/lldap/users").
		To(handler.ListLLdapUsers).
		Doc("List all users sync to lldap").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))

	ws.Route(ws.GET("/lldap/groups").
		To(handler.ListLLdapGroups).
		Doc("List all groups sync to lldap").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))

	// clusterroles
	ws.Route(ws.POST("/clusterroles").
		To(handler.CreateClusterRole).
		Doc("Create cluster role.").
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.DELETE("/clusterroles/{clusterrole}").
		To(handler.DeleteClusterRole).
		Doc("Delete cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.PUT("/clusterroles/{clusterrole}").
		To(handler.UpdateClusterRole).
		Doc("Update cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.PATCH("/clusterroles/{clusterrole}").
		To(handler.PatchClusterRole).
		Doc("Patch cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.GET("/clusterroles").
		To(handler.ListClusterRoles).
		Doc("List all cluster roles.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.ClusterRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.GET("/clusterroles/{clusterrole}").
		To(handler.DescribeClusterRole).
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Doc("Retrieve cluster role details.").
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))

	// globalroles
	ws.Route(ws.POST("/globalroles").
		To(handler.CreateGlobalRole).
		Doc("Create global role.").
		Reads(iamv1alpha2.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))

	ws.Route(ws.DELETE("/globalroles/{globalrole}").
		To(handler.DeleteGlobalRole).
		Doc("Delete global role.").
		Param(ws.PathParameter("globalrole", "global role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))

	ws.Route(ws.PUT("/globalroles/{globalrole}").
		To(handler.UpdateGlobalRole).
		Doc("Update global role.").
		Param(ws.PathParameter("globalrole", "global role name")).
		Reads(iamv1alpha2.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))
	ws.Route(ws.PATCH("/globalroles/{globalrole}").
		To(handler.PatchGlobalRole).
		Doc("Patch global role.").
		Param(ws.PathParameter("globalrole", "global role name")).
		Reads(iamv1alpha2.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))

	ws.Route(ws.GET("/globalroles").
		To(handler.ListGlobalRoles).
		Doc("List all global roles.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.GlobalRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))

	ws.Route(ws.GET("/globalroles/{globalrole}").
		To(handler.DescribeGlobalRole).
		Param(ws.PathParameter("globalrole", "global role name")).
		Doc("Retrieve global role details.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))

	ws.Route(ws.GET("/users/{user}/globalroles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve user's global role templates.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.GlobalRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))

	// roles
	ws.Route(ws.POST("/namespaces/{namespace}/roles").
		To(handler.CreateNamespaceRole).
		Doc("Create role in the specified namespace.").
		Reads(rbacv1.Role{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.DELETE("/namespaces/{namespace}/roles/{role}").
		To(handler.DeleteNamespaceRole).
		Doc("Delete role in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.PUT("/namespaces/{namespace}/roles/{role}").
		To(handler.UpdateNamespaceRole).
		Doc("Update namespace role.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.PATCH("/namespaces/{namespace}/roles/{role}").
		To(handler.PatchNamespaceRole).
		Doc("Patch namespace role.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles").
		To(handler.ListRoles).
		Doc("List all roles in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}").
		To(handler.DescribeNamespaceRole).
		Doc("Retrieve role details.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/members/{member}/roles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve member's role templates in namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("member", "namespace member's username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))

	// namespace rolebinding
	ws.Route(ws.POST("/namespaces/{namespace}/rolebindings").
		To(handler.CreateRoleBinding).
		Doc("Create rolebinding in the specified namespace.").
		Reads([]v1.RoleBinding{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, []v1.RoleBinding{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))

	ws.Route(ws.DELETE("/namespaces/{namespace}/rolebindings/{rolebinding}").
		To(handler.DeleteRoleBinding).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("namespace", "groupbinding name")).
		Param(ws.PathParameter("rolebinding", "groupbinding name")).
		Doc("Delete rolebinding under namespace.").
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	container.Add(ws)
	return nil
}
