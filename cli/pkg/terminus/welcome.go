package terminus

import (
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type WelcomeMessage struct {
	common.KubeAction
}

func (t *WelcomeMessage) Execute(runtime connector.Runtime) error {
	logger.InfoInstallationProgress("Olares system installation is complete")
	logger.Info("No user has been created.")
	// olaresd uses this marker to recognize completion of the CLI install.
	logger.InfoInstallationProgress("All done")

	// If AMD GPU on Ubuntu 22.04/24.04, print warning about reboot for ROCm
	if si := runtime.GetSystemInfo(); si.IsUbuntu() && (si.IsUbuntuVersionEqual(connector.Ubuntu2204) || si.IsUbuntuVersionEqual(connector.Ubuntu2404)) {
		if si.IsRyzenAIMax() || si.IsAmdGPU() {
			logger.Warnf("\x1b[31mWarning: To enable ROCm, please reboot your machine after installation.\x1b[0m")
			fmt.Println()
		}
	}

	return nil
}

type WelcomeModule struct {
	common.KubeModule
}

func (m *WelcomeModule) Init() {
	logger.InfoInstallationProgress("Starting Olares ...")
	m.Name = "Welcome"

	waitServicesReady := &task.LocalTask{
		Name:   "WaitServicesReady",
		Action: new(CheckSystemComponentsReady),
		Retry:  60,
		Delay:  15 * time.Second,
	}

	welcomeMessage := &task.LocalTask{
		Name:   "WelcomeMessage",
		Action: new(WelcomeMessage),
	}

	m.Tasks = append(m.Tasks, waitServicesReady, welcomeMessage)
}
