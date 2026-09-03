package upgrade

import "github.com/beclab/Olares/daemon/pkg/commands"

// Bringing one machine to a release is the same work wherever it happens, so
// it is written down once here.
//
// The control node does it to itself, driven by the upgrade watcher; a compute
// node has it done to it by the orchestrator, as the node prepare stage. Those
// two used to carry their own copies of the list, described in a comment as
// "the same commands the upgrade watcher runs on the control node, minus the
// two that are not a node's business" — which is exactly the kind of agreement
// that stops being true without anyone noticing. It already had: the download
// step picked its architecture with a comparison no arm64 build ever matched,
// and nothing caught it while the control node was the only machine that ran
// it.
//
// Splitting it in two rather than one list is not cosmetic. Downloading is
// safe to repeat and touches nothing; adopting replaces the binaries and
// restarts olaresd. The control node needs them at different points — its
// download runs before the upgrade is even confirmed — so they are separate
// sequences that happen to be adjacent on a compute node.
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
