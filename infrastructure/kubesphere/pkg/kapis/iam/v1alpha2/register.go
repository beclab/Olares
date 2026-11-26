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
	"k8s.io/apimachinery/pkg/runtime/schema"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
)

const (
	GroupName = "iam.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(container *restful.Container, im im.IdentityManagementInterface, am am.AccessManagementInterface, authorizer authorizer.Authorizer) error {
	ws := runtime.NewWebService(GroupVersion)
	handler := newIAMHandler(im, am, authorizer)

	// users
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

	// clusterroles
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

	container.Add(ws)
	return nil
}
