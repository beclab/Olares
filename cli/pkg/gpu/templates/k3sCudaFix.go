package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var (
	K3sCudaFixValues = template.Must(template.New("cuda_lib_fix.sh").Parse(
		dedent.Dedent(`#!/bin/bash
sh_c="sh -c"
real_driver=$($sh_c "find /usr/lib/wsl/drivers/ -name libcuda.so.1.1|head -1")
if [[ x"$real_driver" != x"" ]]; then
    $sh_c "ln -s /usr/lib/wsl/lib/libcuda* /usr/lib/x86_64-linux-gnu/"
    $sh_c "rm -f /usr/lib/x86_64-linux-gnu/libcuda.so"
    $sh_c "rm -f /usr/lib/x86_64-linux-gnu/libcuda.so.1"
    $sh_c "rm -f /usr/lib/x86_64-linux-gnu/libcuda.so.1.1"
    $sh_c "cp -f $real_driver /usr/lib/wsl/lib/libcuda.so"
    $sh_c "cp -f $real_driver /usr/lib/wsl/lib/libcuda.so.1"
    $sh_c "cp -f $real_driver /usr/lib/wsl/lib/libcuda.so.1.1"
    $sh_c "ln -s $real_driver /usr/lib/x86_64-linux-gnu/libcuda.so.1"
    $sh_c "ln -s $real_driver /usr/lib/x86_64-linux-gnu/libcuda.so.1.1"
    $sh_c "ln -s /usr/lib/x86_64-linux-gnu/libcuda.so.1 /usr/lib/x86_64-linux-gnu/libcuda.so"
fi`),
	))
)
