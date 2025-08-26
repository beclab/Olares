package upgrade

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"os"
	"strings"
)

type upgrader_1_12_0_20250702 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_0_20250702) Version() *semver.Version {
	return semver.MustParse("1.12.0-20250702")
}

func (u upgrader_1_12_0_20250702) PrepareForUpgrade() []task.Interface {
	preTasks := []task.Interface{
		&task.LocalTask{
			Name:   "UpdateSysctlReservedPorts",
			Action: new(updateSysctlReservedPorts),
		},
	}
	return append(preTasks, u.upgraderBase.PrepareForUpgrade()...)
}

type updateSysctlReservedPorts struct {
	common.KubeAction
}

func (u *updateSysctlReservedPorts) Execute(runtime connector.Runtime) error {
	const sysctlFile = "/etc/sysctl.conf"
	const reservedPortsKey = "net.ipv4.ip_local_reserved_ports"
	const expectedValue = "30000-32767,46800-50000"

	content, err := os.ReadFile(sysctlFile)
	if err != nil {
		return fmt.Errorf("failed to read sysctl.conf: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var foundKey bool
	var needUpdate bool
	var updatedLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, reservedPortsKey) {
			foundKey = true
			parts := strings.SplitN(trimmedLine, "=", 2)
			if len(parts) == 2 {
				currentValue := strings.TrimSpace(parts[1])
				if currentValue != expectedValue {
					logger.Infof("updating %s from %s to %s", reservedPortsKey, currentValue, expectedValue)
					updatedLines = append(updatedLines, fmt.Sprintf("%s=%s", reservedPortsKey, expectedValue))
					needUpdate = true
				} else {
					updatedLines = append(updatedLines, line)
				}
			} else {
				updatedLines = append(updatedLines, line)
			}
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	if !foundKey {
		logger.Infof("key %s not found in sysctl.conf, adding it", reservedPortsKey)
		updatedLines = append(updatedLines, fmt.Sprintf("%s=%s", reservedPortsKey, expectedValue))
		needUpdate = true
	}

	if needUpdate {
		updatedContent := strings.Join(updatedLines, "\n")
		if err := os.WriteFile(sysctlFile, []byte(updatedContent), 0644); err != nil {
			return fmt.Errorf("failed to write updated sysctl.conf: %v", err)
		}

		if _, err := runtime.GetRunner().SudoCmd("sysctl -p", false, false); err != nil {
			return fmt.Errorf("failed to reload sysctl: %v", err)
		}
		logger.Infof("updated and reloaded sysctl configuration")
	} else {
		logger.Debugf("%s already has the expected value: %s", reservedPortsKey, expectedValue)
	}

	return nil
}
