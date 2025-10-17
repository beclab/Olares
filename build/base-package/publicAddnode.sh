#!/usr/bin/env bash



ERR_EXIT=1
ERR_VALIDATION=2

CURL_TRY="--retry 5 --retry-delay 1 --retry-max-time 10 "
BASE_DIR=$(dirname $(realpath -s $0))
INSTALL_LOG="$BASE_DIR/logs"

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
			exit 1
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

read_tty() {
    printf "\n  %s" "$1"
    read -r "$2" </dev/tty
}

user_prompt() {
    master_ssh_private_ip="$MASTER_SSH_PRIVATE_IP"
    master_ssh_username="$MASTER_SSH_USERNAME"
    ssh_private_keyfile="$SSH_PRIVATE_KEYFILE"
    master_ssh_port="$MASTER_SSH_PORT"

    if [ -z "$master_ssh_private_ip" ]; then
        while :; do
            read_tty 'master node ssh host(private ip): ' master_ssh_private_ip
            [ -n "$master_ssh_private_ip" ] && break
        done
    fi

    if [ -z "$master_ssh_port" ]; then
        local ssh_port
        printf '\n  master ssh port(default: 22): '
        read -r ssh_port </dev/tty
        if [ x"$ssh_port" == x"" ]; then
            master_ssh_port=22
        else
            master_ssh_port=$ssh_port
        fi
    fi

    if [ -z "$master_ssh_username" ]; then
        while :; do
            read_tty 'master node ssh username: ' master_ssh_username
            [ -n "$master_ssh_username" ] && break
        done
    fi

    if [ -z "$ssh_private_keyfile" ]; then
        while :; do
            read_tty 'master node ssh private keyfile: ' ssh_private_keyfile
            [ -n "$ssh_private_keyfile" ] && break
        done
    fi

}

get_master_info() {
    # get remote master info
    if ! command_exists ssh; then
        echo "no ssh client"
        exit $ERR_EXIT
    fi

    if [ ! -f "$ssh_private_keyfile" ]; then
        echo "ssh private keyfile '$ssh_private_keyfile' not exists"
        exit $ERR_EXIT
    fi
    ensure_success $sh_c "chmod 0600 $ssh_private_keyfile"

    ssh_client="ssh -o StrictHostKeyChecking=no -i $ssh_private_keyfile ${master_ssh_username}@${master_ssh_private_ip}"

    REDIS_PASSWORD=$($ssh_client "sudo su -c 'grep ^requirepass /olares/data/redis/etc/redis.conf'"|awk '{print $NF}')
    if [[ $? -ne 0 || x"$REDIS_PASSWORD" == x"" ]]; then
        echo "no master redis password"
        exit $ERR_EXIT
    fi

    local master_node
    master_node=$($ssh_client "sudo su -c '/usr/local/bin/kubectl get node --no-headers'"|grep master|head -n1)
    if [ x"$master_node" == x"" ]; then
        echo "no k8s master node"
        exit $ERR_EXIT
    fi

    k8s_version=$(echo "$master_node"|awk '{print $NF}')
    if [ x"$k8s_version" == x"" ]; then
        echo "no master k8s version"
        exit $ERR_EXIT
    fi

    KUBE_TYPE="k8s"

    if [[ "$k8s_version" =~ "k3s" ]]; then
        KUBE_TYPE="k3s"
        k8s_version=v1.22.16-k3s
    fi

    master_k8s_nodename=$(echo "$master_node" |awk '{print $1}')
    if [ x"$master_k8s_nodename" == x"" ]; then
        echo "no master k8s nodename"
        exit $ERR_EXIT
    fi

    if [ x"$master_k8s_nodename" == x"$HOSTNAME" ]; then
        echo "Duplicate hostname with master node. Please change the hostname"
        exit $ERR_EXIT
    fi 
}

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

build_socat(){
    SOCAT_VERSION="1.7.3.2"
    # Download
    ensure_success $sh_c "curl -LO http://www.dest-unreach.org/socat/download/socat-${SOCAT_VERSION}.tar.gz"
    ensure_success $sh_c "tar xzvf socat-${SOCAT_VERSION}.tar.gz"
    ensure_success $sh_c "cd socat-${SOCAT_VERSION}"

    ensure_success $sh_c "./configure --prefix=/usr && make -j4 && make install && strip socat"
}

build_contrack(){
    ensure_success $sh_c "curl -LO https://github.com/fqrouter/conntrack-tools/archive/refs/tags/conntrack-tools-1.4.1.tar.gz"
    ensure_success $sh_c "tar zxvf conntrack-tools-1.4.1.tar.gz"
    ensure_success $sh_c "cd conntrack-tools-1.4.1"

    ensure_success $sh_c "./configure --prefix=/usr && make -j4 && make install"
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

    if [ x"${os_type}" != x"Linux" ]; then
        log_fatal "unsupported os type '${os_type}', only supported 'Linux' operating system"
    fi

    if [[ x"${os_arch}" != x"x86_64" && x"${os_arch}" != x"amd64" && x"${os_arch}" != x"aarch64" ]]; then
        log_fatal "unsupported os arch '${os_arch}', only supported 'x86_64' or 'aarch64' architecture"
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
                ensure_success $sh_c "systemctl stop systemd-resolved.service &>/dev/null"
                ensure_success $sh_c "systemctl disable systemd-resolved.service &>/dev/null"

                ensure_success $sh_c "mv /usr/bin/systemd-resolve /usr/bin/systemd-resolve.bak >/dev/null"
                if [ ! -d /run/systemd/resolve ]; then
                    ensure_success $sh_c 'mkdir -p /run/systemd/resolve'
                    ensure_success $sh_c 'touch /run/systemd/resolve/stub-resolv.conf'
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

    if ! hostname -i &>/dev/null; then
        ensure_success $sh_c "echo $local_ip  $HOSTNAME >> /etc/hosts"
    fi

    ensure_success $sh_c "hostname -i &>/dev/null"

    # network and dns
    http_code=$(curl --connect-timeout 30 -ksL -o /dev/null -w "%{http_code}" https://download.docker.com/linux/ubuntu)
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
            local aapv_tar="${BASE_DIR}/components/apparmor_4.0.1-0ubuntu1_${ARCH}.deb"
            if [ ! -f "$aapv_tar" ]; then
                if [ x"${ARCH}" == x"arm64" ]; then
                    ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://launchpad.net/ubuntu/+source/apparmor/4.0.1-0ubuntu1/+build/28428841/+files/apparmor_4.0.1-0ubuntu1_arm64.deb"
                else
                    ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://launchpad.net/ubuntu/+source/apparmor/4.0.1-0ubuntu1/+build/28428840/+files/apparmor_4.0.1-0ubuntu1_amd64.deb"
                fi
            else
                ensure_success $sh_c "cp ${aapv_tar} ./"
            fi
            ensure_success $sh_c "dpkg -i apparmor_4.0.1-0ubuntu1_${ARCH}.deb"
        fi
    fi

    # opy pre-installation dependency files 
    if [ -d /opt/deps ]; then
        ensure_success $sh_c "mv /opt/deps/* ${BASE_DIR}"
    fi

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

install_deps() {
    case "$lsb_dist" in
        ubuntu|debian|raspbian)
            pre_reqs="apt-transport-https ca-certificates curl"
			if ! command -v gpg > /dev/null; then
				pre_reqs="$pre_reqs gnupg"
			fi
            ensure_success $sh_c 'apt-get update -qq &>/dev/null'
            ensure_success $sh_c "DEBIAN_FRONTEND=noninteractive apt-get install -y -qq $pre_reqs &>/dev/null"
            ensure_success $sh_c 'apt-get install -y conntrack socat apache2-utils ntpdate net-tools &>/dev/null'
            ;;

        centos|fedora|rhel)
            if [ "$lsb_dist" = "fedora" ]; then
                pkg_manager="dnf"
            else
                pkg_manager="yum"
            fi

            ensure_success $sh_c "$pkg_manager install -y conntrack socat httpd-tools ntpdate net-tools &>/dev/null"
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
    local cloud="$CLOUD_VENDOR"

    if [ "$cloud" == "aliyun" ]; then
        ensure_success $sh_c 'echo "nameserver 100.100.2.136" > /etc/resolv.conf'
        ensure_success $sh_c 'echo "nameserver 1.0.0.1" >> /etc/resolv.conf'
        ensure_success $sh_c 'echo "nameserver 1.1.1.1" >> /etc/resolv.conf'
    else
        ensure_success $sh_c 'echo "nameserver 1.0.0.1" > /etc/resolv.conf'
        ensure_success $sh_c 'echo "nameserver 1.1.1.1" >> /etc/resolv.conf'
    fi
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

parse_get_master_info() {
    # parse parameters from env or stdin
    user_prompt

    log_info 'get master info'
    get_master_info

    echo
    echo "master_ssh_private_ip     :  $master_ssh_private_ip"
    echo "master_ssh_port           :  $master_ssh_port"
    echo "master_ssh_username       :  $master_ssh_username"
    echo "ssh_private_keyfile       :  $ssh_private_keyfile"
    echo "master_k8s_nodename       :  $master_k8s_nodename"
    echo "REDIS_PASSWORD            :  $REDIS_PASSWORD"
    echo
}

prepare_storage() {
    # master info
    parse_get_master_info

    # storage
    TERMINUS_ROOT="/olares"

    if [ x"$PROXY" != x"" ]; then
	    ensure_success $sh_c "echo 'nameserver $PROXY' > /etc/resolv.conf"
    fi

    storage_type="minio"    # or s3

    if [ x"$STORAGE" != x"" ]; then
        storage_type="$STORAGE"
    fi

    echo "storage_type = ${storage_type}"

    case "$storage_type" in
        minio)
            ;;
        s3)
            echo "s3_bucket = ${S3_BUCKET}"

            if [ "x$S3_BUCKET" == "x" ]; then
                echo "s3 bucket is empty."
                exit $ERR_EXIT
            fi
            ;;
        *)
            echo "storage '$storage_type' not supported."
            exit $ERR_EXIT
    esac

    install_juicefs
}

install_juicefs() {
    JFS_VERSION="v11.1.1"

    log_info 'start to install juicefs'
    local juicefs_data="${TERMINUS_ROOT}/data/juicefs"
    if [ ! -d "$juicefs_data" ]; then
        ensure_success $sh_c "mkdir -p $juicefs_data"
    fi

    local fsname="rootfs"
    local metadb="redis://:${REDIS_PASSWORD}@${master_ssh_private_ip}:6379/1"

    local juicefs_bin="/usr/local/bin/juicefs"
    local jfs_mountpoint="${TERMINUS_ROOT}/${fsname}"
    local jfs_cachedir="${TERMINUS_ROOT}/jfscache"
    [ ! -d $jfs_mountpoint ] && ensure_success $sh_c "mkdir -p $jfs_mountpoint"
    [ ! -d $jfs_cachedir ] && ensure_success $sh_c "mkdir -p $jfs_cachedir"

    if [ ! -f "$juicefs_bin" ]; then
        ensure_success $sh_c "curl ${CURL_TRY} -kLO https://github.com/beclab/juicefs-ext/releases/download/${JFS_VERSION}/juicefs-${JFS_VERSION}-linux-${ARCH}.tar.gz"
        ensure_success $sh_c "tar -zxf juicefs-${JFS_VERSION}-linux-${ARCH}.tar.gz"
        ensure_success $sh_c "chmod +x juicefs"
        ensure_success $sh_c "install juicefs /usr/local/bin"
        ensure_success $sh_c "install juicefs /sbin/mount.juicefs"
    fi

    cat > juicefs.service <<_END
[Unit]
Description=JuicefsMount
Documentation=https://juicefs.com/docs/zh/community/introduction/
Wants=redis-online.target
After=redis-online.target
AssertFileIsExecutable=$juicefs_bin

[Service]
WorkingDirectory=/usr/local

EnvironmentFile=
ExecStart=$juicefs_bin mount -o writeback_cache --entry-cache 300 --attr-cache 300 --cache-dir $jfs_cachedir $metadb $jfs_mountpoint

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
    ensure_success $sh_c "cat juicefs.service > /etc/systemd/system/juicefs.service"

    ensure_success $sh_c "systemctl daemon-reload"
    ensure_success $sh_c "systemctl restart juicefs"
    ensure_success $sh_c "systemctl enable juicefs"

    ensure_success $sh_c "systemctl --no-pager status juicefs"
    ensure_success $sh_c "sleep 3 && test -d ${jfs_mountpoint}/.trash"
}

repeat(){
    for _ in $(seq 1 "$1"); do
        echo -n "$2"
    done
}

check_node_ready(){
    status=$($ssh_client "sudo su -c '/usr/local/bin/kubectl get nodes --no-headers'"|awk "/${HOSTNAME}/{print \$2}")
    n=0
    while [ "x${status}" != x"Ready" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 0.5

        status=$($ssh_client "sudo su -c '/usr/local/bin/kubectl get nodes --no-headers'"|awk "/${HOSTNAME}/{print \$2}")
        echo -ne "\rPlease waiting          "
    done

    echo -e "\n"
    $ssh_client "sudo su -c '/usr/local/bin/kubectl get nodes'"
}

install_containerd(){
    if [ x"$KUBE_TYPE" != x"k3s" ]; then
        CONTAINERD_VERSION="1.6.4"
        RUNC_VERSION="1.1.4"
        CNI_PLUGIN_VERSION="1.1.1"

        # preinstall containerd for k8s
        if command_exists containerd && [ -f /etc/systemd/system/containerd.service ];  then
            ctr_cmd=$(command -v ctr)
            if ! system_service_active "containerd"; then
                ensure_success $sh_c "systemctl start containerd"
            fi
        else
            local containerd_tar="${BASE_DIR}/pkg/containerd/${CONTAINERD_VERSION}/${ARCH}/containerd-${CONTAINERD_VERSION}-linux-${ARCH}.tar.gz"
            local runc_tar="${BASE_DIR}/pkg/runc/v${RUNC_VERSION}/${ARCH}/runc.${ARCH}"
            local cni_plugin_tar="${BASE_DIR}/pkg/cni/v${CNI_PLUGIN_VERSION}/${ARCH}/cni-plugins-linux-${ARCH}-v${CNI_PLUGIN_VERSION}.tgz"

            if [ -f "$containerd_tar" ]; then
                ensure_success $sh_c "cp ${containerd_tar} containerd-${CONTAINERD_VERSION}-linux-${ARCH}.tar.gz"
            else
                ensure_success $sh_c "wget https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-${ARCH}.tar.gz"
            fi
            ensure_success $sh_c "tar Cxzvf /usr/local containerd-${CONTAINERD_VERSION}-linux-${ARCH}.tar.gz"

            if [ -f "$runc_tar" ]; then
                ensure_success $sh_c "cp ${runc_tar} runc.${ARCH}"
            else
                ensure_success $sh_c "wget https://github.com/opencontainers/runc/releases/download/v${RUNC_VERSION}/runc.${ARCH}"
            fi
            ensure_success $sh_c "install -m 755 runc.${ARCH} /usr/local/sbin/runc"

            if [ -f "$cni_plugin_tar" ]; then
                ensure_success $sh_c "cp ${cni_plugin_tar} cni-plugins-linux-${ARCH}-v${CNI_PLUGIN_VERSION}.tgz"
            else
                ensure_success $sh_c "wget https://github.com/containernetworking/plugins/releases/download/v${CNI_PLUGIN_VERSION}/cni-plugins-linux-${ARCH}-v${CNI_PLUGIN_VERSION}.tgz"
            fi
            ensure_success $sh_c "mkdir -p /opt/cni/bin"
            ensure_success $sh_c "tar Cxzvf /opt/cni/bin cni-plugins-linux-${ARCH}-v${CNI_PLUGIN_VERSION}.tgz"
            ensure_success $sh_c "mkdir -p /etc/containerd"
            ensure_success $sh_c "containerd config default | tee /etc/containerd/config.toml"
            ensure_success $sh_c "sed -i 's/SystemdCgroup \= false/SystemdCgroup \= true/g' /etc/containerd/config.toml"
            ensure_success $sh_c "sed -i 's/k8s.gcr.io\/pause:3.6/kubesphere\/pause:3.5/g' /etc/containerd/config.toml"
            rm -rf /tmp/registry.toml
            if [ x"$REGISTRY_MIRRORS" != x"" ]; then
                cat << EOF > /tmp/registry.toml
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["$REGISTRY_MIRRORS"]
EOF
            else
                if [ x"$PROXY" != x"" ]; then
                    cat << EOF > /tmp/registry.toml
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://$PROXY:5000"]
EOF
                fi
            fi

            if [ -f /tmp/registry.toml ]; then
                ensure_success $sh_c "cat /tmp/registry.toml >> /etc/containerd/config.toml"
            fi
            # ensure_success $sh_c "curl -L https://raw.githubusercontent.com/containerd/containerd/main/containerd.service -o /etc/systemd/system/containerd.service"
            ensure_success $sh_c "cp $BASE_DIR/deploy/containerd.service /etc/systemd/system/containerd.service"
            ensure_success $sh_c "systemctl daemon-reload"
            ensure_success $sh_c "systemctl enable --now containerd"

            ctr_cmd=$(command -v ctr)
        fi
    fi

    if [ -d $BASE_DIR/images ]; then
        echo "preload images to local ... "
        local tar_count=$(find $BASE_DIR/images -type f -name '*.tar.gz'|wc -l)
        if [ $tar_count -eq 0 ]; then
            if [ -f $BASE_DIR/images/images.node.mf ]; then
                echo "downloading images from olares cloud ..."
                while read img; do
                    local filename=$(echo -n "$img"|md5sum|awk '{print $1}')
                    filename="$filename.tar.gz"
                    echo "downloading ${filename} ..."
                    curl -fsSL https://cdn.olares.com/${filename} -o $BASE_DIR/images/$filename
                done < $BASE_DIR/images/images.node.mf
            fi
        fi

        if [ x"$KUBE_TYPE" == x"k3s" ]; then
            K3S_PRELOAD_IMAGE_PATH="/var/lib/images"
            $sh_c "mkdir -p ${K3S_PRELOAD_IMAGE_PATH} && rm -rf ${K3S_PRELOAD_IMAGE_PATH}/*"
        fi

        while read img; do
            local filename=$(echo -n "$img"|md5sum|awk '{print $1}')
            filename="$filename.tar.gz"
            if [ x"$KUBE_TYPE" == x"k3s" ]; then
                $sh_c "ln -s $BASE_DIR/images/${filename} ${K3S_PRELOAD_IMAGE_PATH}/${filename}"
            else
                $sh_c "gunzip -c $BASE_DIR/images/${filename} | $ctr_cmd -n k8s.io images import -"
            fi
        done < $BASE_DIR/images/images.node.mf
    fi
}

add_worker_node() {
    # download kke
    KKE_VERSION=0.1.24

    log_info 'add this node to k8s cluster'

    if [ x"$PROXY" != x"" ]; then
	    ensure_success $sh_c "echo nameserver $PROXY > /etc/resolv.conf"
	    ensure_success $sh_c "curl ${CURL_TRY} -kLO https://github.com/beclab/kubekey-ext/releases/download/${KKE_VERSION}/kubekey-ext-v${KKE_VERSION}-linux-${ARCH}.tar.gz"
	    ensure_success $sh_c "tar xf kubekey-ext-v${KKE_VERSION}-linux-${ARCH}.tar.gz"
    else
    	ensure_success $sh_c "curl -sfL https://raw.githubusercontent.com/beclab/kubekey-ext/master/downloadKKE.sh | VERSION=${KKE_VERSION} sh -"
    fi
    ensure_success $sh_c "chmod +x kk"

    add_cmd="./kk add nodes --master-node-name $master_k8s_nodename --master-host $master_ssh_private_ip --master-ssh-user $master_ssh_username"
    add_cmd+=" --master-ssh-private-keyfile $ssh_private_keyfile"
    add_cmd+=" --with-kubernetes $k8s_version --skip-master-pull-images --container-manager containerd"

    if [ x"$PROXY" != x"" ]; then
        add_cmd+=" --registry-mirrors http://${PROXY}:5000"
    fi

    # add env OS_LOCALIP
    export OS_LOCALIP="$local_ip"

    ensure_success $sh_c "$add_cmd"

    log_info 'Waiting for node ready ...'
    check_node_ready

    log_info 'Performing the final configuration ...'
    restore_resolv_conf
    ensure_success $sh_c "sed -i '/${local_ip} $HOSTNAME/d' /etc/hosts"

    # cache versions to file
    ensure_success $sh_c "mkdir -p /etc/kke"
    ensure_success $sh_c "echo 'VERSION=${VERSION}' > /etc/kke/version"
    ensure_success $sh_c "echo 'KKE=${KKE_VERSION}' >> /etc/kke/version"
    ensure_success $sh_c "echo 'KUBE=${k8s_version}' >> /etc/kke/version"

    # clean kube config, and master ssh private key
    $sh_c "rm -f /root/.kube/config"

    log_info 'finished add worker node'
}

if [ -d $INSTALL_LOG ]; then
    $sh_c "rm -rf $INSTALL_LOG"
fi

mkdir -p $INSTALL_LOG && cd $INSTALL_LOG || exit
fd_errlog=$INSTALL_LOG/errlog_fd_13

Main() {
    log_info 'Add worker node for Terminus ...\n'
    get_distribution
    get_shell_exec

    (
        log_info 'Precheck and Installing dependencies ...\n'
        precheck_os
        install_deps
        config_system

        log_info 'Preparing and mount storage fs ... \n'
        prepare_storage

        if [[ -z "${TERMINUS_IS_CLOUD_VERSION}" || x"${TERMINUS_IS_CLOUD_VERSION}" != x"true" ]]; then
            log_info 'Installing containerd ...'
            install_containerd
        fi

        log_info 'Installing and Join worker node ...\n'
        add_worker_node
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
