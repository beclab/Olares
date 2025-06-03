package utils

import (
	"context"
	"os"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v4/process"
	"k8s.io/klog/v2"
)

func FindProcByName(ctx context.Context, name string) ([]*process.Process, error) {

	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		klog.Error("list process error, ", err)
		return nil, err
	}

	if name == "" {
		return procs, nil
	}

	name = strings.ToLower(name)
	var filteredProcs []*process.Process
	for _, process := range procs {
		procName, err := process.NameWithContext(ctx)
		if err != nil {
			klog.Info("Unknown process, ", err, ", ", process.Pid)
			continue
		}

		if name == procName {
			filteredProcs = append(filteredProcs, process)
		}
	}

	if len(filteredProcs) > 0 {
		return filteredProcs, nil
	}
	return nil, os.ErrNotExist
}

func ProcessExists(pid int) (bool, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}

	if err = p.Signal(syscall.Signal(0)); err != nil {
		klog.Info(err)
		return false, nil
	}

	return true, nil
}
