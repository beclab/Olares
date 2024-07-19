#!/usr/bin/env bash



set -o pipefail

VERSION="#__VERSION__"
if [ "x${VERSION}" = "x" ]; then
  echo "Unable to get latest Install-Wizard version. Set VERSION env var and re-run. For example: export VERSION=1.0.0"
  echo ""
  exit
fi

# check os type and arch and os vesion
os_type=$(uname -s)
os_arch=$(uname -m)
os_verion=$(lsb_release -d 2>&1 | awk -F'\t' '{print $2}')

case "$os_arch" in 
    arm64) ARCH=arm64; ;; 
    x86_64) ARCH=amd64; ;; 
    armv7l) ARCH=arm; ;; 
    aarch64) ARCH=arm64; ;; 
    ppc64le) ARCH=ppc64le; ;; 
    s390x) ARCH=s390x; ;; 
    *) echo "unsupported arch, exit ..."; 
    exit -1; ;; 
esac 


DOWNLOAD_URL="https://dc3p1870nn3cj.cloudfront.net/install-wizard-v${VERSION}.tar.gz"

if [ x"${ARCH}" == x"arm64" ]; then
  DOWNLOAD_URL="https://dc3p1870nn3cj.cloudfront.net/install-wizard-v${VERSION}-arm64.tar.gz"
fi

echo ""
echo " Downloading Install-Wizard ${VERSION} from ${DOWNLOAD_URL} ... " 
echo ""

foldername="install-wizard-v${VERSION}"
filename="install-wizard-v${VERSION}.tar.gz"

if [ ! -f ${filename} ]; then
  tmpname="install-wizard-v${VERSION}.bak.tar.gz"
  curl -Lo ${tmpname} ${DOWNLOAD_URL}

  if [ $? -ne 0 ] || [ ! -f ${tmpname} ]; then
    echo ""
    echo "Failed to download Install-Wizard ${VERSION} !"
    echo ""
    echo "Please verify the version you are trying to download."
    echo ""
    exit
  fi
  mv ${tmpname} ${filename}
fi

echo ""
echo "Install-Wizard ${VERSION} Download Complete!"
echo ""

if command -v tar &>/dev/null; then
    rm -rf ${foldername} && mkdir -p ${foldername} && cd ${foldername} && tar -xzf "../${filename}"

    if [ $? -eq 0 ]; then
        if [[ x"$os_type" == x"Darwin" ]]; then
          bash  ./uninstall_macos.sh
          touch .installed
          bash ./install_macos.sh
        else
          bash  ./uninstall_cmd.sh
          touch .installed
          bash ./install_cmd.sh
        fi

        exit 0
    fi
else
    echo "Try to unpack the ${filename} failed."
    echo "tar: command not found, please unpack the ${filename} manually."
    exit 1
fi
