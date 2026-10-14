package capi

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/apiserver/authn"
	"bytetrade.io/web3os/bfl/pkg/constants"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	namespaceWorkspaceLabel = "kubesphere.io/workspace"
	devopsProjectWorkspace  = "kubesphere.io/devopsproject"
)

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

	admin, err := h.isAdmin(req.Request.Context(), identity.Username)
	if err != nil {
		writeUpstreamError(resp, err)
		return
	}

	namespaces, err := h.listNamespaceObjects(req.Request.Context())
	if err != nil {
		writeUpstreamError(resp, err)
		return
	}

	users := []namespaceUser{{Name: identity.Username}}
	if admin {
		users, err = h.listNamespaceUsers(req.Request.Context())
		if err != nil {
			writeUpstreamError(resp, err)
			return
		}
	}

	result := buildNamespaceGroups(namespaces, users, identity.Username, admin)
	writeJSON(resp, http.StatusOK, result)
}

func (h *handler) listNamespaceObjects(ctx context.Context) ([]map[string]any, error) {
	namespaces, err := h.kubeClient.Kubernetes().CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	return namespaceObjects(namespaces.Items)
}

func namespaceObjects(namespaces []corev1.Namespace) ([]map[string]any, error) {
	// The legacy KubeSphere endpoint accepts
	// kubesphere.io/workspace!=kubesphere.io/devopsproject even though the
	// comparison value contains a slash and is invalid in a native Kubernetes
	// label selector. Apply the same comparison after the client-go list.
	filtered := namespaces[:0:0]
	for i := range namespaces {
		if namespaces[i].Labels[namespaceWorkspaceLabel] != devopsProjectWorkspace {
			filtered = append(filtered, namespaces[i])
		}
	}
	namespaces = filtered

	// KubeSphere's sortBy=createTime orders newest objects first and uses the
	// name as a descending tie breaker. Preserve that response order.
	sort.SliceStable(namespaces, func(i, j int) bool {
		left := namespaces[i].CreationTimestamp.Time
		right := namespaces[j].CreationTimestamp.Time
		if left.Equal(right) {
			return namespaces[i].Name > namespaces[j].Name
		}
		return left.After(right)
	})

	items := make([]map[string]any, 0, len(namespaces))
	for i := range namespaces {
		item, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&namespaces[i])
		if err != nil {
			return nil, fmt.Errorf("convert namespace %q: %w", namespaces[i].Name, err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (h *handler) listNamespaceUsers(ctx context.Context) ([]namespaceUser, error) {
	var users iamv1alpha2.UserList
	if err := h.kubeClient.CtrlClient().List(ctx, &users); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return namespaceUsers(users.Items), nil
}

func namespaceUsers(users []iamv1alpha2.User) []namespaceUser {
	result := make([]namespaceUser, 0, len(users))
	for i := range users {
		result = append(result, namespaceUser{
			Name:      users[i].Name,
			OwnerRole: users[i].Annotations[constants.UserAnnotationOwnerRole],
		})
	}
	return result
}

func (h *handler) isAdmin(ctx context.Context, username string) (bool, error) {
	var user iamv1alpha2.User
	if err := h.kubeClient.CtrlClient().Get(ctx, types.NamespacedName{Name: username}, &user); err != nil {
		return false, fmt.Errorf("get user role: %w", err)
	}
	role := user.Annotations[constants.UserAnnotationOwnerRole]
	return role == constants.RoleOwner || role == constants.RoleAdmin, nil
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
