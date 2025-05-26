#!/usr/bin/env bash

TERMINUS_CLI_VERSION=0.1.21
ERR_EXIT=1

CURL_TRY="--connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 "

BASE_DIR="$(realpath $(dirname "$0"))/.."
RM=$(command -v rm)
KUBECTL=$(command -v kubectl)
KKE_FILE="/etc/kke/version"
STS_ACCESS_KEY=""
STS_SECRET_KEY=""
STS_TOKEN=""
STS_CLUSTER_ID=""

printf "root: ${BASE_DIR}\nversion: ${TERMINUS_CLI_VERSION}"

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

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

     OS_ARCH="$os_arch"
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

ensure_success() {
    "$@"
    local ret=$?

    if [ $ret -ne 0 ]; then
        echo "Fatal error, command: '$*'"
        exit $ret
    fi

    return $ret
}

log_info() {
    local msg now

    msg="$*"
    now=$(date +'%Y-%m-%d %H:%M:%S.%N %z')
    echo -e "\n\033[38;1m${now} [INFO] ${msg} \033[0m" 
}

get_kubelet_version() {
    [ ! -f $KUBECTL ] && {
        echo "kubectl does not exists"
        exit $ERR_EXIT
    }
    $sh_c "${KUBECTL} get nodes -o jsonpath='{.items[0].status.nodeInfo.kubeletVersion}'"
}

find_version(){
    if [ -f "$KKE_FILE" ]; then
        KKE_VERSION=$(awk -F '=' '/KKE/{printf "%s",$2}' $KKE_FILE)
        KUBE_VERSION=$(awk -F '=' '/KUBE/{printf "%s",$2}' $KKE_FILE)

        [ x"$KKE_VERSION" != x"" ] && [ x"$KUBE_VERSION" != x"" ] && return
    fi

    KKE_VERSION=0.1.21     # don't need to change it, as long as it's greater than 0.1.6

    local kube="$(get_kubelet_version)"
    if [ x"$kube" != x"" ]; then
        KUBE_VERSION="$kube"
        return
    fi

    echo "Warning: file $KKE_FILE does not exists, and kube version not be found"
}

find_storage_key(){
    STS_ACCESS_KEY=$($sh_c "${KUBECTL} get terminus terminus -o jsonpath='{.metadata.annotations.bytetrade\.io/s3-ak}'" &>/dev/null;true)
    STS_SECRET_KEY=$($sh_c "${KUBECTL} get terminus terminus -o jsonpath='{.metadata.annotations.bytetrade\.io/s3-sk}'" &>/dev/null;true)
    STS_TOKEN=$($sh_c "${KUBECTL} get terminus terminus -o jsonpath='{.metadata.annotations.bytetrade\.io/s3-sts}'" &>/dev/null;true)
    STS_CLUSTER_ID=$($sh_c "${KUBECTL} get terminus terminus -o jsonpath='{.metadata.labels.bytetrade\.io/cluster-id}'" &>/dev/null;true)
}

remove_cluster(){
    if [ x"$KUBE_VERSION" == x"" ]; then
        KUBE_VERSION="v1.22.10"
    fi

    if [ x"$KKE_VERSION" == x"" ]; then
        KKE_VERSION="0.1.21"
    fi

    forceUninstall="${FORCE_UNINSTALL_CLUSTER}"

    log_info 'remove kubernetes cluster'

    local kk_tar="${BASE_DIR}/install_wizard/components/kubekey-ext-v${KKE_VERSION}-linux-${ARCH}.tar.gz"

    if [ x"$PROXY" != x"" ]; then
        ensure_success $sh_c "cat /etc/resolv.conf > /etc/resolv.conf.bak"
        ensure_success $sh_c "echo nameserver $PROXY > /etc/resolv.conf"
        # if download failed
        if [ -f "${kk_tar}" ]; then
            ensure_success $sh_c "cp ${kk_tar} ${INSTALL_DIR}"
        else
            ensure_success $sh_c "curl ${CURL_TRY} -kLO https://github.com/beclab/kubekey-ext/releases/download/${KKE_VERSION}/kubekey-ext-v${KKE_VERSION}-linux-${ARCH}.tar.gz"
        fi

        ensure_success $sh_c "tar xf kubekey-ext-v${KKE_VERSION}-linux-${ARCH}.tar.gz"
        ensure_success $sh_c "cat /etc/resolv.conf.bak > /etc/resolv.conf"
    else 
    	ensure_success $sh_c "curl -sfL https://raw.githubusercontent.com/beclab/kubekey-ext/master/downloadKKE.sh | VERSION=${KKE_VERSION} bash -"
    fi
    ensure_success $sh_c "chmod +x kk"

    if [ -z "$forceUninstall" ]; then
        echo
        read -r -p "Are you sure to delete this cluster? [yes/no]: " ans </dev/tty

        if [ x"$ans" != x"yes" ]; then
            echo "exiting..."
            exit
        fi
    fi
    
    $sh_c "./kk delete cluster -A --with-kubernetes $KUBE_VERSION"

    [ -f $KKE_FILE ] && $sh_c "${RM} -f $KKE_FILE"

    if command_exists ipvsadm; then
        $sh_c "ipvsadm -C"
    fi
    $sh_c "iptables -F"

    $sh_c "killall /usr/local/bin/containerd || true"
}

docker_files=(/usr/bin/docker*
/var/lib/docker
/var/run/docker*
/var/lib/dockershim
/usr/local/bin/containerd
/etc/docker
/etc/cni/net.d)

clean_docker() {
    log_info 'destroy docker'

    $sh_c "rm -f /var/run/docker.sock; true"

    for srv in docker containerd; do
        $sh_c "systemctl stop $srv; systemctl disable $srv; true"
    done

    $sh_c "killall -9 containerd dockerd 2>/dev/null; true"

    local pids=$(ps -fea|grep containerd|grep -v grep|awk '{print $2}')
    if [ -n "$pids" ]; then
        $sh_c "kill -9 $pids 2>/dev/null; true"
    fi

    log_info 'clean docker files'

    for i in "${docker_files[@]}"; do
        $sh_c "rm -rf $i >/dev/null; true"
    done
}

terminus_files=(
/usr/local/bin/redis-*
/usr/bin/redis-*
/sbin/mount.juicefs
/etc/init.d/redis-server
/usr/local/bin/juicefs
/usr/local/bin/minio
/usr/local/bin/velero
/etc/systemd/system/redis-server.service
/etc/systemd/system/minio.service
/etc/systemd/system/minio-operator.service
/etc/systemd/system/juicefs.service
/etc/systemd/system/containerd.service
)

remove_storage() {
    log_info 'destroy storage'

    # stop and disable service
    for srv in juicefs minio minio-operator redis-server; do
        $sh_c "systemctl stop $srv 2>/dev/null; systemctl disable $srv 2>/dev/null; true"
    done
    
    $sh_c "killall -9 redis-server 2>/dev/null; true"
    $sh_c "rm -rf /var/jfsCache /terminus/jfscache 2>/dev/null; true"

    # read -r -p "Retain the stored terminus data? [default: yes]: " ans </dev/tty
    # if [[ "$ans" == @("no"|"n"|"N"|"No") ]]; then
    $sh_c "unlink /usr/bin/redis-server 2>/dev/null; unlink /usr/bin/redis-cli 2>/dev/null; true"

    log_info 'clean terminus files'

    for i in "${terminus_files[@]}"; do
        $sh_c "rm -f $i 2>/dev/null; true"
    done

    $sh_c "rm -rf /terminus 2>/dev/null; true"
    # fi
}

remove_mount() {
    version="${TERMINUS_IS_CLOUD_VERSION}"
    storage="${STORAGE}"
    s3_bucket="${S3_BUCKET}"

    if [ -z "$STS_ACCESS_KEY"]; then
        STS_ACCESS_KEY=${AWS_ACCESS_KEY_ID_SETUP}
    fi

    if [ -z "$STS_SECRET_KEY"]; then
        STS_SECRET_KEY=${AWS_SECRET_ACCESS_KEY_SETUP}
    fi

    if [ -z "$STS_TOKEN"]; then
        STS_TOKEN=${AWS_SESSION_TOKEN_SETUP}
    fi

    if [ -z "$STS_CLUSTER_ID" ]; then
        STS_CLUSTER_ID=${CLUSTER_ID}
    fi

    if [ x"$version" == x"true" ]; then
        log_info 'remove juicefs s3 mount'
        ensure_success $sh_c "apt install unzip"
        case "$storage" in
            "s3")
                local awscli_file="awscli-exe-linux-x86_64.zip"
                local awscli_tar="${BASE_DIR}/components/${awscli_file}"
                if ! command_exists aws; then 
                    if [ -f "${awscli_tar}" ]; then
                        ensure_success $sh_c "cp ${awscli_tar} ."
                    else
                        ensure_success $sh_c 'curl ${CURL_TRY} -kLO "https://awscli.amazonaws.com/${awscli_file}"'
                    fi
                    ensure_success $sh_c "unzip -q ${awscli_file}"
                    ensure_success $sh_c "./aws/install --update"
                fi

                AWS=$(command -v aws)

                s3=$($sh_c "echo $s3_bucket | rev | cut -d '.' -f 5 | rev")
                s3=$($sh_c "echo $s3 | sed 's/https/s3/'")

                log_info 'clean juicefs s3 mount'
                ensure_success $sh_c "AWS_ACCESS_KEY_ID=${STS_ACCESS_KEY} AWS_SECRET_ACCESS_KEY=${STS_SECRET_KEY} AWS_SESSION_TOKEN=${STS_TOKEN} ${AWS} s3 rm $s3/${STS_CLUSTER_ID} --recursive"
                ;;
            "oss")
                local osscli_file="ossutil-v1.7.18-linux-${ARCH}.zip"
                local osscli_tar="${BASE_DIR}/components/${osscli_file}"
                if ! command_exists ossutil64; then
                    if [ -f "${osscli_tar}" ]; then
                        ensure_success $sh_c "cp ${osscli_tar} ."
                    else
                        ensure_success $sh_c 'curl ${CURL_TRY} -kLO "https://github.com/aliyun/ossutil/releases/download/v1.7.18/${osscli_file}"'
                    fi

                    ensure_success $sh_c "unzip -q ${osscli_file}"
                    ensure_success $sh_c "mv ./ossutil-v1.7.18-linux-${ARCH}/* /usr/local/sbin/"

                    ensure_success $sh_c "chmod +x /usr/local/bin/ossutil*"
                fi

                oss=$($sh_c "echo $s3_bucket | rev | cut -d '.' -f 4 | rev")
                oss=$($sh_c "echo $oss | sed 's/https/oss/'")
                endpoint=$($sh_c "echo $s3_bucket | awk -F[/.] '{print \"https://\"\$(NF-2)\".\"\$(NF-1)\".\"\$NF}'")
                
                log_info 'clean juicefs oss mount'
                OSSUTIL=$(command -v ossutil64)
                ensure_success $sh_c "${OSSUTIL} rm ${oss}/${STS_CLUSTER_ID}/ --endpoint=${endpoint} --access-key-id=${STS_ACCESS_KEY} --access-key-secret=${STS_SECRET_KEY} --sts-token=${STS_TOKEN} -r -f >/dev/null"
                ;;
            *)
                ;;
        esac
    fi
}

set -o pipefail
set -e

precheck_os
get_shell_exec

INSTALL_DIR=/tmp/install_log

[[ -d ${INSTALL_DIR} ]] && $sh_c "${RM} -rf ${INSTALL_DIR}" 
mkdir -p ${INSTALL_DIR} && cd ${INSTALL_DIR}

log_info 'Uninstalling OS ...'
find_version
find_storage_key
remove_cluster
remove_storage
remove_mount
[[ ! -z $CLEAN_ALL ]] && clean_docker

cd -
$sh_c "${RM} -rf /tmp/install_log"
[[ -d install-wizard ]] && ${RM} -rf install-wizard
set +o pipefail
ls |grep install-wizard*.tar.gz | while read ar; do  ${RM} -f ${ar}; done

[[ -f /usr/local/bin/k3s-uninstall.sh ]] && $sh_c "/usr/local/bin/k3s-uninstall.sh"

log_info 'Uninstall OS success! '
