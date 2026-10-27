package terminus

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
)

type GetNATGatewayIP struct {
	common.KubeAction
}

func (s *GetNATGatewayIP) Execute(runtime connector.Runtime) error {
	var prompt string
	var input string
	var retry bool
	var systemInfo = runtime.GetSystemInfo()
	var hostIP = s.KubeConf.Arg.HostIP
	disableHostIPPrompt := os.Getenv(common.ENV_DISABLE_HOST_IP_PROMPT)
	switch {
	case systemInfo.IsWsl() || systemInfo.IsWindows():
		if strings.EqualFold(disableHostIPPrompt, "") || !util.IsValidIPv4Addr(net.ParseIP(hostIP)) {
			prompt = "the NAT gateway(the Windows host)'s IP is " + hostIP + ", Confirm[Y] or ReEnter [R]: "
		} else {
			input = hostIP
		}
	case systemInfo.IsDarwin():
		if strings.EqualFold(disableHostIPPrompt, "") || !util.IsValidIPv4Addr(net.ParseIP(hostIP)) {
			if hostIP == "" {
				hostIP = systemInfo.GetLocalIp()
			}
			prompt = "the NAT gateway(the MacOS host)'s IP is " + hostIP + ", Confirm[Y] or ReEnter [R]: "
		} else {
			input = hostIP
		}
	case s.KubeConf.Arg.IsOlaresInContainer:
		if hostIP == "" {
			return errors.Errorf("host ip not found")
		}
		input = hostIP
	default:
		return nil
	}

	if prompt != "" {
		reader, err := utils.GetBufIOReaderOfTerminalInput()
		if err != nil {
			return errors.Wrap(err, "failed to get terminal input reader")
		}
	LOOP:
		if !retry {
			fmt.Print(prompt)
		} else {
			fmt.Printf("\nEnter the NAT gateway IP: ")
		}

		input, err = reader.ReadString('\n')
		if input == "" {
			if err != nil && err.Error() != "EOF" {
				return err
			}
		}
		if retry {
			input = strings.TrimSpace(input)
			if !util.IsValidIPv4Addr(net.ParseIP(input)) {
				fmt.Printf("\nsorry, invalid IP, please try again.\n")
				goto LOOP
			}
		} else {
			input = strings.TrimSpace(input)
			switch input {
			case "Y":
				input = hostIP
				break
			case "R":
				retry = true
				fallthrough
			default:
				goto LOOP
			}
		}

		if !util.IsValidIPv4Addr(net.ParseIP(input)) {
			fmt.Printf("\nsorry, invalid IP, please try again.\n")
			goto LOOP
		}
	}

	logger.Infof("Nat Gateway IP: %s", input)
	runtime.GetSystemInfo().SetNATGateway(input)
	return nil
}

type UpdateNATGatewayIP struct {
	common.KubeAction
}

func (a *UpdateNATGatewayIP) Execute(runtime connector.Runtime) error {
	hostIP := runtime.GetSystemInfo().GetNATGateway()
	if hostIP == "" {
		return nil
	}
	var kubectlCMD string
	var kubectlArgs []string
	var err error
	if runtime.GetSystemInfo().IsWindows() {
		kubectlCMD = "cmd"
		kubectlArgs = []string{"/C", "wsl", "-d", a.KubeConf.Arg.WSLDistribution, "-u", "root", common.CommandKubectl}
	} else {
		kubectlCMD, err = util.GetCommand(common.CommandKubectl)
		if err != nil {
			return errors.Wrap(err, "kubectl not found")
		}
	}
	// Server-side apply owns only this key, preserving the controller's config
	// and observed addresses. This also works before the first user exists.
	payload, err := json.Marshal(map[string]interface{}{
		"apiVersion": "v1", "kind": "ConfigMap",
		"metadata": map[string]string{"name": "default-network", "namespace": "os-network"},
		"data":     map[string]string{"nat-gateway-ip": hostIP},
	})
	if err != nil {
		return err
	}
	args := append(kubectlArgs, "apply", "--server-side", "--field-manager=olares-cli-network", "-f", "-")
	cmd := exec.Command(kubectlCMD, args...)
	cmd.Stdin = strings.NewReader(string(payload))
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "failed to update global NAT gateway, output: %s", output)
	}
	return nil
}
