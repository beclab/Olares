#!/bin/bash

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

get_shell_exec
precheck_os

CLI_VERSION="0.1.14"
CLI_FILENAME="terminus-cli-v${CLI_VERSION}_darwin_${ARCH}.tar.gz"
CLI_URL="https://github.com/beclab/Installer/releases/download/${CLI_VERSION}/terminus-cli-v${CLI_VERSION}_darwin_${ARCH}.tar.gz"


cli_tar="terminus-cli-v${CLI_VERSION}_darwin_${ARCH}.tar.gz"
if [ ! -f "${CLI_FILENAME}" ]; then
    curl -Lo ${CLI_FILENAME} ${CLI_URL}
fi
tar xvf terminus-cli-v${CLI_VERSION}_darwin_${ARCH}.tar.gz; chmod +x terminus-cli

./terminus-cli terminus uninstall --minikube