package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) imageInfo(req *restful.Request, resp *restful.Response) {
	imageReq := &api.ImageInfoRequest{}
	err := req.ReadEntity(imageReq)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	if imageReq.AppName == "" || len(imageReq.Images) == 0 {
		api.HandleBadRequest(resp, req, errors.New("empty name or images"))
		return
	}
	start := time.Now()
	klog.Infof("received app %s fetch image info request", imageReq.AppName)

	err = createAppImage(req.Request.Context(), h.ctrlClient, imageReq)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	var am v1alpha1.AppImage
	err = wait.PollImmediate(time.Second, 2*time.Minute, func() (done bool, err error) {
		err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: imageReq.AppName}, &am)
		if err != nil && !apierrors.IsNotFound(err) {
			return false, err
		}
		if am.Status.State == "completed" {
			return true, nil
		}
		if am.Status.State == "failed" {
			return false, errors.New(am.Status.Message)
		}
		klog.Infof("poll app %s image info response", imageReq.AppName)
		return false, nil
	})
	if err != nil {
		klog.Errorf("poll app %s image info failed %v", imageReq.AppName, err)
		api.HandleError(resp, req, err)
		return
	}
	err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: imageReq.AppName}, &am)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	klog.Infof("finished app %s fetch image info request, time elapsed: %v", imageReq.AppName, time.Since(start))

	resp.WriteAsJson(map[string]interface{}{
		"name":   imageReq.AppName,
		"images": am.Status.Images,
	})
}

func createAppImage(ctx context.Context, ctrlClient client.Client, request *api.ImageInfoRequest) error {
	var nodes corev1.NodeList
	err := ctrlClient.List(ctx, &nodes, &client.ListOptions{})
	if err != nil {
		return err
	}
	nodeList := make([]string, 0)
	for _, node := range nodes.Items {
		if !utils.IsNodeReady(&node) || node.Spec.Unschedulable {
			continue
		}
		nodeList = append(nodeList, node.Name)
	}
	if len(nodeList) == 0 {
		return errors.New("cluster has no suitable node to schedule")
	}
	var am v1alpha1.AppImage
	err = ctrlClient.Get(ctx, types.NamespacedName{Name: request.AppName}, &am)
	if err == nil {
		if am.Status.State != "completed" && am.Status.State != "failed" {
			return nil
		}
		err = ctrlClient.Delete(ctx, &am)

		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	imageList := make([]string, 0, len(request.Images))
	for _, image := range request.Images {
		imageList = append(imageList, image.ImageName)
	}
	imagesString, err := json.Marshal(request)
	if err != nil {
		klog.Errorf("marshal appimage request failed %v", err)
		return err
	}
	m := v1alpha1.AppImage{
		ObjectMeta: metav1.ObjectMeta{
			Name: request.AppName,
			Annotations: map[string]string{
				api.AppImagesKey: string(imagesString),
			},
		},
		Spec: v1alpha1.ImageSpec{
			AppName: request.AppName,
			Refs:    imageList,
			Nodes:   nodeList,
		},
	}
	err = ctrlClient.Create(ctx, &m)
	if err != nil {
		return err
	}
	return nil
}
