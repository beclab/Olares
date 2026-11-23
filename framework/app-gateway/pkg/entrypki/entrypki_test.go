package entrypki

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestEnabled(t *testing.T) {
	if Enabled("") || Enabled("false") {
		t.Fatal("default off")
	}
	if !Enabled("true") {
		t.Fatal("opt-in")
	}
}

func TestMaintainDisabledNoop(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	if err := MaintainEntryMTLS(context.Background(), c, DefaultNamespace, false, time.Now()); err != nil {
		t.Fatal(err)
	}
	var sec corev1.Secret
	err := c.Get(context.Background(), types.NamespacedName{Namespace: DefaultNamespace, Name: SecretName}, &sec)
	if err == nil {
		t.Fatal("disabled must not create secret")
	}
}

func TestMaintainBootstrapAndRotate(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	now := time.Now().UTC()
	if err := MaintainEntryMTLS(context.Background(), c, DefaultNamespace, true, now); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	var sec corev1.Secret
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: DefaultNamespace, Name: SecretName}, &sec); err != nil {
		t.Fatal(err)
	}
	if len(sec.Data[keyCACrt]) == 0 || len(sec.Data[keyServerCrt]) == 0 || len(sec.Data[keyClientCrt]) == 0 {
		t.Fatal("missing pem material")
	}
	fp1 := sec.Annotations[annotFingerprint]

	// Force near-expiry leaf so NeedsRotation trips.
	short, err := bootstrapMaterial(now.Add(-200 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	// Re-issue leaves with NotAfter soon by patching server cert via rotateLeaves with old now
	mat := &Material{CACrt: short.CACrt, CAKey: short.CAKey}
	// Build a leaf that expires in 10 days
	caCert, _ := parseCertificate(mat.CACrt)
	caKey, _ := parseECPrivateKey(mat.CAKey)
	leafCrt, leafKey, err := issueLeaf(caCert, caKey, "app-gateway-l4-mtls.os-gateway.svc.cluster.local", now.Add(-355*24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	sec.Data[keyServerCrt] = leafCrt
	sec.Data[keyServerKey] = leafKey
	sec.Data[keyClientCrt] = leafCrt
	sec.Data[keyClientKey] = leafKey
	if err := c.Update(context.Background(), &sec); err != nil {
		t.Fatal(err)
	}
	need, _, err := NeedsRotation(leafCrt, now)
	if err != nil || !need {
		t.Fatalf("expected rotation need=%v err=%v", need, err)
	}
	if err := MaintainEntryMTLS(context.Background(), c, DefaultNamespace, true, now); err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: DefaultNamespace, Name: SecretName}, &sec); err != nil {
		t.Fatal(err)
	}
	if sec.Annotations[annotFingerprint] == fp1 {
		// may collide rarely; ensure previous CA dual-trust key present after rotate
		if len(sec.Data[keyCACrtPrevious]) == 0 && sec.Annotations[annotFingerprint] == fp1 {
			t.Fatal("expected rotation to change material or set previous ca")
		}
	}
	if len(sec.Data[keyCACrtPrevious]) == 0 {
		t.Fatal("dual-trust previous ca expected after leaf rotation")
	}
}

func TestNeedsRotationThreshold(t *testing.T) {
	now := time.Now().UTC()
	mat, err := bootstrapMaterial(now)
	if err != nil {
		t.Fatal(err)
	}
	need, rem, err := NeedsRotation(mat.ServerCrt, now)
	if err != nil {
		t.Fatal(err)
	}
	if need {
		t.Fatalf("fresh leaf should not need rotation, rem=%s", rem)
	}
}
