package containerd

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/containerd/containerd/reference"
	"github.com/gofiber/fiber/v2"
	criruntimev1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/klog/v2"
)

var (
	ParamRegistryName   = "registry"
	DefaultRegistryName = "docker.io"

	ParamImageName = "image"
)

func GetRegistryMirrors(ctx *fiber.Ctx) (map[string]Mirror, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	criPluginConfig, err := getCRIPluginConfig(config)
	if err != nil {
		return nil, err
	}
	return criPluginConfig.Registry.Mirrors, nil
}

func GetRegistryMirror(ctx *fiber.Ctx) (*Mirror, error) {
	registry := ctx.Params(ParamRegistryName)
	if registry == "" {
		registry = DefaultRegistryName
	}

	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	criPluginConfig, err := getCRIPluginConfig(config)
	if err != nil {
		return nil, err
	}

	mirror := criPluginConfig.Registry.Mirrors[registry]
	return &mirror, nil
}

// EnsureRegistryMirror ensures the given endpoint is the first entry for the registry mirror.
// Returns updated=true if configuration changed and was persisted (containerd restarted),
// or updated=false if it was already first. Returns error on failure.
func EnsureRegistryMirror(ctx context.Context, registry string, endpoint string) (bool, error) {
	if registry == "" {
		registry = DefaultRegistryName
	}
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return false, fmt.Errorf("endpoint is required")
	}

	u, err := url.ParseRequestURI(endpoint)
	if err != nil || u == nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return false, fmt.Errorf("invalid mirror endpoint: %s", endpoint)
	}
	endpoint = u.String()

	config, err := getConfig()
	if err != nil {
		return false, err
	}
	criPluginConfig, err := getCRIPluginConfig(config)
	if err != nil {
		return false, err
	}

	if criPluginConfig.Registry.Mirrors == nil {
		criPluginConfig.Registry.Mirrors = make(map[string]Mirror)
	}

	mirror := criPluginConfig.Registry.Mirrors[registry]
	// Build a new endpoints list with the ensured endpoint at the front
	// and without duplicates (preserving the order of the remaining items).
	original := mirror.Endpoints
	var others []string
	for _, ep := range original {
		if ep != endpoint {
			others = append(others, ep)
		}
	}
	newEndpoints := make([]string, 0, 1+len(others))
	newEndpoints = append(newEndpoints, endpoint)
	newEndpoints = append(newEndpoints, others...)

	// If nothing changes (already first and unique), no update needed.
	equal := len(original) == len(newEndpoints)
	if equal {
		for i := range original {
			if original[i] != newEndpoints[i] {
				equal = false
				break
			}
		}
	}
	if equal {
		return false, nil
	}

	mirror.Endpoints = newEndpoints
	criPluginConfig.Registry.Mirrors[registry] = mirror
	if err := updateCRIPluginConfig(config, criPluginConfig); err != nil {
		return false, err
	}
	if err := restartContainerd(ctx); err != nil {
		klog.Errorf("failed to restart containerd: %v", err)
		return false, err
	}
	return true, nil
}

func UpdateRegistryMirror(ctx *fiber.Ctx) (*Mirror, error) {
	registry := ctx.Params(ParamRegistryName)
	if registry == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "registry name is required")
	}

	var mirror Mirror
	if err := ctx.BodyParser(&mirror); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	criPluginConfig, err := getCRIPluginConfig(config)
	if err != nil {
		return nil, err
	}

	if criPluginConfig.Registry.Mirrors == nil {
		criPluginConfig.Registry.Mirrors = make(map[string]Mirror)
	}

	criPluginConfig.Registry.Mirrors[registry] = mirror

	if err := updateCRIPluginConfig(config, criPluginConfig); err != nil {
		return nil, err
	}

	if err := restartContainerd(ctx.Context()); err != nil {
		klog.Errorf("failed to restart containerd: %v", err)
		return nil, err
	}

	return &mirror, nil
}

func DeleteRegistryMirror(ctx *fiber.Ctx) error {
	registry := ctx.Params(ParamRegistryName)
	if registry == "" {
		return fiber.NewError(fiber.StatusBadRequest, "registry name is required")
	}

	config, err := getConfig()
	if err != nil {
		return err
	}
	criPluginConfig, err := getCRIPluginConfig(config)
	if err != nil {
		return err
	}

	if criPluginConfig.Registry.Mirrors == nil {
		criPluginConfig.Registry.Mirrors = make(map[string]Mirror)
	}

	delete(criPluginConfig.Registry.Mirrors, registry)
	if err := updateCRIPluginConfig(config, criPluginConfig); err != nil {
		return err
	}

	if err := restartContainerd(ctx.Context()); err != nil {
		klog.Errorf("failed to restart containerd: %v", err)
		return err
	}

	return nil
}

// merge registries both from parsing images
// and mirror configs
func ListRegistries(ctx *fiber.Ctx) ([]*Registry, error) {
	nameToRegistries := make(map[string]*Registry)
	mirrors, err := GetRegistryMirrors(ctx)
	if err != nil {
		return nil, err
	}
	for registryName, mirror := range mirrors {
		nameToRegistries[registryName] = &Registry{Name: registryName, Endpoints: mirror.Endpoints}
	}
	images, err := ListImages(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			refspec, err := reference.Parse(tag)
			// for any images managed by us or pulled externally
			// this should not happen
			// it only happens if the user tagged a local image
			// to a tag that has an invalid format
			// which is very rare
			// so just log an error and ignore it
			if err != nil {
				klog.Errorf("failed to parse image tag %s: %v", tag, err)
				continue
			}
			host := refspec.Hostname()
			if host == "" {
				klog.Errorf("failed to parse image tag %s: empty host", tag)
				continue
			}
			if registry, ok := nameToRegistries[host]; !ok || registry == nil {
				nameToRegistries[host] = &Registry{Name: host}
			}
			nameToRegistries[host].ImageCount += 1
			nameToRegistries[host].ImageSize += image.Size
		}
	}
	var registries []*Registry
	for _, registry := range nameToRegistries {
		registries = append(registries, registry)
	}
	return registries, nil
}

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
