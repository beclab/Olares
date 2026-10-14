package capi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/apiserver/authn"
	"bytetrade.io/web3os/bfl/pkg/constants"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/types"
)

type objectList struct {
	Items []map[string]any `json:"items"`
}

type namespaceUser struct {
	Name      string
	OwnerRole string
}

type namespaceGroupResponse struct {
	Title string           `json:"title"`
	Data  []map[string]any `json:"data"`
}

func (h *handler) namespaceGroup(req *restful.Request, resp *restful.Response) {
	identity, err := authn.Authenticate(req.Request)
	if err != nil {
		writeJSON(resp, http.StatusUnauthorized, map[string]string{"message": err.Error()})
		return
	}

	query := url.Values{
		"sortBy":        {"createTime"},
		"labelSelector": {"kubesphere.io/workspace!=kubesphere.io/devopsproject"},
	}
	var namespaces objectList
	if err := h.getKSJSON(req.Request.Context(), identity.Token,
		"/kapis/resources.kubesphere.io/v1alpha3/namespaces", query, &namespaces); err != nil {
		writeUpstreamError(resp, err)
		return
	}
	var users objectList
	if err := h.getKSJSON(req.Request.Context(), identity.Token,
		"/kapis/iam.kubesphere.io/v1alpha2/users", nil, &users); err != nil {
		writeUpstreamError(resp, err)
		return
	}

	admin, err := h.isAdmin(req.Request.Context(), identity.Username)
	if err != nil {
		writeUpstreamError(resp, err)
		return
	}
	result := buildNamespaceGroups(namespaces.Items, parseNamespaceUsers(users.Items), identity.Username, admin)
	writeJSON(resp, http.StatusOK, result)
}

func (h *handler) isAdmin(ctx context.Context, username string) (bool, error) {
	var user iamv1alpha2.User
	if err := h.kubeClient.CtrlClient().Get(ctx, types.NamespacedName{Name: username}, &user); err != nil {
		return false, fmt.Errorf("get user role: %w", err)
	}
	role := user.Annotations[constants.UserAnnotationOwnerRole]
	return role == constants.RoleOwner || role == constants.RoleAdmin, nil
}

func parseNamespaceUsers(items []map[string]any) []namespaceUser {
	users := make([]namespaceUser, 0, len(items))
	for _, item := range items {
		name := nestedString(item, "metadata", "name")
		if name == "" {
			continue
		}
		users = append(users, namespaceUser{
			Name:      name,
			OwnerRole: nestedString(item, "metadata", "annotations", constants.UserAnnotationOwnerRole),
		})
	}
	return users
}

func buildNamespaceGroups(namespaces []map[string]any, users []namespaceUser, username string, admin bool) []namespaceGroupResponse {
	if !admin {
		filtered := namespaces[:0:0]
		for _, namespace := range namespaces {
			if ownedNamespace(nestedString(namespace, "metadata", "name"), username) {
				filtered = append(filtered, namespace)
			}
		}
		namespaces = filtered
	}

	mainIndex := -1
	for i, user := range users {
		if user.OwnerRole == constants.RoleOwner {
			mainIndex = i
			break
		}
	}
	if mainIndex < 0 {
		for i, user := range users {
			if user.OwnerRole == constants.RoleAdmin {
				mainIndex = i
				break
			}
		}
	}
	if mainIndex < 0 && len(users) > 0 {
		mainIndex = len(users) - 1
	}
	if mainIndex > 0 {
		users[0], users[mainIndex] = users[mainIndex], users[0]
	}

	buckets := make(map[string][]map[string]any, len(users))
	shared := make([]map[string]any, 0)
	system := make([]map[string]any, 0)
	for _, namespace := range namespaces {
		name := nestedString(namespace, "metadata", "name")
		if sharedNamespace(namespace) {
			shared = append(shared, namespace)
			continue
		}
		matched := false
		for _, user := range users {
			if ownedNamespace(name, user.Name) {
				buckets[user.Name] = append(buckets[user.Name], namespace)
				matched = true
				break
			}
		}
		if !matched {
			system = append(system, namespace)
		}
	}

	result := make([]namespaceGroupResponse, 0, len(users)+2)
	for _, user := range users {
		if len(buckets[user.Name]) > 0 {
			result = append(result, namespaceGroupResponse{Title: user.Name, Data: buckets[user.Name]})
		}
	}
	if admin && len(shared) > 0 {
		result = append(result, namespaceGroupResponse{Title: "Shared", Data: shared})
	}
	if admin && len(system) > 0 {
		result = append(result, namespaceGroupResponse{Title: "System", Data: system})
	}
	return result
}

func ownedNamespace(namespace, username string) bool {
	return namespace != "" && username != "" && strings.HasSuffix(namespace, "-"+username)
}

func sharedNamespace(namespace map[string]any) bool {
	name := nestedString(namespace, "metadata", "name")
	if strings.HasSuffix(name, "-shared") {
		return true
	}
	for _, key := range []string{"bytetrade.io/ns-shared", "app.bytetrade.io/app-shared"} {
		if nestedString(namespace, "metadata", "labels", key) == "true" {
			return true
		}
	}
	return false
}

func nestedString(object map[string]any, path ...string) string {
	var current any = object
	for _, part := range path {
		mapping, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = mapping[part]
	}
	value, _ := current.(string)
	return value
}
