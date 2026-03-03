package images

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appevent "github.com/beclab/Olares/framework/app-service/pkg/event"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	refdocker "github.com/containerd/containerd/reference/docker"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Image struct {
	Name string
	Size int64
}

type ImageManager interface {
	Create(ctx context.Context, am *appv1alpha1.ApplicationManager, refs []appv1alpha1.Ref) error
	UpdateStatus(ctx context.Context, name, state, message string) error
	PollDownloadProgress(ctx context.Context, am *appv1alpha1.ApplicationManager) error
}

type ImageManagerClient struct {
	client.Client
}

func NewImageManager(client client.Client) ImageManager {
	return &ImageManagerClient{client}
}

func (imc *ImageManagerClient) Create(ctx context.Context, am *appv1alpha1.ApplicationManager, refs []appv1alpha1.Ref) error {
	var nodes corev1.NodeList
	err := imc.List(ctx, &nodes, &client.ListOptions{})
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
	var im appv1alpha1.ImageManager
	err = imc.Get(ctx, types.NamespacedName{Name: am.Name}, &im)
	if err == nil {
		err = imc.Delete(ctx, &im)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	labels := make(map[string]string)
	if strings.HasSuffix(am.Spec.AppName, "-dev") && am.Spec.Source == "devbox" {
		labels["dev.bytetrade.io/dev-owner"] = am.Spec.AppOwner
	}
	m := appv1alpha1.ImageManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:   am.Name,
			Labels: labels,
			Annotations: map[string]string{
				constants.ApplicationImageLabel: am.Annotations[constants.ApplicationImageLabel],
			},
		},
		Spec: appv1alpha1.ImageManagerSpec{
			AppName:      am.Spec.AppName,
			AppNamespace: am.Spec.AppNamespace,
			AppOwner:     am.Spec.AppOwner,
			Refs:         refs,
			Nodes:        nodeList,
		},
	}
	err = imc.Client.Create(ctx, &m)
	if err != nil {
		return err
	}
	return nil
}

func (imc *ImageManagerClient) UpdateStatus(ctx context.Context, name, state, message string) error {
	var im appv1alpha1.ImageManager
	err := imc.Get(ctx, types.NamespacedName{Name: name}, &im)
	if err != nil {
		klog.Infof("get im err=%v", err)
		return err
	}

	now := metav1.Now()
	imCopy := im.DeepCopy()
	imCopy.Status.State = state
	imCopy.Status.Message = message
	imCopy.Status.StatusTime = &now
	imCopy.Status.UpdateTime = &now

	err = imc.Status().Patch(ctx, imCopy, client.MergeFrom(&im))
	if err != nil {
		return err
	}
	return nil
}

func (imc *ImageManagerClient) PollDownloadProgress(ctx context.Context, am *appv1alpha1.ApplicationManager) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	var lastProgress float64 = -1
	imageList := make([]Image, 0)
	err := json.Unmarshal([]byte(am.Annotations[constants.ApplicationImageLabel]), &imageList)
	if err != nil {
		klog.Errorf("failed unmarshal to images %v", err)
	}
	for i, ref := range imageList {
		name, err := refdocker.ParseDockerRef(ref.Name)
		if err != nil {
			continue
		}
		imageList[i].Name = name.String()
	}

	for {
		select {
		case <-ticker.C:
			var im appv1alpha1.ImageManager
			err := imc.Get(ctx, types.NamespacedName{Name: am.Name}, &im)
			if err != nil {
				klog.Infof("Failed to get imanagermanagers name=%s err=%v", am.Name, err)
				return err
			}

			if im.Status.State == "failed" {
				return errors.New(im.Status.Message)
			}

			type progress struct {
				offset int64
				total  int64
			}
			maxImageSize := int64(5086840033)

			nodeMap := make(map[string]*progress)
			for _, nodeName := range im.Spec.Nodes {
				for _, ref := range im.Spec.Refs {
					t := im.Status.Conditions[nodeName][ref.Name]["total"]
					if t == "" {
						imageSize := maxImageSize
						info := findImageSize(imageList, ref.Name)
						if info != nil && info.Size != 0 {
							//klog.Infof("get image:%s size:%d", ref.Name, info.Size)
							imageSize = info.Size
						}
						t = strconv.FormatInt(imageSize, 10)
					}

					total, _ := strconv.ParseInt(t, 10, 64)

					t = im.Status.Conditions[nodeName][ref.Name]["offset"]
					if t == "" {
						t = "0"
					}
					offset, _ := strconv.ParseInt(t, 10, 64)

					if _, ok := nodeMap[nodeName]; ok {
						nodeMap[nodeName].offset += offset
						nodeMap[nodeName].total += total
					} else {
						nodeMap[nodeName] = &progress{offset: offset, total: total}
					}
				}
			}
			ret := math.MaxFloat64
			for n, p := range nodeMap {
				var nodeProgress float64
				if p.total != 0 {
					nodeProgress = float64(p.offset) / float64(p.total)
				}
				if len(nodeMap) > 1 {
					klog.Infof("node: %s,app: %s, progress: %.2f", n, am.Spec.AppNamespace, nodeProgress)
				}

				if nodeProgress < ret {
					ret = nodeProgress
				}

			}
			err = imc.updateProgress(ctx, am, &lastProgress, ret*100, am.Spec.OpType == appv1alpha1.UpgradeOp)
			if err == nil {
				return nil
			}

		case <-ctx.Done():
			return context.Canceled
		}
	}
}

func (imc *ImageManagerClient) updateProgress(ctx context.Context, am *appv1alpha1.ApplicationManager, lastProgress *float64, progress float64, isUpgrade bool) error {
	if *lastProgress > progress {
		return errors.New("no need to update progress")
	}

	progressStr := strconv.FormatFloat(progress, 'f', 2, 64)
	if *lastProgress != progress {
		*lastProgress = progress

		appevent.PublishAppEventToQueue(utils.EventParams{
			Owner:  am.Spec.AppOwner,
			Name:   am.Spec.AppName,
			OpType: string(am.Status.OpType),
			OpID:   am.Status.OpID,
			State: func() string {
				if isUpgrade {
					return appv1alpha1.Upgrading.String()
				}
				return appv1alpha1.Downloading.String()
			}(),
			Progress:   progressStr,
			RawAppName: am.Spec.RawAppName,
			Type:       am.Spec.Type.String(),
			Title:      apputils.AppTitle(am.Spec.Config),
			Icon:       apputils.AppIcon(am.Spec.Config),
		})
	}
	klog.Infof("app %s download progress.... %v", am.Spec.AppName, progressStr)
	if progressStr == "100.00" {
		return nil
	}
	return errors.New("under downloading")
}

func findImageSize(imageList []Image, ref string) *Image {
	for _, v := range imageList {
		if v.Name == ref {
			return &v
		}
	}
	return nil
}
