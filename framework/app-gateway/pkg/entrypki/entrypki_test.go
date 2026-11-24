package entrypki

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	if err := MaintainEntryMTLS(ctx, c, DefaultNamespace, true, now.Add(21*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(get(SecretName).Data[keyServerCrt], server.Data[keyServerCrt]) {
		t.Fatal("expiring server credential was not rotated")
	}
}

func TestNeedsRotationThreshold(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	mat, err := bootstrapMaterial(now)
	if err != nil {
		t.Fatal(err)
	}
	expiry := now.Add(30*24*time.Hour - time.Hour)
	for _, tt := range []struct {
		at     time.Time
		rotate bool
	}{
		{now, false},
		{expiry.Add(-10*24*time.Hour - time.Second), false},
		{expiry.Add(-10 * 24 * time.Hour), true},
		{expiry, true},
	} {
		need, remaining, err := NeedsRotation(mat.ServerCrt, tt.at)
		if err != nil || need != tt.rotate || remaining != expiry.Sub(tt.at) {
			t.Fatalf("at %s: rotate=%t remaining=%s err=%v", tt.at, need, remaining, err)
		}
	}
}

func TestMaintainMigratesLongLivedLeavesOnce(t *testing.T) {
	ctx := t.Context()
	now := time.Now().UTC().Truncate(time.Second)
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	maintain := func() {
		t.Helper()
		if err := MaintainEntryMTLS(ctx, c, DefaultNamespace, true, now); err != nil {
			t.Fatal(err)
		}
	}
	get := func(name string) *corev1.Secret {
		t.Helper()
		var sec corev1.Secret
		if err := c.Get(ctx, types.NamespacedName{Namespace: DefaultNamespace, Name: name}, &sec); err != nil {
			t.Fatal(err)
		}
		return &sec
	}
	maintain()
	signer := get(signerSecretName)
	ca, err := parseCertificate(signer.Data[keyCACrt])
	if err != nil {
		t.Fatal(err)
	}
	caKey, err := parseECPrivateKey(signer.Data[keyCAKey])
	if err != nil {
		t.Fatal(err)
	}
	leaves := []struct{ name, cert, key string }{
		{SecretName, keyServerCrt, keyServerKey},
		{ClientSecretName, keyClientCrt, keyClientKey},
	}
	previous := make(map[string]*corev1.Secret, len(leaves))
	for _, leaf := range leaves {
		sec := get(leaf.name)
		template, err := parseCertificate(sec.Data[leaf.cert])
		if err != nil {
			t.Fatal(err)
		}
		template.NotAfter = now.Add(365 * 24 * time.Hour)
		der, err := x509.CreateCertificate(rand.Reader, template, ca, template.PublicKey, caKey)
		if err != nil {
			t.Fatal(err)
		}
		sec.Data[leaf.cert] = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
		if err := c.Update(ctx, sec); err != nil {
			t.Fatal(err)
		}
		previous[leaf.name] = sec
	}

	maintain()
	renewed := make(map[string]*corev1.Secret, len(leaves))
	for _, leaf := range leaves {
		sec := get(leaf.name)
		cert, err := parseCertificate(sec.Data[leaf.cert])
		if err != nil {
			t.Fatal(err)
		}
		old := previous[leaf.name]
		if cert.NotAfter.Sub(cert.NotBefore) != 30*24*time.Hour ||
			bytes.Equal(sec.Data[leaf.key], old.Data[leaf.key]) || !bytes.Equal(sec.Data[keyCACrt], signer.Data[keyCACrt]) {
			t.Fatalf("%s did not move to a new key and 30-day certificate under the same CA", leaf.name)
		}
		renewed[leaf.name] = sec
	}

	maintain()
	for _, leaf := range leaves {
		if get(leaf.name).ResourceVersion != renewed[leaf.name].ResourceVersion {
			t.Fatalf("%s was rewritten after successful migration", leaf.name)
		}
	}
	if get(signerSecretName).ResourceVersion != signer.ResourceVersion {
		t.Fatal("renewal rewrote the signer")
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

func TestMaintainPreservesTrustWhenSignerIsMissing(t *testing.T) {
	ctx := t.Context()
	now := time.Now().UTC()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	if err := MaintainEntryMTLS(ctx, c, "", true, now); err != nil {
		t.Fatal(err)
	}
	get := func(name string) *corev1.Secret {
		t.Helper()
		var sec corev1.Secret
		if err := c.Get(ctx, types.NamespacedName{Namespace: DefaultNamespace, Name: name}, &sec); err != nil {
			t.Fatal(err)
		}
		return &sec
	}
	clientSecret := get(ClientSecretName)
	signer := get(signerSecretName)
	if err := c.Delete(ctx, signer); err != nil {
		t.Fatal(err)
	}
	// An isolated server secret also lacks the signing key.
	if err := MaintainEntryMTLS(ctx, c, "", true, now); err == nil {
		t.Fatal("missing signer with isolated server credential must fail")
	}
	if err := c.Delete(ctx, get(SecretName)); err != nil {
		t.Fatal(err)
	}
	if err := MaintainEntryMTLS(ctx, c, "", true, now); err == nil {
		t.Fatal("missing signer with only a client credential must fail")
	}
	var missing corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: DefaultNamespace, Name: signerSecretName}, &missing); !apierrors.IsNotFound(err) {
		t.Fatalf("unexpected signer lookup result: %v", err)
	}
	if get(ClientSecretName).ResourceVersion != clientSecret.ResourceVersion {
		t.Fatal("failed recovery modified the surviving credential")
	}
	if err := c.Delete(ctx, clientSecret); err != nil {
		t.Fatal(err)
	}
	if err := MaintainEntryMTLS(ctx, c, "", true, now); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(get(signerSecretName).Data[keyCACrt], signer.Data[keyCACrt]) {
		t.Fatal("fresh bootstrap reused the prior CA")
	}
}
