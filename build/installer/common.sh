#!/bin/bash



######## log ########
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
    exit -1
}

######## system ########
get_command() {
    echo $(command -v "$@")
}

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

function dpkg_locked() {
    grep -q 'Could not get lock /var/lib' "$fd_errlog"
    return $?
}

read_tty(){
    echo -n $1
    read $2 < /dev/tty
}

repeat(){
    for _ in $(seq 1 "$1"); do
        echo -n "$2"
    done
}

sleep_waiting(){
    local t=$1
    local n=0
    local max_retries=$((t*2))
    while [ $max_retries -gt 0 ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 0.5
        echo -ne "\rPlease waiting           "

        ((max_retries--))
    done
    echo
    echo "Continue ... "
}

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

function ensure_execute() {
    "$@"
    local ret=$?

    if [ $ret -ne 0 ]; then
        echo "Fatal error, command: '$*'"
        exit $ret
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



######## os ########
get_distribution() {
	lsb_dist=""
	# Every system that we officially support has /etc/os-release
	if [ -r /etc/os-release ]; then
		lsb_dist="$(. /etc/os-release && echo "$ID")"
	fi
	echo "$lsb_dist"
}

get_macos_versioin() {
    version=$(sw_vers -productVersion)
    major_version=$(echo "$version" | cut -d '.' -f 1)
    minor_version=$(echo "$version" | cut -d '.' -f 2)
    case "$major_version.$minor_version" in
        10.14.*)
            os_name="Mojave"
            ;;
        10.15.*)
            os_name="Catalina"
            ;;
        11.*)
            os_name="Big Sur"
            ;;
        12.*)
            os_name="Monterey"
            ;;
        13.*)
            os_name="Ventura"
            ;;
        14.*)
            os_name="Sonoma"
            ;;
        *)
            os_name="Unknown"
            ;;
    esac
    echo "$os_name"
}

precheck_support() {
    os_type=$(uname -s)
    case "$os_type" in
        Linux) OSTYPE=linux; ;;
        Darwin) OSTYPE=darwin; ;;
        *) echo "unsupported os type '${os_type}', exit ...";
        exit -1; ;;
    esac

    os_arch=$(uname -m)
    case "$os_arch" in 
        arm64) ARCH=arm64; ;; 
        x86_64) ARCH=amd64; ;; 
        armv7l) ARCH=arm; ;; 
        aarch64) ARCH=arm64; ;; 
        ppc64le) ARCH=ppc64le; ;; 
        s390x) ARCH=s390x; ;; 
        *) echo "unsupported arch '${os_arch}', exit ..."; 
        exit -1; ;; 
    esac

    if [ "$OSTYPE" == "darwin" ]; then
        OSNAME=$(get_macos_versioin)
        OSVERSION=$(sw_vers -productVersion)
    else
        OSNAME=$(. /etc/os-release && echo "$ID")
        OSVERSION=$(. /etc/os-release && echo "$VERSION_ID")
        lsb_release=$(echo "${OSNAME} ${OSVERSION}" | xargs)

        if [ -z "$lsb_release" ]; then
            echo "unsupported os version '${lsb_release}', exit ..."
            exit -1
        fi

        case "$lsb_release" in
            *Debian* | *debian*)
                case "$lsb_release" in
                    *12* | *11*)
                        ;;
                    *)
                        echo "unsupported os version '${lsb_release}', exit ...";
                        exit -1
                        ;;
                esac
                ;;
            *Ubuntu* | *ubuntu*)
                case "$lsb_release" in
                    *24.*)
                        ;;
                    *22.* | *20.*)
                        ;;
                    *)
                        echo "unsupported os version '${lsb_release}', exit ...";
                        exit -1
                        ;;
                esac
                ;;
            *)
                echo "unsupported os version '${lsb_release}', exit ...";
                exit -1
                ;;
        esac
    fi

    echo "$OSTYPE $OSNAME $OSVERSION $ARCH"
}

get_shell_exec() {
    if [ $(is_darwin) -eq 0 ]; then
        user="$(id -un 2>/dev/null || true)"

    sh_c='sh -c'
	if [ "$user" != 'root' ]; then
		if command_exists sudo && command_exists su; then
			sh_c='sudo -E sh -c'
		else
			cat >&2 <<-'EOF'
			Error: this installer needs the ability to run commands as root.
			We are unable to find either "sudo" or "su" available to make this happen.
			EOF
			exit -1
		fi
	fi
    fi
}

get_os_version() {
  echo ${OSVERSION}
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


is_darwin() {
    os_type=$(uname -s 2>&1)
    if [ "$os_type" == "Darwin" ]; then
        echo 1
        return
    fi
    echo 0
}

is_debian() {
    os_name=$(. /etc/os-release && echo "$ID")
    if [ "$os_name" == "debian" ]; then
        echo 1
        return
    fi
    echo 0
}

is_ubuntu() {
    if [ $OSNAME == *ubuntu* ]; then
        echo 1
        return
    fi
    echo 0
}

is_pve() {
    pve=$(uname -a 2>&1)
    pveversion=$(command -v pveversion)
    if [[ ${pve} == *pve* || ! -z $pveversion ]]; then
        echo 1
        return
    fi
    echo 0
}

is_raspbian() {
    rasp=$(uname -a 2>&1)
    if [[ ${rasp} == *Raspbian* || ${rasp} == *raspbian* || ${rasp} == *raspberry* || ${rasp} == *Raspberry* ]];then
        echo 1
        return
    fi
    echo 0
}

is_wsl() {
    wsl=$(uname -a 2>&1)
    if [[ ${wsl} == *WSL* ]]; then
        echo 1
        return
    fi
    echo 0
}

is_k3s(){
	if [ -f /etc/systemd/system/k3s.service ]; then
		return 0
	fi

	return 1
}

### dns and hosts
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



######## string ########
get_random_string() {
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

get_local_ip(){
    ip=$(ping -c 1 "$HOSTNAME" |awk -F '[()]' '/icmp_seq/{print $2}')
    echo "$ip  $HOSTNAME"

    if [[ x"$ip" == x"" || "$ip" == "172.17.0.1" || "$ip" == "127.0.0.1" || "$ip" == "127.0.1.1" ]]; then
        echo "incorrect ip for hostname '$HOSTNAME', please check"
        exit $ERR_EXIT
    fi

    local_ip="$ip"
}


### in cluster functions
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

### check terminus
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

reverse_proxy_config() {
    # set default-reverse-proxy-config
    local enableCloudflare="1"
    local enableFrp="0"
    local frpServer=""
    local frpPort="0"
    local frpAuthMethod=""
    local frpAuthToken=""

    if [ "${TERMINUS_IS_CLOUD_VERSION}" == "true" ]; then
        enableCloudflare="0"
    elif [ x"${FRP_ENABLE}" == x"1" ]; then
        enableCloudflare="0"
        enableFrp="1"

        frpServer=${DEFAULT_FRP_SERVER}  # default frp server
        if [ ! -z ${frpServer} ]; then
            frpPort="0"
            frpAuthMethod="jws"
        elif [ ! -z ${FRP_SERVER} ]; then
            frpServer=${FRP_SERVER}
            frpPort=${FRP_PORT}
            frpAuthMethod=${FRP_AUTH_METHOD}
            frpAuthToken=${FRP_AUTH_TOKEN}
        fi
    fi
    

    # if [[ x"${enableFrp}" == x"1" && -z "${frpServer}" ]]; then
    #     echo "FrpServer configuration is incorrect, please check ..."
    #     exit $ERR_EXIT
    # fi

    cat > ${BASE_DIR}/deploy/cm-default-reverse-proxy-config.yaml << _END
apiVersion: v1
data:
  cloudflare.enable: "${enableCloudflare}"
  frp.enable: "${enableFrp}"
  frp.server: "${frpServer}"
  frp.port: "${frpPort}"
  frp.auth_method: "${frpAuthMethod}"
  frp.auth_token: "${frpAuthToken}"
kind: ConfigMap
metadata:
  name: default-reverse-proxy-config
  namespace: os-system
_END
    ensure_execute $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/cm-default-reverse-proxy-config.yaml"
}