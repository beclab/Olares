#!/usr/bin/env bash



ERR_EXIT=1

RM=$(command -v rm)
KUBECTL=$(command -v kubectl)
KKE_FILE="/etc/kke/version"

command_exists() {
	command -v "$@" > /dev/null 2>&1
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

    KKE_VERSION=0.1.19       # don't need to change it, as long as it's greater than 0.1.6

    local kube="$(get_kubelet_version)"
    if [ x"$kube" != x"" ]; then
        KUBE_VERSION="$kube"
        return
    fi

    echo "Warning: file $KKE_FILE does not exists, and kube version not be found"
}

remove_cluster(){
    if [ x"$KUBE_VERSION" == x"" ]; then
        KUBE_VERSION="v1.22.10"
    fi

    if [ x"$KKE_VERSION" == x"" ]; then
        KKE_VERSION="0.1.19"
    fi

    forceUninstall="${FORCE_UNINSTALL_CLUSTER}"

    log_info 'remove kubernetes cluster'

    if [ x"$PROXY" != x"" ]; then
        ensure_success $sh_c "cat /etc/resolv.conf > /etc/resolv.conf.bak"
        ensure_success $sh_c "echo nameserver $PROXY > /etc/resolv.conf"
        # if download failed
        if [ -f "${HOME}/kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz" ]; then
            ensure_success $sh_c "cp ${HOME}/kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz ${INSTALL_DIR}"
        else
            ensure_success $sh_c "curl ${CURL_TRY} -kLO https://github.com/beclab/kubekey-ext/releases/download/${KKE_VERSION}/kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz"
        fi
        ensure_success $sh_c "tar xf kubekey-ext-v${KKE_VERSION}-linux-amd64.tar.gz"
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

    if [[ x"$version" == x"true" && x"$storage" == x"s3" ]]; then

        local s3
        log_info 'remove juicefs s3 mount'

        if ! command_exists aws; then 
            ensure_success $sh_c "apt install unzip"
            ensure_success $sh_c 'curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"'
            ensure_success $sh_c "unzip -q awscliv2.zip"
            ensure_success $sh_c "./aws/install --update"
        fi
        
        AWS=$(command -v aws)

        s3=$($sh_c "echo $s3_bucket | rev | cut -d '.' -f 5 | rev")
        s3=$($sh_c "echo $s3 | sed 's/https/s3/'")

        log_info 'clean juicefs s3 mount'
        ensure_success $sh_c "AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID_SETUP} AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY_SETUP} AWS_SESSION_TOKEN=${AWS_SESSION_TOKEN_SETUP} ${AWS} s3 rm $s3/${CLUSTER_ID} --recursive"
    fi
}

set -o pipefail
set -e

get_shell_exec

INSTALL_DIR=/tmp/install_log

[[ -d ${INSTALL_DIR} ]] && $sh_c "${RM} -rf ${INSTALL_DIR}" 
mkdir -p ${INSTALL_DIR} && cd ${INSTALL_DIR}

log_info 'Uninstalling OS ...'
find_version
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
