package capi

import (
	"encoding/json"
	"testing"
	"time"

	"bytetrade.io/web3os/bfl/pkg/constants"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespaceObjectsPreserveKubeSphereOrder(t *testing.T) {
	created := func(value string) metav1.Time {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			t.Fatal(err)
		}
		return metav1.NewTime(parsed)
	}
	namespaces := []corev1.Namespace{
		{ObjectMeta: metav1.ObjectMeta{Name: "older", CreationTimestamp: created("2026-10-14T01:00:00Z")}},
		{ObjectMeta: metav1.ObjectMeta{Name: "alpha", CreationTimestamp: created("2026-10-14T02:00:00Z")}},
		{ObjectMeta: metav1.ObjectMeta{Name: "zeta", CreationTimestamp: created("2026-10-14T02:00:00Z"), Labels: map[string]string{"example": "value"}}},
	}

	items, err := namespaceObjects(namespaces)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"zeta", "alpha", "older"}
	for i := range want {
		if got := nestedString(items[i], "metadata", "name"); got != want[i] {
			t.Fatalf("namespace[%d] = %q, want %q", i, got, want[i])
		}
	}
	if got := nestedString(items[0], "metadata", "labels", "example"); got != "value" {
		t.Fatalf("converted namespace label = %q, want value", got)
	}
}

func TestNamespaceUsersUseTypedUserMetadata(t *testing.T) {
	users := []iamv1alpha2.User{
		{ObjectMeta: metav1.ObjectMeta{Name: "alice", Annotations: map[string]string{constants.UserAnnotationOwnerRole: constants.RoleOwner}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "bob", Annotations: map[string]string{constants.UserAnnotationOwnerRole: constants.RoleNormal}}},
	}

	got := namespaceUsers(users)
	if len(got) != 2 || got[0].Name != "alice" || got[0].OwnerRole != constants.RoleOwner || got[1].Name != "bob" {
		t.Fatalf("namespaceUsers() = %#v", got)
	}
}

func TestBuildNamespaceGroups(t *testing.T) {
	ns := func(name string, labels map[string]any) map[string]any {
		return map[string]any{"metadata": map[string]any{"name": name, "labels": labels}}
	}
	users := []namespaceUser{{Name: "bob", OwnerRole: "normal"}, {Name: "alice", OwnerRole: "owner"}}
	namespaces := []map[string]any{
		ns("app-alice", nil),
		ns("app-bob", nil),
		ns("photos-shared", nil),
		ns("kube-system", nil),
	}

	admin := buildNamespaceGroups(namespaces, append([]namespaceUser(nil), users...), "alice", true)
	if len(admin) != 4 || admin[0].Title != "alice" || admin[1].Title != "bob" || admin[2].Title != "Shared" || admin[3].Title != "System" {
		t.Fatalf("admin groups = %#v", admin)
	}
	normal := buildNamespaceGroups(namespaces, append([]namespaceUser(nil), users...), "bob", false)
	if len(normal) != 1 || normal[0].Title != "bob" || len(normal[0].Data) != 1 {
		t.Fatalf("normal groups = %#v", normal)
	}
}

func TestAppendUtilisationAlignsTimestamps(t *testing.T) {
	decode := func(input string) map[string]any {
		var result map[string]any
		if err := json.Unmarshal([]byte(input), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	metrics := []map[string]any{
		decode(`{"metric_name":"cluster_cpu_usage","data":{"result":[{"values":[[1,"2"],[2,"4"],[3,"5"]]}]}}`),
		decode(`{"metric_name":"cluster_cpu_total","data":{"result":[{"values":[[2,"8"],[1,"4"],[3,"0"]]}]}}`),
	}
	metrics = appendUtilisation(metrics, "cluster_cpu_usage", "cluster_cpu_total", "cluster_cpu_utilisation")
	if len(metrics) != 3 {
		t.Fatalf("metrics len = %d", len(metrics))
	}
	series, ok := firstSeries(metrics[2])
	if !ok {
		t.Fatal("missing utilisation series")
	}
	values := asArray(series["values"])
	want := []float64{0.5, 0.5, 0}
	for i, raw := range values {
		point := asArray(raw)
		if got := number(point[1]); got != want[i] {
			t.Fatalf("ratio[%d] = %v, want %v", i, got, want[i])
		}
	}
}

func TestAppendUtilisationUsesZeroForMissingOrZeroTotal(t *testing.T) {
	decode := func(input string) map[string]any {
		var result map[string]any
		if err := json.Unmarshal([]byte(input), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	metrics := []map[string]any{
		decode(`{"metric_name":"cluster_memory_usage_wo_cache","data":{"result":[{"values":[[1,"4"],[2,"8"],[3,"12"]]}]}}`),
		decode(`{"metric_name":"cluster_memory_total","data":{"result":[{"values":[[1,"0"],[3,"6"]]}]}}`),
	}
	metrics = appendUtilisation(metrics, "cluster_memory_usage_wo_cache", "cluster_memory_total", "cluster_memory_utilisation")
	series, ok := firstSeries(metrics[2])
	if !ok {
		t.Fatal("missing utilisation series")
	}
	values := asArray(series["values"])
	want := []float64{0, 0, 2}
	for i, raw := range values {
		point := asArray(raw)
		if got := number(point[1]); got != want[i] {
			t.Fatalf("ratio[%d] = %v, want %v", i, got, want[i])
		}
	}
}

func TestAppendUtilisationInstantValue(t *testing.T) {
	var response metricResponse
	if err := json.Unmarshal([]byte(`{"results":[{"metric_name":"cluster_cpu_usage","data":{"result":[{"value":[10,"3"]}]}},{"metric_name":"cluster_cpu_total","data":{"result":[{"value":[10,"6"]}]}}]}`), &response); err != nil {
		t.Fatal(err)
	}
	response.Results = appendUtilisation(response.Results, "cluster_cpu_usage", "cluster_cpu_total", "cluster_cpu_utilisation")
	series, ok := firstSeries(response.Results[2])
	if !ok {
		t.Fatal("missing utilisation series")
	}
	value := asArray(series["value"])
	if got := number(value[1]); got != 0.5 {
		t.Fatalf("ratio = %v, want 0.5", got)
	}
}

func TestCollapseSplitMetrics(t *testing.T) {
	var response metricResponse
	if err := json.Unmarshal([]byte(`{"results":[{"metric_name":"user_cpu_total","data":{"result":[{"metric":{"namespace":"n","instance":"a"},"values":[[1,"2"]]},{"metric":{"namespace":"n","instance":"b"},"values":[[2,"3"]]}]}}]}`), &response); err != nil {
		t.Fatal(err)
	}
	collapseSplitMetrics(response.Results)
	series, ok := firstSeries(response.Results[0])
	if !ok || len(asArray(series["values"])) != 2 {
		t.Fatalf("collapsed series = %#v", series)
	}
	metric := series["metric"].(map[string]any)
	if _, exists := metric["instance"]; exists {
		t.Fatalf("volatile label was retained: %#v", metric)
	}
}

func TestSharedNamespaceByLabelOrSuffix(t *testing.T) {
	tests := []struct {
		name      string
		namespace map[string]any
		want      bool
	}{
		{
			name:      "label",
			namespace: map[string]any{"metadata": map[string]any{"name": "photos-alice", "labels": map[string]any{"bytetrade.io/ns-shared": "true"}}},
			want:      true,
		},
		{
			name:      "suffix",
			namespace: map[string]any{"metadata": map[string]any{"name": "photos-shared"}},
			want:      true,
		},
		{
			name:      "owned",
			namespace: map[string]any{"metadata": map[string]any{"name": "photos-alice"}},
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sharedNamespace(tt.namespace); got != tt.want {
				t.Fatalf("sharedNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllowedMetricType(t *testing.T) {
	tests := []struct {
		name       string
		metricType string
		username   string
		admin      bool
		want       bool
	}{
		{name: "admin cluster", metricType: "cluster", username: "alice", admin: true, want: true},
		{name: "admin other user", metricType: "bob", username: "alice", admin: true, want: true},
		{name: "normal own", metricType: "alice", username: "alice", want: true},
		{name: "normal cluster", metricType: "cluster", username: "alice", want: false},
		{name: "normal other user", metricType: "bob", username: "alice", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allowedMetricType(tt.metricType, tt.username, tt.admin); got != tt.want {
				t.Fatalf("allowedMetricType() = %v, want %v", got, tt.want)
			}
		})
	}
}
