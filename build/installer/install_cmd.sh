#!/usr/bin/env bash
source ./common.sh

ENV_BASE_DIR=${BASE_DIR}

ERR_EXIT=1

CURL_TRY="--connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 "

BASE_DIR=$(dirname $(realpath -s $0))
INSTALL_LOG="$BASE_DIR/logs"

[[ -f "${BASE_DIR}/.env" && -z "$DEBUG_VERSION" ]] && . "${BASE_DIR}/.env"


run_install() {
    k8s_version=v1.22.10
    ks_version=v3.3.0

    log_info 'installing k8s and kubesphere'

    # env 'KUBE_TYPE' is specific the special kubernetes (k8s or k3s), default k3s
    if [ x"$KUBE_TYPE" == x"k3s" ]; then
        k8s_version=v1.22.16-k3s
    fi

    ensure_success $sh_c "export OS_LOCALIP=$local_ip && \
        export TERMINUS_IS_CLOUD_VERSION=$TERMINUS_IS_CLOUD_VERSION && \
    $TERMINUS_CLI terminus install $PARAM"

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

    # if [ x"$KUBE_TYPE" == x"k3s" ]; then
    #     retry_cmd $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-k3s.yaml"
    #     if [[ ! -z "${K3S_PRELOAD_IMAGE_PATH}" && -d $K3S_PRELOAD_IMAGE_PATH ]]; then
    #         # remove the preload image path to make sure images will not be reloaded after reboot
    #         ensure_success $sh_c "rm -rf ${K3S_PRELOAD_IMAGE_PATH}"
    #     fi
    # fi

    log_info 'Installing account ...'
    # add the first account
    local xargs=""
    if [[ $(is_wsl) -eq 1 && x"$natgateway" != x"" ]]; then
        echo "annotate bfl with nat gateway ip"
        xargs="--set nat_gateway_ip=${natgateway}"
    fi
    retry_cmd $sh_c "${HELM} upgrade -i account ${BASE_DIR}/wizard/config/account ${xargs}"

    log_info 'Installing settings ...'
    $run_cmd $sh_c "${HELM} upgrade -i settings ${BASE_DIR}/wizard/config/settings"

    # install gpu if necessary
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
    local shared_lib="/terminus/share"
    ensure_success $sh_c "mkdir -p $shared_lib && chown 1000:1000 $shared_lib"

    local ks_redis_pwd=$($sh_c "${KUBECTL} get secret -n kubesphere-system redis-secret -o jsonpath='{.data.auth}' |base64 -d")
    retry_cmd $sh_c "${HELM} upgrade -i system ${BASE_DIR}/wizard/config/system -n os-system \
        --set kubesphere.redis_password=${ks_redis_pwd} --set backup.bucket=\"${BACKUP_CLUSTER_BUCKET}\" \
        --set backup.key_prefix=\"${BACKUP_KEY_PREFIX}\" --set backup.is_cloud_version=\"${TERMINUS_IS_CLOUD_VERSION}\" \
        --set backup.sync_secret=\"${BACKUP_SECRET}\" --set gpu=\"${GPU_TYPE}\" --set s3_bucket=\"${S3_BUCKET}\" \
        --set fs_type=\"${fs_type}\" --set sharedlib=\"$shared_lib\""

    # save backup env to configmap
    cat > ${BASE_DIR}/deploy/cm-backup-config.yaml << _END
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
    $run_cmd $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/cm-backup-config.yaml"

    # patch
    $run_cmd $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-globalrole-workspace-manager.yaml"
    $run_cmd $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-notification-manager.yaml"

    # install app-store charts repo to app sevice
    log_info 'waiting for appservice'
    check_appservice
    appservice_pod=$(get_appservice_pod)

    # set reverse_proxy_config
    reverse_proxy_config

    # gen bfl app key and secret
    bfl_ks=($(get_app_key_secret "bfl"))

    log_info 'Installing launcher ...'
    # install launcher , and init pv
    retry_cmd $sh_c "${HELM} upgrade -i launcher-${username} ${BASE_DIR}/wizard/config/launcher -n user-space-${username} --set bfl.appKey=${bfl_ks[0]} --set bfl.appSecret=${bfl_ks[1]}"

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
    sed -i "s/marketProvider: ''/marketProvider: '${MARKET_PROVIDER}'/g" ${BASE_DIR}/wizard/config/apps/market/values.yaml
    for appdir in "${BASE_DIR}/wizard/config/apps"/*/; do
      if [ -d "$appdir" ]; then
        releasename=$(basename "$appdir")
        $run_cmd $sh_c "${HELM} upgrade -i ${releasename} ${appdir} -n user-space-${username} --set kubesphere.redis_password=${ks_redis_pwd} -f ${BASE_DIR}/wizard/config/apps/values.yaml"
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

# init_minio_cluster(){
#     MINIO_OPERATOR_VERSION="v0.0.1"
#     if [[ ! -f /etc/ssl/etcd/ssl/ca.pem || ! -f /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem || ! -f /etc/ssl/etcd/ssl/node-$HOSTNAME.pem ]]; then
#         echo "cann't find etcd key files"
#         exit $ERR_EXIT
#     fi

#     local minio_operator_tar="${BASE_DIR}/components/minio-operator-${MINIO_OPERATOR_VERSION}-linux-${ARCH}.tar.gz"
#     local minio_operator_bin="/usr/local/bin/minio-operator"

#     if [ ! -f "$minio_operator_bin" ]; then
#         if [ -f "$minio_operator_tar" ]; then
#             ensure_success $sh_c "cp ${minio_operator_tar} minio-operator-${MINIO_OPERATOR_VERSION}-linux-${ARCH}.tar.gz"
#         else
#             ensure_success $sh_c "curl ${CURL_TRY} -k -sfLO https://github.com/beclab/minio-operator/releases/download/${MINIO_OPERATOR_VERSION}/minio-operator-${MINIO_OPERATOR_VERSION}-linux-${ARCH}.tar.gz"
#         fi
# 	      ensure_success $sh_c "tar zxf minio-operator-${MINIO_OPERATOR_VERSION}-linux-${ARCH}.tar.gz"
#         ensure_success $sh_c "install -m 755 minio-operator $minio_operator_bin"
#     fi

#     ensure_success $sh_c "$minio_operator_bin init --address $local_ip --cafile /etc/ssl/etcd/ssl/ca.pem --certfile /etc/ssl/etcd/ssl/node-$HOSTNAME.pem --keyfile /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem --volume $MINIO_VOLUMES --password $MINIO_ROOT_PASSWORD"
# }


install_velero() {
    config_proxy_resolv_conf

    VELERO_VERSION="v1.11.3"
    local velero_tar="$terminus_base_dir/pkg/components/velero-${VELERO_VERSION}.tar.gz"
    if [ -f "$velero_tar" ]; then
        ensure_success $sh_c "cp ${velero_tar} /tmp/velero-${VELERO_VERSION}.tar.gz"
    fi
    ensure_success $sh_c "tar xf /tmp/velero-${VELERO_VERSION}.tar.gz -C /tmp/"
    ensure_success $sh_c "install /tmp/velero-${VELERO_VERSION}-linux-${ARCH}/velero /usr/local/bin"

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

get_natgateway() {
    natgateway=""

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

install_k8s_ks() {
    get_natgateway

    log_info 'Setup your first user ...\n'
    setup_ws

    # generate init config
    ADDON_CONFIG_FILE=${BASE_DIR}/wizard/bin/init-config.yaml
    echo '
    ' > ${ADDON_CONFIG_FILE}

    run_install

    # if [ "$storage_type" == "minio" ]; then
    #     # init minio-operator after etcd installed
    #     init_minio_cluster
    # fi

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
            read_tty "Enter the domain name ( myterminus.com by default ): " domainname
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
            read_tty "Enter the Terminus Name ( registered from TermiPass app ): " username
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
    local terminusd_installed="0"
    if [ ! -z "${AWS_SESSION_TOKEN_SETUP}" ]; then
        s3_sts="${AWS_SESSION_TOKEN_SETUP}"
        s3_ak="${AWS_ACCESS_KEY_ID_SETUP}"
        s3_sk="${AWS_SECRET_ACCESS_KEY_SETUP}"
    fi

    if [[ ! -z "${TERMINUS_BOX}" && x"${TERMINUS_BOX}" == x"1" ]]; then
        terminusd_installed="1"
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
terminusd: '${terminusd_installed}'

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

install_gpu(){

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
    local terminus_base_dir="$HOME/.terminus"
    local manifest_file="$BASE_DIR/installation.manifest"
    local extra
    TERMINUS_CLI=$(command -v terminus-cli)
    if [[ x"$ENV_BASE_DIR" != x"" ]]; then
        terminus_base_dir="$ENV_BASE_DIR"
    fi
    
    PARAM="--base-dir $terminus_base_dir --manifest $manifest_file --kube $KUBE_TYPE --version $VERSION"
    # TODO: install

    get_distribution
    get_shell_exec
        
    (
        # env 'REGISTRY_MIRRORS' is a docker image cache mirrors, separated by commas
        if [ x"$REGISTRY_MIRRORS" != x"" ]; then
            extra=" --registry-mirrors $REGISTRY_MIRRORS"
        fi


        if [ ! -f $terminus_base_dir/.prepared ]; then
            ensure_success $sh_c "export OS_LOCALIP=$local_ip && \
            export TERMINUS_IS_CLOUD_VERSION=$TERMINUS_IS_CLOUD_VERSION && \
            $TERMINUS_CLI terminus download component $PARAM"

            ensure_success $sh_c "export OS_LOCALIP=$local_ip && \
            export TERMINUS_IS_CLOUD_VERSION=$TERMINUS_IS_CLOUD_VERSION && \
            $TERMINUS_CLI terminus prepare $PARAM $extra"
        fi

        if [[ x"$PREINSTALL" != x"" ]]; then
            echo "Success to preinstall !!!"
            exit 0
        fi
        
        get_local_ip
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
