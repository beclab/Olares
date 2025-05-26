package os

import (
	"time"

	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
	"log"
)

func NewCmdStart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Olares OS",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.StartOlares(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}

func NewCmdStop() *cobra.Command {
	var (
		timeout       time.Duration
		checkInterval time.Duration
	)
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the Olares OS",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.StopOlares(timeout, checkInterval); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 1*time.Minute, "Timeout for graceful shutdown before using SIGKILL")
	cmd.Flags().DurationVarP(&checkInterval, "check-interval", "i", 10*time.Second, "Interval between checks for remaining processes")
	return cmd
}
