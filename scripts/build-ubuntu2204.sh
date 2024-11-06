#!/usr/bin/env bash

set -o pipefail
set -xe

curl -Lo Ubuntu2204.appx https://wslstorestorage.blob.core.windows.net/wslblob/Ubuntu2204-221101.AppxBundle
ubuntu2204=$(md5sum Ubuntu2204.appx|awk '{print $1}')

aws s3 cp Ubuntu2204.appx s3://terminus-os-install/${ubuntu2204} --acl=public-read
