#!/usr/bin/env bash
source ./common.sh



ERR_EXIT=1

CURL_TRY="--connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 "

BASE_DIR=$(dirname $(realpath -s $0))
INSTALL_LOG="$BASE_DIR/logs"

[[ -f "${BASE_DIR}/.env" && -z "$DEBUG_VERSION" ]] && . "${BASE_DIR}/.env"

function retry_cmd(){
    wait_k8s_health
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
            
            if [[ $ret -eq 0 ]]; then
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
    wait_k8s_health
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

                if [[ $ret -eq 0 ]]; then
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

# precheck_os() {
#     if [ x"$PREPARED" == x"1" ]; then
#         precheck_localip
#         return
#     fi

#     if [[ -f /boot/cmdline.txt || -f /boot/firmware/cmdline.txt ]]; then
#     # raspbian 
#         SHOULD_RETRY=1

#         if ! command_exists iptables; then 
#             ensure_success $sh_c "apt update && apt install -y iptables"
#         fi

#         systemctl disable --user gvfs-udisks2-volume-monitor
#         systemctl stop --user gvfs-udisks2-volume-monitor

#         local cpu_cgroups_enbaled=$(cat /proc/cgroups |awk '{if($1=="cpu")print $4}')
#         local mem_cgroups_enbaled=$(cat /proc/cgroups |awk '{if($1=="memory")print $4}')
#         if  [[ $cpu_cgroups_enbaled -eq 0 || $mem_cgroups_enbaled -eq 0 ]]; then
#             log_fatal "cpu or memory cgroups disabled, please edit /boot/cmdline.txt or /boot/firmware/cmdline.txt and reboot to enable it."
#         fi
#     fi
    
#     # try to resolv hostname
#     ensure_success $sh_c "hostname -i >/dev/null"

#     precheck_localip

#     # local badHostname
#     # badHostname=$(echo "$HOSTNAME" | grep -E "[A-Z]")
#     # if [ x"$badHostname" != x"" ]; then
#     #     log_fatal "please set the hostname with lowercase ['${badHostname}']"
#     # fi

#     # ip=$(ping -c 1 "$HOSTNAME" |awk -F '[()]' '/icmp_seq/{print $2}')
#     # printf "%s\t%s\n\n" "$ip" "$HOSTNAME"

#     # if [[ x"$ip" == x"" || "$ip" == @("172.17.0.1"|"127.0.0.1"|"127.0.1.1") || ! "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
#     #     log_fatal "incorrect ip for hostname '$HOSTNAME', please check"
#     # fi

#     # local_ip="$ip"

#     # disable local dns
#     case "$OSNAME" in
#         ubuntu|debian|raspbian)
#             if system_service_active "systemd-resolved"; then
#                 ensure_success $sh_c "systemctl stop systemd-resolved.service >/dev/null"
#                 ensure_success $sh_c "systemctl disable systemd-resolved.service >/dev/null"
#                 if [ -e /usr/bin/systemd-resolve ]; then
#                     ensure_success $sh_c "mv /usr/bin/systemd-resolve /usr/bin/systemd-resolve.bak >/dev/null"
#                 fi
#                 if [ -L /etc/resolv.conf ]; then
#                     ensure_success $sh_c 'unlink /etc/resolv.conf && touch /etc/resolv.conf'
#                 fi
#                 config_resolv_conf
#             else
#                 ensure_success $sh_c "cat /etc/resolv.conf > /etc/resolv.conf.bak"
#             fi
#             ;;
#         centos|fedora|rhel)
#             ;;
#         *)
#             ;;
#     esac

#     if ! hostname -i &>/dev/null; then
#         ensure_success $sh_c "echo $local_ip  $HOSTNAME >> /etc/hosts"
#     fi

#     ensure_success $sh_c "hostname -i >/dev/null"

#     # network and dns
#     http_code=$(curl ${CURL_TRY} -sL -o /dev/null -w "%{http_code}" https://download.docker.com/linux/ubuntu)
#     if [ "$http_code" != 200 ]; then
#         config_resolv_conf
#         if [ -f /etc/resolv.conf.bak ]; then
#             ensure_success $sh_c "rm -rf /etc/resolv.conf.bak"
#         fi

#     fi

#     # ubuntu 24 upgrade apparmor
#     if [[ $(is_ubuntu) -eq 1 && $(get_os_version) == *24.* ]]; then
#         aapv=$(apparmor_parser --version)
#         if [[ ! ${aapv} =~ "4.0.1" ]]; then
#             local aapv_tar="${BASE_DIR}/components/apparmor_4.0.1-0ubuntu1_${ARCH}.deb"
#             if [ ! -f "$aapv_tar" ]; then
#                 if [ x"${ARCH}" == x"arm64" ]; then
#                     ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://launchpad.net/ubuntu/+source/apparmor/4.0.1-0ubuntu1/+build/28428841/+files/apparmor_4.0.1-0ubuntu1_arm64.deb"
#                 else
#                     ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://launchpad.net/ubuntu/+source/apparmor/4.0.1-0ubuntu1/+build/28428840/+files/apparmor_4.0.1-0ubuntu1_amd64.deb"
#                 fi
#             else
#                 ensure_success $sh_c "cp ${aapv_tar} ./"
#             fi
#             ensure_success $sh_c "dpkg -i apparmor_4.0.1-0ubuntu1_${ARCH}.deb"
#         fi
#     fi

#     if [[ $(is_wsl) -eq 1 ]]; then
#         $sh_c "chattr -i /etc/hosts"
#         $sh_c "chattr -i /etc/resolv.conf"
#     fi

#     $sh_c "apt remove unattended-upgrades -y"
#     $sh_c "apt install ntpdate -y"

#     local ntpdate=$(get_command ntpdate)
#     local hwclock=$(get_command hwclock)
    
#     $sh_c "$ntpdate -b -u pool.ntp.org"
#     $sh_c "$hwclock -w"
# }

precheck_localip() {
    local ip
    local badHostname

    badHostname=$(echo "$HOSTNAME" | grep -E "[A-Z]")
    if [ x"$badHostname" != x"" ]; then
        log_fatal "please set the hostname with lowercase ['${badHostname}']"
    fi

    ip=$(ping -c 1 "$HOSTNAME" |awk -F '[()]' '/icmp_seq/{print $2}')
    printf "%s\t%s\n\n" "$ip" "$HOSTNAME"

    if [[ x"$ip" == x"" || "$ip" == @("172.17.0.1"|"127.0.0.1"|"127.0.1.1") || ! "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_fatal "incorrect ip for hostname '$HOSTNAME', please check"
    fi

    local_ip="$ip"
}

# install_deps() {
#     if [ x"$PREPARED" == x"1" ]; then
#         return
#     fi
#     case "$OSNAME" in
#         ubuntu|debian|raspbian)
#             pre_reqs="apt-transport-https ca-certificates curl"
#             if [ -z $(get_command gpg) ]; then
#               pre_reqs="$pre_reqs gnupg"
#             fi
#             if [ -z $(get_command sudo) ]; then
#               pre_reqs="$pre_reqs sudo"
#             fi

#             if [ $(is_pve) -eq 0 ]; then
#                 ensure_success $sh_c 'apt-get update -qq >/dev/null'
#             else
#                 $sh_c 'apt-get update -qq >/dev/null'
#             fi
#             ensure_success $sh_c "DEBIAN_FRONTEND=noninteractive apt-get install -y -qq $pre_reqs >/dev/null"
#             ensure_success $sh_c 'DEBIAN_FRONTEND=noninteractive apt-get install -y conntrack socat apache2-utils ntpdate net-tools make gcc openssh-server >/dev/null'
#             ;;

#         centos|fedora|rhel)
#             if [ "$lsb_dist" = "fedora" ]; then
#                 pkg_manager="dnf"
#             else
#                 pkg_manager="yum"
#             fi

#             ensure_success $sh_c "$pkg_manager install -y conntrack socat httpd-tools ntpdate net-tools make gcc openssh-server >/dev/null"
#             ;;
#         *)
#             # build from source code
#             build_socat
#             build_contrack

#             #TODO: install bcrypt tools
#             ;;
#     esac
# }

config_system() {
    if [ x"$PREPARED" == x"1" ]; then
        return
    fi

    local ntpdate hwclock
    natgateway=""

    # kernel printk log level
    # cause SIGSTOP in ubuntu 22.04
    # ensure_success $sh_c 'sysctl -w kernel.printk="3 3 1 7"'

    # ntp sync
    ntpdate=$(get_command ntpdate)
    hwclock=$(get_command hwclock)

    printf '#!/bin/sh\n\n%s -b -u pool.ntp.org && %s -w\n\nexit 0\n' "$ntpdate" "$hwclock" > cron.ntpdate
    ensure_success $sh_c '/bin/sh cron.ntpdate'
    ensure_success $sh_c 'cat cron.ntpdate > /etc/cron.daily/ntpdate && chmod 0700 /etc/cron.daily/ntpdate'
    ensure_success rm -f cron.ntpdate

    if ! system_service_active "ssh"; then
        ensure_success $sh_c 'systemctl enable --now ssh'
    fi 

    if [[ $(is_wsl) -eq 1 ]]; then
        while :; do
            read_tty "Enter the windows host IP: " natgateway
            natgateway=$(echo "$natgateway" | grep -E "[0-9]+(\.[0-9]+){3}" | grep -v "127.0.0.1")
            if [ x"$natgateway" == x"" ]; then
                continue
            fi
            break
        done
    fi
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

k8s_health(){
    if [ ! -z "$KUBECTL" ]; then
        $sh_c "$KUBECTL get --raw='/readyz?verbose' 1>/dev/null"
    fi
}

wait_k8s_health(){
    local max_retry=60
    local ok="n"
    while [ $max_retry -ge 0 ]; do
        if k8s_health; then
            ok="y"
            break
        fi
        sleep 5
        ((max_retry--))
    done

    if [ x"$ok" != x"y" ]; then
        echo "k8s is not health yet, please check it"
        exit $ERR_EXIT
    fi

}


run_install() {
    k8s_version=v1.22.10
    ks_version=v3.3.0

    log_info 'installing k8s and kubesphere'

    # if [[ $(is_wsl) -eq 1 ]]; then
    #     if [ -f /usr/lib/wsl/lib/nvidia-smi ]; then
    #         local device=$(/usr/lib/wsl/lib/nvidia-smi -L|grep 'NVIDIA'|grep UUID)
    #         if [ x"$device" != x"" ]; then
    #             LOCAL_GPU_ENABLE="1"
    #             LOCAL_GPU_SHARE="1"
    #         fi
    #     fi
    # fi

    # env 'KUBE_TYPE' is specific the special kubernetes (k8s or k3s), default k3s
    if [ x"$KUBE_TYPE" == x"k3s" ]; then
        k8s_version=v1.22.16-k3s
    fi


    # TODO: prepare
    # create_cmd="$TERMINUS_CLI terminus prepare --kube $KUBE_TYPE"
    
    local extra

    # env 'REGISTRY_MIRRORS' is a docker image cache mirrors, separated by commas
    # if [ x"$REGISTRY_MIRRORS" != x"" ]; then
    #     extra=" --registry-mirrors $REGISTRY_MIRRORS"
    # fi
    # create_cmd+=" $extra"

    # add env OS_LOCALIP
    # ensure_success $sh_c "export OS_LOCALIP=$local_ip && export TERMINUS_IS_CLOUD_VERSION=$TERMINUS_IS_CLOUD_VERSION && $create_cmd"

    log_info 'k8s and kubesphere installation is complete'

    # cache version to file
    # ensure_success $sh_c "echo 'VERSION=${VERSION}' > /etc/kke/version"
    # ensure_success $sh_c "echo 'KKE=${TERMINUS_CLI_VERSION}' >> /etc/kke/version"
    # ensure_success $sh_c "echo 'KUBE=${k8s_version}' >> /etc/kke/version"

    # setup after kubesphere is installed
    export KUBECONFIG=/root/.kube/config  # for ubuntu
    HELM=$(get_command helm)
    KUBECTL=$(get_command kubectl)

    check_kscm # wait for ks launch

    if [[ $SHOULD_RETRY -eq 1 || $(is_wsl) -eq 1 ]]; then
        run_cmd=retry_cmd
    else
        run_cmd=retry_cmd
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
    local xargs=""
    if [[ $(is_wsl) -eq 1 && x"$natgateway" != x"" ]]; then
        echo "annotate bfl with nat gateway ip"
        xargs="--set nat_gateway_ip=${natgateway}"
    fi
    retry_cmd $sh_c "${HELM} upgrade -i account ${BASE_DIR}/wizard/config/account --force ${xargs}"

    log_info 'Installing settings ...'
    $run_cmd $sh_c "${HELM} upgrade -i settings ${BASE_DIR}/wizard/config/settings --force"

    # install gpu if necessary
    # if [[ "x${GPU_ENABLE}" == "x1" && "x${GPU_DOMAIN}" != "x" ]]; then
    #     log_info 'Installing gpu ...'

    #     if [ x"$KUBE_TYPE" == x"k3s" ]; then
    #         $run_cmd $sh_c "${HELM} upgrade -i gpu ${BASE_DIR}/wizard/config/gpu -n gpu-system --force --set gpu.server=${GPU_DOMAIN} --set container.manager=k3s --create-namespace"
    #         ensure_success $sh_c "mkdir -p /var/lib/rancher/k3s/agent/etc/containerd"
    #         ensure_success $sh_c "cp ${BASE_DIR}/deploy/orion-config.toml.tmpl /var/lib/rancher/k3s/agent/etc/containerd/config.toml.tmpl" 
    #         ensure_success $sh_c "systemctl restart k3s"

    #         check_ksredis
    #         check_kscm
    #         check_ksapi

    #         # waiting for kubesphere webhooks starting
    #         sleep 30
    #     else
    #         $run_cmd $sh_c "${HELM} upgrade -i gpu ${BASE_DIR}/wizard/config/gpu -n gpu-system --force --set gpu.server=${GPU_DOMAIN} --set container.manager=containerd --create-namespace"
    #     fi

    #     check_orion_gpu
    # fi
    GPU_TYPE="none"
    if [ "x${LOCAL_GPU_ENABLE}" == "x1" ]; then  
        GPU_TYPE="nvidia"
        if [ "x${LOCAL_GPU_SHARE}" == "x1" ]; then  
            GPU_TYPE="nvshare"
        fi
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
        --set backup.sync_secret=\"${BACKUP_SECRET}\" --set gpu=\"${GPU_TYPE}\" --set s3_bucket=\"${S3_BUCKET}\" \
        --set fs_type=\"${fs_type}\""

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
    $run_cmd $sh_c "$KUBECTL apply -f cm-backup-config.yaml"

    # patch
    $run_cmd $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-globalrole-workspace-manager.yaml"
    $run_cmd $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-notification-manager.yaml"

    # install app-store charts repo to app sevice
    log_info 'waiting for appservice'
    check_appservice
    appservice_pod=$(get_appservice_pod)

    # gen bfl app key and secret
    bfl_ks=($(get_app_key_secret "bfl"))

    log_info 'Installing launcher ...'
    # install launcher , and init pv
    retry_cmd $sh_c "${HELM} upgrade -i launcher-${username} ${BASE_DIR}/wizard/config/launcher -n user-space-${username} --force --set bfl.appKey=${bfl_ks[0]} --set bfl.appSecret=${bfl_ks[1]}"

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
    fs_type="jfs"
    if [[ $(is_wsl) -eq 1 ]]; then
        fs_type="fs"
    fi

    ensure_success $sh_c "rm -rf ${BASE_DIR}/wizard/config/apps/values.yaml"
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
fs_type: ${fs_type}

os:
  ${app_perm_settings}
EOF

    log_info 'Installing built-in apps ...'
    for appdir in "${BASE_DIR}/wizard/config/apps"/*/; do
      if [ -d "$appdir" ]; then
        releasename=$(basename "$appdir")
        $run_cmd $sh_c "${HELM} upgrade -i ${releasename} ${appdir} -n user-space-${username} --force --set kubesphere.redis_password=${ks_redis_pwd} -f ${BASE_DIR}/wizard/config/apps/values.yaml"
      fi
    done

    # log_info 'Installing user console ...'
    # ensure_success $sh_c "${HELM} upgrade -i console-${username} ${BASE_DIR}/wizard/config/console -n user-space-${username} --set bfl.username=${username}"

    # clear apps values.yaml
    cat /dev/null > ${BASE_DIR}/wizard/config/apps/values.yaml
    cat /dev/null > ${BASE_DIR}/wizard/config/launcher/values.yaml
    copy_charts=("launcher" "apps")
    for cc in "${copy_charts[@]}"; do
        retry_cmd $sh_c "${KUBECTL} cp ${BASE_DIR}/wizard/config/${cc} os-system/${appservice_pod}:/userapps -c app-service"
    done

    log_info 'Performing the final configuration ...'
    # delete admin user after kubesphere installed,
    # admin user creating in the ks-install image should be modified.
    $run_cmd $sh_c "${KUBECTL} patch user admin -p '{\"metadata\":{\"finalizers\":[\"finalizers.kubesphere.io/users\"]}}' --type='merge'"
    $run_cmd $sh_c "${KUBECTL} delete user admin"
    $run_cmd $sh_c "${KUBECTL} delete deployment kubectl-admin -n kubesphere-controls-system"
    # $run_cmd $sh_c "${KUBECTL} scale deployment/ks-installer --replicas=0 -n kubesphere-system"
    $run_cmd $sh_c "${KUBECTL} delete deployment -n kubesphere-controls-system default-http-backend"
    
    # delete storageclass accessor webhook
    # $run_cmd $sh_c "${KUBECTL} delete validatingwebhookconfigurations storageclass-accessor.storage.kubesphere.io"

    # calico config for tailscale
    $run_cmd $sh_c "${KUBECTL} patch felixconfiguration default -p '{\"spec\":{\"featureDetectOverride\": \"SNATFullyRandom=false,MASQFullyRandom=false\"}}' --type='merge'"
}

init_minio_cluster(){
    MINIO_OPERATOR_VERSION="v0.0.1"
    if [[ ! -f /etc/ssl/etcd/ssl/ca.pem || ! -f /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem || ! -f /etc/ssl/etcd/ssl/node-$HOSTNAME.pem ]]; then
        echo "cann't find etcd key files"
        exit $ERR_EXIT
    fi

    local minio_operator_tar="${BASE_DIR}/components/minio-operator-${MINIO_OPERATOR_VERSION}-linux-${ARCH}.tar.gz"
    local minio_operator_bin="/usr/local/bin/minio-operator"

    if [ ! -f "$minio_operator_bin" ]; then
        if [ -f "$minio_operator_tar" ]; then
            ensure_success $sh_c "cp ${minio_operator_tar} minio-operator-${MINIO_OPERATOR_VERSION}-linux-${ARCH}.tar.gz"
        else
            ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://github.com/beclab/minio-operator/releases/download/${MINIO_OPERATOR_VERSION}/minio-operator-${MINIO_OPERATOR_VERSION}-linux-${ARCH}.tar.gz"
        fi
	      ensure_success $sh_c "tar zxf minio-operator-${MINIO_OPERATOR_VERSION}-linux-${ARCH}.tar.gz"
        ensure_success $sh_c "install -m 755 minio-operator $minio_operator_bin"
    fi

    ensure_success $sh_c "$minio_operator_bin init --address $local_ip --cafile /etc/ssl/etcd/ssl/ca.pem --certfile /etc/ssl/etcd/ssl/node-$HOSTNAME.pem --keyfile /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem --volume $MINIO_VOLUMES --password $MINIO_ROOT_PASSWORD"
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
        sleep_waiting 3
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
        sleep_waiting 3
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

    VELERO_VERSION="v1.11.3"
    local velero_tar="${BASE_DIR}/components/velero-${VELERO_VERSION}-linux-${ARCH}.tar.gz"
    if [ -f "$velero_tar" ]; then
        ensure_success $sh_c "cp ${velero_tar} velero-${VELERO_VERSION}-linux-${ARCH}.tar.gz"
    else
        ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://github.com/beclab/velero/releases/download/${VELERO_VERSION}/velero-${VELERO_VERSION}-linux-${ARCH}.tar.gz"
    fi
    ensure_success $sh_c "tar xf velero-${VELERO_VERSION}-linux-${ARCH}.tar.gz"
    ensure_success $sh_c "install velero-${VELERO_VERSION}-linux-${ARCH}/velero /usr/local/bin"

    CRICTL=$(get_command crictl)
    VELERO=$(get_command velero)

    # install velero crds
    ensure_success $sh_c "${VELERO} install --crds-only --retry 10 --delay 5"
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
  velero_ver="v1.11.3"
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
    velero_plugin_install_cmd+=" --velero-pod-cpu-request=10m --velero-pod-cpu-limit=200m"
    velero_plugin_install_cmd+=" --node-agent-pod-cpu-request=10m --node-agent-pod-cpu-limit=200m"
    velero_plugin_install_cmd+=" --wait --wait-minute 30"

    if [[ $(is_raspbian) -eq 1 ]]; then
        velero_plugin_install_cmd+=" --retry 30 --delay 5" # 30 times, 5 seconds delay
    fi

    ensure_success $sh_c "$velero_plugin_install_cmd"
    velero_plugin_install_cmd="${VELERO} plugin add beclab/velero-plugin-for-terminus:$velero_plugin_ver -n os-system"
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

install_k8s_ks() {

    log_info 'Setup your first user ...\n'
    setup_ws

    # generate init config
    ADDON_CONFIG_FILE=${BASE_DIR}/wizard/bin/init-config.yaml
    echo '
    ' > ${ADDON_CONFIG_FILE}

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

    if [[ $(is_wsl) -eq 1 ]]; then
        $sh_c "chattr +i /etc/hosts"
        $sh_c "chattr +i /etc/resolv.conf"
    fi

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
        userpwd=$(get_random_string 8)
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

    $sh_c "rm -rf ${BASE_DIR}/wizard/config/account/values.yaml"
    cat > ${BASE_DIR}/wizard/config/account/values.yaml <<_EOF
user:
  name: '${username}'
  password: '${encryptpwd}'
  email: '${useremail}'
  terminus_name: '${username}@${domainname}'
_EOF

    $sh_c "rm -rf ${BASE_DIR}/wizard/config/settings/values.yaml"
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

  $sh_c "rm -rf ${BASE_DIR}/wizard/config/launcher/values.yaml"
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
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=system-frontend' -o jsonpath='{.items[*].status.phase}'"
}

get_desktop_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=edge-desktop' -o jsonpath='{.items[*].status.phase}'"
}

get_vault_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=vault' -o jsonpath='{.items[*].status.phase}'"
}

get_citus_status(){
    $sh_c "${KUBECTL} get pod  -n os-system -l 'app=citus' -o jsonpath='{.items[*].status.phase}'"
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
    secret=$(get_random_string 16)

    echo "${key} ${secret}"
}

get_app_settings(){
    apps=("portfolio" "vault" "desktop" "message" "wise" "search" "appstore" "notification" "dashboard" "settings" "profile" "agent" "files")
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
    status=$(check_together get_appservice_status get_citus_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for app-service starting ${dot}"
        sleep 0.5

        status=$(check_together get_appservice_status get_citus_status)
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


    if [ x"$PREPARED" != x"1" ]; then
        if [ $(is_wsl) -eq 0 ]; then
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
        fi

        distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
        ensure_success $sh_c "curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | apt-key add -"
        ensure_success $sh_c "curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.list | tee /etc/apt/sources.list.d/libnvidia-container.list"
        ensure_success $sh_c "apt-get update && sudo apt-get install -y nvidia-container-toolkit jq"
    fi

    if [[ x"$KUBE_TYPE" == x"k3s" && x"$PREPARED" != x"1" ]]; then
        if [[ $(is_wsl) -eq 1 ]]; then
            local real_driver=$($sh_c "find /usr/lib/wsl/drivers/ -name libcuda.so.1.1|head -1")
            echo "found cuda driver in $real_driver"
            if [[ x"$real_driver" != x"" ]]; then
                local shellname="cuda_lib_fix.sh"
                cat << EOF > /tmp/${shellname}
#!/bin/bash
sh_c="sh -c"
real_driver=\$(\$sh_c "find /usr/lib/wsl/drivers/ -name libcuda.so.1.1|head -1")
if [[ x"\$real_driver" != x"" ]]; then
    \$sh_c "ln -s /usr/lib/wsl/lib/libcuda* /usr/lib/x86_64-linux-gnu/"
    \$sh_c "rm -f /usr/lib/x86_64-linux-gnu/libcuda.so"
    \$sh_c "rm -f /usr/lib/x86_64-linux-gnu/libcuda.so.1"
    \$sh_c "rm -f /usr/lib/x86_64-linux-gnu/libcuda.so.1.1"
    \$sh_c "cp -f \$real_driver /usr/lib/wsl/lib/libcuda.so"
    \$sh_c "cp -f \$real_driver /usr/lib/wsl/lib/libcuda.so.1"
    \$sh_c "cp -f \$real_driver /usr/lib/wsl/lib/libcuda.so.1.1"
    \$sh_c "ln -s \$real_driver /usr/lib/x86_64-linux-gnu/libcuda.so.1"
    \$sh_c "ln -s \$real_driver /usr/lib/x86_64-linux-gnu/libcuda.so.1.1"
    \$sh_c "ln -s /usr/lib/x86_64-linux-gnu/libcuda.so.1 /usr/lib/x86_64-linux-gnu/libcuda.so"
fi
EOF
                ensure_success $sh_c "mv -f /tmp/${shellname} /usr/local/bin/${shellname}"
                ensure_success $sh_c "chmod +x /usr/local/bin/${shellname}"
                ensure_success $sh_c "echo 'ExecStartPre=-/usr/local/bin/${shellname}' >> /etc/systemd/system/k3s.service"
                ensure_success $sh_c "systemctl daemon-reload"

            fi
        fi
    fi
    
    if [ x"$PREPARED" != x"1" ]; then
        ensure_success $sh_c "nvidia-ctk runtime configure --runtime=containerd --set-as-default"
        ensure_success $sh_c "systemctl restart containerd"
    fi
    

    check_ksredis
    check_kscm
    check_ksapi

    # waiting for kubesphere webhooks starting
    sleep_waiting 30


    ensure_success $sh_c "${KUBECTL} create -f ${BASE_DIR}/deploy/nvidia-device-plugin.yml"

    log_info 'Waiting for Nvidia GPU Driver applied ...\n'

    check_gpu

    if [ "x${LOCAL_GPU_SHARE}" == "x1" ]; then
        log_info 'Installing Nvshare GPU Plugin ...\n'

        ensure_success $sh_c "${KUBECTL} apply -f ${BASE_DIR}/deploy/nvshare-system.yaml"
        ensure_success $sh_c "${KUBECTL} apply -f ${BASE_DIR}/deploy/nvshare-system-quotas.yaml"
        ensure_success $sh_c "${KUBECTL} apply -f ${BASE_DIR}/deploy/device-plugin.yaml"
        ensure_success $sh_c "${KUBECTL} apply -f ${BASE_DIR}/deploy/scheduler.yaml"
    fi
}

source ./wizard/bin/COLORS
PORT="30180"  # desktop port
show_launcher_ip() {
    IP=$(curl ${CURL_TRY} -s http://ifconfig.me/)
    if [ -n "$natgateway" ]; then
        echo -e "http://${natgateway}:$PORT "
    else
        if [ -n "$local_ip" ]; then
            echo -e "http://${local_ip}:$PORT "
        fi
    fi

    if [ -n "$IP" ]; then
        echo -e "http://$IP:$PORT "
    fi
}

if [ -d $INSTALL_LOG ]; then
    $sh_c "rm -rf $INSTALL_LOG"
fi

mkdir -p $INSTALL_LOG && cd $INSTALL_LOG || exit
fd_errlog=$INSTALL_LOG/errlog_fd_13

Main() {

    log_info 'Start to Install Terminus ...\n'
    # TODO: install

    get_distribution
    get_shell_exec
        
    (
        precheck_support
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

touch ${INSTALL_LOG}/install.log
Main | tee ${INSTALL_LOG}/install.log

exit
