#!/usr/bin/env bash

ERR_EXIT=1
ERR_VALIDATION=2

CURL_TRY="--retry 5 --retry-delay 1 --retry-max-time 10 "

# BASE_DIR=$(dirname $(realpath -s $0))

read_tty(){
    echo -n $1
    read $2 < /dev/tty
}

check_backup_password() {
    backupPassword="$BACKUP_REPOSITORY_PASSWORD"

    if [ x"$backupPassword" == x"" ]; then
        while :; do
            read_tty "Enter the backup password: " backupPassword
            if ! validate_backuppwd; then
                continue
            fi
            break
        done
    fi
}

validate_backuppwd() {
    local match
    match=$(echo $backupPassword |egrep -o '^.{3,20}$')

    if [ x"$match" != x"$backupPassword" ]; then
        printf "illegal backup password '$backupPassword', try again\n\n"
        return 1
    fi
    return 0
}

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
            echo "Fatal error, command: '$*'"
            exit $ret
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

build_socat(){
    SOCAT_VERSION="1.7.3.2"
    # Download
    ensure_success $sh_c "curl ${CURL_TRY} -LO http://www.dest-unreach.org/socat/download/socat-${SOCAT_VERSION}.tar.gz"
    ensure_success $sh_c "tar xzvf socat-${SOCAT_VERSION}.tar.gz"
    ensure_success $sh_c "cd socat-${SOCAT_VERSION}"

    ensure_success $sh_c "./configure --prefix=/usr && make -j4 && make install && strip socat"
}

build_contrack(){
    ensure_success $sh_c "curl ${CURL_TRY} -LO https://github.com/fqrouter/conntrack-tools/archive/refs/tags/conntrack-tools-1.4.1.tar.gz"
    ensure_success $sh_c "tar zxvf conntrack-tools-1.4.1.tar.gz"
    ensure_success $sh_c "cd conntrack-tools-1.4.1"

    ensure_success $sh_c "./configure --prefix=/usr && make -j4 && make install"
}

precheck_os() {
    local ip os_type os_arch

    os_type=$(uname -s)
    os_arch=$(uname -m)
    os_verion=$(lsb_release -d 2>&1 | awk -F'\t' '{print $2}')

    if [ x"${os_type}" != x"Linux" ]; then
        log_fatal "unsupported os type '${os_type}', only supported 'Linux' operating system"
    fi

    if [[ x"${os_arch}" != x"x86_64" && x"${os_arch}" != x"amd64" ]]; then
        log_fatal "unsupported os arch '${os_arch}', only supported 'x86_64' architecture"
    fi

    if [[ $(is_ubuntu) -eq 0 && $(is_debian) -eq 0 ]]; then
        log_fatal "unsupported os version '${os_verion}', only supported Ubuntu 20.x, 22.x, 24.x and Debian 11, 12"
    fi

    # try to resolv hostname
    ensure_success $sh_c "hostname -i >/dev/null"

    ip=$(ping -c 1 "$HOSTNAME" |awk -F '[()]' '/icmp_seq/{print $2}')
    echo "$ip  $HOSTNAME"

    if [[ x"$ip" == x"" || "$ip" == @("172.17.0.1"|"127.0.0.1") ]]; then
        echo "incorrect ip for hostname '$HOSTNAME', please check"
        exit $ERR_EXIT
    fi

    local_ip="$ip"

    # disable local dns
    case "$lsb_dist" in
        ubuntu|debian|raspbian)
            if [ -f /usr/bin/systemd-resolve ]; then
                ensure_success $sh_c "systemctl stop systemd-resolved.service >/dev/null"
                ensure_success $sh_c "systemctl disable systemd-resolved.service >/dev/null"

                ensure_success $sh_c "mv /usr/bin/systemd-resolve /usr/bin/systemd-resolve.bak >/dev/null"
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
    http_code=$(curl ${CURL_TRY} --connect-timeout 30 -ksL -o /dev/null -w "%{http_code}" https://download.docker.com/linux/ubuntu)
    if [ "$http_code" != 200 ]; then
        config_resolv_conf
    fi

    # ubuntu 24 upgrade apparmor
    ubuntuversion=$(is_ubuntu)
    if [ ${ubuntuversion} -eq 2 ]; then
        aapv=$(apparmor_parser --version)
        if [[ ! ${aapv} =~ "4.0.1" ]]; then
            ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://launchpad.net/ubuntu/+source/apparmor/4.0.1-0ubuntu1/+build/28428840/+files/apparmor_4.0.1-0ubuntu1_amd64.deb"
            ensure_success $sh_c "dpkg -i apparmor_4.0.1-0ubuntu1_amd64.deb"
        fi
    fi
}

is_debian() {
    lsb_release=$(lsb_release -d 2>&1 | awk -F'\t' '{print $2}')
    if [ -z "$lsb_release" ]; then
        echo 0
        return
    fi
    if [[ ${lsb_release} == *Debian*} ]]; then
        case "$lsb_release" in
            *12.* | *11.*)
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

install_deps() {
    case "$lsb_dist" in
        ubuntu|debian|raspbian)
            pre_reqs="apt-transport-https ca-certificates curl"
			if ! command -v gpg > /dev/null; then
				pre_reqs="$pre_reqs gnupg"
			fi
            ensure_success $sh_c 'apt-get update -qq >/dev/null'
            ensure_success $sh_c "DEBIAN_FRONTEND=noninteractive apt-get install -y -qq $pre_reqs >/dev/null"
            ensure_success $sh_c 'apt-get install -y conntrack socat apache2-utils ntpdate net-tools make gcc tar >/dev/null'
            ;;

        centos|fedora|rhel)
            if [ "$lsb_dist" = "fedora" ]; then
                pkg_manager="dnf"
            else
                pkg_manager="yum"
            fi

            ensure_success $sh_c "$pkg_manager install -y conntrack socat httpd-tools ntpdate net-tools make gcc tar >/dev/null"
            ;;
        *)
            # build from source code
            build_socat
            build_contrack

            #TODO: install bcrypt tools
            ;;
    esac
}

config_system() {
    local ntpdate hwclock

    # ntp sync
    ntpdate=$(command -v ntpdate)
    hwclock=$(command -v hwclock)

    printf '#!/bin/sh\n\n%s -b -u pool.ntp.org && %s -w\n\nexit 0\n' "$ntpdate" "$hwclock" > cron.ntpdate
    ensure_success $sh_c '/bin/sh cron.ntpdate'
    ensure_success $sh_c 'cat cron.ntpdate > /etc/cron.daily/ntpdate && chmod 0700 /etc/cron.daily/ntpdate'
    ensure_success rm -f cron.ntpdate
}

config_resolv_conf() {
    ensure_success $sh_c 'echo -e "nameserver 1.0.0.1\nnameserver 1.1.1.1" > /etc/resolv.conf'
}

restore_resolv_conf() {
    # restore /etc/resolv.conf
    if [ -f /etc/resolv.conf.bak ]; then
        ns=$(awk '/nameserver/{print $NF}' /etc/resolv.conf.bak)
        if [ x"$PROXY" != x"" -a x"$ns" == x"$PROXY" ]; then
            config_resolv_conf
        else
            ensure_success $sh_c "cat /etc/resolv.conf.bak > /etc/resolv.conf"
        fi
    fi
}

install_storage() {
    TERMINUS_ROOT="/olares"

    if [ x"$PROXY" != x"" ]; then
	    ensure_success $sh_c "echo nameserver $PROXY > /etc/resolv.conf"
    fi

    log_info 'Preparing object storage ...\n'

    storage_type="minio"    # or s3

    if [ x"$STORAGE" != x"" ]; then
        storage_type="$STORAGE"
    fi

    echo "storage_type = ${storage_type}"

    case "$storage_type" in
        minio)
            install_minio
            ;;
        s3|oss)
            echo "s3_bucket = ${S3_BUCKET}"

            if [ "x$S3_BUCKET" == "x" ]; then
                echo "s3 bucket is empty."
                exit $ERR_EXIT
            fi
            ;;
        *)
            echo "storage '$storage_type' not supported."
            exit $ERR_EXIT
        ;;
    esac

    # install redis and juicefs filesystem
    install_redis
    install_juicefs
}

install_minio() {
    MINIO_VERSION="RELEASE.2023-05-04T21-44-30Z"
    MINIO_ROOT_USER="minioadmin"
    MINIO_ROOT_PASSWORD=$(random_string 16)

    log_info 'start to install minio'

    local minio_bin="/usr/local/bin/minio"
    local minio_data="${TERMINUS_ROOT}/data/minio/vol1"

    [ ! -d "$minio_data" ] && ensure_success $sh_c "mkdir -p $minio_data"

    if [ ! -f "$minio_bin" ]; then
        ensure_success $sh_c "curl ${CURL_TRY} -kLo minio https://dl.min.io/server/minio/release/linux-amd64/archive/minio.${MINIO_VERSION}"
        ensure_success $sh_c "chmod +x minio"
        ensure_success $sh_c "install minio /usr/local/bin"
    fi

    cat > "${INSTALL_DIR}/minio.service" <<_END
[Unit]
Description=MinIO
Documentation=https://min.io/docs/minio/linux/index.html
Wants=network-online.target
After=network-online.target
AssertFileIsExecutable=$minio_bin

[Service]
WorkingDirectory=/usr/local

User=minio
Group=minio
ProtectProc=invisible

EnvironmentFile=-/etc/default/minio
ExecStartPre=/bin/bash -c "if [ -z \"\${MINIO_VOLUMES}\" ]; then echo \"Variable MINIO_VOLUMES not set in /etc/default/minio\"; exit 1; fi"
ExecStart=$minio_bin server \$MINIO_OPTS \$MINIO_VOLUMES

# MinIO RELEASE.2023-05-04T21-44-30Z adds support for Type=notify (https://www.freedesktop.org/software/systemd/man/systemd.service.html#Type=)
# This may improve systemctl setups where other services use After=minio.server
# Uncomment the line to enable the functionality
# Type=notify

# Let systemd restart this service always
Restart=always

# Specifies the maximum file descriptor number that can be opened by this process
LimitNOFILE=65536

# Specifies the maximum number of threads this process can create
TasksMax=infinity

# Disable timeout logic and wait until process is stopped
TimeoutStopSec=infinity
SendSIGKILL=no

[Install]
WantedBy=multi-user.target

_END

        ensure_success $sh_c "cat ${INSTALL_DIR}/minio.service > /etc/systemd/system/minio.service"
        cat > "${INSTALL_DIR}/minio.env" <<_END
# MINIO_ROOT_USER and MINIO_ROOT_PASSWORD sets the root account for the MinIO server.
# This user has unrestricted permissions to perform S3 and administrative API operations on any resource in the deployment.
# Omit to use the default values 'minioadmin:minioadmin'.
# MinIO recommends setting non-default values as a best practice, regardless of environment
MINIO_VOLUMES="$minio_data"
MINIO_OPTS="--console-address ${local_ip}:9090 --address ${local_ip}:9000"

MINIO_ROOT_USER=$MINIO_ROOT_USER
MINIO_ROOT_PASSWORD=$MINIO_ROOT_PASSWORD
_END
    ensure_success $sh_c "cat "${INSTALL_DIR}/minio.env" > /etc/default/minio"

    $sh_c "groupadd -r minio >/dev/null; true"
    $sh_c "useradd -M -r -g minio minio >/dev/null; true"
    ensure_success $sh_c "chown minio:minio $minio_data"

    ensure_success $sh_c "systemctl daemon-reload"
    ensure_success $sh_c "systemctl restart minio"
    ensure_success $sh_c "systemctl enable minio"
    ensure_success $sh_c "systemctl --no-pager status minio"

    # ensure minio is ready
    local max_retry=60
    local ok="n"
    while [ $max_retry -ge 0 ]; do
        if $sh_c 'systemctl --no-pager status minio >/dev/null'; then
            ok=y
            break
        fi
        sleep 5
        ((max_retry--))
    done

    if [ x"$ok" != x"y" ]; then
        echo "minio is not ready yet, please check it"
        exit $ERR_EXIT
    fi

    # minio password from file
    MINIO_ROOT_PASSWORD=$(awk -F '=' '/^MINIO_ROOT_PASSWORD/{print $2}' /etc/default/minio)
    MINIO_VOLUMES=$minio_data
}

init_minio_cluster(){

    MINIO_OPERATOR_VERSION="v0.0.1"
    if [[ ! -f /etc/ssl/etcd/ssl/ca.pem || ! -f /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem || ! -f /etc/ssl/etcd/ssl/node-$HOSTNAME.pem ]]; then
        echo "cann't find etcd key files"
        exit $ERR_EXIT
    fi

    local minio_operator_bin="/usr/local/bin/minio-operator"

    if [ ! -f "$minio_operator_bin" ]; then
        # TODO: mini-operator public repo
        ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://github.com/beclab/minio-operator/releases/download/${MINIO_OPERATOR_VERSION}/minio-operator-${MINIO_OPERATOR_VERSION}-linux-amd64.tar.gz"
	    ensure_success $sh_c "tar zxf minio-operator-${MINIO_OPERATOR_VERSION}-linux-amd64.tar.gz"
        ensure_success $sh_c "install -m 755 minio-operator $minio_operator_bin"
    fi

    ensure_success $sh_c "$minio_operator_bin init --address $local_ip --cafile /etc/ssl/etcd/ssl/ca.pem --certfile /etc/ssl/etcd/ssl/node-$HOSTNAME.pem --keyfile /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem --volume $MINIO_VOLUMES --password $MINIO_ROOT_PASSWORD"
}

install_redis() {
    REDIS_VERSION=5.0.14
    REDIS_PASSWORD=$(random_string 16)

    log_info 'start to install redis'

    local redis_root="${TERMINUS_ROOT}/data/redis"
    local redis_conf="${redis_root}/etc/redis.conf"
    local redis_bin="/usr/bin/redis-server"
    local cpu_cores

    # install redis, if redis-server not exists
    if [ ! -f "$redis_bin" ]; then
        ensure_success $sh_c "curl -kLO https://download.redis.io/releases/redis-${REDIS_VERSION}.tar.gz"
        ensure_success $sh_c "tar xf redis-${REDIS_VERSION}.tar.gz"

        cpu_cores=$(grep -c processor /proc/cpuinfo)
        if [ -z "$cpu_cores" ] || [ "$cpu_cores" -le 1 ]; then
            cpu_cores=1
        fi
        ensure_success $sh_c "cd redis-${REDIS_VERSION} && make -j${cpu_cores} >/dev/null 2>&1 && make install >/dev/null 2>&1 && cd .."
        ensure_success $sh_c "ln -s /usr/local/bin/redis-server ${redis_bin}"
        ensure_success $sh_c "ln -s /usr/local/bin/redis-cli /usr/bin/redis-cli"
    fi

    # config redis
    ensure_success $sh_c "ls $redis_bin >/dev/null"
    [ ! -d "$redis_root" ] && ensure_success $sh_c "mkdir -p ${redis_root}/etc" \
        && ensure_success $sh_c "mkdir -p ${redis_root}/data" \
        && ensure_success $sh_c "mkdir -p ${redis_root}/log" \
        && ensure_success $sh_c "mkdir -p ${redis_root}/run"

    cat > "${INSTALL_DIR}/redis.conf" <<_END
protected-mode no
bind $local_ip
port 6379
daemonize no
supervised no
pidfile ${redis_root}/run/redis.pid
logfile ${redis_root}/log/redis-server.log
save 900 1
save 300 10
save 60 10000
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir ${redis_root}/data
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
requirepass $REDIS_PASSWORD
_END
    ensure_success $sh_c "cat ${INSTALL_DIR}/redis.conf > $redis_conf"
    ensure_success $sh_c "chmod 0640 $redis_conf"

    cat > "${INSTALL_DIR}/redis-server.service" <<_END
[Unit]
Description=Redis
Documentation=https://redis.io/
Wants=network-online.target
After=network-online.target
AssertFileIsExecutable=$redis_bin

[Service]
WorkingDirectory=$redis_root

User=root
Group=root

EnvironmentFile=
ExecStartPre=/bin/sh -c 'test -f /sys/kernel/mm/transparent_hugepage/enabled && /bin/echo never > /sys/kernel/mm/transparent_hugepage/enabled'
ExecStart=$redis_bin $redis_conf

# Let systemd restart this service always
Restart=no

# Specifies the maximum file descriptor number that can be opened by this process
LimitNOFILE=65536

# Specifies the maximum number of threads this process can create
TasksMax=infinity

# Disable timeout logic and wait until process is stopped
TimeoutStopSec=infinity
SendSIGKILL=no

[Install]
WantedBy=multi-user.target
_END

    ensure_success $sh_c "cat ${INSTALL_DIR}/redis-server.service > /etc/systemd/system/redis-server.service"
    ensure_success $sh_c "sysctl -w vm.overcommit_memory=1 net.core.somaxconn=10240 >/dev/null"

    ensure_success $sh_c "systemctl daemon-reload >/dev/null"
    ensure_success $sh_c "systemctl restart redis-server >/dev/null; true"
    ensure_success $sh_c "systemctl enable redis-server >/dev/null"

    log_info 'redis service enabled'

    # eusure redis is started
    ensure_success $sh_c "( sleep 10 && systemctl --no-pager status redis-server ) || \
    ( systemctl restart redis-server && sleep 3 && systemctl --no-pager status redis-server ) || \
    ( systemctl restart redis-server && sleep 3 && systemctl --no-pager status redis-server )"

    REDIS_PASSWORD=$($sh_c "awk '/requirepass/{print \$NF}' $redis_conf")
    if [ x"$REDIS_PASSWORD" == x"" ]; then
        echo "no redis password found in $redis_conf"
        exit $ERR_EXIT
    fi

    log_info 'try to connect redis'

    pong=$(/usr/bin/redis-cli -h "$local_ip" -a "$REDIS_PASSWORD" ping 2>/dev/null)
    if [ x"$pong" != x"PONG" ]; then
        echo "failed to connect redis server: ${local_ip}:6379"
        exit $ERR_EXIT
    fi

    log_info 'success to install redis'
}

install_juicefs() {
    JFS_VERSION="v11.0.1"

    log_info 'start to install juicefs'

    local juicefs_data="${TERMINUS_ROOT}/data/juicefs"
    if [ ! -d "$juicefs_data" ]; then
        ensure_success $sh_c "mkdir -p $juicefs_data"
    fi


    local format_cmd
    local fsname="rootfs"
    local bucket="olares"
    local metadb="redis://:${REDIS_PASSWORD}@${local_ip}:6379/1"

    local juicefs_bin="/usr/local/bin/juicefs"
    local jfs_mountpoint="${TERMINUS_ROOT}/${fsname}"
    local jfs_cachedir="${TERMINUS_ROOT}/jfscache"
    [ ! -d $jfs_mountpoint ] && ensure_success $sh_c "mkdir -p $jfs_mountpoint"
    [ ! -d $jfs_cachedir ] && ensure_success $sh_c "mkdir -p $jfs_cachedir"

    if [ ! -f "$juicefs_bin" ]; then
        ensure_success $sh_c "curl ${CURL_TRY} -kLO https://github.com/beclab/juicefs-ext/releases/download/${JFS_VERSION}/juicefs-${JFS_VERSION}-linux-amd64.tar.gz"
        ensure_success $sh_c "tar -zxf juicefs-${JFS_VERSION}-linux-amd64.tar.gz"
        ensure_success $sh_c "chmod +x juicefs"
        ensure_success $sh_c "install juicefs /usr/local/bin"
        ensure_success $sh_c "install juicefs /sbin/mount.juicefs"

        # format minio or s3
        format_cmd="$juicefs_bin format $metadb --storage $storage_type"
        if [ "$storage_type" == "minio" ]; then
            format_cmd+=" --bucket http://${local_ip}:9000/${bucket} --access-key $MINIO_ROOT_USER --secret-key $MINIO_ROOT_PASSWORD"
        elif [[ "$storage_type" == @("s3"|"oss") ]]; then
            format_cmd+=" --bucket $S3_BUCKET"

            if [[ ! -z "${TERMINUS_IS_CLOUD_VERSION}" && x"${TERMINUS_IS_CLOUD_VERSION}" == x"true" ]]; then
                ak="${AWS_ACCESS_KEY_ID_SETUP}"
                sk="${AWS_SECRET_ACCESS_KEY_SETUP}"

                if [ ! -z "${AWS_SESSION_TOKEN_SETUP}" ]; then
                    format_cmd+=" --session-token ${AWS_SESSION_TOKEN_SETUP}"
                fi

                fsname="${CLUSTER_ID}"
            fi


            if [[ x"$ak" != x"" && x"$sk" != x"" ]]; then
                format_cmd+=" --access-key $ak --secret-key $sk"
            fi
        fi
        # format_cmd+=" $fsname &>/dev/null"
        format_cmd+=" $fsname" # debug
        ensure_success $sh_c "$format_cmd"
    fi

    cat > "${INSTALL_DIR}/juicefs.service" <<_END
[Unit]
Description=JuicefsMount
Documentation=https://juicefs.com/docs/zh/community/introduction/
Wants=redis-online.target
After=redis-online.target
AssertFileIsExecutable=$juicefs_bin

[Service]
WorkingDirectory=/usr/local

EnvironmentFile=
ExecStart=$juicefs_bin mount --cache-dir $jfs_cachedir $metadb $jfs_mountpoint

# Let systemd restart this service always
Restart=always

# Specifies the maximum file descriptor number that can be opened by this process
LimitNOFILE=65536

# Specifies the maximum number of threads this process can create
TasksMax=infinity

# Disable timeout logic and wait until process is stopped
TimeoutStopSec=infinity
SendSIGKILL=no

[Install]
WantedBy=multi-user.target
_END
    ensure_success $sh_c "cat ${INSTALL_DIR}/juicefs.service > /etc/systemd/system/juicefs.service"

    ensure_success $sh_c "systemctl daemon-reload"
    ensure_success $sh_c "systemctl restart juicefs"
    ensure_success $sh_c "systemctl enable juicefs"

    ensure_success $sh_c "systemctl --no-pager status juicefs"
    ensure_success $sh_c "sleep 3 && test -d ${jfs_mountpoint}/.trash"
}

random_string() {
    local length=12
    local alphanumeric="0abc1de2fg3hi4jkl5mno6pqr7st8uvw9xyz"

    if [[ -n "$1" && $1 -gt 0 ]]; then
        length=$1
    fi

    local text n
    for ((i=0,l=${#alphanumeric}; i<$length; i++)); do
        n=$[RANDOM%l]
        text+="${alphanumeric:n:1}"
    done
    echo -n "$text"
}

repeat() {
    for _ in $(seq 1 "$1"); do
        echo -n "$2"
    done
}

_get_pod_not_runnings() {
    $sh_c "${KUBECTL} get pod -A --no-headers"|grep -Ev 'headscale|Running|Terminating|infisical-deployment'|wc -l
}

_get_backup_status() {
    local status
    local backupName=$($sh_c "echo $TERMINUS_BACKUP_NAME | rev | cut -d'-' -f2- | rev")
    status=$($sh_c "${VELERO} -n os-system backup get $backupName" 2>/dev/null)

    if [ $? -ne 0 ]; then
        return
    fi

    if echo "$status"|grep -q NAME; then
        echo "$status" |grep -v NAME|awk '{print $2}'
    fi
}

_get_restore_status() {
    local count=0
    local backupName=$($sh_c "echo $TERMINUS_BACKUP_NAME | rev | cut -d'-' -f2- | rev")
    local status=$($sh_c "${VELERO} -n os-system restore get" 2>/dev/null)

    if [ $? -ne 0 ]; then
        return
    fi

    local st
    echo "$status" | while IFS= read line; do
        if [ "$count" -ge 1 ]; then
            return
        fi
        if echo "$line" |grep -q "$backupName"; then
            count=1
            st=$(echo "$line" |awk '{print $3}')
            echo "$st"
        fi
    done
}

check_pods_running() {
    local not_runnings

    not_runnings=$(_get_pod_not_runnings)
    n=0
    while [ "$not_runnings" -gt 0 ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 3

        not_runnings=$(_get_pod_not_runnings)
        echo -ne "\rPlease waiting          "
    done
    echo
}

check_backup_storage_location_available() {
    local status
    status=$(_get_backup_storage_location_status)

    n=0
    while [[ ! -z $status && x"${status}" != x"Available" ]]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 2

        status=$(_get_backup_storage_location_status)
        echo -ne "\rPlease waiting          "
    done
    echo
}

_get_backup_storage_location_status() {
    local status
    status=$($sh_c "${VELERO} -n os-system backup-location get" 2>/dev/null)

    if [ $? -ne 0 ]; then
        return
    fi

    if echo "$status"|grep -q NAME; then
        echo "$status" |grep -v NAME|awk '{print $4}'
    fi
}

check_backup_available() {
    local status
    status=$(_get_backup_status)

    n=0
    while [[ ! -z $status && x"${status}" != x"Completed" ]]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 2

        status=$(_get_backup_status)
        echo -ne "\rPlease waiting          "
    done
    echo
}

check_restore_available() {
    local status
    status=$(_get_restore_status)

    n=0
    while [[ ! -z $status && x"${status}" != x"Completed" ]]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 2

        status=$(_get_restore_status)
        echo -ne "\rPlease waiting          "
    done
    echo
}

run_install() {
    k8s_version=v1.22.10

    # env 'KUBE_TYPE' is specific the special kubernetes (k8s or k3s), default k3s
    [[ -z $KUBE_TYPE ]] && KUBE_TYPE="k3s"
    if [ x"$KUBE_TYPE" == x"k3s" ]; then
        k8s_version=v1.21.5-k3s
    fi
    create_cmd="./kk create cluster --with-kubernetes $k8s_version --container-manager containerd"  # --with-addon ${ADDON_CONFIG_FILE}

    # env 'REGISTRY_MIRRORS' is a docker image cache mirrors, separated by commas
    if [ x"$REGISTRY_MIRRORS" != x"" ]; then
        create_cmd+=" --registry-mirrors $REGISTRY_MIRRORS"
    # env 'PROXY' is a cache proxy server, to download binaries and container images
    elif [ x"$PROXY" != x"" ]; then
        create_cmd+=" --registry-mirrors http://${PROXY}:5000 --download-cmd 'curl ${CURL_TRY} -kL -o %s %s'"
    fi

    ensure_success $sh_c "$create_cmd"
    log_info 'k8s cluster installation is complete'

    ensure_success $sh_c "sed -i '/${local_ip} $HOSTNAME/d' /etc/hosts"

    CRICTL=$(command -v crictl)
    ensure_success $sh_c "${CRICTL} config runtime-endpoint unix:///run/containerd/containerd.sock"
    ensure_success $sh_c "${CRICTL} config image-endpoint unix:///run/containerd/containerd.sock"

    KUBECTL=$(command -v kubectl)
    ensure_success $sh_c "${KUBECTL} get nodes -o wide"
}

install_velero() {
  VELERO_IMAGE="beclab/velero:v1.11.1"
  ensure_success $sh_c "${VELERO} install --image ${VELERO_IMAGE} --crds-only"
}

install_velero_plugin_terminus() {
  local region provider namespace bucket storage_location
  local plugin velero_storage_location_install_cmd velero_plugin_install_cmd
  local msg
  provider="terminus"
  namespace="os-system"
  storage_location="terminus-cloud"
  bucket="terminus-cloud"
  plugin="beclab/velero-plugin-for-terminus:v1.0.2"

  if [[ "$provider" == x"" || "$namespace" == x"" || "$bucket" == x"" || "$plugin" == x"" ]]; then
    echo "velero plugin install params invalid."
    exit $ERR_EXIT
  fi

  velero_plugin_terminus=$($sh_c "${VELERO} plugin get -n $namespace |grep 'velero.io/terminus' |wc -l")
  if [[ ${velero_plugin_terminus} == x"" || ${velero_plugin_terminus} -lt 1 ]]; then
    velero_plugin_install_cmd="${VELERO} install"
    velero_plugin_install_cmd+=" --no-default-backup-location --namespace $namespace"
    velero_plugin_install_cmd+=" --image ${VELERO_IMAGE} --use-volume-snapshots=false"
    velero_plugin_install_cmd+=" --no-secret --plugins $plugin"
    velero_plugin_install_cmd+=" --velero-pod-cpu-request=50m --velero-pod-cpu-limit=500m"
    velero_plugin_install_cmd+=" --node-agent-pod-cpu-request=50m --node-agent-pod-cpu-limit=500m"
    velero_plugin_install_cmd+=" --wait"
    ensure_success $sh_c "$velero_plugin_install_cmd"
    velero_plugin_install_cmd="${VELERO} plugin add $plugin -n $namespace"
    msg=$($sh_c "$velero_plugin_install_cmd 2>&1")
  fi

  if [[ ! -z $msg && $msg != *"Duplicate"*  ]]; then
    log_info "$msg"
  fi

  check_velero_ready

  local velero_patch
  velero_patch='[{"op":"replace","path":"/spec/template/spec/volumes","value": [{"name":"plugins","emptyDir":{}},{"name":"scratch","emptyDir":{}},{"name":"terminus-cloud","hostPath":{"path":"/terminus/rootfs/k8s-backup", "type":"DirectoryOrCreate"}}]},{"op": "replace", "path": "/spec/template/spec/containers/0/volumeMounts", "value": [{"name":"plugins","mountPath":"/plugins"},{"name":"scratch","mountPath":"/scratch"},{"mountPath":"/data","name":"terminus-cloud"}]},{"op": "replace", "path": "/spec/template/spec/containers/0/securityContext", "value": {"privileged": true, "runAsNonRoot": false, "runAsUser": 0}}]'

  msg=$($sh_c "${KUBECTL} patch deploy velero -n $namespace --type='json' -p='$velero_patch'")
  if [[ ! -z $msg && $msg != *"patched"* ]]; then
    log_info "Velero plugin patched error: $msg"
  else
    echo "Velero plugin patched succeed"
  fi

  sleep 1

  terminus_backup_location=$($sh_c "${VELERO} backup-location get -n $namespace | awk '\$1 == \"${storage_location}\" {count++} END{print count}'")
  if [[ ${terminus_backup_location} == x"" || ${terminus_backup_location} -lt 1 ]]; then
    velero_storage_location_install_cmd="${VELERO} backup-location create $storage_location"
    velero_storage_location_install_cmd+=" --provider $provider --namespace $namespace"
    velero_storage_location_install_cmd+=" --prefix \"\" --bucket $bucket"
    msg=$($sh_c "$velero_storage_location_install_cmd 2>&1")
  fi

  if [[ ! -z $msg && $msg != *"successfully"* && $msg != *"exists"* ]]; then
    log_info "$msg"
  fi

  sleep 1
}

restore_k8s_os() {
    log_info 'Check backup is avaliable ...'
    local backupName=$($sh_c "echo $TERMINUS_BACKUP_NAME | rev | cut -d'-' -f2- | rev")

    if [ x"$backupName" == x"" ]; then
        echo "no env 'TERMINUS_BACKUP_NAME' provided"
        exit $ERR_EXIT
    fi

    check_backup_storage_location_available

    check_backup_available

    local include_status_rgs="applicationpermissions.sys.bytetrade.io,providerregistries.sys.bytetrade.io"
    include_status_rgs+=",applications.app.bytetrade.io,terminus.sys.bytetrade.io,middlewarerequests.apr.bytetrade.io"
    include_status_rgs+=",pgclusterbackups.apr.bytetrade.io,pgclusterrestores.apr.bytetrade.io"
    include_status_rgs+=",redisclusterbackups.redis.kun"

    log_info 'Creating k8s restore task ...'
    ensure_success $sh_c "${VELERO} -n os-system restore create --status-include-resources $include_status_rgs --status-exclude-resources perconaservermongodbs.psmdb.percona.com,distributedredisclusters.redis.kun --selector 'managered-by notin (mongo-backup-mongo-cluster,mongo-restore-mongo-cluster)' --from-backup $backupName"

    check_restore_available
}

_get_sts_bfl() {
    local res
    res=$($sh_c "${KUBECTL} get sts -A -l tier=bfl 2>/dev/null")

    if [[ x"$res" == x"" ]]; then
        echo 0
    elif ! echo "$res"|grep -q NAME; then
        echo 0
    else
        echo "$res"|grep -v NAME|wc -l
    fi
}

_get_deployment_backup_server() {
    local res
    res=$($sh_c "${KUBECTL} -n os-framework get deployment backup 2>/dev/null")
    if [ "$?" -ne 0 ]; then
        echo 0
    fi

    if ! echo "$res"|grep -q NAME; then
        echo 0
    else
        echo "$res"|grep -v NAME|wc -l
    fi
}

check_velero_ready() {
    local status
    status=$(_get_deployment_velero)

    n=0
    while [[ ! -z $status && x"${status}" != x"Running" ]]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 2

        status=$(_get_deployment_velero)
        echo -ne "\rPlease waiting          "
    done
    echo
}

_get_deployment_velero() {
    local count=0
    local status=$($sh_c "${KUBECTL} get pod -n os-system |grep velero" 2>/dev/null)

    if [ $? -ne 0 ]; then
        return
    fi

    local st
    echo "$status" | while IFS= read line; do
        if [ "$count" -ge 1 ]; then
            return
        fi
        count=1
        st=$(echo "$line" |awk '{print $3}')
        echo "$st"
    done
}

check_backup_server_ready() {
    local count

    count=$(_get_deployment_backup_server)
    n=0
    while [ "$count" -lt 1 ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 3

        count=$(_get_deployment_backup_server)
        echo -ne "\rPlease waiting          "
    done
    echo
}

check_sts_bfl_ready() {
    local count

    count=$(_get_sts_bfl)
    n=0
    while [ "$count" -lt 1 ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 3

        count=$(_get_sts_bfl)
        echo -ne "\rPlease waiting          "
    done
    echo
}

_get_sts_app_service() {
    local res
    res=$($sh_c "${KUBECTL} -n os-system get sts app-service 2>/dev/null")

    if ! echo "$res"|grep -q NAME; then
        echo 0
    else
        echo "$res"|grep -v NAME|wc -l
    fi
}

check_sts_app_service_ready() {
    local count

    count=$(_get_sts_app_service)
    n=0
    while [ "$count" -lt 1 ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 3

        count=$(_get_sts_app_service)
        echo -ne "\rPlease waiting          "
    done
    echo
}

get_k8s_annotation() {
    if [ $# -ne 4 ]; then
        echo "get annotation, invalid parameters"
        exit $ERR_EXIT
    fi

    local ns resource_type resource_name key
    ns="$1"
    resource_type="$2"
    resource_name="$3"
    key="$4"

    local res

    res=$($sh_c "${KUBECTL} -n $ns get $resource_type $resource_name -o jsonpath='{.metadata.annotations.$key}' 2>/dev/null")
    if [[ $? -eq 0 && x"$res" != x"" ]]; then
        echo "$res"
        return
    fi
    echo "can not to get $ns ${resource_type}/${resource_name} annotation '$key', got value '$res'"
    exit $ERR_EXIT
}

check_together(){
    local all=$@
    local s=""
    for f in "${all[@]}"; do
        s=$($f)
        if [ "x${s}" != "xRunning" ]; then
            break
        fi
    done

    echo "${s}"
}

get_bfl_status(){
    $sh_c "${KUBECTL} get pod -A -l 'tier=bfl' -o jsonpath='{.items[0].status.phase}'"
}

get_auth_status(){
    $sh_c "${KUBECTL} get pod -A -l 'app=authelia' -o jsonpath='{.items[0].status.phase}'"
}

get_profile_status(){
    $sh_c "${KUBECTL} get pod -A -l 'app=profile' -o jsonpath='{.items[0].status.phase}'"
}

get_desktop_status(){
    $sh_c "${KUBECTL} get pod -A -l 'app=edge-desktop' -o jsonpath='{.items[0].status.phase}'"
}

get_vault_status(){
    $sh_c "${KUBECTL} get pod -A -l 'app=vault' -o jsonpath='{.items[0].status.phase}'"
}

get_settings_status(){
    $sh_c "${KUBECTL} get pod -A -l 'app=settings' -o jsonpath='{.items[0].status.phase}'"
}

check_desktop(){
    status=$(check_together get_profile_status get_auth_status get_desktop_status get_settings_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 0.5

        status=$(check_together get_profile_status get_auth_status get_desktop_status get_settings_status)
        echo -ne "\rPlease waiting          "
    done
    echo
}

check_bfl(){
    status=$(get_bfl_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for bfl starting ${dot}"
        sleep 0.5

        status=$(get_bfl_status)
        echo -ne "\rWaiting for bfl starting          "
    done
    echo
}

init_mongo_crds(){
    local mongoclusterfile
    mongoclusterfile="${INSTALL_DIR}/crd_mongo_cluster.yaml"

    if [ ! -f ${mongoclusterfile} ]; then
        $sh_c "${KUBECTL} get perconaservermongodbs.psmdb.percona.com -n os-system mongo-cluster -o yaml > ${mongoclusterfile}"
    fi
}

restore_mongo(){
    local pbmName pbmNamePattern restorePattern rsstate cfgstate
    local backupName=$($sh_c "echo $TERMINUS_BACKUP_NAME | rev | cut -d'-' -f2- | rev")
    pbmNamePattern="${KUBECTL} get backupconfigs.sys.bytetrade.io -n os-system ${backupName} -o jsonpath='{.metadata.annotations.percona/psmdb-last-backup-pbmname}'"
    pbmName=$($sh_c "$pbmNamePattern")
    if [  -z "$pbmName" ]; then
        echo "restore mongodb skip ..."
        return
    fi

    echo "preparing for restore mongo backup ${pbmName} ..."

    restorePattern='{"apiVersion":"psmdb.percona.com/v1","kind":"PerconaServerMongoDBRestore","metadata":{"labels": {"managered-by": "mongo-restore-mongo-cluster"},"name":"mongocluster-restore","namespace":"os-system"},"spec":{"backupSource":{"destination":"s3://mongo-backup/'"${pbmName}"'","pbmName":"'"${pbmName}"'","s3":{"bucket":"mongo-backup","credentialsSecret":"mongo-cluster-backup-fakes3","endpointUrl":"http://tapr-s3-svc.os-system:4568","insecureSkipTLSVerify":false,"maxUploadParts":10000,"prefix":"","storageClass":"STANDARD","uploadPartSize":10485760},"type":"physical"},"clusterName":"mongo-cluster","storageName":"s3-local"}}'

    echo "${restorePattern}" > "${INSTALL_DIR}/restore.yaml"

    mstate=$(_get_mongo_cluster_state)
    rsstate=$(_get_sts_mongo_rs_state)
    cfgstate=$(_get_sts_mongo_cfg_state)
    done="0"
    first="1"
    while [ x"$done" != x"1" ]; do
        mstate=$(_get_mongo_cluster_state)
        rstate=$(_get_mongo_restore_state)
        if [ x"$first" == x"1" ]; then
            rsstate=$(_get_sts_mongo_rs_state)
            cfgstate=$(_get_sts_mongo_cfg_state)
        fi

        if [ x"$mstate" == x"ready" ] && [ x"$rstate" == x"ready" ]; then
            done="1"
            break
        fi

        if [ x"$mstate" == x"ready" ] && [ -z "$rstate" ] && [ x"$first" == x"1" ]; then
            if [ -z "$rsstate" ] || [ -z "$cfgstate" ]; then
                sleep 10
                continue
            fi
            sleep 8
            $sh_c "${KUBECTL} apply -f ${INSTALL_DIR}/restore.yaml"
            first="0"
            rsstate=""
            cfgstate=""
            continue
        fi
        if [ -n "$rstate" ]; then
            echo "restore state ${rstate}, please waiting ..."
        fi
        sleep 20
    done
}

_get_sts_mongo_rs_state(){
  local query="${KUBECTL} logs mongo-cluster-rs0-0 -n os-system -c backup-agent |grep 'listening for the commands'"
  local state=$($sh_c "${query}" 2>/dev/null)
  if [ $? -ne 0 ]; then
      echo ""
      return
  fi

  echo "$state"
}

_get_sts_mongo_cfg_state(){
  local query="${KUBECTL} logs mongo-cluster-cfg-0 -n os-system -c backup-agent |grep 'listening for the commands'"
  local state=$($sh_c "${query}" 2>/dev/null)
  if [ $? -ne 0 ]; then
      echo ""
      return
  fi

  echo "$state"
}

_get_mongo_restore_state(){
    sleep 0.5
    local state
    state=$($sh_c "${KUBECTL} get perconaservermongodbrestores.psmdb.percona.com -n os-system mongocluster-restore -o jsonpath='{.status.state}'" 2>/dev/null)
    if [ $? -ne 0 ]; then
        echo ""
        return
    fi
    echo "$state" || "none"
}

_get_mongo_cluster_state(){
    sleep 0.5
    local state
    state=$($sh_c "${KUBECTL} get perconaservermongodbs.psmdb.percona.com -n os-system mongo-cluster -o jsonpath='{.status.state}'" 2>/dev/null)
    if [ $? -ne 0 ]; then
        echo ""
        return
    fi
    echo "$state" || "unknown"
}

check_vault(){
    status=$(get_vault_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 0.5

        status=$(get_vault_status)
        echo -ne "\rPlease waiting          "
    done
    echo
}

repaire_pvs() {
    log_info 'waiting for bfl ready and create pv'
    check_sts_bfl_ready

    local pv sc path storage

    $sh_c "${KUBECTL} get sts -A -l 'tier=bfl' -o jsonpath='{range .items[]}{.metadata.namespace}{\"\n\"}{end}' 2>/dev/null" | \
    while read -r ns; do
        for vl in userspace appcache dbdata; do
            pv=$(get_k8s_annotation "$ns" sts bfl ${vl}_pv)
            path=$(get_k8s_annotation "$ns" sts bfl ${vl}_hostpath)
            sc=$(get_k8s_annotation "$ns" sts bfl ${vl}_sc)
            storage=$(get_k8s_annotation "$ns" sts bfl ${vl}_storage)

            cat > "${INSTALL_DIR}/bfl-${pv}.yaml" <<_END
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: $pv
spec:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: $storage
  hostPath:
    path: $path
    type: DirectoryOrCreate
  persistentVolumeReclaimPolicy: Delete
  volumeMode: Filesystem
  storageClassName: $sc
_END
            ensure_success $sh_c "${KUBECTL} apply -f ${INSTALL_DIR}/bfl-${pv}.yaml"
        done
    done

    log_info 'waiting for app-service ready and create pv'
    check_sts_app_service_ready

    for vl in charts usertmpl; do
        pv=$(get_k8s_annotation os-system sts app-service ${vl}_pv)
        sc=$(get_k8s_annotation os-system sts app-service ${vl}_sc)
        path=$(get_k8s_annotation os-system sts app-service ${vl}_hostpath)
        storage=$(get_k8s_annotation os-system sts app-service ${vl}_storage)

        cat > "${INSTALL_DIR}/appservice-pv-${vl}.yaml" <<_END
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: $pv
spec:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: $storage
  hostPath:
    path: $path
    type: DirectoryOrCreate
  persistentVolumeReclaimPolicy: Delete
  volumeMode: Filesystem
  storageClassName: $sc
_END
        ensure_success $sh_c "${KUBECTL} apply -f ${INSTALL_DIR}/appservice-pv-${vl}.yaml"
    done

    log_info 'waiting for create citus-data-pv'
    
    cat > "${INSTALL_DIR}/citus-data-${pv}.yaml" <<_END
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: citus-data-pv
spec:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: '50Gi'
  hostPath:
    path: /terminus/userdata/dbdata/pg_data
    type: DirectoryOrCreate
  persistentVolumeReclaimPolicy: Delete
  volumeMode: Filesystem
  storageClassName: citus-data-sc

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: citus-data-pvc
  namespace: os-system
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: '50Gi'
  volumeMode: Filesystem
  volumeName: citus-data-pv
  storageClassName: citus-data-sc

---
apiVersion: apr.bytetrade.io/v1alpha1
kind: PGCluster
metadata:
  name: citus
  namespace: os-system
spec:
  replicas: 1
  owner: system
  backupStorage: /terminus/rootfs/middleware-backup/pg_backup
_END
    ensure_success $sh_c "${KUBECTL} apply -f ${INSTALL_DIR}/citus-data-${pv}.yaml"
}

repaire_crd_status() {
  local namespace crd_providerregistries_name crd_applicationpermissions_name patch

  crd_providerregistries_name="providerregistries.sys.bytetrade.io"
  crd_applicationpermissions_name="applicationpermissions.sys.bytetrade.io"
  patch='[{"op":"add","path":"/status","value":{"state":"active"}}]'

  local status=$($sh_c "${KUBECTL} get $crd_providerregistries_name -A --no-headers")

  echo "$status" | while IFS= read line; do
    ns=$(echo "$line" |awk '{print $1}')
    st=$(echo "$line" |awk '{print $6}')
    name=$(echo "$line" |awk '{print $2}')
    if [ x"$st" != x"active" ]; then
      res=$($sh_c "${KUBECTL} patch $crd_providerregistries_name $name -n $ns --type='json' -p='$patch'")
      echo "$res"
    fi
  done

  status=$($sh_c "${KUBECTL} get $crd_applicationpermissions_name -A --no-headers")

  echo "$status" | while IFS= read line; do
    ns=$(echo "$line" |awk '{print $1}')
    st=$(echo "$line" |awk '{print $6}')
    name=$(echo "$line" |awk '{print $2}')
    if [ x"$st" != x"active" ]; then
      res=$($sh_c "${KUBECTL} patch $crd_applicationpermissions_name $name -n $ns --type='json' -p='$patch'")
      echo "$res"
    fi
  done
}

repaire_crd_terminus() {
    local patch
    if [ ! -z "${AWS_SESSION_TOKEN_SETUP}" ]; then
        patch='[{"op":"add","path":"/metadata/annotations/bytetrade.io~1s3-sts","value":"'"$AWS_SESSION_TOKEN_SETUP"'"},{"op":"add","path":"/metadata/annotations/bytetrade.io~1s3-ak","value":"'"$AWS_ACCESS_KEY_ID_SETUP"'"},{"op":"add","path":"/metadata/annotations/bytetrade.io~1s3-sk","value":"'"$AWS_SECRET_ACCESS_KEY_SETUP"'"},{"op":"add","path":"/metadata/annotations/bytetrade.io~1cluster-id","value":"'"$CLUSTER_ID"'"}]'
        $sh_c "${KUBECTL} patch terminus.sys.bytetrade.io terminus -n os-system --type='json' -p='$patch'"
    fi
}

restore_osdata() {
    RESTIC_VERSION=0.15.2

    local rootfs=/terminus/rootfs

    if [ ! -d "${rootfs}/.trash" ]; then
        echo 'juicefs not ready, please check'
        exit $ERR_EXIT
    fi

    log_info 'installing restic ...'
    ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://github.com/restic/restic/releases/download/v${RESTIC_VERSION}/restic_${RESTIC_VERSION}_linux_amd64.bz2"
    ensure_success $sh_c "bunzip2 --force restic_${RESTIC_VERSION}_linux_amd64.bz2"
    ensure_success $sh_c "install restic_${RESTIC_VERSION}_linux_amd64 /usr/local/bin/restic"

    log_info "restoring osdata to ${rootfs}/"

    local target s3Repository snapshotId passwordHash 
    s3Repository="$BACKUP_S3_REPOSITORY"
    snapshotId="$BACKUP_SNAPSHOT_ID"

    if [[ x"$s3Repository" == x"" || x"$snapshotId" == x"" ]]; then
        echo "no env 'BACKUP_S3_REPOSITORY' or 'BACKUP_SNAPSHOT_ID' provided"
        exit $ERR_EXIT
    fi

    # verification is better on the cloud frontend
    # if [[ x"$BACKUP_REPOSITORY_PASSWORD" == x"" || x"$BACKUP_REPOSITORY_PASSWORD_HASH" == x"" ]]; then
    #     echo "no env 'BACKUP_REPOSITORY_PASSWORD' or 'BACKUP_REPOSITORY_PASSWORD_HASH' provided"
    #     exit $ERR_EXIT
    # fi

    # passwordHash=$(printf "$(echo -n "$BACKUP_REPOSITORY_PASSWORD" | base64)" | sha256sum |awk '{print $1}')

    # if [[ x"$passwordHash" != x"$BACKUP_REPOSITORY_PASSWORD_HASH" ]]; then
    #     echo "incorrect password for restore using backup repository"
    #     exit $ERR_EXIT
    # fi

    target="${INSTALL_DIR}/restore"

    ensure_success $sh_c "RESTIC_PASSWORD=$backupPassword AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN=$AWS_SESSION_TOKEN /usr/local/bin/restic -r $s3Repository restore $snapshotId --target ${target}/"
    ensure_success $sh_c "chmod 0755 $target"

    if [ ! -d "${target}/rootfs" ]; then
        echo 'failed to restore juicefs osdata'
        exit $ERR_EXIT
    fi

    ensure_success $sh_c "mv ${target}/rootfs/* ${rootfs}/"
}

install_containerd(){
    if [ x"$KUBE_TYPE" != x"k3s" ]; then
        CONTAINERD_VERSION="1.6.4"
        # preinstall containerd for k8s
        if command_exists containerd && [ -f /etc/systemd/system/containerd.service ];  then
            echo "restart containerd"
            ensure_success $sh_c "rm -rf /etc/containerd"
            ensure_success $sh_c "mkdir -p /etc/containerd"
            ensure_success $sh_c "containerd config default | tee /etc/containerd/config.toml"
            ensure_success $sh_c "sed -i 's/SystemdCgroup \= false/SystemdCgroup \= true/g' /etc/containerd/config.toml"
            ensure_success $sh_c "sed -i 's/k8s.gcr.io\/pause:3.6/kubesphere\/pause:3.5/g' /etc/containerd/config.toml"
            ensure_success $sh_c "sed -i 's/\(\[plugins\.\"io\.containerd\.grpc\.v1\.cri\"\.registry\.mirrors\]\)/\1\n        [plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors.\"docker.io\"]\n          endpoint = [\"http:\/\/$PROXY:5000\"]/' /etc/containerd/config.toml"

            ensure_success $sh_c "curl -L https://raw.githubusercontent.com/containerd/containerd/main/containerd.service -o /etc/systemd/system/containerd.service"
            ensure_success $sh_c "sed -i 's/\(LimitCORE=infinity\)/\1\nLimitNOFILE=1048576/' /etc/systemd/system/containerd.service"
            ensure_success $sh_c "systemctl daemon-reload"
            ensure_success $sh_c "systemctl restart containerd"
            ensure_success $sh_c "systemctl enable --now containerd"

            ctr_cmd=$(command -v ctr)
        else
            echo "install containerd"
            ensure_success $sh_c "wget https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz --no-check-certificate"
            ensure_success $sh_c "tar Cxzvf /usr/local containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz"
            ensure_success $sh_c "wget https://github.com/opencontainers/runc/releases/download/v1.1.3/runc.amd64 --no-check-certificate"
            ensure_success $sh_c "install -m 755 runc.amd64 /usr/local/sbin/runc"
            ensure_success $sh_c "wget https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz --no-check-certificate"
            ensure_success $sh_c "mkdir -p /opt/cni/bin"
            ensure_success $sh_c "tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.1.1.tgz"
            ensure_success $sh_c "rm -rf /etc/containerd"
            ensure_success $sh_c "mkdir -p /etc/containerd"
            ensure_success $sh_c "containerd config default | tee /etc/containerd/config.toml"
            ensure_success $sh_c "sed -i 's/SystemdCgroup \= false/SystemdCgroup \= true/g' /etc/containerd/config.toml"
            ensure_success $sh_c "sed -i 's/k8s.gcr.io\/pause:3.6/kubesphere\/pause:3.5/g' /etc/containerd/config.toml"
            ensure_success $sh_c "sed -i 's/\(\[plugins\.\"io\.containerd\.grpc\.v1\.cri\"\.registry\.mirrors\]\)/\1\n        [plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors.\"docker.io\"]\n          endpoint = [\"http:\/\/$PROXY:5000\"]/' /etc/containerd/config.toml"

            ensure_success $sh_c "curl -L https://raw.githubusercontent.com/containerd/containerd/main/containerd.service -o /etc/systemd/system/containerd.service"
            ensure_success $sh_c "sed -i 's/\(LimitCORE=infinity\)/\1\nLimitNOFILE=1048576/' /etc/systemd/system/containerd.service"
            # ensure_success $sh_c "cp $BASE_DIR/deploy/containerd.service /etc/systemd/system/containerd.service"
            ensure_success $sh_c "systemctl daemon-reload"
            ensure_success $sh_c "systemctl restart containerd"
            ensure_success $sh_c "systemctl enable --now containerd"

            ctr_cmd=$(command -v ctr)
        fi
    fi

    # if [ -d $BASE_DIR/images ]; then
    #     echo "preload images to local ... "
    #     local tar_count=$(find $BASE_DIR/images -type f -name '*.tar.gz'|wc -l)
    #     if [ $tar_count -eq 0 ]; then
    #         if [ -f $BASE_DIR/images/images.mf ]; then
    #             echo "downloading images from terminus cloud ..."
    #             while read img; do
    #                 local filename=$(echo -n "$img"|md5sum|awk '{print $1}')
    #                 filename="$filename.tar.gz"
    #                 echo "downloading ${filename} ..."
    #                 curl -fsSL https://dc3p1870nn3cj.cloudfront.net/${filename} -o $BASE_DIR/images/$filename
    #             done < $BASE_DIR/images/images.mf
    #         fi
    #     fi

    #     if [ x"$KUBE_TYPE" == x"k3s" ]; then
    #         K3S_PRELOAD_IMAGE_PATH="/var/lib/rancher/k3s/agent/images"
    #         $sh_c "mkdir -p ${K3S_PRELOAD_IMAGE_PATH} && rm -rf ${K3S_PRELOAD_IMAGE_PATH}/*"
    #     fi

    #     find $BASE_DIR/images -type f -name '*.tar.gz' | while read filename; do
    #         if [ x"$KUBE_TYPE" == x"k3s" ]; then
    #             local tgz=$(echo "${filename}"|awk -F'/' '{print $NF}')
    #             $sh_c "ln -s ${filename} ${K3S_PRELOAD_IMAGE_PATH}/${tgz}"
    #         else
    #             $sh_c "gunzip -c ${filename} | $ctr_cmd -n k8s.io images import -"
    #         fi
    #     done
    # fi
}

install_k8s() {
    KKE_VERSION=0.1.19

    ensure_success $sh_c "mkdir -p /etc/kke"

    log_info 'Downloading kke installer ...'
    if [ x"$PROXY" != x"" ]; then
	    if [ -f "${HOME}/kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz" ]; then
          ensure_success $sh_c "cp ${HOME}/kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz ${INSTALL_DIR}"
      else
          ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://github.com/beclab/kubekey-ext/releases/download/${KKE_VERSION}/kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz"
      fi
	    ensure_success $sh_c "tar xf kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz"
    else
    	ensure_success $sh_c "curl ${CURL_TRY} -sfL https://raw.githubusercontent.com/beclab/kubekey-ext/master/downloadKKE.sh | VERSION=${KKE_VERSION} bash -"
    fi
    ensure_success $sh_c "chmod +x kk"

    if [[ -z "${TERMINUS_IS_CLOUD_VERSION}" || x"${TERMINUS_IS_CLOUD_VERSION}" != x"true" ]]; then
        log_info 'Installing containerd ...'
        install_containerd
    fi

    log_info 'Installing k8s cluster ...'
    run_install

    log_info 'Waiting for kube pods ready ...'
    check_pods_running

    log_info 'k8s installation is complete'
}

restore_terminus() {
    log_info 'Installing minimal k8s cluster ...'
    install_k8s

    if [ "$storage_type" == "minio" ]; then
        # init minio-operator after etcd installed
        init_minio_cluster
    fi

    log_info 'Installing backup/restore component velero ...'
    ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://github.com/vmware-tanzu/velero/releases/download/v1.11.0/velero-v1.11.0-linux-amd64.tar.gz"
    ensure_success $sh_c "tar xf velero-v1.11.0-linux-amd64.tar.gz"
    ensure_success $sh_c "install velero-v1.11.0-linux-amd64/velero /usr/local/bin"

    VELERO=$(command -v velero)
    
    install_velero

    install_velero_plugin_terminus

    log_info 'Waiting for velero ready ...'
    check_pods_running

    log_info "Restoring k8s resources ..."
    restore_k8s_os

    log_info 'Waiting for ready and repairing pvs ...'
    repaire_pvs

    log_info 'Waiting for ready and repairing crd ...'
    repaire_crd_status
    repaire_crd_terminus

    log_info 'Waiting for bfl ready ...'
    check_bfl

    log_info 'Waiting for vault ready ...'
    check_vault

    log_info 'Waiting for desktop ready ...'
    check_desktop

    log_info 'Waiting for mongo ready ...'
    init_mongo_crds
    restore_mongo
}

INSTALL_DIR=$HOME/.terminus
INSTALL_LOG=$INSTALL_DIR/logs

if [ -d "$INSTALL_LOG" ]; then
    $sh_c "rm -rf $INSTALL_LOG"
fi

mkdir -p $INSTALL_LOG && cd $INSTALL_LOG || exit
fd_errlog=$INSTALL_LOG/errlog_fd_13

Main() {
    log_info 'Restoring Terminus ...\n'
    check_backup_password
    get_distribution
    get_shell_exec

    (
        log_info 'Precheck and Installing dependencies ...\n'
        precheck_os
        install_deps
        config_system

        log_info 'Starting install storage ...\n'
        install_storage

        # restore juicefs data
        log_info 'Restoring juicefs osdata ...'
        restore_osdata

        log_info 'Restoring k8s and os apps ...'
        restore_terminus
        restore_resolv_conf
    ) 2>&1

    ret=$?
    if [ $ret -ne 0 ]; then
        msg="command error occurs, exit with '$ret' directly"
        if [ -f $fd_errlog ]; then
            fderr="$(<$fd_errlog)"
            if [[ x"$fderr" != x"" ]]; then
                msg="$fderr"
            fi
        fi
        log_fatal "$msg"
    fi

    log_info 'All done\n'
}

Main | tee install.log

exit
