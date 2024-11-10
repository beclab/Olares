#!/usr/bin/env bash

set -o pipefail
set -e

function command_exists() {
	  command -v "$@" > /dev/null 2>&1
}

if [[ x"$VERSION" == x"" ]]; then
    export VERSION="#__VERSION__"
fi

if [[ "x${VERSION}" == "x" || "x${VERSION:3}" == "xVERSION__" ]]; then
    echo "error: unable to get the wanted Olares version, please set the VERSION env var and rerun this script."
    echo "for example: VERSION=1.0.0 bash ./install.sh"
    exit 1
fi

# check os type and arch
os_type=$(uname -s)
os_arch=$(uname -m)

case "$os_arch" in 
    arm64) ARCH=arm64; ;; 
    x86_64) ARCH=amd64; ;; 
    armv7l) ARCH=arm; ;; 
    aarch64) ARCH=arm64; ;; 
    ppc64le) ARCH=ppc64le; ;; 
    s390x) ARCH=s390x; ;; 
    *) echo "error: unsupported arch \"$os_arch\"";
    exit 1; ;;
esac 

# set shell execute command
user="$(id -un 2>/dev/null || true)"
sh_c='sh -c'
if [ "$user" != 'root' ]; then
    if command_exists sudo && command_exists su; then
        if [[ "$os_type" != "Darwin" ]]; then
            sh_c='sudo -E sh -c'
        fi
    else
        echo "error: this installer needs the ability to run as root, but the command \"sudo\" and  \"su\" can not be found"
        exit 1
    fi
fi

if ! command_exists tar; then
    echo "error: the \"tar\" command is needed by installer to unpack installation files, but can not be found"
    exit 1
fi

if [[ x"$KUBE_TYPE" == x"" ]]; then
    echo "the KUBE_TYPE env var is not set, defaulting to \"k3s\""
    echo ""
    export KUBE_TYPE="k3s"
fi

BASE_DIR="$HOME/.olares"
if [ ! -d $BASE_DIR ]; then
    mkdir -p $BASE_DIR
fi

CLI_VERSION="0.1.49"
CLI_FILE="olares-cli-v${CLI_VERSION}_linux_${ARCH}.tar.gz"
if [[ x"$os_type" == x"Darwin" ]]; then
    CLI_FILE="olares-cli-v${CLI_VERSION}_darwin_${ARCH}.tar.gz"
fi

if [[ ! -f ${CLI_FILE} ]]; then
    CLI_URL="https://dc3p1870nn3cj.cloudfront.net/${CLI_FILE}"

    echo "downloading Olares installer from ${CLI_URL} ..."
    echo ""

    curl -Lo ${CLI_FILE} ${CLI_URL}

    if [[ $? -ne 0 ]]; then
        echo "error: failed to download Olares installer"
        exit 1
    else
        echo "Olares installer ${CLI_VERSION} download complete!"
        echo ""
    fi
fi

INSTALL_OLARES_CLI="/usr/local/bin/olares-cli"
echo "unpacking Olares installer to $INSTALL_OLARES_CLI..."
echo ""
tar -zxf ${CLI_FILE} && chmod +x olares-cli
if [[ x"$os_type" == x"Darwin" ]]; then
    if [ ! -f "/usr/local/Cellar/olares" ]; then
        current_user=$(whoami)
        $sh_c "sudo mkdir -p /usr/local/Cellar/olares && sudo chown ${current_user}:staff /usr/local/Cellar/olares"
    fi
    $sh_c "mv olares-cli /usr/local/Cellar/olares/olares-cli && \
           sudo rm -rf /usr/local/bin/olares-cli && \
           sudo ln -s /usr/local/Cellar/olares/olares-cli $INSTALL_OLARES_CLI"
else
    $sh_c "mv olares-cli $INSTALL_OLARES_CLI"
fi

if [[ $? -ne 0 ]]; then
    echo "error: failed to unpack Olares installer"
    exit 1
fi

PARAMS="--version $VERSION --base-dir $BASE_DIR --kube $KUBE_TYPE"

if [ -f $BASE_DIR/.prepared ]; then
    echo "file $BASE_DIR/.prepared detected, skip preparing phase"
    echo ""
else
    echo "downloading installation wizard..."
    echo ""
    $sh_c "$INSTALL_OLARES_CLI olares download wizard $PARAMS"
    if [[ $? -ne 0 ]]; then
        echo "error: failed to download installation wizard"
        exit 1
    fi

    echo "downloading installation packages..."
    echo ""
    $sh_c "$INSTALL_OLARES_CLI olares download component $PARAMS"
    if [[ $? -ne 0 ]]; then
        echo "error: failed to download installation packages"
        exit 1
    fi

    echo "preparing installation environment..."
    echo ""
    # env 'REGISTRY_MIRRORS' is a docker image cache mirrors, separated by commas
    if [ x"$REGISTRY_MIRRORS" != x"" ]; then
        extra="--registry-mirrors $REGISTRY_MIRRORS"
    fi
    $sh_c "$INSTALL_OLARES_CLI olares prepare $PARAMS $extra"
    if [[ $? -ne 0 ]]; then
        echo "error: failed to prepare installation environment"
        exit 1
    fi
fi

if [ -f $BASE_DIR/.installed ]; then
    echo "file $BASE_DIR/.installed detected, skip installing"
    echo "if it is left by an unclean uninstallation, please manually remove it and invoke the installer again"
    exit 0
fi
if [ "$PREINSTALL" == "1" ]; then
    echo "Pre Install mode is specified by the \"PREINSTALL\" env var, skip installing"
    exit 0
fi
echo "installing Olares..."
echo ""
$sh_c "$INSTALL_OLARES_CLI olares install $PARAMS"

if [[ $? -ne 0 ]]; then
    echo "error: failed to install Olares"
    exit 1
fi
