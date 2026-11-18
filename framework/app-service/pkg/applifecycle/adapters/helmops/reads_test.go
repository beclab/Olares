package helmops_test

import (
	"context"
	"errors"
	"testing"

	"k8s.io/client-go/rest"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/helmops"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

// -----------------------------------------------------------------
// GetDeployedVersion — pure delegation through the injected
// GetVersionFn. Verifies argument threading (namespace + release
// + adapter kubeConfig) and return propagation.
// -----------------------------------------------------------------

func TestGetDeployedVersion_Delegates(t *testing.T) {
	var gotKC *rest.Config
	var gotNS, gotRelease string
	kc := &rest.Config{Host: "https://apiserver"}
	o := helmops.New(testutil.NewFakeClient(), kc,
		helmops.WithGetVersion(func(_ context.Context, kubeConfig *rest.Config, ns, release string) (string, int, error) {
			gotKC = kubeConfig
			gotNS = ns
			gotRelease = release
			return "1.2.3", 7, nil
		}),
	)

	version, revision, err := o.GetDeployedVersion(context.Background(), "user-x-app", "app-x")
	if err != nil {
		t.Fatalf("GetDeployedVersion: %v", err)
	}
	if version != "1.2.3" {
		t.Errorf("version: %q want 1.2.3", version)
	}
	if revision != 7 {
		t.Errorf("revision: %d want 7", revision)
	}
	if gotKC != kc {
		t.Errorf("kubeConfig: got %p want %p (adapter-injected)", gotKC, kc)
	}
	if gotNS != "user-x-app" {
		t.Errorf("namespace: %q want user-x-app", gotNS)
	}
	if gotRelease != "app-x" {
		t.Errorf("release: %q want app-x", gotRelease)
	}
}

func TestGetDeployedVersion_ErrorPropagates(t *testing.T) {
	wantErr := errors.New("no such release")
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithGetVersion(func(_ context.Context, _ *rest.Config, _, _ string) (string, int, error) {
			return "", 0, wantErr
		}),
	)
	_, _, err := o.GetDeployedVersion(context.Background(), "ns", "rel")
	if !errors.Is(err, wantErr) {
		t.Errorf("err: %v want %v", err, wantErr)
	}
}
