/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package addons

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/cmd/apply"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var (
	defaultCacheDir = filepath.Join(homedir.HomeDir(), ".kube", "cache")
)

func InstallYaml(manifests []string, namespace, kubeConfig, version string) error {

	configFlags := NewConfigFlags(kubeConfig, namespace)
	o, err := CreateApplyOptions(configFlags, manifests, version)
	if err != nil {
		return err
	}

	if err := o.Run(); err != nil {
		return err
	}

	return nil
}

func CreateApplyOptions(configFlags *genericclioptions.ConfigFlags, manifests []string, version string) (*apply.ApplyOptions, error) {
	matchVersionKubeConfigFlags := NewMatchVersionFlags(configFlags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	ioStreams := genericclioptions.IOStreams{In: nil, Out: os.Stdout, ErrOut: os.Stderr}

	flags := apply.NewApplyFlags(f, ioStreams)
	return ToOptions(flags, manifests, version)
}

func ToOptions(flags *apply.ApplyFlags, manifests []string, version string) (*apply.ApplyOptions, error) {
	serverSideApply := false
	cmp, err := versionutil.MustParseSemantic(version).Compare("v1.16.0")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to compare version: %v", err))
	}
	if cmp == 0 || cmp == 1 {
		serverSideApply = true
	} else {
		serverSideApply = false
	}

	dryRunStrategy := cmdutil.DryRunNone

	dynamicClient, err := flags.Factory.DynamicClient()
	if err != nil {
		return nil, err
	}

	dryRunVerifier := resource.NewQueryParamVerifier(dynamicClient, flags.Factory.OpenAPIGetter(), resource.QueryParamDryRun)
	fieldValidationVerifier := resource.NewQueryParamVerifier(dynamicClient, flags.Factory.OpenAPIGetter(), resource.QueryParamFieldValidation)
	fieldManager := "client-side-apply"

	// allow for a success message operation to be specified at print time
	toPrinter := func(operation string) (printers.ResourcePrinter, error) {
		flags.PrintFlags.NamePrintFlags.Operation = operation
		cmdutil.PrintFlagsWithDryRunStrategy(flags.PrintFlags, dryRunStrategy)
		return flags.PrintFlags.ToPrinter()
	}
	_ = flags.RecordFlags.CompleteWithChangeCause("")

	recorder, err := flags.RecordFlags.ToRecorder()
	if err != nil {
		return nil, err
	}

	filenames := manifests
	flags.DeleteFlags.FileNameFlags.Filenames = &filenames

	deleteOptions, err := flags.DeleteFlags.ToOptions(dynamicClient, flags.IOStreams)
	if err != nil {
		return nil, err
	}

	err = deleteOptions.FilenameOptions.RequireFilenameOrKustomize()
	if err != nil {
		return nil, err
	}

	openAPISchema, _ := flags.Factory.OpenAPISchema()
	validator, err := flags.Factory.Validator("Ignore", fieldValidationVerifier)
	if err != nil {
		return nil, err
	}
	builder := flags.Factory.NewBuilder()
	mapper, err := flags.Factory.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	namespace, enforceNamespace, err := flags.Factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return nil, err
	}

	o := &apply.ApplyOptions{
		PrintFlags: flags.PrintFlags,

		DeleteOptions:   deleteOptions,
		ToPrinter:       toPrinter,
		ServerSideApply: serverSideApply,
		ForceConflicts:  true,
		FieldManager:    fieldManager,
		Selector:        flags.Selector,
		DryRunStrategy:  dryRunStrategy,
		DryRunVerifier:  dryRunVerifier,
		Prune:           flags.Prune,
		PruneResources:  flags.PruneResources,
		All:             flags.All,
		Overwrite:       flags.Overwrite,
		OpenAPIPatch:    flags.OpenAPIPatch,
		PruneWhitelist:  flags.PruneWhitelist,

		Recorder:         recorder,
		Namespace:        namespace,
		EnforceNamespace: enforceNamespace,
		Validator:        validator,
		Builder:          builder,
		Mapper:           mapper,
		DynamicClient:    dynamicClient,
		OpenAPISchema:    openAPISchema,

		IOStreams: flags.IOStreams,

		VisitedUids:       sets.NewString(),
		VisitedNamespaces: sets.NewString(),
	}

	o.PostProcessorFn = o.PrintAndPrunePostProcessor()

	return o, nil
}

func NewConfigFlags(kubeconfig, namespace string) *genericclioptions.ConfigFlags {
	var impersonateGroup []string
	insecure := false

	return &genericclioptions.ConfigFlags{
		Insecure:   &insecure,
		Timeout:    stringptr("0"),
		KubeConfig: stringptr(kubeconfig),

		CacheDir:         stringptr(defaultCacheDir),
		ClusterName:      stringptr(""),
		AuthInfoName:     stringptr(""),
		Context:          stringptr(""),
		Namespace:        stringptr(namespace),
		APIServer:        stringptr(""),
		TLSServerName:    stringptr(""),
		CertFile:         stringptr(""),
		KeyFile:          stringptr(""),
		CAFile:           stringptr(""),
		BearerToken:      stringptr(""),
		Impersonate:      stringptr(""),
		ImpersonateGroup: &impersonateGroup,
	}
}

func stringptr(val string) *string {
	return &val
}

func NewMatchVersionFlags(delegate genericclioptions.RESTClientGetter) *cmdutil.MatchVersionFlags {
	return &cmdutil.MatchVersionFlags{
		Delegate: delegate,
	}
}

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func DoServerSideApply(ctx context.Context, cfg *rest.Config, objectYAML []byte) error {
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(objectYAML, nil, obj)
	if err != nil {
		return err
	}

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		dr = dyn.Resource(mapping.Resource)
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	force := true
	if _, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{Force: &force, FieldManager: "sample-controller"}); err != nil {
		return err
	}
	fmt.Println(strings.ToLower(obj.GetKind()) + "/" + obj.GetName() + "  " + "created")
	return nil
}

func DoPatchCluster(client dynamic.Interface, name string, data []byte) error {
	var gvr = schema.GroupVersionResource{
		Group:    "kubekey.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "clusters",
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	if _, err := dyn.Resource(gvr).Patch(context.TODO(), name, types.MergePatchType, data, metav1.PatchOptions{}); err != nil {
		fmt.Println("error")
		return err
	}
	dyn.Resource(gvr).Get(context.TODO(), "aaa", metav1.GetOptions{})
	return nil
}

func GetCluster(name string) (*unstructured.Unstructured, error) {
	var gvr = schema.GroupVersionResource{
		Group:    "cluster.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "clusters",
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	cluster, err := dyn.Resource(gvr).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return cluster, nil
}
