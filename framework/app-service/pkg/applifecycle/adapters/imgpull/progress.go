package imgpull

import (
	"context"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/images"
)

// IM Status.State string constants. Kept as local consts (rather
// than imported) because they are historically raw strings in the
// legacy code (see pkg/images/client.go:147 and
// controllers/image_controller.go:187-190) — never promoted to
// typed enums on the CRD. Duplicating them here keeps the adapter
// independent of pkg/images internals; if the state names ever get
// promoted to CRD constants we swap in one place.
const (
	imStateFailed    = "failed"
	imStateCompleted = "completed"
)

// Progress implements ports.ImagePuller.
//
// Contract per ports/image.go:
//   - Zero value + nil error means "no IM CR exists yet" (i.e. the
//     Runner just called CreateJob and the CR is not yet visible
//     via the informer, OR the AM has never been Downloading).
//   - Percent is 0..100 inclusive.
//   - Percent=100 with Complete=true means success.
//   - Percent=100 with Failed=true means "all refs decided, at
//     least one failed permanently".
//
// Numeric progress is images.DownloadPercent — the same formula
// PollDownloadProgress and the app-status HTTP endpoint use
// (annotation size fallback when a ref has not reported total).
// This adapter only adds the terminal-state flags and does not
// poll or publish events; the Runner owns those.
func (p *Puller) Progress(ctx context.Context, amName string) (ports.ImageProgress, error) {
	if amName == "" {
		return ports.ImageProgress{}, nil
	}
	var im appv1.ImageManager
	if err := p.c.Get(ctx, types.NamespacedName{Name: amName}, &im); err != nil {
		if apierrors.IsNotFound(err) {
			return ports.ImageProgress{}, nil
		}
		return ports.ImageProgress{}, err
	}
	return computeProgress(&im, progressImageListAM(ctx, p, amName, &im)), nil
}

// progressImageListAM is the AM whose ApplicationImageLabel feeds
// DownloadPercent's size fallback. Prefer the live AM; if it is
// missing or has no image list, use the copy CreateJob stamped
// onto the IM CR.
func progressImageListAM(ctx context.Context, p *Puller, amName string, im *appv1.ImageManager) *appv1.ApplicationManager {
	var am appv1.ApplicationManager
	if err := p.c.Get(ctx, types.NamespacedName{Name: amName}, &am); err == nil {
		if am.Annotations[constants.ApplicationImageLabel] != "" {
			return &am
		}
	}
	return &appv1.ApplicationManager{
		ObjectMeta: im.ObjectMeta,
	}
}

// computeProgress folds IM terminal state into ImageProgress and
// delegates the numeric percent to images.DownloadPercent.
func computeProgress(im *appv1.ImageManager, am *appv1.ApplicationManager) ports.ImageProgress {
	prog := ports.ImageProgress{
		Reason:  im.Status.State,
		Message: im.Status.Message,
	}

	switch im.Status.State {
	case imStateFailed:
		prog.Failed = true
		return prog
	case appv1.DownloadingCanceled.String():
		// Written by CancelJob above, and by the legacy cancel call
		// sites. Terminal for the puller's purposes — reported as
		// Failed so a poll loop stops rather than spinning on a CR
		// that will never progress. Reason carries the distinction
		// for callers that need to tell cancel from failure.
		prog.Failed = true
		return prog
	case imStateCompleted:
		prog.Percent = 100
		prog.Complete = true
		return prog
	}

	percent := images.DownloadPercent(im, am)
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	prog.Percent = percent
	return prog
}
