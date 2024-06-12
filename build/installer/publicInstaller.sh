#!/usr/bin/env bash



set -o pipefail

VERSION="#__VERSION__"
if [ "x${VERSION}" = "x" ]; then
  echo "Unable to get latest Install-Wizard version. Set VERSION env var and re-run. For example: export VERSION=1.0.0"
  echo ""
  exit
fi

DOWNLOAD_URL="https://dc3p1870nn3cj.cloudfront.net/install-wizard-v${VERSION}.tar.gz"

echo ""
echo " Downloading Install-Wizard ${VERSION} from ${DOWNLOAD_URL} ... " 
echo ""

filename="install-wizard-v${VERSION}.tar.gz"
curl -Lo ${filename} ${DOWNLOAD_URL}
if [ $? -ne 0 ] || [ ! -f ${filename} ]; then
  echo ""
  echo "Failed to download Install-Wizard ${VERSION} !"
  echo ""
  echo "Please verify the version you are trying to download."
  echo ""
  exit
fi

ret='0'
command -v tar >/dev/null 2>&1 || { ret='1'; }
if [ "$ret" -eq 0 ]; then
    mkdir -p install-wizard && cd install-wizard && tar -xzf "../${filename}"
else
    echo "Install-Wizard ${VERSION} Download Complete!"
    echo ""
    echo "Try to unpack the ${filename} failed."
    echo "tar: command not found, please unpack the ${filename} manually."
    exit
fi

echo ""
echo "Install-Wizard ${VERSION} Download Complete!"
echo ""


bash ./install_cmd.sh
