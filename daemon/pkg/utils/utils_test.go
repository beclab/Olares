package utils

import (
	"context"
	"testing"

	"bytetrade.io/web3os/terminusd/pkg/nets"
)

func TestFindProc(t *testing.T) {

	procs, err := FindProcByName(context.Background(), "login")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(procs)
}

func TestTerminusInit(t *testing.T) {
	client, err := GetDynamicClient()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	init, _, err := IsTerminusInitialized(context.Background(), client)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log("result: ", init)

}

func TestIPs(t *testing.T) {
	ips, err := nets.GetInternalIpv4Addr()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(ips)
}

func TestMasterNodeIP(t *testing.T) {
	add, err := MasterNodeIp(false)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(add)
}

func TestDetectedUse(t *testing.T) {
	DetectdUsbDevices(context.Background())
}

func TestGetGpu(t *testing.T) {
	s, err := GetGpuInfo()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(*s)
}

func TestFindCommand(t *testing.T) {
	cmd, err := findCommand(context.Background(), "ls")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(cmd)
}
