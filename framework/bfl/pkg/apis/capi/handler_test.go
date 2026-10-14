package capi

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"bytetrade.io/web3os/bfl/pkg/constants"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestGlobalRole(t *testing.T) {
	tests := map[string]string{
		constants.RoleOwner:  "platform-admin",
		constants.RoleAdmin:  "platform-admin",
		constants.RoleNormal: constants.RoleNormal,
		"":                   "",
	}
	for input, want := range tests {
		if got := globalRole(input); got != want {
			t.Fatalf("globalRole(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestUnwrapData(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "envelope", body: `{"code":0,"data":[{"name":"eth0"}]}`, want: `[{"name":"eth0"}]`},
		{name: "raw object", body: `{"name":"eth0"}`, want: `{"name":"eth0"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unwrapData([]byte(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			if !json.Valid(got) || string(got) != tt.want {
				t.Fatalf("unwrapData() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAppDetailResponseOmitsRemovedFields(t *testing.T) {
	payload, err := json.Marshal(appDetailResponse{})
	if err != nil {
		t.Fatal(err)
	}
	var response map[string]any
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"clusterRole", "localeManifest", "gitopsEngine"} {
		if _, exists := response[field]; exists {
			t.Errorf("app detail response still contains removed field %q", field)
		}
	}
}

func TestGlobalRulesFromClientMatchesKubeSphereResponse(t *testing.T) {
	binding := iamObject("GlobalRoleBinding", "olaresid")
	binding.Object["subjects"] = []any{map[string]any{"kind": "User", "name": "olaresid"}}
	binding.Object["roleRef"] = map[string]any{"kind": "GlobalRole", "name": "platform-admin"}

	platformAdmin := iamObject("GlobalRole", "platform-admin")
	platformAdmin.SetAnnotations(map[string]string{
		iamv1alpha2.AggregationRolesAnnotation: `["manage-clusters","view-clusters","manage-users","view-users","manage-settings"]`,
	})
	manageClusters := roleTemplate("manage-clusters", `{"clusters":"manage"}`, time.Unix(1, 0))
	viewClusters := roleTemplate("view-clusters", `{"clusters":"view"}`, time.Unix(2, 0))
	manageUsers := roleTemplate("manage-users", `{"users":"manage"}`, time.Unix(1, 0))
	viewUsers := roleTemplate("view-users", `{"users":"view"}`, time.Unix(2, 0))
	manageSettings := roleTemplate("manage-settings", `{"platform-settings":"manage"}`, time.Unix(1, 0))

	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			globalRoleBindingGVR: "GlobalRoleBindingList",
			globalRoleGVR:        "GlobalRoleList",
		},
		binding, platformAdmin, manageClusters, viewClusters, manageUsers, viewUsers, manageSettings,
	)
	got, err := globalRulesFromClient(context.Background(), client, "olaresid", "host")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		"clusters":          {"view", "manage"},
		"users":             {"view", "manage"},
		"platform-settings": {"manage"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("global rules = %#v, want %#v", got, want)
	}
}

func iamObject(kind, name string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(schema.GroupVersionKind{Group: "iam.kubesphere.io", Version: "v1alpha2", Kind: kind})
	object.SetName(name)
	return object
}

func roleTemplate(name, rules string, createdAt time.Time) *unstructured.Unstructured {
	role := iamObject("GlobalRole", name)
	role.SetAnnotations(map[string]string{roleTemplateRulesAnnotation: rules})
	role.SetCreationTimestamp(metav1.NewTime(createdAt))
	return role
}
