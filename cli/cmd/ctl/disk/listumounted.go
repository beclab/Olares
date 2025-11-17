package disk

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/beclab/Olares/cli/pkg/utils/lvm"
	"github.com/spf13/cobra"
)

func NewListUnmountedDisksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-unmounted",
		Short: "List unmounted disks",
		Run: func(cmd *cobra.Command, args []string) {
			unmountedDevices, err := lvm.FindUnmountedDevices()
			if err != nil {
				log.Fatalf("Error finding unmounted devices: %v\n", err)
			}

			// print header
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

			fmt.Fprint(w, "Device\tSize\n")
			for path, device := range unmountedDevices {
				fmt.Fprintf(w, "%s\t%s\n", path, device.Size)
			}
			w.Flush()
		},
	}
	return cmd
}
