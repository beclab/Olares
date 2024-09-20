#!/usr/bin/env bash

set -o pipefail
set -e


if [[ x"$VERSION" == x"" ]]; then
  export VERSION="#__VERSION__"
  MD5SUM="#__MD5SUM__"
  # export VERSION="1.8.0-99995"
  # MD5SUM="4c0e0dd59ec5e334e89374af4f89d411"
fi

if [[ "x${VERSION}" == "x" || "x${VERSION:3}" == "xVERSION__" ]]; then
  echo "Unable to get latest Install-Wizard version. Set VERSION env var and re-run. For example: export VERSION=1.0.0"
  echo ""
  exit
fi

# check os type and arch and os vesion
os_type=$(uname -s)
os_arch=$(uname -m)
# os_verion=$(lsb_release -d 2>&1 | awk -F'\t' '{print $2}')

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


if [[ x"$os_type" != x"Darwin" ]]; then
  if command -v sudo > /dev/null; then
    SUDO=$(command -v sudo)
  else
    cat >&2 <<-'EOF'
Error: this installer needs the ability to run commands as root.
We are unable to find either "sudo" or "su" available to make this happen.
EOF
    exit -1
  fi
fi


if command -v tar >/dev/null; then
    if [[ x"$KUBE_TYPE" == x"" ]]; then
      export KUBE_TYPE="k3s"
    fi

    foldername="install-wizard-v${VERSION}"
    $SUDO rm -rf $HOME/.terminus/${foldername} && \
    mkdir -p $HOME/.terminus/${foldername}

    CLI_VERSION="0.1.24"
    CLI_FILE="terminus-cli-v${CLI_VERSION}_linux_${ARCH}.tar.gz"
    INSTALL_TERMINUS_CLI="/usr/local/bin/terminus-cli"
    if [[ x"$os_type" == x"Darwin" ]]; then
        INSTALL_TERMINUS_CLI="/usr/local/Cellar/terminus/terminus-cli"
        CLI_FILE="terminus-cli-v${CLI_VERSION}_darwin_${ARCH}.tar.gz"
    fi

    if [[ ! -f ${CLI_FILE} ]]; then
        CLI_URL="https://dc3p1870nn3cj.cloudfront.net/${CLI_FILE}"
        
        echo ""
        echo " Downloading Terminus Installer ${CLI_VERSION} from ${CLI_URL} ... " 
        echo ""

        curl -Lo ${CLI_FILE} ${CLI_URL}
    
    fi


    if [[ $? -eq 0 ]]; then
        echo ""
        echo "Terminus Installer ${CLI_VERSION} Download Complete!"
        echo ""

        if [[ "x${MD5SUM}" == "x" || "x${MD5SUM:3}" == "xMD5SUM__" ]]; then
          MD5SUM=$(curl -sSfL "https://dc3p1870nn3cj.cloudfront.net/install-wizard-v${VERSION}.md5sum.txt"|awk '{print $1}')
        fi

        if [[ x"$os_type" == x"Darwin" ]]; then
          if [ ! -f "/usr/local/Cellar" ]; then
            current_user=$(whoami)
            sh -c "sudo mkdir -p /usr/local/Cellar && sudo chown ${current_user}:staff /usr/local/Cellar"
          fi
          sh -c "tar -zxvf ${CLI_FILE} && chmod +x terminus-cli && \
          mkdir -p /usr/local/Cellar/terminus && \
          mv terminus-cli $INSTALL_TERMINUS_CLI && \
          sudo rm -rf /usr/local/bin/terminus-cli && \
          sudo ln -s $INSTALL_TERMINUS_CLI /usr/local/bin/terminus-cli"
          
          # TODO: download install-wizard

          sh -c "cd $HOME/.terminus/${foldername} && \
          $INSTALL_TERMINUS_CLI terminus download-wizard --version $VERSION --md5sum $MD5SUM --base-dir $HOME/.terminus"

          if [[ $? -ne 0 ]]; then
            exit -1
          fi

          cd $HOME/.terminus/${foldername} && \
          bash  ./install_macos.sh
        else
          $SUDO -E sh -c "tar -zxvf ${CLI_FILE} && chmod +x terminus-cli && \
          mv terminus-cli $INSTALL_TERMINUS_CLI"

          $SUDO -E sh -c "cd $HOME/.terminus/${foldername} && \
          $INSTALL_TERMINUS_CLI terminus download-wizard --version $VERSION --md5sum $MD5SUM --base-dir $HOME/.terminus"

          if [[ $? -ne 0 ]]; then
            exit -1
          fi

          cd $HOME/.terminus/${foldername} && \
          $SUDO -E sh -c "bash  ./install_cmd.sh"
        fi

        exit 0
    fi
else
    echo "Try to unpack the ${filename} failed."
    echo "tar: command not found, please unpack the ${filename} manually."
    exit 1
fi
