package appstate

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMakeRecordNil(t *testing.T) {
	if rec := makeRecord(nil, appsv1.InstallFailed, "boom"); rec != nil {
		t.Fatalf("makeRecord(nil) = %+v, want nil", rec)
	}
}

func TestMakeRecordMapsFields(t *testing.T) {
	am := &appsv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "nginx-mgr",
			Annotations: map[string]string{api.AppVersionKey: "1.2.3"},
		},
		Spec: appsv1.ApplicationManagerSpec{Source: "market"},
		Status: appsv1.ApplicationManagerStatus{
			OpType: appsv1.InstallOp,
			OpID:   "op-123",
		},
	}

	rec := makeRecord(am, appsv1.Running, "ok")
	if rec == nil {
		t.Fatal("makeRecord returned nil")
	}
	if rec.OpType != appsv1.InstallOp {
		t.Errorf("OpType=%q want install", rec.OpType)
	}
	if rec.OpID != "op-123" {
		t.Errorf("OpID=%q want op-123", rec.OpID)
	}
	if rec.Source != "market" {
		t.Errorf("Source=%q want market", rec.Source)
	}
	if rec.Version != "1.2.3" {
		t.Errorf("Version=%q want 1.2.3", rec.Version)
	}
	if rec.Status != appsv1.Running {
		t.Errorf("Status=%q want Running", rec.Status)
	}
	if rec.Message != "ok" {
		t.Errorf("Message=%q want ok", rec.Message)
	}
	if rec.StateTime == nil {
		t.Error("StateTime should be set")
	}
}
