package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdInstallOs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Olares",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.CliInstallTerminusPipeline(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)
	flagSetter.Add(common.FlagOSUserName,
		"",
		"",
		"Set the username for the Olares instance, if not set, will be prompted for input",
	).WithEnv(common.EnvLegacyOSUserName)
	flagSetter.Add(common.FlagOSDomainName,
		"",
		"",
		"Set the domain name for the Olares instance, if not set, will be prompted for input",
	).WithEnv(common.EnvLegacyOSDomainName)
	flagSetter.Add(common.FlagOSPassword,
		"",
		"",
		"Set the inital password for the first user of the Olares instance, if not set, a randomly generated password will be used",
	)
	flagSetter.Add(common.FlagEnableReverseProxy,
		"",
		false,
		"Enable reverse proxy, if not set, will be dynamically enabled if public IP is not detected, and disabled otherwise",
	)
	flagSetter.Add(common.FlagEnableJuiceFS,
		"",
		false,
		"Use JuiceFS as the rootfs for Olares workloads, rather than the local disk.",
	).WithAlias(common.FlagLegacyEnableJuiceFS).WithEnv(common.EnvLegacyEnableJuiceFS)
	flagSetter.Add(common.FlagEnablePodSwap,
		"",
		false,
		"Enable pods on Kubernetes cluster to use swap, setting --enable-zram, --zram-size or --zram-swap-priority implicitly enables this option, regardless of the command line args, note that only pods of the BestEffort QOS group can use swap due to K8s design",
	)
	flagSetter.Add(common.FlagSwappiness,
		"",
		false,
		"Configure the Linux swappiness value, if not set, the current configuration is remained",
	)
	flagSetter.Add(common.FlagEnableZRAM,
		"",
		false,
		"Set up a ZRAM device to be used for swap, setting --zram-size or --zram-swap-priority implicitly enables this option, regardless of the command line args",
	)
	flagSetter.Add(common.FlagZRAMSize,
		"",
		"",
		"Set the size of the ZRAM device, takes a format of https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity, defaults to half of the total RAM",
	)
	flagSetter.Add(common.FlagZRAMSwapPriority,
		"",
		false,
		"Set the swap priority of the ZRAM device, between -1 and 32767, defaults to 100",
	)

	config.AddCDNServiceFlagBy(flagSetter)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddStorageFlagsBy(flagSetter)
	config.AddKubeTypeFlagBy(flagSetter)
	config.AddMiniKubeProfileFlagBy(flagSetter)

	cmd.AddCommand(NewCmdInstallStorage())
	return cmd
}
