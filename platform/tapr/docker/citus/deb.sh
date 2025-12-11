#!/bin/bash

arch=$1

set -ex

# if [[ "$arch" == "arm64" ]]; then 
    curl -fsSL https://repo.pigsty.cc/key | gpg --dearmor -o /etc/apt/keyrings/pigsty.gpg

    # Get Debian distribution codename (distro_codename=jammy, focal, bullseye, bookworm), and write the corresponding upstream repository address to the APT List file
    distro_codename=$(lsb_release -cs)
    tee /etc/apt/sources.list.d/pigsty-io.list > /dev/null <<EOF
deb [signed-by=/etc/apt/keyrings/pigsty.gpg] https://repo.pigsty.cc/apt/infra generic main
EOF

# deb [signed-by=/etc/apt/keyrings/pigsty.gpg] https://repo.pigsty.cc/apt/pgsql ${distro_codename} main
    # Refresh APT repository cache
    apt update 
# else 
#     curl -s https://packagecloud.io/install/repositories/citusdata/community/script.deb.sh 
# fi