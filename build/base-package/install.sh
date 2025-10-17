#!/usr/bin/env bash

set -o pipefail
set -e

function command_exists() {
	  command -v "$@" > /dev/null 2>&1
}

if [[ x"$VERSION" == x"" ]]; then
    if [[ "$LOCAL_RELEASE" == "1" ]]; then
        ts=$(date +%Y%m%d%H%M%S)
        export VERSION="1.12.2-$ts"
        echo "will build and use a local release of Olares with version: $VERSION"
        echo ""
    else
        export VERSION="#__VERSION__"
    fi
fi

if [[ "x${VERSION}" == "x" || "x${VERSION:3}" == "xVERSION__" ]]; then
    echo "error: Olares version is unspecified, please set the VERSION env var and rerun this script."
    echo "for example: VERSION=1.12.2-20241124 bash $0"
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

cdn_url=${OLARES_CDN_SERVICE}
if [ -z ${cdn_url} ]; then
    cdn_url="https://cdn.olares.com"
fi

RELEASE_ID="#__RELEASE_ID__"
if [[ $RELEASE_ID == "" || "${RELEASE_ID:3}" == "RELEASE_ID__" ]]; then
  RELEASE_ID_SUFFIX=""
else
  RELEASE_ID_SUFFIX=".$RELEASE_ID"
fi
CLI_FILE="olares-cli-v${VERSION}_linux_${ARCH}${RELEASE_ID_SUFFIX}.tar.gz"
if [[ x"$os_type" == x"Darwin" ]]; then
    CLI_FILE="olares-cli-v${VERSION}_darwin_${ARCH}${RELEASE_ID_SUFFIX}.tar.gz"
fi

if [[ "$LOCAL_RELEASE" == "1" ]]; then
    if ! command_exists olares-cli ; then
        echo "error: LOCAL_RELEASE specified but olares-cli not found"
        exit 1
    fi
    INSTALL_OLARES_CLI=$(which olares-cli)
else
    if command_exists olares-cli && [[ "$(olares-cli -v | awk '{print $3}')" == "$VERSION" ]]; then
        INSTALL_OLARES_CLI=$(which olares-cli)
        echo "olares-cli already installed and is the expected version"
        echo ""
    else
        if [[ ! -f ${CLI_FILE} ]]; then
            CLI_URL="${cdn_url}/${CLI_FILE}"

            echo "downloading Olares installer from ${CLI_URL} ..."
            echo ""

            curl -Lo ${CLI_FILE} ${CLI_URL}

            if [[ $? -ne 0 ]]; then
                echo "error: failed to download Olares installer"
                exit 1
            else
                echo "Olares installer ${VERSION} download complete!"
                echo ""
            fi
        fi
        INSTALL_OLARES_CLI="/usr/local/bin/olares-cli"
        echo "unpacking Olares installer to $INSTALL_OLARES_CLI..."
        echo ""
        tar -zxf ${CLI_FILE} olares-cli && chmod +x olares-cli
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
    fi
fi

PARAMS="--version $VERSION --base-dir $BASE_DIR"
KUBE_PARAM="--kube $KUBE_TYPE"
CDN="--cdn-service ${cdn_url}"

if [[ -f $BASE_DIR/.prepared ]]; then
    echo "file $BASE_DIR/.prepared detected, skip preparing phase"
    echo ""
else
    if [[ "$LOCAL_RELEASE" == "1" ]]; then
        if [[ -d $BASE_DIR/versions/v$VERSION  ]]; then
            echo "local release already exists, skip building"
            echo ""
        else
            echo "building local release ..."
            $sh_c "$INSTALL_OLARES_CLI release $PARAMS $CDN"
            if [[ $? -ne 0 ]]; then
                echo "error: failed to build local release"
                exit 1
            fi
        fi
    else
        echo "running system prechecks ..."
        echo ""
        $sh_c "$INSTALL_OLARES_CLI precheck $PARAMS"
        if [[ $? -ne 0 ]]; then
            exit 1
        fi
        echo "downloading installation wizard..."
        echo ""
        if [[ ! -z "$RELEASE_ID_SUFFIX" ]]; then
            DOWNLOAD_WIZARD_RELEASE_ID_PARAM="--release-id $RELEASE_ID"
        fi
        $sh_c "$INSTALL_OLARES_CLI download wizard $PARAMS $KUBE_PARAM $CDN $DOWNLOAD_WIZARD_RELEASE_ID_PARAM"
        if [[ $? -ne 0 ]]; then
            echo "error: failed to download installation wizard"
            exit 1
        fi
    fi

    echo "downloading installation packages..."
    echo ""
    $sh_c "$INSTALL_OLARES_CLI download component $PARAMS $KUBE_PARAM $CDN"
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
    $sh_c "$INSTALL_OLARES_CLI prepare $PARAMS $KUBE_PARAM $extra"
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

if [[ "$JUICEFS" == "1" ]]; then
    echo "JuiceFS is enabled"
    fsflag="--with-juicefs=true"
    if [[ "$STORAGE" == "" ]]; then
        echo "installing MinIO ..."
    else
        echo "checking storage config ..."
    fi
    $sh_c "$INSTALL_OLARES_CLI install storage $PARAMS"
    if [[ $? -ne 0 ]]; then
      exit 1
    fi
fi

if [[ -n "$SWAPPINESS" ]]; then
    swapflag="$swapflag --swappiness $SWAPPINESS"
fi
if [[ "$ENABLE_POD_SWAP" == "1" ]]; then
    swapflag="$swapflag --enable-pod-swap"
fi
if [[ "$ENABLE_ZRAM" == "1" ]]; then
    swapflag="$swapflag --enable-zram"
fi
if [[ -n "$ZRAM_SIZE" ]]; then
    swapflag="$swapflag --zram-size $ZRAM_SIZE"
fi
if [[ -n "$ZRAM_SWAP_PRIORITY" ]]; then
    swapflag="$swapflag --zram-swap-priority $ZRAM_SWAP_PRIORITY"
fi
echo "installing Olares..."
echo ""
$sh_c "$INSTALL_OLARES_CLI install $PARAMS $KUBE_PARAM $fsflag $swapflag"

if [[ $? -ne 0 ]]; then
    echo "error: failed to install Olares"
    exit 1
fi
