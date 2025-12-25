package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"github.com/beclab/Olares/framework/app-service/pkg/utils/registry"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	refdocker "github.com/containerd/containerd/reference/docker"
	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/transports/alltransports"
	imagetypes "github.com/containers/image/v5/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AppImageInfoController represents a controller for managing the lifecycle of appimage.
type AppImageInfoController struct {
	client.Client
	imageClient *containerd.Client
}

// SetupWithManager sets up the ImageManagerController with the provided controller manager
func (r *AppImageInfoController) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("app-image-controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		klog.Errorf("set up app-image-controller failed %v", err)
		return err
	}
	if r.imageClient == nil {
		r.imageClient, err = containerd.New("/var/run/containerd/containerd.sock", containerd.WithDefaultNamespace("k8s.io"),
			containerd.WithTimeout(10*time.Second))
		if err != nil {
			klog.Errorf("create image client failed %v", err)
			return err
		}
	}

	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&appv1alpha1.AppImage{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, app *appv1alpha1.AppImage) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name: app.Name,
				}}}
			}),
		predicate.TypedFuncs[*appv1alpha1.AppImage]{
			CreateFunc: func(e event.TypedCreateEvent[*appv1alpha1.AppImage]) bool {
				return r.preEnqueueCheckForCreate(e.Object)
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*appv1alpha1.AppImage]) bool {
				return r.preEnqueueCheckForUpdate(e.ObjectOld, e.ObjectNew)
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*appv1alpha1.AppImage]) bool {
				return false
			},
		},
	))

	if err != nil {
		klog.Errorf("Failed to add watch err=%v", err)
		return err
	}

	return nil
}

// Reconcile implements the reconciliation loop for the ImageManagerController
func (r *AppImageInfoController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	ctrl.Log.Info("reconcile app image request", "name", req.Name)

	var am appv1alpha1.AppImage
	err := r.Get(ctx, req.NamespacedName, &am)

	if err != nil {
		ctrl.Log.Error(err, "get app image error", "name", req.Name)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		// unexpected error, retry after 5s
		return ctrl.Result{}, err
	}

	err = r.reconcile(ctx, &am)
	if err != nil {
		ctrl.Log.Error(err, "get app image info error", "name", req.Name)
	}

	return reconcile.Result{}, nil
}

func (r *AppImageInfoController) reconcile(ctx context.Context, instance *appv1alpha1.AppImage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	var err error

	var cur appv1alpha1.AppImage
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name}, &cur)
	if err != nil {
		klog.Errorf("Failed to get app manager name=%s err=%v", instance.Name, err)
		return err
	}
	if areAllNodesCompleted(cur.Spec, cur.Status.Conditions) {
		klog.Infof("all node completed app %s", instance.Name)
		err = r.updateStatus(ctx, &cur, []appv1alpha1.ImageInfo{}, "completed", "completed")
		if err != nil {
			klog.Errorf("update appimage status failed %v", err)
		}
		return nil
	}

	start := time.Now()
	klog.Infof("get image app %s request start", instance.Name)

	state, message := "completed", "completed"
	imageInfos, err := r.GetImageInfo(ctx, &cur)
	if err != nil {
		state = "failed"
		message = err.Error()
		klog.Errorf("get image info failed %v", err)
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err = r.Get(ctx, types.NamespacedName{Name: instance.Name}, &cur)
		if err != nil {
			return err
		}
		err = r.updateStatus(ctx, &cur, imageInfos, state, message)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		klog.Infof("update app image status failed %v", err)
		return err
	}
	klog.Infof("get app %s image: %v info success, time elapsed: %v", instance.Name, instance.Spec.Refs, time.Since(start))
	return nil
}

func (r *AppImageInfoController) preEnqueueCheckForCreate(obj client.Object) bool {
	am, _ := obj.(*appv1alpha1.AppImage)
	klog.Infof("enqueue check: %v", am.Status.State)
	if am.Status.State == "completed" || am.Status.State == "failed" {
		return false
	}
	return true
}

func (r *AppImageInfoController) preEnqueueCheckForUpdate(old, new client.Object) bool {
	return false
}

func (r *AppImageInfoController) updateStatus(ctx context.Context, am *appv1alpha1.AppImage, imageInfos []appv1alpha1.ImageInfo, state, message string) error {
	var err error
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err = r.Get(ctx, types.NamespacedName{Name: am.Name}, am)
		if err != nil {
			return err
		}

		now := metav1.Now()
		amCopy := am.DeepCopy()
		amCopy.Status.State = state
		amCopy.Status.StatueTime = &now
		amCopy.Status.Images = append(amCopy.Status.Images, imageInfos...)
		amCopy.Status.Message = message
		node := os.Getenv("NODE_NAME")
		amCopy.Status.Conditions = append(amCopy.Status.Conditions, appv1alpha1.Condition{
			Node:      node,
			Completed: true,
		})

		err = r.Status().Patch(ctx, amCopy, client.MergeFrom(am))
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func areAllNodesCompleted(spec appv1alpha1.ImageSpec, conditions []appv1alpha1.Condition) bool {
	conditionMap := make(map[string]bool)
	for _, condition := range conditions {
		conditionMap[condition.Node] = condition.Completed
	}
	for _, node := range spec.Nodes {
		completed := conditionMap[node]
		if !completed {
			return false
		}
	}
	return true
}

func parseImageSource(ctx context.Context, name string) (imagetypes.ImageSource, error) {
	ref, err := alltransports.ParseImageName(name)
	if err != nil {
		return nil, err
	}
	sys := newSystemContext()
	return ref.NewImageSource(ctx, sys)
}

func (r *AppImageInfoController) GetManifest(ctx context.Context, instance *appv1alpha1.AppImage, imageName string) (*imagetypes.ImageInspectInfo, error) {
	if instance.Annotations == nil || instance.Annotations[api.AppImagesKey] == "" {
		return r.getManifestViaNetwork(ctx, imageName)
	}
	imageInfoReqData := instance.Annotations[api.AppImagesKey]
	var imageInfoReq api.ImageInfoRequest
	err := json.Unmarshal([]byte(imageInfoReqData), &imageInfoReq)
	if err != nil {
		klog.Infof("failed to unmarshal image info %v", err)
		return r.getManifestViaNetwork(ctx, imageName)
	}
	var manifest *api.ImageInfoV2
	imageRef, err := refdocker.ParseDockerRef(imageName)
	if err != nil {
		klog.Errorf("invalid docker ref %s %v", imageName, err)
		return nil, err
	}
	for _, imageInfo := range imageInfoReq.Images {
		name, err := refdocker.ParseDockerRef(imageInfo.ImageName)
		if err != nil {
			return nil, err
		}
		for _, l := range imageInfo.InfoV2 {
			if name.String() == imageRef.String() {
				if l.Os == runtime.GOOS && l.Architecture == runtime.GOARCH {
					manifest = &l
					break
				}
			}
		}

	}
	if manifest == nil || len(manifest.LayersData) == 0 {
		return r.getManifestViaNetwork(ctx, imageName)
	}
	klog.Infof("get app %s image manifest from annotations", imageName)
	return r.buildImageInspectFromManifest(manifest), nil
}

func newSystemContext() *imagetypes.SystemContext {
	return &imagetypes.SystemContext{}
}

type imageInfoResult struct {
	info appv1alpha1.ImageInfo
	err  error
}

func (r *AppImageInfoController) GetImageInfo(ctx context.Context, instance *appv1alpha1.AppImage) ([]appv1alpha1.ImageInfo, error) {
	nodeName := os.Getenv("NODE_NAME")

	var wg sync.WaitGroup
	results := make(chan imageInfoResult, len(instance.Spec.Refs))
	tokens := make(chan struct{}, 5)
	klog.Infof("refs: %d", len(instance.Spec.Refs))
	for _, originRef := range instance.Spec.Refs {
		wg.Add(1)
		go func(originRef string) {
			defer wg.Done()
			tokens <- struct{}{}
			defer func() { <-tokens }()
			name, err := refdocker.ParseDockerRef(originRef)
			if err != nil {
				klog.Errorf("invalid image ref %s %v", originRef, err)
				results <- imageInfoResult{err: err}
				return
			}
			manifest, err := r.GetManifest(ctx, instance, originRef)
			if err != nil {
				klog.Infof("get image %s manifest failed %v", name.String(), err)
				results <- imageInfoResult{err: err}
				return
			}

			imageInfo := appv1alpha1.ImageInfo{
				Node:         nodeName,
				Name:         originRef,
				Architecture: manifest.Architecture,
				Variant:      manifest.Variant,
				Os:           manifest.Os,
			}
			imageLayers := make([]appv1alpha1.ImageLayer, 0)
			for _, layer := range manifest.LayersData {
				imageLayer := appv1alpha1.ImageLayer{
					MediaType:   layer.MIMEType,
					Digest:      layer.Digest.String(),
					Size:        layer.Size,
					Annotations: layer.Annotations,
				}
				_, err = r.imageClient.ContentStore().Info(ctx, layer.Digest)
				if err == nil {
					imageLayer.Offset = layer.Size
					imageLayers = append(imageLayers, imageLayer)
					// go next layer
					continue
				}
				if errors.Is(err, errdefs.ErrNotFound) {
					statuses, err := r.imageClient.ContentStore().ListStatuses(ctx)
					if err != nil {
						klog.Errorf("list statuses failed %v", err)
						results <- imageInfoResult{err: err}
						return
					}
					for _, status := range statuses {
						s := "layer-" + layer.Digest.String()
						if s == status.Ref {
							imageLayer.Offset = status.Offset
							break
						}
					}
				} else {
					klog.Infof("get content info failed %v", err)
					results <- imageInfoResult{err: err}
					return
				}
				imageLayers = append(imageLayers, imageLayer)
			}
			imageInfo.LayersData = imageLayers
			results <- imageInfoResult{info: imageInfo}
		}(originRef)
	}

	wg.Wait()
	close(results)
	imageInfos := make([]appv1alpha1.ImageInfo, 0, len(instance.Spec.Refs))
	var firstErr error
	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}
		} else {
			imageInfos = append(imageInfos, result.info)
		}
	}
	return imageInfos, firstErr
}

func (r *AppImageInfoController) buildImageInspectFromManifest(manifest *api.ImageInfoV2) *imagetypes.ImageInspectInfo {
	layersData := make([]imagetypes.ImageInspectLayer, 0, len(manifest.Layers))
	for _, layer := range manifest.LayersData {
		layersData = append(layersData, imagetypes.ImageInspectLayer{
			MIMEType: layer.MIMEType,
			Digest:   layer.Digest,
			Size:     layer.Size,
		})
	}
	klog.Infof("buildImageInspectFromManifest: os: %s, arch: %s", runtime.GOOS, runtime.GOARCH)
	return &imagetypes.ImageInspectInfo{
		Os:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		LayersData:   layersData,
	}
}

func (r *AppImageInfoController) getManifestViaNetwork(ctx context.Context, originRef string) (*imagetypes.ImageInspectInfo, error) {

	name, err := refdocker.ParseDockerRef(originRef)
	if err != nil {
		return nil, err
	}
	replacedRef, _ := utils.ReplacedImageRef(registry.GetMirrors(), name.String(), false)

	var src imagetypes.ImageSource
	srcImageName := "docker://" + replacedRef
	sysCtx := newSystemContext()
	fmt.Printf("imageName: %s\n", replacedRef)
	src, err = parseImageSource(ctx, srcImageName)
	if err != nil {
		klog.Infof("parse Image Source: %v", err)
		return nil, err
	}
	unparsedInstance := image.UnparsedInstance(src, nil)
	_, _, err = unparsedInstance.Manifest(ctx)
	if err != nil {
		klog.Infof("parse manifest: %v", err)

		return nil, err
	}
	img, err := image.FromUnparsedImage(ctx, sysCtx, unparsedInstance)
	if err != nil {
		klog.Infof("from unparsed image: %v", err)

		return nil, err
	}
	imgInspect, err := img.Inspect(ctx)
	if err != nil {
		klog.Infof("inspect image failed: %v", err)

		return nil, err
	}

	return imgInspect, err
}
