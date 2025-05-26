#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))
source $BASE_DIR/common.sh

lsb_dist=$1
SOCAT_VERSION=$2
FLEX_VERSION=$3
CONNTRACK_VERSION=$4

build_socat(){
    if [ -z $SOCAT_VERSION ]; then
        SOCAT_VERSION="1.7.3.4"
    fi
    local socat_tar="${BASE_DIR}/components/socat-${SOCAT_VERSION}.tar.gz"
    ensure_success $sh_c "tar xzvf socat-${SOCAT_VERSION}.tar.gz"
    ensure_success $sh_c "cd socat-${SOCAT_VERSION}"
    ensure_success $sh_c "./configure --prefix=/usr && make -j4 && make install && strip socat"
}

build_contrack(){
    if [ -z $CONNTRACK_VERSION ]; then
        CONNTRACK_VERSION="1.4.1"
    fi
    local contrack_tar="${BASE_DIR}/components/conntrack-tools-${CONNTRACK_VERSION}.tar.gz"
    ensure_success $sh_c "tar zxvf conntrack-tools-${CONNTRACK_VERSION}.tar.gz"
    ensure_success $sh_c "cd conntrack-tools-${CONNTRACK_VERSION}"
    ensure_success $sh_c "./configure --prefix=/usr && make -j4 && make install"
}

install_deps() {
    case "$lsb_dist" in
        ubuntu|debian|raspbian)
            pre_reqs="apt-transport-https ca-certificates curl"
			if ! command -v gpg > /dev/null; then
				pre_reqs="$pre_reqs gnupg"
			fi
            ensure_success $sh_c 'apt-get update -qq >/dev/null'
            ensure_success $sh_c "DEBIAN_FRONTEND=noninteractive apt-get install -y -qq $pre_reqs >/dev/null"
            ensure_success $sh_c 'DEBIAN_FRONTEND=noninteractive apt-get install -y conntrack socat apache2-utils ntpdate net-tools make gcc openssh-server >/dev/null'
            ;;

        centos|fedora|rhel)
            if [ "$lsb_dist" = "fedora" ]; then
                pkg_manager="dnf"
            else
                pkg_manager="yum"
            fi

            ensure_success $sh_c "$pkg_manager install -y conntrack socat httpd-tools ntpdate net-tools make gcc openssh-server >/dev/null"
            ;;
        *)
            # build from source code
            build_socat
            build_contrack

            #TODO: install bcrypt tools
            ;;
    esac
}

echo ">>> install_deps os: ${lsb_dist}, socat: ${SOCAT_VERSION}, flex: ${FLEX_VERSION}, conntrack: ${CONNTRACK_VERSION}"
install_deps
exit