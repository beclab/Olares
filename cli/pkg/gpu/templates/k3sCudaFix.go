package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var (
	K3sCudaFixValues = template.Must(template.New("cuda_lib_fix.sh").Parse(
		dedent.Dedent(`#!/bin/bash
sh_c="sh -c"
real_driver=""
real_nvml=""

# Try to find the real driver path via strace
real_driver_path=$($sh_c "strace -qq -e trace=openat /usr/lib/wsl/lib/nvidia-smi 2>&1|grep '/usr/lib/wsl/drivers'|grep libnvidia-ml.so.1|awk '{print \$2}'|sed 's/[\",]//g'|sed 's/libnvidia-ml.so.1//g'")
if [[ x"$real_driver_path" != x"" ]]; then
    real_driver="${real_driver_path}libcuda.so.1.1"
    real_nvml="${real_driver_path}libnvidia-ml.so.1"
else
    driver_path=$($sh_c "strace -qq -e trace=openat /usr/lib/wsl/lib/nvidia-smi 2>&1|grep '/usr/lib/wsl/'|grep libnvidia-ml.so.1")
    if [[ x"$driver_path" != x"" ]]; then
        echo "already fixed cuda libs, exit now."
        exit 0
    fi
fi

if [[ x"$real_driver" == x"" ]]; then
    real_driver=$($sh_c "find /usr/lib/wsl/drivers/ -name libcuda.so.1.1|head -1")
    real_nvml=$($sh_c "find /usr/lib/wsl/drivers/ -name libnvidia-ml.so.1|head -1")
fi

if [[ x"$real_driver" != x"" ]]; then
    $sh_c "rm -f /usr/lib/x86_64-linux-gnu/libcuda.so"
    $sh_c "rm -f /usr/lib/x86_64-linux-gnu/libcuda.so.1"
    $sh_c "rm -f /usr/lib/x86_64-linux-gnu/libcuda.so.1.1"
    $sh_c "rm -f /usr/lib/x86_64-linux-gnu/libnvidia-ml.so.1"
    $sh_c "cp -f $real_driver /usr/lib/wsl/lib/libcuda.so"
    $sh_c "cp -f $real_driver /usr/lib/wsl/lib/libcuda.so.1"
    $sh_c "cp -f $real_driver /usr/lib/wsl/lib/libcuda.so.1.1"
    $sh_c "cp -f $real_nvml /usr/lib/wsl/lib/libnvidia-ml.so.1"
    $sh_c "cp -f $real_driver /usr/lib/x86_64-linux-gnu/"
    $sh_c "cp -f $real_nvml /usr/lib/x86_64-linux-gnu/"
    $sh_c "ln -s /usr/lib/x86_64-linux-gnu/libcuda.so.1.1 /usr/lib/x86_64-linux-gnu/libcuda.so.1"
    $sh_c "ln -s /usr/lib/x86_64-linux-gnu/libcuda.so.1 /usr/lib/x86_64-linux-gnu/libcuda.so"
fi`),
	))
)
