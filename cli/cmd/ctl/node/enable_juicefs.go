package node

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdEnableJuiceFS() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable-juicefs",
		Short: "convert a single-node master's local rootfs to a JuiceFS-backed shared rootfs",
		Long: `Convert an already-installed single-node Olares master from a local-filesystem
rootfs to a JuiceFS-backed shared rootfs, in place, without uninstalling.

Olares must be stopped before running this command: run "olares-cli stop"
first. If Olares is still running, the command aborts and asks you to stop it.
With the cluster stopped the rootfs is quiescent, so the data is migrated in a
single full sync.

By default this installs the bundled object storage (MinIO) as the JuiceFS
backend. To use an external S3-compatible object store instead, pass the
storage flags (--storage-type s3/oss/cos, --s3-bucket, keys, etc.); in that
case MinIO is not installed and the credentials are validated instead.

It also installs the metadata engine (Redis), formats a JuiceFS filesystem,
migrates the existing rootfs data into it, swaps the rootfs directory for the
JuiceFS mount, regenerates the cluster services, and starts Olares back up.

After it completes, the master satisfies the JuiceFS precondition required to
add worker nodes, so you can run "olares-cli node add" on the workers.

If JuiceFS is already enabled, the command reports that the migration is
already complete and exits without touching Olares.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.EnableJuiceFSPipeline(cmd.Context()); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddStorageFlagsBy(flagSetter)

	return cmd
}
