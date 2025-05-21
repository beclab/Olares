#!/usr/bin/env bash

set -o pipefail
set -xe

curl -Lo wsl.2.3.26.0.amd64.msi https://github.com/microsoft/WSL/releases/download/2.3.26/wsl.2.3.26.0.x64.msi
wsl_2_3_26=$(md5sum wsl.2.3.26.0.amd64.msi|awk '{print $1}')

aws s3 cp wsl.2.3.26.0.amd64.msi s3://terminus-os-install/${wsl_2_3_26} --acl=public-read

curl -Lo wsl.2.3.26.0.arm64.msi https://github.com/microsoft/WSL/releases/download/2.3.26/wsl.2.3.26.0.arm64.msi
wsl_2_3_26_arm64=$(md5sum wsl.2.3.26.0.arm64.msi|awk '{print $1}')

aws s3 cp wsl.2.3.26.0.arm64.msi s3://terminus-os-install/arm64/${wsl_2_3_26_arm64} --acl=public-read