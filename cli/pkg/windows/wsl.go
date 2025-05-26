package windows

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

type Wsl struct {
	version string
}

func (w *Wsl) GetVersion() error {
	var checkversion = &utils.DefaultCommandExecutor{
		Commands: []string{"-v"},
	}
	versionOutput, err := checkversion.RunCmd("wsl", utils.DEFAULT)
	if err != nil {
		return err
	}

	if version := w.tidyVersions(versionOutput); version != "" {
		w.version = version
		return nil
	}

	return fmt.Errorf("wsl version get invalid, output: %s", versionOutput)
}

func (w *Wsl) IsInstalled() bool {
	return !strings.EqualFold(w.version, "")
}

func (w *Wsl) CompareTo(version2 string) int {
	parts1 := strings.Split(w.version, ".")
	parts2 := strings.Split(version2, ".")

	maxLength := len(parts1)
	if len(parts2) > maxLength {
		maxLength = len(parts2)
	}

	for i := 0; i < maxLength; i++ {
		var num1, num2 int
		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}
	return 0
}

func (w *Wsl) PrintVersion() {
	if w.IsInstalled() {
		logger.Infof("WSL Version: %s", w.version)
		return
	}
	logger.Info("WSL not found")
}

func (w *Wsl) Install(packagePath string) (string, error) {
	var installCmd = &utils.DefaultCommandExecutor{
		Commands: []string{"/i", packagePath, "/quiet", "/norestart"},
	}

	installoutput, err := installCmd.RunCmd("msiexec", utils.UTF8)
	if err != nil {
		return installoutput, errors.Wrap(errors.WithStack(err), fmt.Sprintf("Install WSL failed, message: %s", installoutput))
	}

	return installoutput, nil
}

func (w *Wsl) tidyVersions(content string) string {
	var result = ""
	content = strings.ReplaceAll(content, "\n", "\r")
	lines := strings.Split(content, "\r")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "WSL") {
			continue
		}
		if strings.Contains(line, "WSLg") {
			continue
		}
		if strings.Contains(line, "：") || strings.Contains(line, ":") {
			tmp := strings.ReplaceAll(line, "：", ":")
			v := strings.Split(tmp, ":")
			if v == nil || len(v) != 2 {
				break
			}
			result = v[1]
			result = strings.TrimSpace(result)
			break
		}
	}
	if result == "" || !w.isVersionValid(result) {
		return ""
	}
	return result
}

func (w *Wsl) isVersionValid(version string) bool {
	versionRegex := `^\d+(\.\d+)*$`
	re := regexp.MustCompile(versionRegex)
	if re.MatchString(version) {
		return true
	}
	return false
}
