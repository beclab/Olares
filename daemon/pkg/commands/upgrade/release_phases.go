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
	// downloaded, so that the orchestrator it starts is the target's.
	// Images are loaded later, by the import-images stage on every node.
	//
	// NewInstallOlaresd ends this process: it replaces the daemon and
	// restarts it, and the watcher runs the sequence again afterwards. That
	// is why both phases are reentrant, and why NewInstallOlaresd does
	// nothing when it is already the target.
	ReleaseAdoptPhases = []func() commands.Interface{
		NewInstallCLI,
		NewInstallOlaresd,
	}
)
