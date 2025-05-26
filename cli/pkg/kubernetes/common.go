package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

func CheckKubeExists() (string, string, bool) {
	var kubectl, err = util.GetCommand(common.CommandKubectl)
	if err != nil || kubectl == "" {
		return "", "", false
	}

	var ver string
	var ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var cmd = fmt.Sprintf("%s get nodes -o jsonpath='{.items[0].status.nodeInfo.kubeletVersion}'", kubectl)
	stdout, _, err := util.ExecWithContext(ctx, cmd, false, false)

	if err != nil || stdout == "" {
		return "", "", false
	}

	if strings.Contains(stdout, "k3s") {
		if strings.Contains(stdout, "-") {
			stdout = strings.ReplaceAll(stdout, "-", "+")
		}

		var v1 = strings.Split(stdout, "+")
		if len(v1) != 2 {
			return ver, "k3s", true
		}
		ver = fmt.Sprintf("%s-k3s", v1[0])
		return ver, "k3s", true
	} else {
		ver = stdout
		return ver, "k8s", true
	}
}
