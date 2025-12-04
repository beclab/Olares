package images

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"bytetrade.io/web3os/app-service/pkg/utils"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

// StatusInfo holds the status info for an upload or download
type StatusInfo struct {
	Ref       string
	Status    string
	Offset    int64
	Total     int64
	StartedAt time.Time
	UpdatedAt time.Time
}

var retryStrategy = wait.Backoff{
	Steps:    20,
	Duration: 10 * time.Millisecond,
	Factor:   3.0,
	Jitter:   0.1,
}

func showProgress(ctx context.Context, rCtx *containerd.RemoteContext, ongoing *jobs, cs content.Store, needPullImage bool, opts PullOptions) {
	var (
		interval = rand.Float64() + float64(1)
		ticker   = time.NewTicker(time.Duration(interval * float64(time.Second)))
		start    = time.Now()
		statuses = map[string]StatusInfo{}
		done     bool
		ordered  []StatusInfo
	)
	defer ticker.Stop()

	// no need to pull image, just update image manager status
	if !needPullImage {
		err := setPulledImageStatus(ongoing.originRef, opts)
		if err != nil {
			klog.Infof("setPulledImageStatus name=%v, err=%v", ongoing.name, err)
		}
		return
	}

	if rCtx == nil {
		klog.Infof("show progress with nil remote context")
		return
	}

	attempt := 0
	var descs []ocispec.Descriptor
	err := retry.OnError(wait.Backoff{
		Steps:    10,
		Duration: time.Second,
		Factor:   2.0,
		Jitter:   0.1,
	}, func(error) bool { return true }, func() error {
		_, desc, err := rCtx.Resolver.Resolve(context.TODO(), ongoing.name)
		if err != nil {
			klog.Infof("resolve ref failed err=%v", err)
			return err
		}
		descs, err = getAllLayerDescriptor(rCtx, desc, cs)
		if err != nil {
			attempt++
			klog.Infof("image %s,attempt %d to get layer descriptor err=%v", ongoing.name, attempt, err)
			return err
		}
		for _, desc := range descs {
			if images.IsLayerType(desc.MediaType) {
				return nil
			}
		}
		return errors.New("not get completed layers")
	})
	if err != nil {
		klog.Infof("get all layer descriptor failed err=%v", err)
		return
	}
	var imageSize int64
	descSet := make(map[string]int64)
	for _, d := range descs {
		if _, ok := descSet[d.Digest.String()]; ok {
			continue
		}
		if images.IsLayerType(d.MediaType) {
			klog.Infof("typ: %s, Digest: %s, size: %d", d.MediaType, d.Digest.String(), d.Size)
			imageSize += d.Size
		}

		descSet[d.Digest.String()] = d.Size
	}
	klog.Infof("imagesize: %v", imageSize)
outer:
	for {
		select {
		case <-ticker.C:
			resolved := "resolved"
			if !ongoing.isResolved() {
				resolved = "resolving"
			}
			statuses[ongoing.name] = StatusInfo{
				Ref:    ongoing.name,
				Status: resolved,
			}

			keys := []string{ongoing.name}

			activeSeen := map[string]struct{}{}
			if !done {
				active, err := cs.ListStatuses(ctx, "")
				if err != nil {
					klog.ErrorS(err, "active check failed")
					continue
				}
				// update status of active entries!
				for _, a := range active {
					statuses[a.Ref] = StatusInfo{
						Ref:       a.Ref,
						Status:    "downloading",
						Offset:    a.Offset,
						Total:     a.Total,
						StartedAt: a.StartedAt,
						UpdatedAt: a.UpdatedAt,
					}
					activeSeen[a.Ref] = struct{}{}
				}
			}

			// now, update the items in jobs that are not in active
			for _, j := range ongoing.jobs() {
				key := remotes.MakeRefKey(ctx, j)
				keys = append(keys, key)
				if _, ok := activeSeen[key]; ok {
					continue
				}

				status, ok := statuses[key]
				if !done && (!ok || status.Status == "downloading") {
					info, err := cs.Info(ctx, j.Digest)
					if err != nil {
						if !errdefs.IsNotFound(err) {
							klog.Errorf("Failed to get content info err=%v", err)
							continue outer
						} else {
							statuses[key] = StatusInfo{
								Ref:    key,
								Status: "waiting",
							}
						}
					} else if info.CreatedAt.After(start) {
						statuses[key] = StatusInfo{
							Ref:       key,
							Status:    "done",
							Offset:    info.Size,
							Total:     info.Size,
							UpdatedAt: info.CreatedAt,
						}
					} else {
						statuses[key] = StatusInfo{
							Ref:    key,
							Status: "exists",
						}
					}
				} else if done {
					if ok {
						if status.Status != "done" && status.Status != "exists" {
							status.Status = "done"
							statuses[key] = status
						}
					} else {
						statuses[key] = StatusInfo{
							Ref:    key,
							Status: "exists",
						}
					}
				}
			}

			ordered = []StatusInfo{}
			for _, key := range keys {
				if _, ok := statuses[key]; ok {
					if key == ongoing.name {
						continue
					}
				}
				ordered = append(ordered, statuses[key])
			}
			klog.Infof("downloading image %v", ongoing.name)
			err := updateProgress(ordered, ongoing, descSet, imageSize, opts)
			if err != nil {
				klog.Infof("update progress failed err=%v", err)
			}

			if done {
				klog.Infof("image %s progress is done", ongoing.name)
				return
			}
		case <-ctx.Done():
			done = true // allow ui to update once more
		}
	}
}

func updateProgress(statuses []StatusInfo, ongoing *jobs, seen map[string]int64, imageSize int64, opts PullOptions) error {
	client, err := utils.GetClient()
	if err != nil {
		return err
	}
	var offset int64
	var progress float64

	klog.Infof("imageName=%s", ongoing.name)
	statusesLen := len(statuses)
	doneLayer := 0
	isLayerType := func(mediaType string) bool {
		if strings.HasPrefix(mediaType, "manifest-") || strings.HasPrefix(mediaType, "index-") || strings.HasPrefix(mediaType, "config-") {
			return false
		}
		return true
	}
	for _, status := range statuses {
		klog.Infof("status: %s,ref: %v, offset: %v, Total: %v", status.Status, status.Ref, status.Offset, status.Total)
		if status.Status == "exists" {
			key := strings.Split(status.Ref, "-")[1]
			offset += seen[key]
			doneLayer++
			continue
		}

		if !isLayerType(status.Ref) {
			statusesLen--
			continue
		}
		if status.Status == "done" {
			offset += status.Total
			doneLayer++
			continue
		}

		offset += status.Offset
	}
	if doneLayer == statusesLen && doneLayer != 0 {
		offset = imageSize
	}
	if imageSize != 0 {
		progress = float64(offset) / float64(imageSize) * float64(100)
	}

	klog.Infof("download image %s progress=%v, imageSize=%d, offset=%d", ongoing.name, progress, imageSize, offset)
	klog.Infof("#######################################")

	err = retry.RetryOnConflict(retryStrategy, func() error {
		name, _ := apputils.FmtAppMgrName(opts.AppName, opts.OwnerName, opts.AppNamespace)
		im, err := client.AppV1alpha1().ImageManagers().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			klog.Infof("cannot found image manager err=%v", err)
			return err
		}

		now := metav1.Now()
		imCopy := im.DeepCopy()
		imCopy.Status.StatusTime = &now
		imCopy.Status.UpdateTime = &now
		thisNode := os.Getenv("NODE_NAME")
		p := im.Status.Conditions[thisNode][ongoing.originRef]["offset"]
		oldOffset, _ := strconv.ParseInt(p, 10, 64)
		if oldOffset >= offset || offset > imageSize {
			return nil
		}

		if imCopy.Status.Conditions == nil {
			imCopy.Status.Conditions = make(map[string]map[string]map[string]string)
		}
		if imCopy.Status.Conditions[thisNode] == nil {
			imCopy.Status.Conditions[thisNode] = make(map[string]map[string]string)

		}
		imCopy.Status.Conditions[thisNode][ongoing.originRef] = map[string]string{
			"total":  strconv.FormatInt(imageSize, 10),
			"offset": strconv.FormatInt(offset, 10),
		}

		_, err = client.AppV1alpha1().ImageManagers().UpdateStatus(context.TODO(), imCopy, metav1.UpdateOptions{})
		if err != nil {
			klog.Infof("update imagemanager name=%s status err=%v", imCopy.Name, err)
			return err
		}

		return nil
	})
	if err != nil {
		klog.Infof("update status in showprogress error=%v", err)
		return err
	}
	return nil
}

func setPulledImageStatus(imageRef string, opts PullOptions) error {
	client, err := utils.GetClient()
	if err != nil {
		return err
	}
	thisNode := os.Getenv("NODE_NAME")
	imageManagerName, _ := apputils.FmtAppMgrName(opts.AppName, opts.OwnerName, opts.AppNamespace)
	err = retry.RetryOnConflict(retryStrategy, func() error {
		im, err := client.AppV1alpha1().ImageManagers().Get(context.TODO(), imageManagerName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		now := metav1.Now()
		imCopy := im.DeepCopy()
		imCopy.Status.StatusTime = &now
		imCopy.Status.UpdateTime = &now

		if imCopy.Status.Conditions == nil {
			imCopy.Status.Conditions = make(map[string]map[string]map[string]string)
		}
		if imCopy.Status.Conditions[thisNode] == nil {
			imCopy.Status.Conditions[thisNode] = make(map[string]map[string]string)
		}
		imCopy.Status.Conditions[thisNode][imageRef] = map[string]string{
			"offset": "56782302",
			"total":  "56782302",
		}
		_, err = client.AppV1alpha1().ImageManagers().UpdateStatus(context.TODO(), imCopy, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// jobs provides a way of identifying the download keys for a particular task
// encountering during the pull walk.
//
// This is very minimal and will probably be replaced with something more
// featured.
type jobs struct {
	name      string
	added     map[digest.Digest]struct{}
	descs     []ocispec.Descriptor
	mu        sync.Mutex
	resolved  bool
	originRef string
}

func newJobs(name string, originRef string) *jobs {
	return &jobs{
		name:      name,
		originRef: originRef,
		added:     map[digest.Digest]struct{}{},
	}
}

func (j *jobs) add(desc ocispec.Descriptor) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.resolved = true

	if _, ok := j.added[desc.Digest]; ok {
		return
	}
	j.descs = append(j.descs, desc)
	j.added[desc.Digest] = struct{}{}
}

func (j *jobs) jobs() []ocispec.Descriptor {
	j.mu.Lock()
	defer j.mu.Unlock()

	var descs []ocispec.Descriptor
	return append(descs, j.descs...)
}

func (j *jobs) isResolved() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.resolved
}

func getAllLayerDescriptor(rCtx *containerd.RemoteContext, root ocispec.Descriptor, cs content.Store) ([]ocispec.Descriptor, error) {
	ans := make([]ocispec.Descriptor, 0)
	childrenHandler := images.ChildrenHandler(cs)
	childrenHandler = images.SetChildrenMappedLabels(cs, childrenHandler, rCtx.ChildLabelMap)
	if rCtx.AllMetadata {
		childrenHandler = remotes.FilterManifestByPlatformHandler(childrenHandler, rCtx.PlatformMatcher)
	} else {
		childrenHandler = images.FilterPlatforms(childrenHandler, rCtx.PlatformMatcher)
	}

	var getDescriptor func(descs ...ocispec.Descriptor) error
	getDescriptor = func(descs ...ocispec.Descriptor) error {
		if len(descs) == 0 {
			return nil
		}
		for _, c := range descs {
			ans = append(ans, c)
			children, err := childrenHandler.Handle(context.TODO(), c)
			if err != nil {
				return err
			}
			getDescriptor(children...)
		}
		return nil
	}
	err := getDescriptor(root)
	return ans, err
}
