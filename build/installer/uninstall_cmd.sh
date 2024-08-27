#!/usr/bin/env bash

ERR_EXIT=1
RM=$(command -v rm)
BASE_DIR=$(dirname $(realpath -s $0))
CURL_TRY="--connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 "
KKE_FILE="/etc/kke/version"

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

precheck_os() {
    local ip os_type os_arch

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

     OS_ARCH="$os_arch"
}


get_shell_exec(){
    user="$(id -un 2>/dev/null || true)"

	sh_c='sh -c'
	if [ "$user" != 'root' ]; then
		if command_exists sudo && command_exists su; then
			sh_c='sudo su -c'
		else
			cat >&2 <<-'EOF'
			Error: this installer needs the ability to run commands as root.
			We are unable to find either "sudo" or "su" available to make this happen.
			EOF
			exit $ERR_EXIT
		fi
	fi
}

ensure_success() {
    "$@"
    local ret=$?

    if [ $ret -ne 0 ]; then
        echo "Fatal error, command: '$*'"
        exit $ret
    fi

    return $ret
}

log_info() {
    local msg now

    msg="$*"
    now=$(date +'%Y-%m-%d %H:%M:%S.%N %z')
    echo -e "\n\033[38;1m${now} [INFO] ${msg} \033[0m" 
}

remove_cluster(){
    CLI_VERSION="0.1.13"
    forceUninstall="${FORCE_UNINSTALL_CLUSTER}"
    forceDeleteCache="false"

    version="${TERMINUS_IS_CLOUD_VERSION}"
    storage="${STORAGE}"
    s3_bucket="${S3_BUCKET}"

    log_info 'remove kubernetes cluster'

    local cli_tar="${BASE_DIR}/terminus-cli-v${CLI_VERSION}_linux_${ARCH}.tar.gz"
    if [ ! -f "${cli_tar}" ]; then
        ensure_success $sh_c "curl ${CURL_TRY} -kL -o ${BASE_DIR}/terminus-cli-v${CLI_VERSION}_linux_${ARCH}.tar.gz https://github.com/beclab/Installer/releases/download/${CLI_VERSION}/terminus-cli-v${CLI_VERSION}_linux_${ARCH}.tar.gz"
    fi
    ensure_success $sh_c "tar xvf ${BASE_DIR}/terminus-cli-v${CLI_VERSION}_linux_${ARCH}.tar.gz -C ${BASE_DIR}"
    ensure_success $sh_c "chmod +x ${BASE_DIR}/terminus-cli"

    if [ x"$PREPARED" != x"1" ]; then
      if [ -z "$forceUninstall" ]; then
        echo
        read -r -p "Are you sure to delete this cluster? [yes/no]: " ans </dev/tty

        if [ x"$ans" != x"yes" ]; then
            echo "exiting..."
            exit
        fi
      fi
    fi

    if [ ! -z "$forceUninstall" ]; then
      forceDeleteCache="true"
    fi

    local extra
    if [ x"$PREPARED" == x"1" ]; then
        extra=" --quiet "
    else
        extra=" --storage-type=${storage} --storage-bucket=${s3_bucket} "
    fi

    $sh_c "export DELETE_CACHE=${forceDeleteCache} && export TERMINUS_IS_CLOUD_VERSION=${version} && ${BASE_DIR}/terminus-cli terminus uninstall ${extra}"

    [ -f $KKE_FILE ] && $sh_c "${RM} -f $KKE_FILE"
}

set -o pipefail
# set -e

if [ ! -f '/var/run/lock/.installed' ]; then
    exit 0
fi

[[ -f /var/run/lock/.prepared ]] && PREPARED=1

get_shell_exec
precheck_os

INSTALL_DIR=/tmp/install_log

[[ -d ${INSTALL_DIR} ]] && $sh_c "${RM} -rf ${INSTALL_DIR}" 
mkdir -p ${INSTALL_DIR} && cd ${INSTALL_DIR}

Main() {
    log_info 'Uninstalling OS ...'
    remove_cluster

    cd -
    $sh_c "${RM} -rf /tmp/install_log"
    [[ -d install-wizard ]] && ${RM} -rf install-wizard
    set +o pipefail
    ls |grep install-wizard*.tar.gz | while read ar; do  ${RM} -f ${ar}; done

    ${RM} -rf /var/run/lock/.installed
    log_info 'Uninstall OS success! '
}



Main | tee ${BASE_DIR}/uninstall.log