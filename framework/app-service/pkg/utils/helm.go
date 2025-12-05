package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/constants"
	refdocker "github.com/containerd/containerd/reference/docker"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	helmLoader "helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

func actionConfig() (*action.Configuration, error) {
	registryClient, err := registry.NewClient()
	if err != nil {
		return nil, err
	}
	configuration := action.Configuration{
		Releases:       storage.Init(driver.NewMemory()),
		KubeClient:     &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: ioutil.Discard}},
		Capabilities:   chartutil.DefaultCapabilities,
		RegistryClient: registryClient,
		Log:            func(s string, i ...interface{}) { klog.Infof(s, i...) },
	}
	return &configuration, nil
}

func InitAction() (*action.Install, error) {
	config, err := actionConfig()
	if err != nil {
		klog.Infof("actionConfig err=%v", err)
		return nil, err
	}
	instAction := action.NewInstall(config)
	instAction.Namespace = "spaced"
	instAction.ReleaseName = "test-release"
	instAction.DryRun = true
	return instAction, nil
}

func getChart(instAction *action.Install, filepath string) (*chart.Chart, error) {
	settings := cli.New()
	cp, err := instAction.ChartPathOptions.LocateChart(filepath, settings)
	if err != nil {
		klog.Errorf("locate chart [%s] error: %v", filepath, err)
		return nil, err
	}
	p := getter.All(settings)
	chartRequested, err := helmLoader.Load(cp)
	if err != nil {
		klog.Errorf("Load err=%v", err)
		return nil, err
	}
	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if instAction.DependencyUpdate {
				man := &downloader.Manager{
					ChartPath:  cp,
					Keyring:    instAction.ChartPathOptions.Keyring,
					SkipUpdate: false,
					Getters:    p,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = helmLoader.Load(cp); err != nil {
					return nil, err
				}
			} else {

			}
		}
	}
	return chartRequested, nil
}

func GetResourceListFromChart(chartPath string, values map[string]interface{}) (resources kube.ResourceList, err error) {
	instAction, err := InitAction()
	if err != nil {
		return nil, err
	}
	instAction.Namespace = "app-namespace"
	chartRequested, err := getChart(instAction, chartPath)
	if err != nil {
		klog.Infof("getchart err=%v", err)
		return nil, err
	}

	// fake values for helm dry run
	//values := make(map[string]interface{})
	//values["bfl"] = map[string]interface{}{
	//	"username": "bfl-username",
	//}
	values["user"] = map[string]interface{}{
		"zone": "user-zone",
	}
	values["schedule"] = map[string]interface{}{
		"nodeName": "node",
	}
	values["oidc"] = map[string]interface{}{
		"client": map[string]interface{}{},
		"issuer": "issuer",
	}
	values["userspace"] = map[string]interface{}{
		"appCache": "appcache",
		"userData": "userspace/Home",
	}
	values["os"] = map[string]interface{}{
		"appKey":    "appKey",
		"appSecret": "appSecret",
	}

	values["domain"] = map[string]string{}
	values["cluster"] = map[string]string{}
	values["dep"] = map[string]interface{}{}
	values["postgres"] = map[string]interface{}{
		"databases": map[string]interface{}{},
	}
	values["mariadb"] = map[string]interface{}{
		"databases": map[string]interface{}{},
	}
	values["mysql"] = map[string]interface{}{
		"databases": map[string]interface{}{},
	}
	values["redis"] = map[string]interface{}{}
	values["mongodb"] = map[string]interface{}{
		"databases": map[string]interface{}{},
	}
	values["minio"] = map[string]interface{}{
		"buckets": map[string]interface{}{},
	}
	values["rabbitmq"] = map[string]interface{}{
		"vhosts": map[string]interface{}{},
	}
	values["elasticsearch"] = map[string]interface{}{
		"indexes": map[string]interface{}{},
	}
	values["svcs"] = map[string]interface{}{}
	values["nats"] = map[string]interface{}{
		"subjects": map[string]interface{}{},
		"refs":     map[string]interface{}{},
	}
	values[constants.OlaresEnvHelmValuesKey] = map[string]interface{}{}

	ret, err := instAction.RunWithContext(context.Background(), chartRequested, values)
	if err != nil {
		return nil, err
	}

	var metadataAccessor = meta.NewAccessor()
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewBufferString(ret.Manifest), 4096)
	for {
		ext := runtime.RawExtension{}
		if err := d.Decode(&ext); err != nil {
			if err == io.EOF {
				return resources, nil
			}
			return nil, fmt.Errorf("error parsing")
		}
		ext.Raw = bytes.TrimSpace(ext.Raw)
		if len(ext.Raw) == 0 || bytes.Equal(ext.Raw, []byte("null")) {
			continue
		}
		obj, _, err := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, nil)
		if err != nil {
			return nil, err
		}
		name, _ := metadataAccessor.Name(obj)
		namespace, _ := metadataAccessor.Namespace(obj)
		info := &resource.Info{
			Namespace: namespace,
			Name:      name,
			Object:    obj,
		}
		resources = append(resources, info)
	}
}

func GetRefFromResourceList(chartPath string, values map[string]interface{}, images []string) (refs []v1alpha1.Ref, err error) {
	resources, err := GetResourceListFromChart(chartPath, values)
	if err != nil {
		klog.Infof("get resourcelist from chart err=%v", err)
		return refs, err
	}
	getPullPolicy := func(imageName string, pullPolicy corev1.PullPolicy) corev1.PullPolicy {
		if pullPolicy != "" {
			return pullPolicy
		}
		if strings.HasSuffix(imageName, ":latest") {
			return corev1.PullAlways
		}
		return corev1.PullIfNotPresent
	}
	seen := make(map[string]struct{})
	for _, r := range resources {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		if kind == "Deployment" {
			var deployment v1.Deployment
			err = scheme.Scheme.Convert(r.Object, &deployment, nil)
			if err != nil {
				return refs, err
			}
			for _, c := range deployment.Spec.Template.Spec.InitContainers {
				refs = append(refs, v1alpha1.Ref{
					Name:            c.Image,
					ImagePullPolicy: getPullPolicy(c.Image, c.ImagePullPolicy),
				})
			}
			for _, c := range deployment.Spec.Template.Spec.Containers {

				refs = append(refs, v1alpha1.Ref{
					Name:            c.Image,
					ImagePullPolicy: getPullPolicy(c.Image, c.ImagePullPolicy),
				})
			}
		}
		if kind == "StatefulSet" {
			var sts v1.StatefulSet
			err = scheme.Scheme.Convert(r.Object, &sts, nil)
			if err != nil {
				return refs, err
			}
			for _, c := range sts.Spec.Template.Spec.InitContainers {
				refs = append(refs, v1alpha1.Ref{
					Name:            c.Image,
					ImagePullPolicy: getPullPolicy(c.Image, c.ImagePullPolicy),
				})
			}
			for _, c := range sts.Spec.Template.Spec.Containers {
				refs = append(refs, v1alpha1.Ref{
					Name:            c.Image,
					ImagePullPolicy: getPullPolicy(c.Image, c.ImagePullPolicy),
				})
			}
		}
	}
	for _, image := range images {
		refs = append(refs, v1alpha1.Ref{
			Name:            image,
			ImagePullPolicy: getPullPolicy(image, ""),
		})
	}
	filteredRefs := make([]v1alpha1.Ref, 0)
	for _, imageRef := range refs {
		if _, ok := seen[imageRef.Name]; !ok {
			named, err := refdocker.ParseDockerRef(imageRef.Name)
			namedRef := imageRef.Name
			if err == nil {
				namedRef = named.String()
			}
			if _, exists := seen[namedRef]; !exists {
				filteredRefs = append(filteredRefs, v1alpha1.Ref{
					Name:            namedRef,
					ImagePullPolicy: imageRef.ImagePullPolicy,
				})
			}

			seen[namedRef] = struct{}{}
		}
	}
	return filteredRefs, nil
}

func RenderManifest(filepath, owner, admin string, isAdmin bool) (string, error) {
	templateData, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return RenderManifestFromContent(templateData, owner, admin, isAdmin)
}

func RenderManifestFromContent(content []byte, owner, admin string, isAdmin bool) (string, error) {
	c := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "chart",
			Version: "0.0.1",
		},
		Templates: []*chart.File{
			{
				Name: "OlaresManifest.yaml",
				Data: content,
			},
		},
	}
	//admin := owner
	//if isAdmin {
	//
	//}
	values := map[string]interface{}{
		"admin": admin,
		"bfl": map[string]string{
			"username": owner,
		},
		"isAdmin": isAdmin,
	}

	valuesToRender, err := chartutil.ToRenderValues(c, values, chartutil.ReleaseOptions{}, nil)
	if err != nil {
		return "", err
	}

	e := engine.Engine{}
	renderedTemplates, err := e.Render(c, valuesToRender)
	if err != nil {
		return "", err
	}

	renderedYAML := renderedTemplates["chart/OlaresManifest.yaml"]

	return renderedYAML, nil
}
