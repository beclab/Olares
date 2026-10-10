package upgrade

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	kubefake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestFixedMACPreflightPredictsBudgetAfterNetworkLoss(t *testing.T) {
	for _, tc := range []struct {
		name    string
		always  bool
		spare   bool
		wantErr bool
	}{
		{"default policy blocks affected-only budget", false, false, true},
		{"explicit unhealthy eviction policy", true, false, false},
		{"unaffected healthy capacity", false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pod := fixedMACFixturePod("media-0", "old")
			pod.Labels["budget"] = "media"
			min := intstr.FromInt32(1)
			pdb := &policyv1.PodDisruptionBudget{ObjectMeta: metav1.ObjectMeta{Name: "media", Namespace: "apps", Generation: 1}, Spec: policyv1.PodDisruptionBudgetSpec{MinAvailable: &min, Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"budget": "media"}}}, Status: policyv1.PodDisruptionBudgetStatus{ObservedGeneration: 1, DesiredHealthy: 1, CurrentHealthy: 2, DisruptionsAllowed: 1}}
			if tc.always {
				policy := policyv1.AlwaysAllow
				pdb.Spec.UnhealthyPodEvictionPolicy = &policy
			}
			kube := kubefake.NewSimpleClientset(pod, pdb)
			if tc.spare {
				spare := pod.DeepCopy()
				spare.Name = "other"
				spare.UID = "other"
				if _, err := kube.CoreV1().Pods("apps").Create(t.Context(), spare, metav1.CreateOptions{}); err != nil {
					t.Fatal(err)
				}
			}
			err := checkFixedMACDisruptionBudget(t.Context(), kube, &fixedMACProgress{Instances: []fixedMACInstance{{OldUID: "old"}}})
			if (err != nil) != tc.wantErr {
				t.Fatalf("error=%v wantError=%v", err, tc.wantErr)
			}
		})
	}
}

func TestFixedMACEvictionRetriesBudgetAndNeverEvictsReplacement(t *testing.T) {
	old := fixedMACFixturePod("media-0", "old")
	kube := kubefake.NewSimpleClientset(old)
	calls := 0
	kube.PrependReactor("create", "pods", func(a ktesting.Action) (bool, runtime.Object, error) {
		if a.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		calls++
		if calls == 1 {
			return true, nil, apierrors.NewTooManyRequests("budget unavailable", 0)
		}
		eviction := a.(ktesting.CreateAction).GetObject().(*policyv1.Eviction)
		if *eviction.DeleteOptions.Preconditions.UID != "old" {
			t.Fatal("missing UID protection")
		}
		return true, nil, nil
	})
	item := fixedMACInstance{Namespace: "apps", Name: old.Name, OldUID: old.UID}
	if err := evictFixedMACPod(t.Context(), kube, item, time.Millisecond); err != nil || calls != 2 {
		t.Fatalf("calls=%d error=%v", calls, err)
	}
	fresh := old.DeepCopy()
	fresh.UID = "replacement"
	if _, err := kube.CoreV1().Pods("apps").Update(t.Context(), fresh, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := evictFixedMACPod(t.Context(), kube, item, time.Millisecond); err != nil || calls != 2 {
		t.Fatalf("replacement eviction attempted: %v", err)
	}
}

func TestFixedMACEvictionHonorsCancellation(t *testing.T) {
	old := fixedMACFixturePod("media-0", "old")
	kube := kubefake.NewSimpleClientset(old)
	ctx, cancel := context.WithCancel(t.Context())
	kube.PrependReactor("create", "pods", func(a ktesting.Action) (bool, runtime.Object, error) {
		cancel()
		return true, nil, apierrors.NewTooManyRequests("budget unavailable", 0)
	})
	defer cancel()
	if err := evictFixedMACPod(ctx, kube, fixedMACInstance{Namespace: "apps", Name: old.Name, OldUID: old.UID}, time.Millisecond); err == nil {
		t.Fatal("cancelled eviction accepted")
	}
}

func TestFixedMACLegacyPreflightRejectsAmbiguityBeforeMutation(t *testing.T) {
	dc := fixedMACDynamic()
	app, err := dc.Resource(fixedMACApplications).Get(t.Context(), "apps-media", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err = unstructured.SetNestedStringMap(app.Object, map[string]string{"overlayMacvlanMacByOrdinal": `{"0":"02:00:00:00:00:01"}`}, "spec", "settings"); err != nil {
		t.Fatal(err)
	}
	if _, err = dc.Resource(fixedMACApplications).Update(t.Context(), app, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	template := corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{fixedMACLabel: "true", "applications.app.bytetrade.io/name": "media"}}}
	first := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "media", Namespace: "apps", UID: "sts"}, Spec: appsv1.StatefulSetSpec{Template: template}}
	second := first.DeepCopy()
	second.Name = "second"
	second.UID = "second"
	kube := kubefake.NewSimpleClientset(first, second)
	p := &fixedMACProgress{Instances: []fixedMACInstance{{Namespace: "apps", Name: "media-0", Kind: "StatefulSet", Workload: "media", WorkloadUID: "sts"}}}
	if err := checkFixedMACLegacy(t.Context(), kube, dc, p); err == nil {
		t.Fatal("ambiguous legacy identity accepted")
	}
	for _, a := range kube.Actions() {
		if a.GetVerb() != "get" && a.GetVerb() != "list" {
			t.Fatalf("preflight mutated %s", a.GetVerb())
		}
	}
}

func TestFixedMACLegacyPreflightValidatesClaimOwnership(t *testing.T) {
	for _, foreign := range []bool{false, true} {
		t.Run(fmt.Sprint(foreign), func(t *testing.T) {
			dc := fixedMACDynamic()
			app, err := dc.Resource(fixedMACApplications).Get(t.Context(), "apps-media", metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if err = unstructured.SetNestedStringMap(app.Object, map[string]string{"overlayMacvlanMacByOrdinal": `{"0":"02:00:00:00:00:01"}`}, "spec", "settings"); err != nil {
				t.Fatal(err)
			}
			if _, err = dc.Resource(fixedMACApplications).Update(t.Context(), app, metav1.UpdateOptions{}); err != nil {
				t.Fatal(err)
			}
			claim, err := dc.Resource(fixedMACAllocations).Get(t.Context(), "020000000001", metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			uid := "app"
			if foreign {
				uid = "other-app"
			}
			for key, value := range map[string]string{"mac": "02:00:00:00:00:01", "instanceKey": "apps/media/0", "applicationUID": uid} {
				if err = unstructured.SetNestedField(claim.Object, value, "spec", key); err != nil {
					t.Fatal(err)
				}
			}
			yes := true
			claim.SetOwnerReferences([]metav1.OwnerReference{{APIVersion: "app.bytetrade.io/v1alpha1", Kind: "Application", Name: app.GetName(), UID: app.GetUID(), BlockOwnerDeletion: &yes}})
			if _, err = dc.Resource(fixedMACAllocations).Update(t.Context(), claim, metav1.UpdateOptions{}); err != nil {
				t.Fatal(err)
			}
			template := corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{fixedMACLabel: "true", "applications.app.bytetrade.io/name": "media"}}}
			sts := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "media", Namespace: "apps", UID: "sts"}, Spec: appsv1.StatefulSetSpec{Template: template}}
			auxiliary := sts.DeepCopy()
			auxiliary.Name = "auxiliary"
			delete(auxiliary.Spec.Template.Labels, fixedMACLabel)
			kube := kubefake.NewSimpleClientset(sts, auxiliary, fixedMACFixturePod("media-0", "old"))
			p := &fixedMACProgress{Instances: []fixedMACInstance{{Namespace: "apps", Name: "media-0", Kind: "StatefulSet", Workload: "media", WorkloadUID: "sts", OldUID: "old"}}}
			err = checkFixedMACLegacy(t.Context(), kube, dc, p)
			if (err != nil) != foreign {
				t.Fatalf("foreign=%v error=%v", foreign, err)
			}
		})
	}
}

func TestFixedMACResumePrecedesKubernetesTasks(t *testing.T) {
	precheck := &PrecheckModule{}
	precheck.Init()
	upgrade := &Module{TargetVersion: semver.MustParse("1.12.8-20261010")}
	upgrade.Init()
	for _, tasks := range [][]task.Interface{precheck.Tasks, upgrade.Tasks} {
		first, ok := tasks[0].(*task.LocalTask)
		if !ok || first.Name != "ResumeOverlayNetwork" {
			t.Fatalf("network recovery does not precede API work: %v", tasks[0])
		}
	}
}
