package terminus

import (
	"context"
	"testing"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/rest"
)

func TestApplyNetworkCRDs_SkipsWhenAllCRDsPresent(t *testing.T) {
	t.Setenv("APP_GATEWAY_STACK_ENABLED", "1")
	t.Setenv("OLARES_INSTALLER_DIR", t.TempDir())

	oldValidate := validateAppGatewaySystemInstallerArtifactsFunc
	oldGetConfig := getConfigForNetworkCRDs
	oldInit := initConfigForNetworkCRDs
	oldLoadValues := loadValuesForNetworkCRDs
	oldApply := applyCRDChartFunc
	oldLinkerdPresent := linkerdCRDsPresentFunc
	oldEnvoyPresent := envoyCRDsPresentFunc
	defer func() {
		validateAppGatewaySystemInstallerArtifactsFunc = oldValidate
		getConfigForNetworkCRDs = oldGetConfig
		initConfigForNetworkCRDs = oldInit
		loadValuesForNetworkCRDs = oldLoadValues
		applyCRDChartFunc = oldApply
		linkerdCRDsPresentFunc = oldLinkerdPresent
		envoyCRDsPresentFunc = oldEnvoyPresent
	}()

	validateAppGatewaySystemInstallerArtifactsFunc = func(string) error { return nil }
	getConfigForNetworkCRDs = func() (*rest.Config, error) { return &rest.Config{}, nil }
	linkerdCRDsPresentFunc = func(*rest.Config) bool { return true }
	envoyCRDsPresentFunc = func(*rest.Config) bool { return true }
	initConfigForNetworkCRDs = func(*rest.Config, string) (*action.Configuration, *cli.EnvSettings, error) {
		t.Fatal("initConfig should not be called when all CRDs are present")
		return nil, nil, nil
	}
	loadValuesForNetworkCRDs = func(string) (map[string]interface{}, error) {
		t.Fatal("loadValues should not be called when all CRDs are present")
		return nil, nil
	}
	applyCRDChartFunc = func(context.Context, *action.Configuration, *cli.EnvSettings, string, string, string, map[string]interface{}) error {
		t.Fatal("apply should not be called when all CRDs are present")
		return nil
	}

	if err := (&ApplyNetworkCRDs{}).Execute(nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestApplyNetworkCRDs_AppliesMissingCRDs(t *testing.T) {
	t.Setenv("APP_GATEWAY_STACK_ENABLED", "1")
	t.Setenv("OLARES_INSTALLER_DIR", t.TempDir())

	oldValidate := validateAppGatewaySystemInstallerArtifactsFunc
	oldGetConfig := getConfigForNetworkCRDs
	oldInit := initConfigForNetworkCRDs
	oldLoadValues := loadValuesForNetworkCRDs
	oldApply := applyCRDChartFunc
	oldLinkerdPresent := linkerdCRDsPresentFunc
	oldEnvoyPresent := envoyCRDsPresentFunc
	defer func() {
		validateAppGatewaySystemInstallerArtifactsFunc = oldValidate
		getConfigForNetworkCRDs = oldGetConfig
		initConfigForNetworkCRDs = oldInit
		loadValuesForNetworkCRDs = oldLoadValues
		applyCRDChartFunc = oldApply
		linkerdCRDsPresentFunc = oldLinkerdPresent
		envoyCRDsPresentFunc = oldEnvoyPresent
	}()

	validateAppGatewaySystemInstallerArtifactsFunc = func(string) error { return nil }
	getConfigForNetworkCRDs = func() (*rest.Config, error) { return &rest.Config{}, nil }
	initConfigForNetworkCRDs = func(*rest.Config, string) (*action.Configuration, *cli.EnvSettings, error) {
		return &action.Configuration{}, &cli.EnvSettings{}, nil
	}
	loadValuesForNetworkCRDs = func(string) (map[string]interface{}, error) {
		return map[string]interface{}{}, nil
	}
	linkerdCRDsPresentFunc = func(*rest.Config) bool { return true }
	envoyCRDsPresentFunc = func(*rest.Config) bool { return false }

	var calls int
	var gotRelease string
	applyCRDChartFunc = func(_ context.Context, _ *action.Configuration, _ *cli.EnvSettings, releaseName, _ string, _ string, _ map[string]interface{}) error {
		calls++
		gotRelease = releaseName
		return nil
	}

	if err := (&ApplyNetworkCRDs{}).Execute(nil); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected one apply call, got %d", calls)
	}
	if gotRelease != "app-gateway-envoy-gateway-crds" {
		t.Fatalf("unexpected release name %q", gotRelease)
	}
}
