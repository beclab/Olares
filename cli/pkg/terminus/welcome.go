package terminus

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"fmt"
	"net"
	"time"
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
	publicNetworkInfo := t.KubeConf.Arg.PublicNetworkInfo
	publicIPs = append(publicIPs, publicNetworkInfo.OSPublicIPs...)
	if publicNetworkInfo.AWSPublicIP != nil {
		publicIPs = append(publicIPs, publicNetworkInfo.AWSPublicIP)
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
	} else if publicNetworkInfo.PubliclyAccessible && publicNetworkInfo.ExternalPublicIP != nil {
		fmt.Println()
		logger.Info("this machine is explicitly specified as publicly accessible")
		logger.Info("but no public IP address can be found from the system")
		logger.Info("a reflected public IP as seen by others on the internet is determined on a best effort basis:")
		logger.Infof("http://%s:%d", publicNetworkInfo.ExternalPublicIP, port)
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
