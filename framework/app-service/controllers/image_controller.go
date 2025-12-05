package controllers

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	appv1alpha1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/images"

	"github.com/hashicorp/go-multierror"
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

var thisNode string

func init() {
	thisNode = os.Getenv("NODE_NAME")
}

// ImageManagerController represents a controller for managing the lifecycle of applicationmanager.
type ImageManagerController struct {
	client.Client
}

// SetupWithManager sets up the ImageManagerController with the provided controller manager
func (r *ImageManagerController) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("image-controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&appv1alpha1.ImageManager{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, app *appv1alpha1.ImageManager) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name:      app.Name,
					Namespace: app.Spec.AppOwner,
				}}}
			}),
		predicate.TypedFuncs[*appv1alpha1.ImageManager]{
			CreateFunc: func(e event.TypedCreateEvent[*appv1alpha1.ImageManager]) bool {
				return r.preEnqueueCheckForCreate(e.Object)
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*appv1alpha1.ImageManager]) bool {
				return r.preEnqueueCheckForUpdate(e.ObjectOld, e.ObjectNew)
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*appv1alpha1.ImageManager]) bool {
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

var imageManager map[string]context.CancelFunc

// Reconcile implements the reconciliation loop for the ImageManagerController
func (r *ImageManagerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	ctrl.Log.Info("reconcile image manager request", "name", req.Name)

	var im appv1alpha1.ImageManager
	err := r.Get(ctx, req.NamespacedName, &im)

	if err != nil {
		ctrl.Log.Error(err, "get application manager error", "name", req.Name)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		// unexpected error, retry after 5s
		return ctrl.Result{}, err
	}

	err = r.reconcile(ctx, &im)
	if err != nil {
		ctrl.Log.Error(err, "download image error", "name", req.Name)
	}

	return reconcile.Result{}, err
}

func (r *ImageManagerController) reconcile(ctx context.Context, instance *appv1alpha1.ImageManager) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer func() {
		delete(imageManager, instance.Name)
	}()
	var err error

	var cur appv1alpha1.ImageManager
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name}, &cur)
	if err != nil {
		klog.Errorf("Failed to get imagemanagers name=%s err=%v", instance.Name, err)
		return err
	}

	if imageManager == nil {
		imageManager = make(map[string]context.CancelFunc)
	}
	if _, ok := imageManager[instance.Name]; ok {
		return nil
	}
	imageManager[instance.Name] = cancel
	if cur.Status.State != appv1alpha1.Downloading.String() {
		err = r.updateStatus(ctx, &cur, appv1alpha1.Downloading.String(), "start downloading")
		if err != nil {
			klog.Infof("Failed to update imagemanager status name=%v, err=%v", cur.Name, err)
			return err
		}
	}

	err = r.download(ctx, cur.Spec.Refs,
		images.PullOptions{
			AppName:      instance.Spec.AppName,
			OwnerName:    instance.Spec.AppOwner,
			AppNamespace: instance.Spec.AppNamespace,
		})
	if err != nil {
		klog.Infof("download failed err=%v", err)

		state := "failed"
		if errors.Is(err, context.Canceled) {
			state = appv1alpha1.DownloadingCanceled.String()
		}
		err = r.updateStatus(context.TODO(), instance, state, err.Error())
		if err != nil {
			klog.Infof("Failed to update status err=%v", err)
		}
		return err
	}
	time.Sleep(2 * time.Second)
	err = r.updateStatus(context.TODO(), instance, "completed", "image download completed")
	if err != nil {
		klog.Infof("Failed to update status err=%v", err)
		return err
	}

	klog.Infof("download app: %s image success", instance.Spec.AppName)
	return nil
}

func (r *ImageManagerController) preEnqueueCheckForCreate(obj client.Object) bool {
	im, _ := obj.(*appv1alpha1.ImageManager)
	if im.Status.State == "failed" || im.Status.State == appv1alpha1.DownloadingCanceled.String() ||
		im.Status.State == "completed" {
		return false
	}
	klog.Infof("enqueue check: %v", im.Status.State)
	return true
}

func (r *ImageManagerController) preEnqueueCheckForUpdate(old, new client.Object) bool {
	im, _ := new.(*appv1alpha1.ImageManager)
	if im.Status.State == appv1alpha1.DownloadingCanceled.String() {
		go r.cancel(im)
	}
	return false
}

func (r *ImageManagerController) download(ctx context.Context, refs []appv1alpha1.Ref, opts images.PullOptions) (err error) {
	if len(refs) == 0 {
		return errors.New("no image to download")
	}
	var wg sync.WaitGroup
	var errs error
	tokens := make(chan struct{}, 3)
	for _, ref := range refs {
		wg.Add(1)
		go func(ref appv1alpha1.Ref) {
			tokens <- struct{}{}
			defer wg.Done()
			defer func() {
				<-tokens
			}()
			iClient, ctx, cancel, err := images.NewClient(ctx)
			if err != nil {
				errs = multierror.Append(errs, err)
				return
			}
			defer cancel()
			_, err = iClient.PullImage(ctx, ref, opts)
			if err != nil {
				errs = multierror.Append(errs, err)
				klog.Infof("pull image failed name=%v err=%v", ref, err)
			}
		}(ref)
	}
	klog.Infof("waiting image %v to download", refs)
	wg.Wait()
	return errs
}

func (r *ImageManagerController) updateStatus(ctx context.Context, im *appv1alpha1.ImageManager, state string, message string) error {
	var err error
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err = r.Get(ctx, types.NamespacedName{Name: im.Name}, im)
		if err != nil {
			return err
		}

		now := metav1.Now()
		imCopy := im.DeepCopy()
		if state != "completed" {
			imCopy.Status.State = state
		}
		imCopy.Status.Message = message
		imCopy.Status.StatusTime = &now
		imCopy.Status.UpdateTime = &now

		if state == "completed" {
			for _, node := range imCopy.Spec.Nodes {
				if node != thisNode {
					continue
				}
				if imCopy.Status.Conditions == nil {
					imCopy.Status.Conditions = make(map[string]map[string]map[string]string)
				}
				if imCopy.Status.Conditions[thisNode] == nil {
					imCopy.Status.Conditions[thisNode] = make(map[string]map[string]string)
				}
				for _, ref := range imCopy.Spec.Refs {
					if _, ok := imCopy.Status.Conditions[thisNode][ref.Name]; !ok {
						imCopy.Status.Conditions[thisNode][ref.Name] = map[string]string{
							"offset": "56782302",
							"total":  "56782302",
						}
					}
				}
			}

			checkAllCompleted := func() bool {
				for _, node := range imCopy.Spec.Nodes {
					conditionsNode, ok := imCopy.Status.Conditions[node]
					if !ok {
						return false
					}
					for _, ref := range imCopy.Spec.Refs {
						if _, ok := conditionsNode[ref.Name]["offset"]; !ok {
							return false
						}
						if _, ok := conditionsNode[ref.Name]["total"]; !ok {
							return false
						}
						if conditionsNode[ref.Name]["offset"] != conditionsNode[ref.Name]["total"] {
							return false
						}
					}
				}
				return true
			}
			if checkAllCompleted() {
				klog.Errorf("checkallcompleted............")
				imCopy.Status.State = state
			}

		}

		err = r.Status().Patch(ctx, imCopy, client.MergeFrom(im))
		if err != nil {
			klog.Errorf("failed to patch im %s status with state=%s %v", imCopy.Name, imCopy.Status.State, err)
			return err
		}
		return nil
	})

	return err
}

func (r *ImageManagerController) cancel(im *appv1alpha1.ImageManager) error {
	cancel, ok := imageManager[im.Name]
	if !ok {
		return errors.New("can not execute cancel")
	}
	cancel()
	return nil
}
