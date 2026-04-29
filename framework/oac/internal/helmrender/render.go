package helmrender

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	helmLoader "helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/resource"
)

// RenderNamespace is the fake namespace used during dry-run. oac's
// downstream resource checks refer to this constant when validating that all
// workload manifests land in the app namespace.
const RenderNamespace = "app-namespace"

// defaultReleaseName is the fallback used when Render is called without a
// caller-supplied release name. Only exercised by tests / tooling that don't
// have a manifest handy; production call sites always pass metadata.name so
// chart templates using `{{ .Release.Name }}` render to the app name.
const defaultReleaseName = "test-release"

// Render performs a helm dry-run of the chart rooted at oacPath and returns
// the decoded kube.ResourceList. values must be the fully-constructed values
// map (typically from BuildValues). releaseName drives `{{ .Release.Name }}`
// substitutions; pass the manifest's metadata.name to mirror how Olares
// installs apps in production (release name == app name). An empty string
// falls back to defaultReleaseName.
//
// io.EOF from the inner YAML decoder is treated as a clean termination by
// the caller; the original legacy code did the same.
func Render(oacPath string, values map[string]interface{}, releaseName string) (kube.ResourceList, error) {
	instAction, err := newInstallAction()
	if err != nil {
		return nil, err
	}
	instAction.Namespace = RenderNamespace
	if releaseName != "" {
		instAction.ReleaseName = releaseName
	}
	chartRequested, err := loadChart(instAction, oacPath)
	if err != nil {
		return nil, err
	}

	ret, err := instAction.RunWithContext(context.Background(), chartRequested, values)
	if err != nil {
		return nil, err
	}

	accessor := meta.NewAccessor()
	var resources kube.ResourceList
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewBufferString(ret.Manifest), 4096)
	for {
		ext := runtime.RawExtension{}
		if err := decoder.Decode(&ext); err != nil {
			if err == io.EOF {
				return resources, nil
			}
			return nil, fmt.Errorf("error parsing rendered manifest: %w", err)
		}
		ext.Raw = bytes.TrimSpace(ext.Raw)
		if len(ext.Raw) == 0 || bytes.Equal(ext.Raw, []byte("null")) {
			continue
		}
		obj, _, err := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, nil)
		if err != nil {
			return nil, err
		}
		name, _ := accessor.Name(obj)
		namespace, _ := accessor.Namespace(obj)
		resources = append(resources, &resource.Info{
			Namespace: namespace,
			Name:      name,
			Object:    obj,
		})
	}
}

func newActionConfig() (*action.Configuration, error) {
	registryClient, err := registry.NewClient()
	if err != nil {
		return nil, err
	}
	cfg := action.Configuration{
		Releases:       storage.Init(driver.NewMemory()),
		KubeClient:     &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: ioutil.Discard}},
		Capabilities:   chartutil.DefaultCapabilities,
		RegistryClient: registryClient,
	}
	return &cfg, nil
}

func newInstallAction() (*action.Install, error) {
	cfg, err := newActionConfig()
	if err != nil {
		return nil, err
	}
	instAction := action.NewInstall(cfg)
	instAction.Namespace = "spaced"
	instAction.ReleaseName = defaultReleaseName
	instAction.DryRun = true
	return instAction, nil
}

func loadChart(instAction *action.Install, path string) (*chart.Chart, error) {
	cp, err := instAction.ChartPathOptions.LocateChart(path, &cli.EnvSettings{})
	if err != nil {
		return nil, err
	}
	getters := getter.All(&cli.EnvSettings{})
	loaded, err := helmLoader.Load(cp)
	if err != nil {
		return nil, err
	}
	if req := loaded.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(loaded, req); err != nil && instAction.DependencyUpdate {
			mgr := &downloader.Manager{
				ChartPath:  cp,
				Keyring:    instAction.ChartPathOptions.Keyring,
				SkipUpdate: false,
				Getters:    getters,
			}
			if err := mgr.Update(); err != nil {
				return nil, err
			}
			if loaded, err = helmLoader.Load(cp); err != nil {
				return nil, err
			}
		}
	}
	return loaded, nil
}
