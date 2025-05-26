#!/usr/bin/env bash


ERR_EXIT=1

CURL_TRY="--connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 "

BASE_DIR=$(dirname $(realpath -s $0))

[[ -f "${BASE_DIR}/.env" && -z "$DEBUG_VERSION" ]] && . "${BASE_DIR}/.env"

if [ ! -d "/tmp/install_log" ]; then
    $(mkdir -p /tmp/install_log)
fi

fd_errlog=/tmp/install_log/errlog_fd_13

get_distribution() {
	lsb_dist=""
	# Every system that we officially support has /etc/os-release
	if [ -r /etc/os-release ]; then
		lsb_dist="$(. /etc/os-release && echo "$ID")"
	fi
	echo "$lsb_dist"
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

function dpkg_locked() {
    grep -q 'Could not get lock /var/lib' "$fd_errlog"
    return $?
}

function retry_cmd(){
    "$@"
    local ret=$?
    if [ $ret -ne 0 ];then
        local max_retries=50
        local delay=3
        while [ $max_retries -gt 0 ]; do
            printf "retry to execute command '%s', after %d seconds\n" "$*" $delay
            ((delay+=2))
            sleep $delay

            "$@"
            ret=$?
            
            if [ $ret -eq 0 ]; then
                break
            fi
            
            ((max_retries--))

        done

        if [ $ret -ne 0 ]; then
            log_fatal "command: '$*'"
        fi
    fi

    return $ret
}

function ensure_success() {
    exec 13> "$fd_errlog"

    "$@" 2>&13
    local ret=$?

    if [ $ret -ne 0 ]; then
        local max_retries=50
        local delay=3

        if dpkg_locked; then
            while [ $max_retries -gt 0 ]; do
                printf "retry to execute command '%s', after %d seconds\n" "$*" $delay
                ((delay+=2))
                sleep $delay

                exec 13> "$fd_errlog"
                "$@" 2>&13
                ret=$?

                local r=""

                if [ $ret -eq 0 ]; then
                    r=y
                fi

                if ! dpkg_locked; then
                    r+=y
                fi

                if [[ x"$r" == x"yy" ]]; then
                    printf "execute command '%s' successed.\n\n" "$*"
                    break
                fi
                ((max_retries--))
            done
        else
            log_fatal "command: '$*'"
        fi
    fi

    return $ret
}

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

log_info() {
    local msg now

    msg="$*"
    now=$(date +'%Y-%m-%d %H:%M:%S.%N %z')
    echo -e "\n\033[38;1m${now} [INFO] ${msg} \033[0m" 
}

log_fatal() {
    local msg now

    msg="$*"
    now=$(date +'%Y-%m-%d %H:%M:%S.%N %z')
    echo -e "\n\033[31;1m${now} [FATAL] ${msg} \033[0m" 
    exit $ERR_EXIT
}

system_service_active() {
    if [[ $# -ne 1 || x"$1" == x"" ]]; then
        return 1
    fi

    local ret
    ret=$($sh_c "systemctl is-active $1")
    if [ "$ret" == "active" ]; then
        return 0
    fi
    return 1
}


is_debian() {
    lsb_release=$(lsb_release -d 2>&1 | awk -F'\t' '{print $2}')
    if [ -z "$lsb_release" ]; then
        echo 0
        return
    fi
    if [[ ${lsb_release} == *Debian* ]]; then
        case "$lsb_release" in
            *12* | *11*)
                echo 1
                ;;
            *)
                echo 0
                ;;
        esac
    else
        echo 0
    fi
}

is_ubuntu() {
    lsb_release=$(lsb_release -d 2>&1 | awk -F'\t' '{print $2}')
    if [ -z "$lsb_release" ]; then
        echo 0
        return
    fi
    if [[ ${lsb_release} == *Ubuntu* ]];then 
        case "$lsb_release" in
            *24.*)
                echo 2
                ;;
            *22.* | *20.*)
                echo 1
                ;;
            *)
                echo 0
                ;;
        esac
    else
        echo 0
    fi
}

is_raspbian(){
    lsb_release=$(lsb_release -d 2>&1 | awk -F'\t' '{print $2}')
    if [ -z "$lsb_release" ]; then
        echo 0
        return
    fi
    if [[ ${lsb_release} == *Raspbian* ]];then 
        case "$lsb_release" in
            *11* | *12*)
                echo 1
                ;;
            *)
                echo 0
                ;;
        esac
    else
        echo 0
    fi
}

is_wsl(){
    wsl=$(uname -a 2>&1)
    if [[ ${wsl} == *WSL* ]]; then
        echo 1
        return
    fi

    echo 0
}



get_distribution
get_shell_exec