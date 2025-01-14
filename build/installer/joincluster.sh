#!/usr/bin/env bash

set -o pipefail
set -e

function command_exists() {
	  command -v "$@" > /dev/null 2>&1
}

function read_tty() {
    echo -n $1
    read $2 < /dev/tty
}

function confirm() {
    if [[ "$QUIET" == "1" ]]; then
        return 0
    fi
    answer=""
    while :; do
        read_tty "Do you confirm to continue? (y/n): " answer
        if [[ "$answer" != "y" && "$answer" != "n" ]]; then
            echo "Please input the letter y or n"
            continue
        fi
        if [[ "$answer" == "y" ]]; then
            return 0
        fi
        if [[ "$answer" == "n" ]]; then
            exit 0
        fi
    done
}

function validate_ip() {
    if [[ ! "$1" ]]; then
        echo "invalid IP: empty address"
        return 1
    elif [[ ! $1 =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "invalid IP: illegal format"
        return 1
    elif [[ $1 =~ ^127 ]]; then
        echo "invalid IP: loopback address"
        return 1
    else
        return 0
    fi
}

MASTER_SSH_OPTIONS=""

function add_master_host_ssh_options() {
    MASTER_SSH_OPTIONS="$MASTER_SSH_OPTIONS --$1 $2"
}

function set_master_host_ssh_options() {
    master_host="$MASTER_HOST"
    if [[ ! "$master_host" ]]; then
        read_tty "Enter the master node's IP: " master_host
    fi

    while :; do
        if ! validate_ip "$master_host"; then
            read_tty "Enter the master node's IP: " master_host
        else
            break
        fi
    done

    add_master_host_ssh_options master-host "$master_host"

    if [[ "$MASTER_NODE_NAME" ]]; then
        add_master_host_ssh_options master-node-name "$MASTER_NODE_NAME"
    fi

    if [[ "$MASTER_SSH_USER" ]]; then
        add_master_host_ssh_options master-ssh-user "$MASTER_SSH_USER"
    else
        echo "the environment variable \$MASTER_SSH_USER is not set"
        echo "the default remote user \"root\" on the master node will be used to authenticate"
        echo "if this is unexpected, please set it explicitly"
        confirm
    fi

    if [[ "$MASTER_SSH_PASSWORD" ]]; then
        add_master_host_ssh_options master-ssh-password "$MASTER_SSH_PASSWORD"
    fi

    if [[ "$MASTER_SSH_PRIVATE_KEY_PATH" ]]; then
        add_master_host_ssh_options master-ssh-private-key-path "$MASTER_SSH_PRIVATE_KEY_PATH"
    elif [[ ! "$MASTER_SSH_PASSWORD" ]]; then
        echo "the environment variable \$MASTER_SSH_PRIVATE_KEY_PATH is not set"
        echo "the default key in the local path /root/.ssh/id_rsa will be used to authenticate to the master"
        echo "please make sure the key exists and the public key has already been added to the master node"
        echo "if this is unexpected, please set it explicitly"
        confirm
    fi

    if [[ "$MASTER_SSH_PORT" ]]; then
        add_master_host_ssh_options master-ssh-port "$MASTER_SSH_PORT"
    fi
}

function getmasterinfo() {
    $sh_c "$INSTALL_OLARES_CLI node masterinfo $MASTER_SSH_OPTIONS" | tee /proc/$$/fd/1
    if [[ $? -ne 0 ]]; then
        exit 1
    fi
    echo "" > /proc/$$/fd/1
}

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

if [[ "$os_type" != "Linux" ]]; then
    echo "error: only Linux machine can be added to the cluster"
    exit 1
fi

# set shell execute command
user="$(id -un 2>/dev/null || true)"
sh_c='sh -c'
if [ "$user" != 'root' ]; then
    if ! command_exists sudo; then
        echo "error: the ability to run as root is needed, but the command \"sudo\" can not be found"
        exit 1
    fi
    sh_c='sudo -E sh -c'
fi

if ! command_exists tar; then
    echo "error: the \"tar\" command is needed to unpack installation files, but can not be found"
    exit 1
fi

BASE_DIR="$HOME/.olares"
if [ ! -d $BASE_DIR ]; then
    mkdir -p $BASE_DIR
fi

cdn_url=${DOWNLOAD_CDN_URL}
if [[ -z "${cdn_url}" ]]; then
    cdn_url="https://dc3p1870nn3cj.cloudfront.net"
fi

set_master_host_ssh_options

CLI_VERSION="0.1.107"
CLI_FILE="olares-cli-v${CLI_VERSION}_linux_${ARCH}.tar.gz"

if command_exists olares-cli && [[ "$(olares-cli -v | awk '{print $3}')" == "$CLI_VERSION" ]]; then
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
            echo "Olares installer ${CLI_VERSION} download complete!"
            echo ""
        fi
    fi
    INSTALL_OLARES_CLI="/usr/local/bin/olares-cli"
    echo "unpacking Olares installer to $INSTALL_OLARES_CLI..."
    echo ""
    tar -zxf ${CLI_FILE} olares-cli && chmod +x olares-cli
    $sh_c "mv olares-cli $INSTALL_OLARES_CLI"

    if [[ $? -ne 0 ]]; then
        echo "error: failed to unpack Olares installer"
        exit 1
    fi
fi

echo "getting master info and checking current machine's eligibility to join the cluster"
echo ""
master_olares_version="$( getmasterinfo | grep OlaresVersion | awk '{print $2}' )"
if [[ ! "$master_olares_version" ]]; then
    echo "failed to fetch the version of Olares installed on master node"
    exit 1
fi
PARAMS="--version $master_olares_version --base-dir $BASE_DIR"
CDN="--download-cdn-url ${cdn_url}"

if [[ -f $BASE_DIR/.prepared ]]; then
    echo "file $BASE_DIR/.prepared detected, skip preparing phase"
    echo ""
    echo "please make sure the prepared Olares version is the same as the master, or there might be compatibility issues"
    echo ""
else
    echo "running system prechecks ..."
    echo ""
    $sh_c "$INSTALL_OLARES_CLI olares precheck $PARAMS"
    if [[ $? -ne 0 ]]; then
        exit 1
    fi

    echo "downloading installation wizard..."
    echo ""
    $sh_c "$INSTALL_OLARES_CLI olares download wizard $PARAMS $CDN"
    if [[ $? -ne 0 ]]; then
        echo "error: failed to download installation wizard"
        exit 1
    fi

    echo "downloading installation packages..."
    echo ""
    $sh_c "$INSTALL_OLARES_CLI olares download component $PARAMS $CDN"
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

echo "installing Kubernetes and joining Olares cluster..."
echo ""
$sh_c "$INSTALL_OLARES_CLI node add $PARAMS $MASTER_SSH_OPTIONS"

if [[ $? -ne 0 ]]; then
    echo "error: failed to install Olares"
    exit 1
fi
