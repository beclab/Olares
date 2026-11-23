package entrypki

import (
	"bytes"
	"context"
	"crypto/x509"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestMaintainSeparatesLegacyCredentialsAndRepairsMissingClient(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	now := time.Now().UTC()
	legacy, err := bootstrapMaterial(now)
	if err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: DefaultNamespace, Name: SecretName},
		Data: map[string][]byte{keyCACrt: legacy.CACrt, keyCAKey: legacy.CAKey,
			keyServerCrt: legacy.ServerCrt, keyServerKey: legacy.ServerKey,
			keyClientCrt: legacy.ClientCrt, keyClientKey: legacy.ClientKey},
	}).Build()
	if err := MaintainEntryMTLS(ctx, c, DefaultNamespace, true, now); err != nil {
		t.Fatal(err)
	}
	get := func(name string) *corev1.Secret {
		t.Helper()
		sec := &corev1.Secret{}
		if err := c.Get(ctx, types.NamespacedName{Namespace: DefaultNamespace, Name: name}, sec); err != nil {
			t.Fatal(err)
		}
		return sec
	}
	signer, server, client := get(signerSecretName), get(SecretName), get(ClientSecretName)
	if len(signer.Data) != 2 || len(signer.Data[keyCAKey]) == 0 ||
		len(server.Data) != 3 || len(server.Data[keyServerKey]) == 0 ||
		len(client.Data) != 3 || len(client.Data[keyClientKey]) == 0 {
		t.Fatal("credentials were not separated by consumer")
	}
	if !bytes.Equal(server.Data[keyCACrt], legacy.CACrt) || !bytes.Equal(client.Data[keyCACrt], legacy.CACrt) {
		t.Fatal("migration changed the trust anchor")
	}
	if err := c.Delete(ctx, client); err != nil {
		t.Fatal(err)
	}
	if err := MaintainEntryMTLS(ctx, c, DefaultNamespace, true, now); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(get(ClientSecretName).Data[keyClientCrt], client.Data[keyClientCrt]) {
		t.Fatal("missing client credential was not reissued")
	}
	if err := MaintainEntryMTLS(ctx, c, DefaultNamespace, true, now.Add(200*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(get(SecretName).Data[keyServerCrt], server.Data[keyServerCrt]) {
		t.Fatal("expiring server credential was not rotated")
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

func TestBootstrapLeafIdentityAndPurpose(t *testing.T) {
	now := time.Now().UTC()
	mat, err := bootstrapMaterial(now)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(mat.CACrt) {
		t.Fatal("could not load issuer")
	}
	for _, tc := range []struct {
		pem      []byte
		name     string
		usage    x509.ExtKeyUsage
		wrongUse x509.ExtKeyUsage
	}{
		{mat.ServerCrt, "app-gateway-l4-mtls.os-gateway.svc.cluster.local", x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		{mat.ClientCrt, "l4-bfl-proxy.entry.client", x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	} {
		cert, err := parseCertificate(tc.pem)
		if err != nil {
			t.Fatal(err)
		}
		opts := x509.VerifyOptions{Roots: roots, DNSName: tc.name, KeyUsages: []x509.ExtKeyUsage{tc.usage}, CurrentTime: now}
		if _, err := cert.Verify(opts); err != nil {
			t.Fatalf("%s certificate unexpectedly rejected: %v", tc.name, err)
		}
		opts.DNSName = "unrelated.example"
		if _, err := cert.Verify(opts); err == nil {
			t.Fatalf("%s certificate accepted for the wrong peer", tc.name)
		}
		opts.DNSName = tc.name
		opts.KeyUsages = []x509.ExtKeyUsage{tc.wrongUse}
		if _, err := cert.Verify(opts); err == nil {
			t.Fatalf("%s certificate accepted for the wrong TLS role", tc.name)
		}
	}
}
