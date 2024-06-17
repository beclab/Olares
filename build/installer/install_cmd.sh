#!/usr/bin/env bash




ERR_EXIT=1

CURL_TRY="--connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 "

BASE_DIR=$(dirname $(realpath -s $0))

[[ -f "${BASE_DIR}/.env" && -z "$DEBUG_VERSION" ]] && . "${BASE_DIR}/.env"

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

build_socat(){
    SOCAT_VERSION="1.7.3.2"
    local socat_tar="${BASE_DIR}/components/socat-${SOCAT_VERSION}.tar.gz"

    if [ -f "$socat_tar" ]; then
        ensure_success $sh_c "cp ${socat_tar} ./"
    else
        ensure_success $sh_c "curl ${CURL_TRY} -LO http://www.dest-unreach.org/socat/download/socat-${SOCAT_VERSION}.tar.gz"
    fi
    
    ensure_success $sh_c "tar xzvf socat-${SOCAT_VERSION}.tar.gz"
    ensure_success $sh_c "cd socat-${SOCAT_VERSION}"

    ensure_success $sh_c "./configure --prefix=/usr && make -j4 && make install && strip socat"
}

build_contrack(){
    local contrack_tar="${BASE_DIR}/components/conntrack-tools-1.4.1.tar.gz"
    if [ -f "$contrack_tar" ]; then
        ensure_success $sh_c "cp ${contrack_tar} ./"
    else
        ensure_success $sh_c "curl ${CURL_TRY} -LO https://github.com/fqrouter/conntrack-tools/archive/refs/tags/conntrack-tools-1.4.1.tar.gz"
    fi
    
    ensure_success $sh_c "tar zxvf conntrack-tools-1.4.1.tar.gz"
    ensure_success $sh_c "cd conntrack-tools-1.4.1"

    ensure_success $sh_c "./configure --prefix=/usr && make -j4 && make install"
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

precheck_os() {
    local ip os_type os_arch

    # check os type and arch and os vesion
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
    printf "%s\t%s\n\n" "$ip" "$HOSTNAME"

    if [[ x"$ip" == x"" || "$ip" == @("172.17.0.1"|"127.0.0.1"|"127.0.1.1") || ! "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_fatal "incorrect ip for hostname '$HOSTNAME', please check"
    fi

    local_ip="$ip"

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
            local aapv_tar="${BASE_DIR}/components/apparmor_4.0.1-0ubuntu1_amd64.deb"
            if [ ! -f "$aapv_tar" ]; then
                ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://launchpad.net/ubuntu/+source/apparmor/4.0.1-0ubuntu1/+build/28428840/+files/apparmor_4.0.1-0ubuntu1_amd64.deb"
            else
                ensure_success $sh_c "cp ${aapv_tar} ./"
            fi
            ensure_success $sh_c "dpkg -i apparmor_4.0.1-0ubuntu1_amd64.deb"
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

config_system() {
    local ntpdate hwclock

    # kernel printk log level
    ensure_success $sh_c 'sysctl -w kernel.printk="3 3 1 7"'

    # ntp sync
    ntpdate=$(command -v ntpdate)
    hwclock=$(command -v hwclock)

    printf '#!/bin/sh\n\n%s -b -u pool.ntp.org && %s -w\n\nexit 0\n' "$ntpdate" "$hwclock" > cron.ntpdate
    ensure_success $sh_c '/bin/sh cron.ntpdate'
    ensure_success $sh_c 'cat cron.ntpdate > /etc/cron.daily/ntpdate && chmod 0700 /etc/cron.daily/ntpdate'
    ensure_success rm -f cron.ntpdate
}

config_proxy_resolv_conf() {
    if [ x"$PROXY" == x"" ]; then
        return
    fi
	ensure_success $sh_c "echo nameserver $PROXY > /etc/resolv.conf"
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
        if [[ x"$PROXY" != x"" && x"$ns" == x"$PROXY" ]]; then
            config_resolv_conf
        else
            ensure_success $sh_c "cat /etc/resolv.conf.bak > /etc/resolv.conf"
        fi
    fi
}

run_install() {
    k8s_version=v1.22.10
    ks_version=v3.3.0

    log_info 'installing k8s and kubesphere'

    if [ -d "$BASE_DIR/pkg" ]; then
        ensure_success $sh_c "cp -a ${BASE_DIR}/pkg ./"
    fi

    # env 'KUBE_TYPE' is specific the special kubernetes (k8s or k3s), default k3s
    if [ x"$KUBE_TYPE" == x"k3s" ]; then
        k8s_version=v1.21.4-k3s
    fi
    create_cmd="./kk create cluster --with-kubernetes $k8s_version --with-kubesphere $ks_version --container-manager containerd"  # --with-addon ${ADDON_CONFIG_FILE}

    local extra

    # env 'REGISTRY_MIRRORS' is a docker image cache mirrors, separated by commas
    if [ x"$REGISTRY_MIRRORS" != x"" ]; then
        extra=" --registry-mirrors $REGISTRY_MIRRORS"
    fi
    # env 'PROXY' is a cache proxy server, to download binaries and container images
    if [ x"$PROXY" != x"" ]; then
        # download binary with cache proxy
        if [ x"$KUBE_TYPE" != x"k3s" ];then
            ensure_success $sh_c "./kk create phase os"
            ensure_success $sh_c "./kk create phase binary --with-kubernetes $k8s_version --download-cmd 'curl ${CURL_TRY} -kL -o %s %s'"
        else
            create_cmd+=" --download-cmd 'curl ${CURL_TRY} -kL -o %s %s'"
        fi

        restore_resolv_conf
        extra=" --registry-mirrors http://${PROXY}:5000"
    fi
    create_cmd+=" $extra"

    # add env OS_LOCALIP
    export OS_LOCALIP="$local_ip"

    ensure_success $sh_c "$create_cmd"

    log_info 'k8s and kubesphere installation is complete'

    # cache version to file
    ensure_success $sh_c "echo 'VERSION=${VERSION}' > /etc/kke/version"
    ensure_success $sh_c "echo 'KKE=${KKE_VERSION}' >> /etc/kke/version"
    ensure_success $sh_c "echo 'KUBE=${k8s_version}' >> /etc/kke/version"

    # setup after kubesphere is installed
    export KUBECONFIG=/root/.kube/config  # for ubuntu
    HELM=$(command -v helm)
    KUBECTL=$(command -v kubectl)

    check_kscm # wait for ks launch

    if [ "x${LOCAL_GPU_ENABLE}" == "x1" ]; then
        install_gpu
    fi

    ensure_success $sh_c "sed -i '/${local_ip} $HOSTNAME/d' /etc/hosts"

    if [ x"$KUBE_TYPE" == x"k3s" ]; then
        retry_cmd $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-k3s.yaml"
        if [[ ! -z "${K3S_PRELOAD_IMAGE_PATH}" && -d $K3S_PRELOAD_IMAGE_PATH ]]; then
            # remove the preload image path to make sure images will not be reloaded after reboot
            ensure_success $sh_c "rm -rf ${K3S_PRELOAD_IMAGE_PATH}"
        fi
    fi

    log_info 'Installing account ...'
    # add the first account
    retry_cmd $sh_c "${HELM} upgrade -i account ${BASE_DIR}/wizard/config/account --force"

    log_info 'Installing settings ...'
    ensure_success $sh_c "${HELM} upgrade -i settings ${BASE_DIR}/wizard/config/settings --force"

    # install gpu if necessary
    if [[ "x${GPU_ENABLE}" == "x1" && "x${GPU_DOMAIN}" != "x" ]]; then
        log_info 'Installing gpu ...'

        if [ x"$KUBE_TYPE" == x"k3s" ]; then
            ensure_success $sh_c "${HELM} upgrade -i gpu ${BASE_DIR}/wizard/config/gpu -n gpu-system --force --set gpu.server=${GPU_DOMAIN} --set container.manager=k3s --create-namespace"
            ensure_success $sh_c "mkdir -p /var/lib/rancher/k3s/agent/etc/containerd"
            ensure_success $sh_c "cp ${BASE_DIR}/deploy/orion-config.toml.tmpl /var/lib/rancher/k3s/agent/etc/containerd/config.toml.tmpl" 
            ensure_success $sh_c "systemctl restart k3s"

            check_ksredis
            check_kscm
            check_ksapi

            # waiting for kubesphere webhooks starting
            sleep 30
        else
            ensure_success $sh_c "${HELM} upgrade -i gpu ${BASE_DIR}/wizard/config/gpu -n gpu-system --force --set gpu.server=${GPU_DOMAIN} --set container.manager=containerd --create-namespace"
        fi

        check_orion_gpu
    fi

    GPU_TYPE="none"
    if [ "x${LOCAL_GPU_ENABLE}" == "x1" ]; then  
        GPU_TYPE="nvidia"
        if [ "x${LOCAL_GPU_SHARE}" == "x1" ]; then  
            GPU_TYPE="nvshare"
        fi
    fi
    if [ "x${GPU_ENABLE}" == "x1" ]; then
        GPU_TYPE="virtaitech"
    fi

    local bucket="none"
    if [ "x${S3_BUCKET}" != "x" ]; then
        bucket="${S3_BUCKET}"
    fi

    # add ownerReferences of user
    log_info 'Installing appservice ...'
    local ks_redis_pwd=$($sh_c "${KUBECTL} get secret -n kubesphere-system redis-secret -o jsonpath='{.data.auth}' |base64 -d")
    retry_cmd $sh_c "${HELM} upgrade -i system ${BASE_DIR}/wizard/config/system -n os-system --force \
        --set kubesphere.redis_password=${ks_redis_pwd} --set backup.bucket=\"${BACKUP_CLUSTER_BUCKET}\" \
        --set backup.key_prefix=\"${BACKUP_KEY_PREFIX}\" --set backup.is_cloud_version=\"${TERMINUS_IS_CLOUD_VERSION}\" \
        --set backup.sync_secret=\"${BACKUP_SECRET}\" --set gpu=\"${GPU_TYPE}\" --set s3_bucket=\"${S3_BUCKET}\""

    # save backup env to configmap
    cat > cm-backup-config.yaml << _END
apiVersion: v1
data:
  terminus.cloudVersion: "${TERMINUS_IS_CLOUD_VERSION}"
  backup.clusterBucket: "${BACKUP_CLUSTER_BUCKET}"
  backup.keyPrefix: "${BACKUP_KEY_PREFIX}"
  backup.secret: "${BACKUP_SECRET}"
kind: ConfigMap
metadata:
  name: backup-config
  namespace: os-system
_END
    ensure_success $sh_c "$KUBECTL apply -f cm-backup-config.yaml"

    # patch
    ensure_success $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-globalrole-workspace-manager.yaml"
    ensure_success $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-notification-manager.yaml"

    # install app-store charts repo to app sevice
    log_info 'waiting for appservice'
    check_appservice
    appservice_pod=$(get_appservice_pod)

    # gen bfl app key and secret
    bfl_ks=($(get_app_key_secret "bfl"))

    log_info 'Installing launcher ...'
    # install launcher , and init pv
    ensure_success $sh_c "${HELM} upgrade -i launcher-${username} ${BASE_DIR}/wizard/config/launcher -n user-space-${username} --force --set bfl.appKey=${bfl_ks[0]} --set bfl.appSecret=${bfl_ks[1]}"

    log_info 'waiting for bfl'
    check_bfl
    bfl_node=$(get_bfl_node)
    bfl_doc_url=$(get_bfl_url)

    ns="user-space-${username}"

    log_info 'Try to find pv ...'
    userspace_pvc=$(get_k8s_annotation "$ns" sts bfl userspace_pvc)
    userspace_hostpath=$(get_k8s_annotation "$ns" sts bfl userspace_hostpath)
    appcache_hostpath=$(get_k8s_annotation "$ns" sts bfl appcache_hostpath)
    dbdata_hostpath=$(get_k8s_annotation "$ns" sts bfl dbdata_hostpath)

    # generate apps charts values.yaml
    # TODO: infisical password
    app_perm_settings=$(get_app_settings)
    cat ${BASE_DIR}/wizard/config/launcher/values.yaml > ${BASE_DIR}/wizard/config/apps/values.yaml
    cat << EOF >> ${BASE_DIR}/wizard/config/apps/values.yaml
  url: '${bfl_doc_url}'
  nodeName: ${bfl_node}
pvc:
  userspace: ${userspace_pvc}
userspace:
  userData: ${userspace_hostpath}/Home
  appData: ${userspace_hostpath}/Data
  appCache: ${appcache_hostpath}
  dbdata: ${dbdata_hostpath}
desktop:
  nodeport: 30180
global:
  bfl:
    username: '${username}'


debugVersion: ${DEBUG_VERSION}
gpu: ${GPU_TYPE}

os:
  ${app_perm_settings}
EOF

    log_info 'Installing built-in apps ...'

    for appdir in "${BASE_DIR}/wizard/config/apps"/*/; do
      if [ -d "$appdir" ]; then
        releasename=$(basename "$appdir")
        ensure_success $sh_c "${HELM} upgrade -i ${releasename} ${appdir} -n user-space-${username} --force --set kubesphere.redis_password=${ks_redis_pwd} -f ${BASE_DIR}/wizard/config/apps/values.yaml"
      fi
    done

    # log_info 'Installing user console ...'
    # ensure_success $sh_c "${HELM} upgrade -i console-${username} ${BASE_DIR}/wizard/config/console -n user-space-${username} --set bfl.username=${username}"

    # clear apps values.yaml
    cat /dev/null > ${BASE_DIR}/wizard/config/apps/values.yaml
    cat /dev/null > ${BASE_DIR}/wizard/config/launcher/values.yaml
    copy_charts=("launcher" "apps")
    for cc in "${copy_charts[@]}"; do
        ensure_success $sh_c "${KUBECTL} cp ${BASE_DIR}/wizard/config/${cc} os-system/${appservice_pod}:/userapps"
    done

    log_info 'Performing the final configuration ...'
    # delete admin user after kubesphere installed,
    # admin user creating in the ks-install image should be modified.
    ensure_success $sh_c "${KUBECTL} patch user admin -p '{\"metadata\":{\"finalizers\":[\"finalizers.kubesphere.io/users\"]}}' --type='merge'"
    ensure_success $sh_c "${KUBECTL} delete user admin"
    ensure_success $sh_c "${KUBECTL} delete deployment kubectl-admin -n kubesphere-controls-system"
    ensure_success $sh_c "${KUBECTL} scale deployment/ks-installer --replicas=0 -n kubesphere-system"

    # delete storageclass accessor webhook
    ensure_success $sh_c "${KUBECTL} delete validatingwebhookconfigurations storageclass-accessor.storage.kubesphere.io"

    # calico config for tailscale
    ensure_success $sh_c "${KUBECTL} patch felixconfiguration default -p '{\"spec\":{\"featureDetectOverride\": \"SNATFullyRandom=false,MASQFullyRandom=false\"}}' --type='merge'"
}

install_storage() {
    TERMINUS_ROOT="/terminus"
    storage_type="minio"    # or s3

    if [[ ! -z "${TERMINUS_IS_CLOUD_VERSION}" && x"${TERMINUS_IS_CLOUD_VERSION}" == x"true" ]]; then
        local DATA_DIR="/osdata"
        if [ -d $DATA_DIR ]; then
            if [[ -d $TERMINUS_ROOT || -f $TERMINUS_ROOT ]]; then
                $sh_c "rm -rf $TERMINUS_ROOT"
            fi

            ensure_success $sh_c "mkdir -p $DATA_DIR$TERMINUS_ROOT"
            ensure_success $sh_c "ln -s $DATA_DIR$TERMINUS_ROOT $TERMINUS_ROOT"

        fi
    fi


    log_info 'Preparing object storage ...\n'

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

    local minio_tar="${BASE_DIR}/components/minio.${MINIO_VERSION}"
    local minio_bin="/usr/local/bin/minio"
    local minio_data="${TERMINUS_ROOT}/data/minio/vol1"

    [ ! -d "$minio_data" ] && ensure_success $sh_c "mkdir -p $minio_data"

    if [ ! -f "$minio_bin" ]; then
        if [ -f "$minio_tar" ]; then
            ensure_success $sh_c "cp ${minio_tar} minio"
        else
            ensure_success $sh_c "curl ${CURL_TRY} -kLo minio https://dl.min.io/server/minio/release/linux-amd64/archive/minio.${MINIO_VERSION}"
        fi
        ensure_success $sh_c "chmod +x minio"
        ensure_success $sh_c "install minio /usr/local/bin"
    fi

    cat > minio.service <<_END
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

        ensure_success $sh_c "cat minio.service > /etc/systemd/system/minio.service"
        cat > minio.env <<_END
# MINIO_ROOT_USER and MINIO_ROOT_PASSWORD sets the root account for the MinIO server.
# This user has unrestricted permissions to perform S3 and administrative API operations on any resource in the deployment.
# Omit to use the default values 'minioadmin:minioadmin'.
# MinIO recommends setting non-default values as a best practice, regardless of environment
MINIO_VOLUMES="$minio_data"
MINIO_OPTS="--console-address ${local_ip}:9090 --address ${local_ip}:9000"

MINIO_ROOT_USER=$MINIO_ROOT_USER
MINIO_ROOT_PASSWORD=$MINIO_ROOT_PASSWORD
_END
    ensure_success $sh_c "cat minio.env > /etc/default/minio"

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

    local minio_operator_tar="${BASE_DIR}/components/minio-operator-${MINIO_OPERATOR_VERSION}-linux-amd64.tar.gz"
    local minio_operator_bin="/usr/local/bin/minio-operator"

    if [ ! -f "$minio_operator_bin" ]; then
        if [ -f "$minio_operator_tar" ]; then
            ensure_success $sh_c "cp ${minio_operator_tar} minio-operator-${MINIO_OPERATOR_VERSION}-linux-amd64.tar.gz"
        else
            ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://github.com/beclab/minio-operator/releases/download/${MINIO_OPERATOR_VERSION}/minio-operator-${MINIO_OPERATOR_VERSION}-linux-amd64.tar.gz"
        fi
	      ensure_success $sh_c "tar zxf minio-operator-${MINIO_OPERATOR_VERSION}-linux-amd64.tar.gz"
        ensure_success $sh_c "install -m 755 minio-operator $minio_operator_bin"
    fi

    ensure_success $sh_c "$minio_operator_bin init --address $local_ip --cafile /etc/ssl/etcd/ssl/ca.pem --certfile /etc/ssl/etcd/ssl/node-$HOSTNAME.pem --keyfile /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem --volume $MINIO_VOLUMES --password $MINIO_ROOT_PASSWORD"
}

install_redis() {
    REDIS_VERSION=5.0.14
    REDIS_PASSWORD=$(random_string 16)

    log_info 'start to install redis'

    local redis_tar="${BASE_DIR}/components/redis-${REDIS_VERSION}.tar.gz"
    local redis_root="${TERMINUS_ROOT}/data/redis"
    local redis_conf="${redis_root}/etc/redis.conf"
    local redis_bin="/usr/bin/redis-server"
    local cpu_cores

    # install redis, if redis-server not exists
    if [ ! -f "$redis_bin" ]; then
        if [ -f "$redis_tar" ]; then
            ensure_success $sh_c "cp ${redis_tar} redis-${REDIS_VERSION}.tar.gz"
        else
            ensure_success $sh_c "curl -kLO https://download.redis.io/releases/redis-${REDIS_VERSION}.tar.gz"
        fi
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

    cat > redis.conf <<_END
protected-mode no
bind $local_ip
port 6379
daemonize no
supervised no
pidfile ${redis_root}/run/redis.pid
logfile ${redis_root}/log/redis-server.log
save 900 1
save 600 50
save 300 100
save 180 300
save 60 1000
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir ${redis_root}/data
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 32mb
requirepass $REDIS_PASSWORD
_END
    ensure_success $sh_c "cat redis.conf > $redis_conf"
    ensure_success $sh_c "chmod 0640 $redis_conf"

    cat > redis-server.service <<_END
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
ExecStartPre=/bin/sh -c "test -f /sys/kernel/mm/transparent_hugepage/enabled && /bin/echo never > /sys/kernel/mm/transparent_hugepage/enabled; test -f ${redis_root}/data/appendonly.aof && (echo y | /usr/local/bin/redis-check-aof --fix ${redis_root}/data/appendonly.aof); true"
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

    ensure_success $sh_c "cat redis-server.service > /etc/systemd/system/redis-server.service"
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
    JFS_VERSION="v11.1.0"

    log_info 'start to install juicefs'

    local juicefs_data="${TERMINUS_ROOT}/data/juicefs"
    if [ ! -d "$juicefs_data" ]; then
        ensure_success $sh_c "mkdir -p $juicefs_data"
    fi

    bucket="terminus"

    local format_cmd
    local fsname="rootfs"
    local metadb="redis://:${REDIS_PASSWORD}@${local_ip}:6379/1"
    local ak="$ACCESS_KEY"
    local sk="$SECRET_KEY"

    local juicefs_tar="${BASE_DIR}/components/juicefs-${JFS_VERSION}-linux-amd64.tar.gz"
    local juicefs_bin="/usr/local/bin/juicefs"
    local jfs_mountpoint="${TERMINUS_ROOT}/${fsname}"
    local jfs_cachedir="${TERMINUS_ROOT}/jfscache"
    [ ! -d $jfs_mountpoint ] && ensure_success $sh_c "mkdir -p $jfs_mountpoint"
    [ ! -d $jfs_cachedir ] && ensure_success $sh_c "mkdir -p $jfs_cachedir"

    if [ ! -f "$juicefs_bin" ]; then
        if [ -f "$juicefs_tar" ]; then
            ensure_success $sh_c "cp ${juicefs_tar} juicefs-${JFS_VERSION}-linux-amd64.tar.gz"
        else
            ensure_success $sh_c "curl ${CURL_TRY} -kLO https://github.com/beclab/juicefs-ext/releases/download/${JFS_VERSION}/juicefs-${JFS_VERSION}-linux-amd64.tar.gz"
        fi
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
        format_cmd+=" $fsname --trash-days 0" # debug
        ensure_success $sh_c "$format_cmd"
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

random_string() {
    local length=12
    local alphanumeric="abc2def3gh4jk5mn6pqr7st8uvw9xyz"

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

pull_velero_image() {
    local count
    local velero_ver=$1
    count=$(_check_velero_image_exists "$velero_ver")
    if [ x"$count" == x"0" ]; then
        echo "pull velero image $velero_ver ..."
        ensure_success $sh_c "$CRICTL pull docker.io/beclab/velero:${velero_ver} &>/dev/null;true"
    fi

    while [ "$count" -lt 1 ]; do
        sleep 3
        count=$(_check_velero_image_exists "$velero_ver")
    done
    echo
}

_check_velero_image_exists() {
  local exists=0
  local ver=$1
  local res=$($sh_c "${CRICTL} images |grep 'velero ' 2>/dev/null")
  if [ "$?" -ne 0 ]; then
      echo "0"
  fi
  exists=$(echo "$res" | while IFS= read -r line; do
      linev=$(echo $line |awk '{print $2}')
      if [ "$linev" == "$ver" ]; then
          echo 1
          break
      fi
  done)

  if [ -z "$exists" ]; then
      exists=0
  fi

  echo "${exists}"
}

pull_velero_plugin_image() {
    local count
    local velero_plugin_ver=$1
    count=$(_check_velero_plugin_image_exists "$velero_plugin_ver")
    if [ x"$count" == x"0" ]; then
        echo "pull velero-plugin image $velero_plugin_ver ..."
        ensure_success $sh_c "$CRICTL pull docker.io/beclab/velero-plugin-for-terminus:${velero_plugin_ver} &>/dev/null;true"
    fi

    while [ "$count" -lt 1 ]; do
        sleep 3
        count=$(_check_velero_plugin_image_exists "$velero_plugin_ver")
    done
    echo
}

_check_velero_plugin_image_exists() {
  local exists=0
  local ver=$1
  local query="${CRICTL} images"
  local res=$($sh_c "${CRICTL} images |grep 'velero-plugin-for-terminus' 2>/dev/null")
  if [ "$?" -ne 0 ]; then
      echo "0"
  fi

  exists=$(echo "$res" | while IFS= read -r line; do
      linev=$(echo $line |awk '{print $2}')
      if [ "$linev" == "$ver" ]; then
          echo 1
          break
      fi
  done)

  if [ -z "$exists" ]; then
      exists=0
  fi

  echo "$exists"
}

install_velero() {
    config_proxy_resolv_conf

    VELERO_VERSION="v1.11.0"
    local velero_tar="${BASE_DIR}/components/velero-${VELERO_VERSION}-linux-amd64.tar.gz"
    if [ -f "$velero_tar" ]; then
        ensure_success $sh_c "cp ${velero_tar} velero-${VELERO_VERSION}-linux-amd64.tar.gz"
    else
        ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://github.com/vmware-tanzu/velero/releases/download/v1.11.0/velero-v1.11.0-linux-amd64.tar.gz"
    fi
    ensure_success $sh_c "tar xf velero-v1.11.0-linux-amd64.tar.gz"
    ensure_success $sh_c "install velero-v1.11.0-linux-amd64/velero /usr/local/bin"

    CRICTL=$(command -v crictl)
    VELERO=$(command -v velero)

    # install velero crds
    ensure_success $sh_c "${VELERO} install --crds-only"
    restore_resolv_conf
}

install_velero_plugin_terminus() {
  local region provider namespace bucket storage_location
  local plugin velero_storage_location_install_cmd velero_plugin_install_cmd
  local msg
  provider="terminus"
  namespace="os-system"
  storage_location="terminus-cloud"
  bucket="terminus-cloud"
  velero_ver="v1.11.1"
  velero_plugin_ver="v1.0.2"

  if [[ "$provider" == x"" || "$namespace" == x"" || "$bucket" == x"" || "$velero_ver" == x"" || "$velero_plugin_ver" == x"" ]]; then
    echo "Backup plugin install params invalid."
    exit $ERR_EXIT
  fi

  pull_velero_image "$velero_ver"
  pull_velero_plugin_image "$velero_plugin_ver"

  terminus_backup_location=$($sh_c "${VELERO} backup-location get -n os-system | awk '\$1 == \"${storage_location}\" {count++} END{print count}'")
  if [[ ${terminus_backup_location} == x"" || ${terminus_backup_location} -lt 1 ]]; then
    velero_storage_location_install_cmd="${VELERO} backup-location create $storage_location"
    velero_storage_location_install_cmd+=" --provider $provider --namespace $namespace"
    velero_storage_location_install_cmd+=" --prefix \"\" --bucket $bucket"
    msg=$($sh_c "$velero_storage_location_install_cmd 2>&1")
  fi

  if [[ ! -z $msg && $msg != *"successfully"* && $msg != *"exists"* ]]; then
    log_info "$msg"
  fi

  sleep 0.5

  velero_plugin_terminus=$($sh_c "${VELERO} plugin get -n os-system |grep 'velero.io/terminus' |wc -l")
  if [[ ${velero_plugin_terminus} == x"" || ${velero_plugin_terminus} -lt 1 ]]; then
    velero_plugin_install_cmd="${VELERO} install"
    velero_plugin_install_cmd+=" --no-default-backup-location --namespace $namespace"
    velero_plugin_install_cmd+=" --image beclab/velero:$velero_ver --use-volume-snapshots=false"
    velero_plugin_install_cmd+=" --no-secret --plugins beclab/velero-plugin-for-terminus:$velero_plugin_ver"
    velero_plugin_install_cmd+=" --velero-pod-cpu-request=50m --velero-pod-cpu-limit=500m"
    velero_plugin_install_cmd+=" --node-agent-pod-cpu-request=50m --node-agent-pod-cpu-limit=500m"
    velero_plugin_install_cmd+=" --wait"
    ensure_success $sh_c "$velero_plugin_install_cmd"
    velero_plugin_install_cmd="${VELERO} plugin add beclab/velero-plugin-for-terminus:$plugin -n os-system"
    msg=$($sh_c "$velero_plugin_install_cmd 2>&1")
  fi

  if [[ ! -z $msg && $msg != *"Duplicate"*  ]]; then
    log_info "$msg"
  fi

  local velero_patch
  velero_patch='[{"op":"replace","path":"/spec/template/spec/volumes","value": [{"name":"plugins","emptyDir":{}},{"name":"scratch","emptyDir":{}},{"name":"terminus-cloud","hostPath":{"path":"/terminus/rootfs/k8s-backup", "type":"DirectoryOrCreate"}}]},{"op": "replace", "path": "/spec/template/spec/containers/0/volumeMounts", "value": [{"name":"plugins","mountPath":"/plugins"},{"name":"scratch","mountPath":"/scratch"},{"mountPath":"/data","name":"terminus-cloud"}]},{"op": "replace", "path": "/spec/template/spec/containers/0/securityContext", "value": {"privileged": true, "runAsNonRoot": false, "runAsUser": 0}}]'

  msg=$($sh_c "${KUBECTL} patch deploy velero -n os-system --type='json' -p='$velero_patch'")
  if [[ ! -z $msg && $msg != *"patched"* ]]; then
    log_info "Backup plugin patched error: $msg"
  else
    echo "Backup plugin patched succeed"
  fi
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
            local containerd_tar="${BASE_DIR}/pkg/containerd/${CONTAINERD_VERSION}/amd64/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz"
            local runc_tar="${BASE_DIR}/pkg/runc/v${RUNC_VERSION}/amd64/runc.amd64"
            local cni_plugin_tar="${BASE_DIR}/pkg/cni/v${CNI_PLUGIN_VERSION}/amd64/cni-plugins-linux-amd64-v${CNI_PLUGIN_VERSION}.tgz"

            if [ -f "$containerd_tar" ]; then
                ensure_success $sh_c "cp ${containerd_tar} containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz"
            else
                ensure_success $sh_c "wget https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz"
            fi
            ensure_success $sh_c "tar Cxzvf /usr/local containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz"

            if [ -f "$runc_tar" ]; then
                ensure_success $sh_c "cp ${runc_tar} runc.amd64"
            else
                ensure_success $sh_c "wget https://github.com/opencontainers/runc/releases/download/v${RUNC_VERSION}/runc.amd64"
            fi
            ensure_success $sh_c "install -m 755 runc.amd64 /usr/local/sbin/runc"

            if [ -f "$cni_plugin_tar" ]; then
                ensure_success $sh_c "cp ${cni_plugin_tar} cni-plugins-linux-amd64-v${CNI_PLUGIN_VERSION}.tgz"
            else
                ensure_success $sh_c "wget https://github.com/containernetworking/plugins/releases/download/v${CNI_PLUGIN_VERSION}/cni-plugins-linux-amd64-v${CNI_PLUGIN_VERSION}.tgz"
            fi
            ensure_success $sh_c "mkdir -p /opt/cni/bin"
            ensure_success $sh_c "tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v${CNI_PLUGIN_VERSION}.tgz"
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
            if [ -f $BASE_DIR/images/images.mf ]; then
                echo "downloading images from terminus cloud ..."
                while read img; do
                    local filename=$(echo -n "$img"|md5sum|awk '{print $1}')
                    filename="$filename.tar.gz"
                    echo "downloading ${filename} ..."
                    curl -fsSL https://dc3p1870nn3cj.cloudfront.net/${filename} -o $BASE_DIR/images/$filename
                done < $BASE_DIR/images/images.mf
            fi
        fi

        if [ x"$KUBE_TYPE" == x"k3s" ]; then
            K3S_PRELOAD_IMAGE_PATH="/var/lib/rancher/k3s/agent/images"
            $sh_c "mkdir -p ${K3S_PRELOAD_IMAGE_PATH} && rm -rf ${K3S_PRELOAD_IMAGE_PATH}/*"
        fi

        find $BASE_DIR/images -type f -name '*.tar.gz' | while read filename; do
            if [ x"$KUBE_TYPE" == x"k3s" ]; then
                local tgz=$(echo "${filename}"|awk -F'/' '{print $NF}')
                $sh_c "ln -s ${filename} ${K3S_PRELOAD_IMAGE_PATH}/${tgz}"
            else
                $sh_c "gunzip -c ${filename} | $ctr_cmd -n k8s.io images import -"
            fi
        done
    fi
}

install_k8s_ks() {
    KKE_VERSION=0.1.20

    ensure_success $sh_c "mkdir -p /etc/kke"
    local kk_bin="${BASE_DIR}/components/kk"
    local kk_tar="${BASE_DIR}/components/kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz"

    if [ ! -f "$kk_bin" ]; then
        if [ ! -f "$kk_tar" ]; then
            if [ x"$PROXY" != x"" ]; then
              ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://github.com/beclab/kubekey-ext/releases/download/${KKE_VERSION}/kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz"
              ensure_success $sh_c "tar xf kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz"
            else
              ensure_success $sh_c "curl ${CURL_TRY} -sfL https://raw.githubusercontent.com/beclab/kubekey-ext/master/downloadKKE.sh | VERSION=${KKE_VERSION} sh -"
            fi
        else
            ensure_success $sh_c "cp ${kk_tar} kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz"
            ensure_success $sh_c "tar xf kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz"
        fi
    else 
        ensure_success $sh_c "cp ${kk_bin} ./"
    fi
    ensure_success $sh_c "chmod +x kk"

    log_info 'Setup your first user ...\n'
    setup_ws

    # generate init config
    ADDON_CONFIG_FILE=${BASE_DIR}/wizard/bin/init-config.yaml
    echo '
    ' > ${ADDON_CONFIG_FILE}

    if [[ -z "${TERMINUS_IS_CLOUD_VERSION}" || x"${TERMINUS_IS_CLOUD_VERSION}" != x"true" ]]; then
        log_info 'Installing containerd ...'
        install_containerd
    fi

    run_install

    if [ "$storage_type" == "minio" ]; then
        # init minio-operator after etcd installed
        init_minio_cluster
    fi

    log_info 'Installing backup component ...'
    install_velero

    install_velero_plugin_terminus

    log_info 'Waiting for Vault ...'
    check_vault

    log_info 'Starting Terminus ...'
    check_desktop

    log_info 'Installation wizard is complete\n'

    # install complete
    echo -e " Terminus is running at"
    echo -e "${GREEN_LINE}"
    show_launcher_ip
    echo -e "${GREEN_LINE}"
    echo -e " Open your browser and visit the above address."
    echo -e " "
    echo -e " User: ${username} "
    echo -e " Password: ${userpwd} "
    echo -e " "
    echo -e " Please change the default password after login."
}

read_tty(){
    echo -n $1
    read $2 < /dev/tty
}

validate_username() {
    local min=2
    local max=250
    local usermatch
    local keywords=(user system space default os kubesphere kube kubekey kubernetes gpu tapr bfl bytetrade project pod)

    shopt -s nocasematch
    for k in "${keywords[@]}"; do
        if [[ "$username" == "$k" ]]; then
            printf "'$username' is a system reserved keyword and cannot be set as a username.\n\n"
            return 1
        fi
    done
    shopt -u nocasematch

    usermatch=$(echo $username |egrep -o '^[a-z0-9]([a-z0-9]*[a-z0-9])?([a-z0-9]([a-z0-9]*[a-z0-9])?)*')

    if [ x"$usermatch" != x"$username" ]; then
        printf "illegal username '$username', try again\n\n"
        return 1
    fi

    if [[ ${#username} -lt $min || ${#username} -gt $max ]]; then
        printf "illegal username '$username', cannot be less than $min and cannot exceed $max characters. try again\n\n"
        return 1
    fi

    return 0
}

validate_useremail() {
    local match
    match=$(echo $useremail |egrep -o '^(([A-Za-z0-9]+((\.|\-|\_|\+)?[A-Za-z0-9]?)*[A-Za-z0-9]+)|[A-Za-z0-9]+)@(([A-Za-z0-9]+)+((\.|\-|\_)?([A-Za-z0-9]+)+)*)+\.([A-Za-z]{2,})+$')

    if [ x"$match" != x"$useremail" ]; then
        printf "illegal email '$useremail', try again\n\n"
        return 1
    fi
    return 0
}

validate_domainname() {
    local match
    match=$(echo $domainname |egrep -o '^([a-z0-9])(([a-z0-9-]{1,61})?[a-z0-9]{1})?(\.[a-z0-9](([a-z0-9-]{1,61})?[a-z0-9]{1})?)?(\.[a-zA-Z]{2,10})+$')

    if [ x"$match" != x"$domainname" ]; then
        printf "illegal domain name '$domainname', try again\n\n"
        return 1
    fi
    return 0
}

validate_userpwd() {
    local min=6
    local max=32

    if [[ ${#userpwd} -lt $min || ${#userpwd} -gt $max ]]; then
        printf "illegal password '$userpwd', cannot be less than $min and cannot exceed $max characters. try again\n\n"
        return 1
    fi
    return 0
}

setup_ws() {
    # username, email, password from env
    username="$TERMINUS_OS_USERNAME"
    userpwd="$TERMINUS_OS_PASSWORD"
    useremail="$TERMINUS_OS_EMAIL"
    domainname="$TERMINUS_OS_DOMAINNAME"

    log_info 'parse user info from env or stdin\n'
    if [ -z "$domainname" ]; then
        while :; do
            read_tty "Enter the domain name ( default myterminus.com ): " domainname
            [[ -z "$domainname" ]] && domainname="myterminus.com"

            if ! validate_domainname; then
                continue
            fi
            break
        done
    fi

    if ! validate_domainname; then
        log_fatal "illegal domain name '$domainname'"
    fi

    if [ -z "$username" ]; then
        while :; do
            read_tty "Enter the terminus name: " username
            local domain=$(echo "$username"|awk -F'@' '{print $2}')
            if [[ ! -z "${domain}" && x"${domain}" != x"${domainname}" ]]; then
                printf "illegal domain name '$domain', try again\n\n"
                continue
            fi

            username=$(echo "$username"|awk -F'@' '{print $1}')

            if ! validate_username; then
                continue
            fi
            break
        done
    fi

    if ! validate_username; then
        log_fatal "illegal username '$username'"
    fi

    if [ -z "$useremail" ]; then
        # while :; do
        #     read_tty "Enter the email: " useremail
        #     if ! validate_useremail; then
        #         continue
        #     fi
        #     break
        # done
        useremail="${username}@${domainname}"
    fi

    if ! validate_useremail; then
        log_fatal "illegal user email '$useremail'"
    fi

    if [ -z "$userpwd" ]; then
        # while :; do
        #     read_tty "Enter the password: " userpwd
        #     if ! validate_userpwd; then
        #         continue
        #     fi
        #     break
        # done
        userpwd=$(random_string 8)
    fi

    if ! validate_userpwd; then
        log_fatal "illegal user password '$userpwd'"
    fi

    encryptpwd=$(htpasswd -nbBC 10 USER "${userpwd}"|awk -F":" '{print $2}')

    log_info 'generate app values'

    # generate values
    local s3_sts="none"
    local s3_ak="none"
    local s3_sk="none"
    if [ ! -z "${AWS_SESSION_TOKEN_SETUP}" ]; then
        s3_sts="${AWS_SESSION_TOKEN_SETUP}"
        s3_ak="${AWS_ACCESS_KEY_ID_SETUP}"
        s3_sk="${AWS_SECRET_ACCESS_KEY_SETUP}"
    fi

    cat > ${BASE_DIR}/wizard/config/account/values.yaml <<_EOF
user:
  name: '${username}'
  password: '${encryptpwd}'
  email: '${useremail}'
  terminus_name: '${username}@${domainname}'
_EOF

    cat > ${BASE_DIR}/wizard/config/settings/values.yaml <<_EOF
namespace:
  name: 'user-space-${username}'
  role: admin

cluster_id: ${CLUSTER_ID}
s3_sts: ${s3_sts}
s3_ak: ${s3_ak}
s3_sk: ${s3_sk}

user:
  name: '${username}'
_EOF

  cat > ${BASE_DIR}/wizard/config/launcher/values.yaml <<_EOF
bfl:
  nodeport: 30883
  nodeport_ingress_http: 30083
  nodeport_ingress_https: 30082
  username: '${username}'
  admin_user: true
_EOF

  sed -i "s/#__DOMAIN_NAME__/${domainname}/" ${BASE_DIR}/wizard/config/settings/templates/terminus_cr.yaml

  publicIp=$(curl --connect-timeout 5 -sL http://169.254.169.254/latest/meta-data/public-ipv4 2>&1)
  publicHostname=$(curl --connect-timeout 5 -sL http://169.254.169.254/latest/meta-data/public-hostname 2>&1)

  local selfhosted="true"
  if [[ ! -z "${TERMINUS_IS_CLOUD_VERSION}" && x"${TERMINUS_IS_CLOUD_VERSION}" == x"true" ]]; then
    selfhosted="false"
  fi
  if [[ x"$publicHostname" =~ "amazonaws" && -n "$publicIp" && ! x"$publicIp" =~ "Not Found" ]]; then
    selfhosted="false"
  fi
  sed -i "s/#__SELFHOSTED__/${selfhosted}/" ${BASE_DIR}/wizard/config/settings/templates/terminus_cr.yaml
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

get_auth_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=authelia' -o jsonpath='{.items[*].status.phase}'"
}

get_profile_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=profile' -o jsonpath='{.items[*].status.phase}'"
}

get_desktop_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=edge-desktop' -o jsonpath='{.items[*].status.phase}'"
}

get_vault_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=vault' -o jsonpath='{.items[*].status.phase}'"
}

get_appservice_status(){
    $sh_c "${KUBECTL} get pod  -n os-system -l 'tier=app-service' -o jsonpath='{.items[*].status.phase}'"
}

get_appservice_pod(){
    $sh_c "${KUBECTL} get pod  -n os-system -l 'tier=app-service' -o jsonpath='{.items[*].metadata.name}'"
}

get_bfl_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'tier=bfl' -o jsonpath='{.items[*].status.phase}'"
}

get_bfl_node(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'tier=bfl' -o jsonpath='{.items[*].spec.nodeName}'"
}

get_kscm_status(){
    $sh_c "${KUBECTL} get pod  -n kubesphere-system -l 'app=ks-controller-manager' -o jsonpath='{.items[*].status.phase}' 2>/dev/null"
}

get_ksapi_status(){
    $sh_c "${KUBECTL} get pod  -n kubesphere-system -l 'app=ks-apiserver' -o jsonpath='{.items[*].status.phase}' 2>/dev/null"
}

get_ksredis_status(){
    $sh_c "${KUBECTL} get pod  -n kubesphere-system -l 'app=redis' -o jsonpath='{.items[*].status.phase}' 2>/dev/null"
}

get_gpu_status(){
    $sh_c "${KUBECTL} get pod  -n kube-system -l 'name=nvidia-device-plugin-ds' -o jsonpath='{.items[*].status.phase}'"
}

get_orion_gpu_status(){
    $sh_c "${KUBECTL} get pod  -n gpu-system -l 'app=orionx-container-runtime' -o jsonpath='{.items[*].status.phase}'"
}

get_userspace_dir(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'tier=bfl' -o \
    jsonpath='{range .items[0].spec.volumes[*]}{.name}{\" \"}{.persistentVolumeClaim.claimName}{\"\\n\"}{end}}'" | \
    while read pvc; do
        pvc_data=($pvc)
        if [ ${#pvc_data[@]} -gt 1 ]; then
            if [ "x${pvc_data[0]}" == "xuserspace-dir" ]; then
                USERSPACE_PVC="${pvc_data[1]}"
                pv=$($sh_c "${KUBECTL} get pvc -n user-space-${username} ${pvc_data[1]} -o jsonpath='{.spec.volumeName}'")
                pv_path=$($sh_c "${KUBECTL} get pv ${pv} -o jsonpath='{.spec.hostPath.path}'")
                USERSPACE_PV_PATH="${pv_path}"

                echo "${USERSPACE_PVC} ${USERSPACE_PV_PATH}"
                break
            fi
        fi
    done 
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

    res=$($sh_c "${KUBECTL} -n $ns get $resource_type $resource_name -o jsonpath='{.metadata.annotations.$key}'")
    if [[ $? -eq 0 && x"$res" != x"" ]]; then
        echo "$res"
        return
    fi
    echo "can not to get $ns ${resource_type}/${resource_name} annotation '$key', got value '$res'"
    exit $ERR_EXIT
}

get_bfl_url() {
    bfl_ip=$(curl ${CURL_TRY} -s http://checkip.dyndns.org/ | grep -o "[[:digit:].]\+")
    echo "http://$bfl_ip:30883/bfl/apidocs.json"
}

get_app_key_secret(){
    app=$1
    key="bytetrade_${app}_${RANDOM}"
    secret=$(random_string 16)

    echo "${key} ${secret}"
}

get_app_settings(){
    apps=("portfolio" "vault" "desktop" "message" "wise" "search" "appstore" "notification" "dashboard" "settings" "devbox" "profile" "agent" "files")
    for a in "${apps[@]}";do
        ks=($(get_app_key_secret $a))
        echo '
  '${a}':
    appKey: '${ks[0]}'    
    appSecret: "'${ks[1]}'"    
        '
    done
}

repeat(){
    for _ in $(seq 1 "$1"); do
        echo -n "$2"
    done
}

check_desktop(){
    status=$(check_together get_profile_status get_auth_status get_desktop_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 0.5

        status=$(check_together get_profile_status get_auth_status get_desktop_status)
        echo -ne "\rPlease waiting          "

    done
    echo
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

check_appservice(){
    status=$(get_appservice_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for app-service starting ${dot}"
        sleep 0.5

        status=$(get_appservice_status)
        echo -ne "\rWaiting for app-service starting          "

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

check_kscm(){
    status=$(get_kscm_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for ks-controller-manager starting ${dot}"
        sleep 0.5

        status=$(get_kscm_status)
        echo -ne "\rWaiting for ks-controller-manager starting          "

    done
    echo
}

check_ksapi(){
    status=$(get_ksapi_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for ks-apiserver starting ${dot}"
        sleep 0.5

        status=$(get_ksapi_status)
        echo -ne "\rWaiting for ks-apiserver starting          "

    done
    echo
}

check_ksredis(){
    status=$(get_ksredis_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for ks-redis starting ${dot}"
        sleep 0.5

        status=$(get_ksredis_status)
        echo -ne "\rWaiting for ks-redis starting          "

    done
    echo
}

check_gpu(){
    status=$(get_gpu_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for nvidia-device-plugin starting ${dot}"
        sleep 0.5

        status=$(get_gpu_status)
        echo -ne "\rWaiting for nvidia-device-plugin starting          "

    done
    echo
}

check_orion_gpu(){
    status=$(get_orion_gpu_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for orionx-container-runtime starting ${dot}"
        sleep 0.5

        status=$(get_orion_gpu_status)
        echo -ne "\rWaiting for orionx-container-runtime starting          "

    done
    echo
}

install_gpu(){
    # only for leishen mix
    # to be tested
    log_info 'Installing Nvidia GPU Driver ...\n'

    distribution=$(. /etc/os-release;echo $ID$VERSION_ID|sed 's/\.//g')

    if [ "$distribution" == "ubuntu2404" ]; then
        echo "Not supported Ubuntu 24.04"
        return
    fi

    if [[ "$distribution" =~ "ubuntu" ]]; then
        case "$distribution" in
            ubuntu2404)
                local u24_cude_keyring_deb="${BASE_DIR}/components/ubuntu2404_cuda-keyring_1.1-1_all.deb"
                if [ -f "$u24_cude_keyring_deb" ]; then
                    ensure_success $sh_c "cp ${u24_cude_keyring_deb} cuda-keyring_1.1-1_all.deb"
                else 
                    ensure_success $sh_c "wget https://developer.download.nvidia.com/compute/cuda/repos/$distribution/x86_64/cuda-keyring_1.1-1_all.deb"
                fi
                ensure_success $sh_c "dpkg -i cuda-keyring_1.1-1_all.deb"
                ;;
            ubuntu2204|ubuntu2004)
                local cude_keyring_deb="${BASE_DIR}/components/${distribution}_cuda-keyring_1.0-1_all.deb"
                if [ -f "$cude_keyring_deb" ]; then
                    ensure_success $sh_c "cp ${cude_keyring_deb} cuda-keyring_1.0-1_all.deb"
                else
                    ensure_success $sh_c "wget https://developer.download.nvidia.com/compute/cuda/repos/$distribution/x86_64/cuda-keyring_1.0-1_all.deb"
                fi
                ensure_success $sh_c "dpkg -i cuda-keyring_1.0-1_all.deb"
                ;;
            *)
                ;;
        esac
    fi
    
    ensure_success $sh_c "apt-get update"
    ensure_success $sh_c "apt-get -y install cuda-12-1"
    ensure_success $sh_c "apt-get -y install nvidia-kernel-open-545"
    ensure_success $sh_c "apt-get -y install nvidia-driver-545"

    distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
    ensure_success $sh_c "curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | apt-key add -"
    ensure_success $sh_c "curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.list | tee /etc/apt/sources.list.d/libnvidia-container.list"
    ensure_success $sh_c "apt-get update && sudo apt-get install -y nvidia-container-toolkit jq"

    if [ x"$KUBE_TYPE" == x"k3s" ]; then
        ensure_success $sh_c "cp /var/lib/rancher/k3s/agent/etc/containerd/config.toml /var/lib/rancher/k3s/agent/etc/containerd/config.toml.tmpl" 
        ensure_success $sh_c "nvidia-ctk runtime configure --runtime=containerd --set-as-default --config=/var/lib/rancher/k3s/agent/etc/containerd/config.toml.tmpl"
        ensure_success $sh_c "systemctl restart k3s"
    else
        ensure_success $sh_c "nvidia-ctk runtime configure --runtime=containerd --set-as-default"
        ensure_success $sh_c "systemctl restart containerd"
    fi

    check_ksredis
    check_kscm
    check_ksapi

    # waiting for kubesphere webhooks starting
    sleep 30

    ensure_success $sh_c "${KUBECTL} create -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/v0.14.0/nvidia-device-plugin.yml"

    log_info 'Waiting for Nvidia GPU Driver applied ...\n'

    check_gpu

    if [ "x${LOCAL_GPU_SHARE}" == "x1" ]; then
        log_info 'Installing Nvshare GPU Plugin ...\n'

        ensure_success $sh_c "${KUBECTL} apply -f https://raw.githubusercontent.com/grgalex/nvshare/v0.1/kubernetes/manifests/nvshare-system.yaml"
        ensure_success $sh_c "${KUBECTL} apply -f https://raw.githubusercontent.com/grgalex/nvshare/v0.1/kubernetes/manifests/nvshare-system-quotas.yaml"
        ensure_success $sh_c "${KUBECTL} apply -f https://raw.githubusercontent.com/grgalex/nvshare/v0.1/kubernetes/manifests/device-plugin.yaml"
        ensure_success $sh_c "${KUBECTL} apply -f https://raw.githubusercontent.com/grgalex/nvshare/v0.1/kubernetes/manifests/scheduler.yaml"
    fi
}

source ./wizard/bin/COLORS
PORT="30180"  # desktop port
show_launcher_ip() {
    IP=$(curl ${CURL_TRY} -s http://ifconfig.me/)
    if [ -n "$local_ip" ]; then
        echo -e "http://${local_ip}:$PORT "
    fi

    if [ -n "$IP" ]; then
        echo -e "http://$IP:$PORT "
    fi
}

if [ -d /tmp/install_log ]; then
    $sh_c "rm -rf /tmp/install_log"
fi

mkdir -p /tmp/install_log && cd /tmp/install_log || exit
fd_errlog=/tmp/install_log/errlog_fd_13

Main() {
    [[ -z $KUBE_TYPE ]] && KUBE_TYPE="k3s"

    log_info 'Start to Install Terminus ...\n'
    get_distribution
    get_shell_exec

    (
        log_info 'Precheck and Installing dependencies ...\n'
        precheck_os
        install_deps
        config_system

        log_info 'Installing terminus ...\n'
        config_proxy_resolv_conf
        install_storage
        install_k8s_ks
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
