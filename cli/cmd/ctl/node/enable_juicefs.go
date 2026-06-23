package node

import (
	"log"
	"time"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdEnableJuiceFS() *cobra.Command {
	var (
		stopTimeout       time.Duration
		stopCheckInterval time.Duration
	)
	cmd := &cobra.Command{
		Use:   "enable-juicefs",
		Short: "convert a single-node master's local rootfs to a JuiceFS-backed shared rootfs",
		Long: `Convert an already-installed single-node Olares master from a local-filesystem
rootfs to a JuiceFS-backed shared rootfs, in place, without uninstalling.

By default this installs the bundled object storage (MinIO) as the JuiceFS
backend. To use an external S3-compatible object store instead, pass the
storage flags (--storage-type s3/oss/cos, --s3-bucket, keys, etc.); in that
case MinIO is not installed and the credentials are validated instead.

It also installs the metadata engine (Redis), formats a JuiceFS filesystem,
migrates the existing rootfs data into it (an online bulk copy followed by a
brief offline incremental copy), swaps the rootfs directory for the JuiceFS
mount, and regenerates the cluster services.

After it completes, the master satisfies the JuiceFS precondition required to
add worker nodes, so you can run "olares-cli node add" on the workers.

The command is idempotent and resumable: if JuiceFS is already enabled it only
ensures Olares is running.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.EnableJuiceFSPipeline(cmd.Context(), stopTimeout, stopCheckInterval); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddStorageFlagsBy(flagSetter)
	cmd.Flags().DurationVarP(&stopTimeout, "timeout", "t", 1*time.Minute, "Timeout for graceful container shutdown before using SIGKILL during the offline sync window")
	cmd.Flags().DurationVarP(&stopCheckInterval, "check-interval", "i", 10*time.Second, "Interval between checks for remaining container processes")

	return cmd
}
