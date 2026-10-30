package upgrade

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/storage"
	migration "github.com/beclab/Olares/cli/pkg/upgrade/entranceids"
	"github.com/beclab/api/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

// EntranceIDsBreakingChangeTasks is appended by the chosen release's breaking
// upgrader after the new CRD, before resuming writers. Version is intentionally
// not registered until the release version has been confirmed.
func EntranceIDsBreakingChangeTasks() []task.Interface {
	return []task.Interface{&task.LocalTask{Name: "PersistEntranceIDs", Action: new(persistEntranceIDsBreakingChange), Retry: 3}}
}

type persistEntranceIDsBreakingChange struct{ common.KubeAction }

func (a *persistEntranceIDsBreakingChange) Execute(_ connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	cs, err := versioned.NewForConfig(config)
	if err != nil {
		return err
	}
	kube, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	ctx := context.Background()
	crd, err := dyn.Resource(schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}).Get(ctx, "applications.app.bytetrade.io", metav1.GetOptions{})
	if err != nil {
		return err
	}
	versions, _, err := unstructured.NestedSlice(crd.Object, "spec", "versions")
	if err != nil {
		return err
	}
	supported := false
	for _, v := range versions {
		version, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if version["name"] != "v1alpha1" {
			continue
		}
		supported = true
		for _, field := range []string{"entrances", "sharedEntrances"} {
			typ, found, _ := unstructured.NestedString(version, "schema", "openAPIV3Schema", "properties", "spec", "properties", field, "items", "properties", "id", "type")
			if !found || typ != "string" {
				return fmt.Errorf("install Application CRD with %s[].id before migration", field)
			}
		}
	}
	if !supported {
		return fmt.Errorf("Application v1alpha1 CRD schema is unavailable")
	}
	return migration.Run(ctx, cs, kube, filepath.Join(storage.OlaresRootDir, "upgrade", migration.Migration+".json"))
}
