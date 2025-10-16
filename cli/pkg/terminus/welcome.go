package terminus

import (
	"fmt"
	"net"
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
	port := 30180
	localIP := runtime.GetSystemInfo().GetLocalIp()
	if si := runtime.GetSystemInfo(); si.GetNATGateway() != "" {
		localIP = si.GetNATGateway()
	}
	var publicIPs []net.IP
	networkSettings := t.KubeConf.Arg.NetworkSettings
	publicIPs = append(publicIPs, networkSettings.OSPublicIPs...)
	if networkSettings.CloudProviderPublicIP != nil {
		publicIPs = append(publicIPs, networkSettings.CloudProviderPublicIP)
	}
	var filteredPublicIPs []net.IP
	for _, publicIP := range publicIPs {
		if publicIP == nil {
			continue
		}
		if publicIP.String() == localIP {
			continue
		}
		for _, filteredIP := range filteredPublicIPs {
			if filteredIP.String() == publicIP.String() {
				continue
			}
		}
		filteredPublicIPs = append(filteredPublicIPs, publicIP)
	}

	logger.InfoInstallationProgress("Installation wizard is complete")
	logger.InfoInstallationProgress("All done")
	fmt.Printf("\n\n\n\n------------------------------------------------\n\n")
	logger.Info("Olares is running locally at:")
	logger.Infof("http://%s:%d", localIP, port)
	if len(filteredPublicIPs) > 0 {
		fmt.Println()
		logger.Info("and publicly accessible at:")
		for _, publicIP := range filteredPublicIPs {
			logger.Infof("http://%s:%d", publicIP, port)
		}
	} else if networkSettings.EnableReverseProxy != nil && !*networkSettings.EnableReverseProxy && networkSettings.ExternalPublicIP != nil {
		fmt.Println()
		logger.Info("this installation is explicitly configured to disable reverse proxy")
		logger.Info("but no public IP address can be found from the system")
		logger.Info("a reflected public IP as seen by internet peers is determined on a best effort basis:")
		logger.Infof("http://%s:%d", networkSettings.ExternalPublicIP, port)
	}
	fmt.Println()
	logger.Info("Open your browser and visit the above address")
	logger.Info("with the following credentials:")
	fmt.Println()
	logger.Infof("Username: %s", t.KubeConf.Arg.User.UserName)
	logger.Infof("Password: %s", t.KubeConf.Arg.User.Password)
	fmt.Printf("\n------------------------------------------------\n\n\n\n\n")

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
		Action: new(CheckKeyPodsRunning),
		Retry:  60,
		Delay:  15 * time.Second,
	}

	welcomeMessage := &task.LocalTask{
		Name:   "WelcomeMessage",
		Action: new(WelcomeMessage),
	}

	m.Tasks = append(m.Tasks, waitServicesReady, welcomeMessage)
}
