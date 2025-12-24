package images

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/utils/registry"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cmd/ctr/commands"
	"github.com/containerd/containerd/cmd/ctr/commands/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	refdocker "github.com/containerd/containerd/reference/docker"
	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const maxRetries = 6

var (
	sock = "/var/run/containerd/containerd.sock"
)

type PullOptions struct {
	AppName      string
	OwnerName    string
	AppNamespace string
}

type ImageService interface {
	PullImage(ctx context.Context, ref string, opts PullOptions) (string, error)
	Progress(ctx context.Context, ref string, opts PullOptions) (string, error)
}

type imageService struct {
	client *containerd.Client
}

func NewClient(ctx context.Context) (*imageService, context.Context, context.CancelFunc, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	_, err := os.Stat(sock)
	if err != nil {
		return nil, ctx, cancel, err
	}
	client, err := containerd.New(sock, containerd.WithDefaultNamespace("k8s.io"),
		containerd.WithTimeout(10*time.Second))
	if err != nil {
		return nil, ctx, cancel, err
	}
	return &imageService{
		client: client,
	}, ctx, cancel, nil
}

func (is *imageService) PullImage(ctx context.Context, ref appv1alpha1.Ref, opts PullOptions) (string, error) {
	originNamed, err := refdocker.ParseDockerRef(ref.Name)
	if err != nil {
		return ref.Name, err
	}
	ref.Name = originNamed.String()
	// replaced image ref
	currentMirrors := registry.GetMirrors()
	replacedRef, plainHTTP := utils.ReplacedImageRef(currentMirrors, originNamed.String(), false)
	config := newFetchConfig(plainHTTP)

	ongoing := newJobs(replacedRef, originNamed.String())

	imageName, err := is.GetExistsImage(originNamed.String())
	if err != nil && !errors.Is(err, errdefs.ErrNotFound) {
		klog.Infof("Failed to get image status err=%v", err)
		return "", err
	}
	present := imageName != ""

	pctx, stopProgress := context.WithCancel(ctx)

	progress := make(chan struct{})

	h := images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		if desc.MediaType != images.MediaTypeDockerSchema1Manifest {
			ongoing.add(desc)
		}
		return nil, nil
	})

	labels := commands.LabelArgs(config.Labels)
	remoteOpts := []containerd.RemoteOpt{
		containerd.WithPullLabels(labels),
		containerd.WithResolver(config.Resolver),
		containerd.WithImageHandler(h),
		containerd.WithPullUnpack,
		containerd.WithSchema1Conversion,
	}
	if config.AllMetadata {
		remoteOpts = append(remoteOpts, containerd.WithAllMetadata())
	}

	if config.PlatformMatcher != nil {
		remoteOpts = append(remoteOpts, containerd.WithPlatformMatcher(config.PlatformMatcher))
	} else {
		for _, platform := range config.Platforms {
			remoteOpts = append(remoteOpts, containerd.WithPlatform(platform))
		}
	}

	rCtx, _ := fetchCtx(is.client, remoteOpts...)

	go func() {
		showProgress(pctx, rCtx, ongoing, is.client.ContentStore(), shouldPUllImage(ref, present), opts)
		close(progress)
	}()

	if shouldPUllImage(ref, present) {
		downloadFunc := func() error {
			attempt := 1

			for {
				_, err = is.client.Pull(pctx, replacedRef, remoteOpts...)
				if err == nil {
					break
				}

				select {
				case <-pctx.Done():
					return pctx.Err()
				default:
				}

				if attempt >= maxRetries {
					return fmt.Errorf("download failed after %d attempts: %v", attempt, err)
				}

				delay := attempt * 5
				ticker := time.NewTicker(time.Second)
				klog.Infof("attempt %d", attempt)
				attempt++
			selectLoop:
				for {
					klog.Infof("Retrying in %d seconds", delay)
					select {
					case <-ticker.C:
						delay--
						if delay == 0 {
							ticker.Stop()
							break selectLoop
						}
					case <-pctx.Done():
						ticker.Stop()
						return pctx.Err()
					}
				}
			}
			err = is.tag(pctx, replacedRef, originNamed.String())
			if err != nil {
				return err
			}
			return nil
		}

		err = downloadFunc()
	}

	stopProgress()
	if err != nil {
		klog.Infof("fetch image name=%s err=%v", ref, err)
		return "", err
	}

	<-progress
	return originNamed.String(), nil
}

func (is *imageService) Progress(ctx context.Context, ref string, opts PullOptions) (string, error) {
	return "", nil
}

func (is *imageService) GetExistsImage(ref string) (string, error) {
	name, err := refdocker.ParseDockerRef(ref)
	if err != nil {
		return "", err
	}
	image, err := is.client.GetImage(context.TODO(), name.String())
	if err != nil {
		return "", err
	}
	return image.Name(), nil
}

func (is *imageService) tag(ctx context.Context, ref, targetRef string) error {
	if ref == targetRef {
		return nil
	}
	ctx, done, err := is.client.WithLease(ctx)
	if err != nil {
		return err
	}
	defer done(ctx)
	imgService := is.client.ImageService()
	image, err := imgService.Get(ctx, ref)
	if err != nil {
		return err
	}
	image.Name = targetRef
	if _, err = imgService.Create(ctx, image); err != nil {
		if errdefs.IsAlreadyExists(err) {
			if err = imgService.Delete(ctx, targetRef); err != nil {
				return err
			}
			if _, err = imgService.Create(ctx, image); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func shouldPUllImage(ref appv1alpha1.Ref, imagePresent bool) bool {
	if ref.ImagePullPolicy == corev1.PullNever {
		return false
	}
	if ref.ImagePullPolicy == corev1.PullAlways ||
		(ref.ImagePullPolicy == corev1.PullIfNotPresent && (!imagePresent)) {
		return true
	}
	return false
}

func newFetchConfig(plainHTTP bool) *content.FetchConfig {
	options := docker.ResolverOptions{
		PlainHTTP: plainHTTP,
	}
	resolver := docker.NewResolver(options)
	config := &content.FetchConfig{
		Resolver:        resolver,
		PlatformMatcher: platforms.Default(),
	}
	return config
}

func fetchCtx(client *containerd.Client, remoteOpts ...containerd.RemoteOpt) (*containerd.RemoteContext, error) {
	rCtx := &containerd.RemoteContext{
		Resolver: docker.NewResolver(docker.ResolverOptions{}),
	}
	for _, o := range remoteOpts {
		if err := o(client, rCtx); err != nil {
			return nil, err
		}
	}
	if rCtx.PlatformMatcher == nil {
		if len(rCtx.Platforms) == 0 {
			rCtx.PlatformMatcher = platforms.All
		} else {
			var ps []ocispec.Platform
			for _, s := range rCtx.Platforms {
				p, err := platforms.Parse(s)
				if err != nil {
					return nil, fmt.Errorf("invalid platform %s: %w", s, err)
				}
				ps = append(ps, p)
			}

			rCtx.PlatformMatcher = platforms.Any(ps...)
		}
	}
	return rCtx, nil
}
