#!/binbash
source ./common.sh

ERR_EXIT=1

CURL_TRY="--connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 "

BASE_DIR=$(dirname $(realpath -s $0))
BASE_DIR=${BASE_DIR:-.}
CLUSTER_NAME=$1
PROFILE_NAME="terminus-${CLUSTER_NAME:-0}"

[[ -f "${BASE_DIR}/.env" && -z "$DEBUG_VERSION" ]] && . "${BASE_DIR}/.env"

read_tty(){
    echo -n $1
    read $2 < /dev/tty
}


function ensure_success() {
    "$@"
    local ret=$?

    if [ $ret -ne 0 ]; then
        echo "Fatal error, command: '$*'"
        exit $ret
    fi

    return $ret
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

install_helm() {
    if ! command_exists helm; then
        echo "Installing helm ..."
        curl -sSfL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
    fi
    if ! command_exists helm; then
        echo "Helm installation failed, please manually download and install the corresponding version of Helm."
        echo ""
        echo ""
        exit -1
    fi
}

install_cli(){
    KUBE_TYPE=${KUBE_TYPE}
    CLI_VERSION="0.1.14"
    if [ -z $KUBE_TYPE ]; then
        KUBE_TYPE="k3s"
    fi
    
    local cli_name="terminus-cli-v${CLI_VERSION}_${OSTYPE}_${ARCH}.tar.gz"
    local cli_tar="${BASE_DIR}/${cli_name}"
    if [ ! -f "$cli_tar" ]; then
        echo "Installing terminus-cli ..."
        ensure_success $sh_c "curl ${CURL_TRY} -k -sfL -o ${BASE_DIR}/${cli_name} https://github.com/beclab/Installer/releases/download/${CLI_VERSION}/${cli_name}"
    fi
    ensure_success $sh_c "tar xf ${BASE_DIR}/${cli_name} -C ${BASE_DIR}/"
}

install_ks(){
    # cmd="${BASE_DIR}/terminus-cli terminus install --kube ${KUBE_TYPE} --minikube --profile ${PROFILE_NAME}"
    # ensure_success $sh_c "${cmd}"
    ensure_success $sh_c "$TERMINUS_CLI terminus install $PARAM"
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

get_kscm_status(){
    $sh_c "${KUBECTL} get pod  -n kubesphere-system -l 'app=ks-controller-manager' -o jsonpath='{.items[*].status.phase}' 2>/dev/null"
}

get_vault_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=vault' -o jsonpath='{.items[*].status.phase}'"
}

get_appservice_status(){
    $sh_c "${KUBECTL} get pod  -n os-system -l 'tier=app-service' -o jsonpath='{.items[*].status.phase}'"
}

get_bfl_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'tier=bfl' -o jsonpath='{.items[*].status.phase}'"
}

get_bfl_node(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'tier=bfl' -o jsonpath='{.items[*].spec.nodeName}'"
}

get_appservice_pod(){
    $sh_c "${KUBECTL} get pod  -n os-system -l 'tier=app-service' -o jsonpath='{.items[*].metadata.name}'"
}

get_ksapi_status(){
    $sh_c "${KUBECTL} get pod  -n kubesphere-system -l 'app=ks-apiserver' -o jsonpath='{.items[*].status.phase}' 2>/dev/null"
}

get_settings_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=settings' -o jsonpath='{.items[*].status.phase}'"
}

get_app_key_secret(){
    app=$1
    key="bytetrade_${app}_${RANDOM}"
    secret=$(get_random_string 16)

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

repeat(){
    for _ in $(seq 1 "$1"); do
        echo -n "$2"
    done
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

check_settings(){
    status=$(get_settings_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for settings starting ${dot}"
        sleep 0.5

        status=$(get_settings_status)
        echo -ne "\rWaiting for settings starting          "

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

validate_domainname() {
    local match
    match=$(echo $domainname |egrep -o '^([a-z0-9])(([a-z0-9-]{1,61})?[a-z0-9]{1})?(\.[a-z0-9](([a-z0-9-]{1,61})?[a-z0-9]{1})?)?(\.[a-zA-Z]{2,10})+$')

    if [ x"$match" != x"$domainname" ]; then
        printf "illegal domain name '$domainname', try again\n\n"
        return 1
    fi
    return 0
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

    if ! command_exists htpasswd; then
        log_fatal "Please install htpasswd"
    fi

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
        useremail="${username}@${domainname}"
    fi

    if ! validate_useremail; then
        log_fatal "illegal user email '$useremail'"
    fi

    if [ -z "$userpwd" ]; then
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

  ensure_success $sh_c "$SED 's/#__DOMAIN_NAME__/${domainname}/' ${BASE_DIR}/wizard/config/settings/templates/terminus_cr.yaml"

  publicIp=$(curl --connect-timeout 5 -sL http://169.254.169.254/latest/meta-data/public-ipv4 2>&1)
  publicHostname=$(curl --connect-timeout 5 -sL http://169.254.169.254/latest/meta-data/public-hostname 2>&1)

  local selfhosted="true"
  if [[ ! -z "${TERMINUS_IS_CLOUD_VERSION}" && x"${TERMINUS_IS_CLOUD_VERSION}" == x"true" ]]; then
    selfhosted="false"
  fi
  if [[ x"$publicHostname" =~ "amazonaws" && -n "$publicIp" && ! x"$publicIp" =~ "Not Found" ]]; then
    selfhosted="false"
  fi
  ensure_success $sh_c "$SED 's/#__SELFHOSTED__/${selfhosted}/' ${BASE_DIR}/wizard/config/settings/templates/terminus_cr.yaml"
}

run_install(){
    GPU_TYPE="none"
    HELM=$(get_command helm)
    KUBECTL=$(get_command kubectl)

    install_ks

    check_kscm # wait for ks launch
    check_ksapi

    retry_cmd $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-k3s.yaml"

    log_info 'Installing account ...'
    # add the first account
    local xargs=""
    if [[ x"$natgateway" != x"" ]]; then
        echo "annotate bfl with nat gateway ip"
        xargs="--set nat_gateway_ip=${natgateway}"
    fi
    retry_cmd $sh_c "${HELM} upgrade -i account ${BASE_DIR}/wizard/config/account --force ${xargs}"

    log_info 'Installing settings ...'
    ensure_success $sh_c "${HELM} upgrade -i settings ${BASE_DIR}/wizard/config/settings --force"

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
    retry_cmd $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-globalrole-workspace-manager.yaml"
    retry_cmd $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-notification-manager.yaml"

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
    ensure_success $sh_c "${HELM} upgrade -i launcher-${username} ${BASE_DIR}/wizard/config/launcher -n user-space-${username} --force --set bfl.appKey=${bfl_ks[0]} --set bfl.appSecret=${bfl_ks[1]}"

    log_info 'waiting for bfl'
    check_bfl
    bfl_node=$(get_bfl_node)

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
  url: ''
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
fs_type: fs

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
    # ensure_success $sh_c "${KUBECTL} scale deployment/ks-installer --replicas=0 -n kubesphere-system"
    ensure_success $sh_c "${KUBECTL} delete deployment -n kubesphere-controls-system default-http-backend"


    # delete storageclass accessor webhook
    # ensure_success $sh_c "${KUBECTL} delete validatingwebhookconfigurations storageclass-accessor.storage.kubesphere.io"

    # calico config for tailscale
    ensure_success $sh_c "${KUBECTL} patch felixconfiguration default -p '{\"spec\":{\"featureDetectOverride\": \"SNATFullyRandom=false,MASQFullyRandom=false\"}}' --type='merge'"
}


main(){
    log_info 'Start to Install Terminus ...\n'
    HOSTNAME=$(hostname)
    natgateway=$(ping -c 1 "$HOSTNAME" |awk -F '[()]' '/PING/{print $2}')
    natgateway=$(echo "$natgateway" | grep -E "[0-9]+(\.[0-9]+){3}" | grep -v "127.0.0.1")

    precheck_support

    if [ x"$natgateway" == x"" ]; then
        while :; do
            read_tty "Enter the host IP: " natgateway
            natgateway=$(echo "$natgateway" | grep -E "[0-9]+(\.[0-9]+){3}" | grep -v "127.0.0.1")
            if [ x"$natgateway" == x"" ]; then
                continue
            fi
            break
        done
    fi

    sh_c="sh -c"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        TAR=gtar
        SED="sed -i '' -e"
    else
        TAR=tar
        SED="sed -i"
    fi

    install_helm

    # install_cli
    local terminus_base_dir="$HOME/.terminus"
    local manifest_file="$BASE_DIR/installation.manifest"
    local extra
    TERMINUS_CLI=$(command -v terminus-cli)
    if [[ x"$ENV_BASE_DIR" != x"" ]]; then
        terminus_base_dir="$ENV_BASE_DIR"
    fi

    PARAM="--base-dir $terminus_base_dir --manifest $manifest_file --version $VERSION --minikube --profile ${PROFILE_NAME}"

    if [ ! -f $terminus_base_dir/.prepared ]; then
        ensure_success $sh_c "export OS_LOCALIP=$local_ip && \
            export TERMINUS_IS_CLOUD_VERSION=$TERMINUS_IS_CLOUD_VERSION && \
            $TERMINUS_CLI terminus download --base-dir $terminus_base_dir --manifest $manifest_file --version $VERSION"

        if command_exists minikube ; then
            running=$(minikube profile list|grep "${PROFILE_NAME}"|grep Running)
            if [ x"$running" == x"" ]; then
                ensure_success $sh_c "minikube start -p '${PROFILE_NAME}' --kubernetes-version=v1.22.10 --network-plugin=cni --cni=calico --cpus='4' --memory='8g' --ports=30180:30180,443:443,80:80"
            fi
        else
            log_fatal "Please install minikube on your machine"
        fi
        touch $terminus_base_dir/.prepared
    fi


    setup_ws

    run_install

    log_info 'Waiting for Vault ...'
    check_vault

    log_info 'Starting Terminus ...'
    ensure_success $sh_c "${KUBECTL} rollout restart sts bfl -n user-space-${username}"
    check_desktop

    check_settings

    log_info 'Installation wizard is complete\n'


    # install complete
    echo -e " Terminus is running"
    echo -e " Open your browser and visit."
    echo -e "${GREEN_LINE}"
    echo -e " http://${natgateway}:30180/"
    echo -e "${GREEN_LINE}"
    echo -e " "
    echo -e " User: ${username} "
    echo -e " Password: ${userpwd} "
    echo -e " "
    echo -e " Please change the default password after login."
}

main | tee macos_install.log
