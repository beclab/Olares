package utils

import (
	"fmt"
	"path/filepath"
	"testing"

	"bytetrade.io/web3os/installer/pkg/common"
)

func TestA(t *testing.T) {
	var a = "/home/ubuntu/.terminus/versions/v1.8.0-20240928/wizard/config/apps/argo"
	var b = filepath.Base(a)
	fmt.Println("---b---", b)
}

func TestExecNvidiaSmi(t *testing.T) {

	runtime := common.LocalRuntime{}

	info, installed, err := ExecNvidiaSmi(&runtime)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(installed)
	t.Log(info)
}
