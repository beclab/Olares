package upgrade

import "github.com/beclab/Olares/daemon/pkg/commands"

// The upgrade watcher brings the control node to a release with these before
// the orchestrator starts: downloading runs before the upgrade is even
// confirmed, adopting once it is. Every other node gets the same work from the
// orchestrator's olaresd stages (see olaresdStages).
//
// They are two sequences rather than one because downloading is safe to repeat
// and touches nothing, while adopting replaces the binaries and restarts
// olaresd.
var (
	// ReleaseDownloadPhases fetch a release onto this machine. They change
	// nothing that is running.
	ReleaseDownloadPhases = []func() commands.Interface{
		NewDownloadCLI,
		NewDownloadWizard,
		NewDownloadSpaceCheck,
		NewDownloadComponent,
	}

	// ReleaseAdoptPhases make this machine run a release it has already
	// downloaded: the two binaries, then the images they will need.
	//
	// The middle one ends this process. NewInstallOlaresd replaces the daemon
	// and restarts it, so on both paths the sequence is expected not to
	// return, and both paths are expected to run it again afterwards — which
	// is why every phase here is reentrant, and why NewInstallOlaresd checks
	// its own version and does nothing when it is already the target.
	ReleaseAdoptPhases = []func() commands.Interface{
		NewInstallCLI,
		NewInstallOlaresd,
		NewImportImages,
	}
)
