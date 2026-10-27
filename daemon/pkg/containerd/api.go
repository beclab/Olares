package containerd

import (
	"fmt"
	"strings"

	"github.com/containerd/containerd/reference"
	"github.com/gofiber/fiber/v2"
	criruntimev1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/klog/v2"
)

// olaresd is a node-level daemon: it reports/manages THIS node's local
// containerd images (this file) and its registry mirror configuration (see
// registry.go).

var (
	ParamImageName = "image"
)

func ListImages(ctx *fiber.Ctx, registry string) ([]*criruntimev1.Image, error) {
	criImageService, err := NewCRIImageService()
	if err != nil {
		return nil, fmt.Errorf("create CRI image service failed: %v", err)
	}
	images, err := criImageService.ListImages(ctx.Context(), &criruntimev1.ImageFilter{})
	if err != nil {
		return nil, err
	}
	if registry == "" {
		return images, nil
	}
	var filteredImages []*criruntimev1.Image
	for _, image := range images {
		for _, tag := range image.RepoTags {
			refspec, err := reference.Parse(tag)
			if err != nil {
				klog.Errorf("failed to parse image tag %s: %v", tag, err)
				continue
			}
			if refspec.Hostname() == registry {
				filteredImages = append(filteredImages, image)
			}
		}
	}
	return filteredImages, nil
}

func DeleteImage(ctx *fiber.Ctx) error {
	image := ctx.Params(ParamImageName)
	criImageService, err := NewCRIImageService()
	if err != nil {
		return fmt.Errorf("create CRI image service failed: %v", err)
	}
	return criImageService.RemoveImage(ctx.Context(), &criruntimev1.ImageSpec{Image: image})
}

func PruneImages(ctx *fiber.Ctx) (*PruneImageResult, error) {
	criImageService, err := NewCRIImageService()
	if err != nil {
		return nil, fmt.Errorf("create CRI image service failed: %v", err)
	}
	images, err := criImageService.ListImages(ctx.Context(), &criruntimev1.ImageFilter{})
	if err != nil {
		return nil, fmt.Errorf("list all images failed: %v", err)
	}
	idsToImages := make(map[string]*criruntimev1.Image)
	for _, image := range images {
		if image.Pinned {
			continue
		}
		idsToImages[image.Id] = image
	}
	criRuntimeService, err := NewCRIRuntimeService()
	if err != nil {
		return nil, fmt.Errorf("create CRI runtime service failed: %v", err)
	}
	containers, err := criRuntimeService.ListContainers(ctx.Context(), &criruntimev1.ContainerFilter{})
	if err != nil {
		return nil, fmt.Errorf("list all containers failed: %v", err)
	}
	for _, container := range containers {
		delete(idsToImages, container.ImageRef)
	}
	res := &PruneImageResult{}
	for id, image := range idsToImages {
		for _, tag := range image.RepoTags {
			// temporary hack to avoid prune critical sandbox images
			// it can be removed later when we upgrade containerd to at least v1.6.30
			// and adds image pinning logics to olares-cli and/or app-service
			if strings.Contains(tag, "pause") {
				continue
			}
		}
		err := criImageService.RemoveImage(ctx.Context(), &criruntimev1.ImageSpec{Image: id})
		if err != nil {
			klog.Errorf("failed to remove image %s: %v", id, err)
			continue
		}
		res.Images = append(res.Images, image)
		res.Count += 1
		res.Size += image.Size
	}
	return res, nil
}
