/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package util

import (
	"bytes"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"sort"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type Data map[string]interface{}

// Render text template with given `variables` Render-context
func Render(tmpl *template.Template, variables any) (string, error) {

	var buf strings.Builder

	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", errors.Wrap(err, "Failed to render template")
	}
	return buf.String(), nil
}

// Home returns the home directory for the executing user.
func Home() (string, error) {
	u, err := user.Current()
	if nil == err {
		return u.HomeDir, nil
	}

	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	return homeUnix()
}

func homeUnix() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

func GetArgs(argsMap map[string]string, args []string) ([]string, map[string]string) {
	targetMap := make(map[string]string, len(argsMap))
	for k, v := range argsMap {
		targetMap[k] = v
	}
	targetSlice := make([]string, len(args))
	copy(targetSlice, args)

	for _, arg := range targetSlice {
		splitArg := strings.SplitN(arg, "=", 2)
		if len(splitArg) < 2 {
			continue
		}
		targetMap[splitArg[0]] = splitArg[1]
	}

	for arg, value := range targetMap {
		cmd := fmt.Sprintf("%s=%s", arg, value)
		targetSlice = append(targetSlice, cmd)
	}
	sort.Strings(targetSlice)
	return targetSlice, targetMap
}

// Round returns the result of rounding 'val' according to the specified 'precision' precision (the number of digits after the decimal point)。
// and precision can be negative number or zero
func Round(val float64, precision int) float64 {
	p := math.Pow10(precision)
	return math.Floor(val*p+0.5) / p
}

// ArchAlias returns the alias of cpu's architecture.
// amd64: x86_64
// arm64: aarch64
func ArchAlias(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return ""
	}
}

func FormatSed(darwin bool) string {
	var res = "sed -i "
	if darwin {
		return fmt.Sprintf("%s '' ", res)
	}

	return res
}

func FormatBytes(bytes int64) string {
	const (
		KB = 1 << 10 // 1024
		MB = 1 << 20 // 1024 * 1024
		GB = 1 << 30 // 1024 * 1024 * 1024
		TB = 1 << 40 // 1024 * 1024 * 1024 * 1024
	)

	var result string
	switch {
	case bytes >= TB:
		result = fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		result = fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		result = fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		result = fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		result = fmt.Sprintf("%d Byte", bytes)
	}

	return result
}

func RemoveHTTPPrefix(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return url
}

func IsOnAWSEC2() bool {
	vmUUIDFile := "/sys/hypervisor/uuid"
	if IsExist("/sys/hypervisor/uuid") {
		if content, err := os.ReadFile(vmUUIDFile); err == nil {
			logger.Debugf("read content of aws vm uuid file: %s", string(content))
			if strings.EqualFold(string(content)[:3], "ec2") {
				return true
			}
			return false
		} else {
			logger.Debugf("failed to read aws vm uuid file: %v", err)
		}
	} else {
		logger.Debug("aws vm uuid file does not exits")
	}
	productUUIDFile := "/sys/devices/virtual/dmi/id/product_uuid"
	if IsExist(productUUIDFile) {
		if content, err := os.ReadFile(productUUIDFile); err == nil {
			logger.Debugf("read content of aws product uuid file: %s", string(content))
			if strings.EqualFold(string(content)[:3], "ec2") {
				return true
			}
			return false
		} else {
			logger.Debugf("failed to read product uuid file: %v", err)
		}
	} else {
		logger.Debug("aws product uuid file does not exits")
	}
	resp, err := http.Get("http://169.254.169.254/latest/dynamic/instance-identity/document")
	if err != nil {
		logger.Debugf("failed to get aws instance identity document: %v", err)
		return false
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Debugf("failed to read aws instance identity document: %v", err)
		return false
	}
	logger.Debugf("got aws instance identity document: %s", string(body))
	if strings.Contains(string(body), "instanceID") {
		return true
	}
	return false
}
