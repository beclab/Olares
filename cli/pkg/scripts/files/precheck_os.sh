#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))
source $BASE_DIR/common.sh

precheck_os() {
    local ip os_type os_arch

    # check os type and arch and os vesion
    os_type=$(uname -s)
    os_arch=$(uname -m)
    os_verion=$(lsb_release -d 2>&1 | awk -F'\t' '{print $2}')

    case "$os_arch" in 
        x86_64) ARCH=amd64; ;; 
        armv7l) ARCH=arm; ;; 
        aarch64) ARCH=arm64; ;; 
        ppc64le) ARCH=ppc64le; ;; 
        s390x) ARCH=s390x; ;; 
        *) echo "unsupported arch, exit ..."; 
        exit -1; ;; 
    esac 

    if [ x"${os_type}" != x"Linux" ]; then
        log_fatal "unsupported os type '${os_type}', only supported 'Linux' operating system"
    fi

    if [[ x"${os_arch}" != x"x86_64" && x"${os_arch}" != x"amd64" && x"${os_arch}" != x"aarch64" ]]; then
        log_fatal "unsupported os arch '${os_arch}', only supported 'x86_64' or 'aarch64' architecture"
    fi

    if [[ $(is_ubuntu) -eq 0 && $(is_debian) -eq 0 && $(is_raspbian) -eq 0 ]]; then
        log_fatal "unsupported os version '${os_verion}'"
    fi

    if [[ -f /boot/cmdline.txt || -f /boot/firmware/cmdline.txt ]]; then
     # raspbian 
        SHOULD_RETRY=1

        if ! command_exists iptables; then 
            ensure_success $sh_c "apt update && apt install -y iptables"
        fi

        systemctl disable --user gvfs-udisks2-volume-monitor
        systemctl stop --user gvfs-udisks2-volume-monitor

        local cpu_cgroups_enbaled=$(cat /proc/cgroups |awk '{if($1=="cpu")print $4}')
        local mem_cgroups_enbaled=$(cat /proc/cgroups |awk '{if($1=="memory")print $4}')
        if  [[ $cpu_cgroups_enbaled -eq 0 || $mem_cgroups_enbaled -eq 0 ]]; then
            log_fatal "cpu or memory cgroups disabled, please edit /boot/cmdline.txt or /boot/firmware/cmdline.txt and reboot to enable it."
        fi
    fi

    # try to resolv hostname
    ensure_success $sh_c "hostname -i >/dev/null"

    ip=$(ping -c 1 "$HOSTNAME" |awk -F '[()]' '/icmp_seq/{print $2}')
    printf "%s\t%s\n\n" "$ip" "$HOSTNAME"

    if [[ x"$ip" == x"" || "$ip" == @("172.17.0.1"|"127.0.0.1"|"127.0.1.1") || ! "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_fatal "incorrect ip for hostname '$HOSTNAME', please check"
    fi

    local_ip="$ip"
    OS_ARCH="$os_arch"

    # disable local dns
    case "$lsb_dist" in
        ubuntu|debian|raspbian)
            if system_service_active "systemd-resolved"; then
                ensure_success $sh_c "systemctl stop systemd-resolved.service >/dev/null"
                ensure_success $sh_c "systemctl disable systemd-resolved.service >/dev/null"
                if [ -e /usr/bin/systemd-resolve ]; then
                    ensure_success $sh_c "mv /usr/bin/systemd-resolve /usr/bin/systemd-resolve.bak >/dev/null"
                fi
                if [ -L /etc/resolv.conf ]; then
                    ensure_success $sh_c 'unlink /etc/resolv.conf && touch /etc/resolv.conf'
                fi
                config_resolv_conf
            else
                ensure_success $sh_c "cat /etc/resolv.conf > /etc/resolv.conf.bak"
            fi
            ;;
        centos|fedora|rhel)
            ;;
        *)
            ;;
    esac

    if ! hostname -i &>/dev/null; then
        ensure_success $sh_c "echo $local_ip  $HOSTNAME >> /etc/hosts"
    fi

    ensure_success $sh_c "hostname -i >/dev/null"

    # network and dns
    http_code=$(curl ${CURL_TRY} -sL -o /dev/null -w "%{http_code}" https://download.docker.com/linux/ubuntu)
    if [ "$http_code" != 200 ]; then
        config_resolv_conf
        if [ -f /etc/resolv.conf.bak ]; then
            ensure_success $sh_c "rm -rf /etc/resolv.conf.bak"
        fi

    fi

    # ubuntu 24 upgrade apparmor
    ubuntuversion=$(is_ubuntu)
    if [ ${ubuntuversion} -eq 2 ]; then
        aapv=$(apparmor_parser --version)
        if [[ ! ${aapv} =~ "4.0.1" ]]; then
            ensure_success $sh_c "dpkg -i ${BASE_DIR}/components/apparmor_4.0.1-0ubuntu1_${ARCH}.deb"
        fi
    fi
}

log_info 'Precheck and Installing dependencies ...'
precheck_os
exit