#!/usr/bin/env bash

set -o pipefail
set -xe

apt install -y bzip2

# darwin
curl -Lo restic_0.17.3_darwin_amd64.bz2 https://github.com/restic/restic/releases/download/v0.17.3/restic_0.17.3_darwin_amd64.bz2
bzip2 -d restic_0.17.3_darwin_amd64.bz2
hash_restic_0_17_3_darwin_amd64=$(md5sum restic_0.17.3_darwin_amd64 |awk '{print $1}')
aws s3 cp restic_0.17.3_darwin_amd64 s3://terminus-os-install/${hash_restic_0_17_3_darwin_amd64} --acl=public-read

curl -Lo restic_0.17.3_darwin_arm64.bz2 https://github.com/restic/restic/releases/download/v0.17.3/restic_0.17.3_darwin_arm64.bz2
bzip2 -d restic_0.17.3_darwin_arm64.bz2
hash_restic_0_17_3_darwin_arm64=$(md5sum restic_0.17.3_darwin_arm64 |awk '{print $1}')
aws s3 cp restic_0.17.3_darwin_arm64 s3://terminus-os-install/${hash_restic_0_17_3_darwin_arm64} --acl=public-read

# linux
curl -Lo restic_0.17.3_linux_amd64.bz2 https://github.com/restic/restic/releases/download/v0.17.3/restic_0.17.3_linux_amd64.bz2
bzip2 -d restic_0.17.3_linux_amd64.bz2
hash_restic_0_17_3_linux_amd64=$(md5sum restic_0.17.3_linux_amd64 |awk '{print $1}')
aws s3 cp restic_0.17.3_linux_amd64 s3://terminus-os-install/${hash_restic_0_17_3_linux_amd64} --acl=public-read

curl -Lo restic_0.17.3_linux_arm64.bz2 https://github.com/restic/restic/releases/download/v0.17.3/restic_0.17.3_linux_arm64.bz2
bzip2 -d restic_0.17.3_linux_arm64.bz2
hash_restic_0_17_3_linux_arm64=$(md5sum restic_0.17.3_linux_arm64 |awk '{print $1}')
aws s3 cp restic_0.17.3_linux_arm64 s3://terminus-os-install/${hash_restic_0_17_3_linux_arm64} --acl=public-read

