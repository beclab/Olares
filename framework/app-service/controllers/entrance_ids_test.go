package controllers

import (
	"context"
	app "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sys "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	fakeapi "github.com/beclab/api/pkg/generated/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestAssignIDsUsesOriginalDomainUser(t *testing.T) {
	r := &ApplicationReconciler{AppClientset: fakeapi.NewSimpleClientset(&sys.Terminus{ObjectMeta: metav1.ObjectMeta{Name: "terminus"}, Spec: sys.TerminusSpec{LocalName: "alice"}})}
	for _, owner := range []string{"alice", "bob"} {
		a := &app.Application{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{app.AppSharedLabel: app.AppSharedTrue}}, Spec: app.ApplicationSpec{Appid: "abcd1234", Owner: owner, Entrances: []app.Entrance{{Name: "web"}}, SharedEntrances: []app.Entrance{{Name: "shared"}}, UserSettings: map[string]map[string]string{"viewer": {"addedEntrances": `[{"name":"dev-9000","type":"dev","port":9000}]`}}}}
		saved, err := r.assignApplicationIDs(context.Background(), a, nil)
		if err != nil {
			t.Fatal(err)
		}
		want := app.NewUserEntranceID("abcd1234", owner, "alice", 0, 1)
		if saved.Spec.Entrances[0].ID != want {
			t.Fatal(saved.Spec.Entrances)
		}
		if saved.Spec.SharedEntrances[0].ID != app.SharedEntranceID("abcd1234", 0, 1) {
			t.Fatal(saved.Spec.SharedEntrances)
		}
		// No identity read is required just to retain already allocated IDs.
		stable, err := (&ApplicationReconciler{}).assignApplicationIDs(context.Background(), saved, saved)
		if err != nil || stable.Spec.Entrances[0].ID != want {
			t.Fatal(stable, err)
		}
	}
}
func TestExistingMissingIDsAreNotReconciled(t *testing.T) {
	r := &ApplicationReconciler{AppClientset: fakeapi.NewSimpleClientset(&sys.Terminus{ObjectMeta: metav1.ObjectMeta{Name: "terminus"}, Spec: sys.TerminusSpec{LocalName: "alice"}})}
	a := &app.Application{Spec: app.ApplicationSpec{Appid: "abcd1234", Owner: "alice", Entrances: []app.Entrance{{Name: "web"}}}}
	if _, err := r.assignApplicationIDs(context.Background(), a, a); err == nil {
		t.Fatal("reconcile filled a legacy ID")
	}
}
