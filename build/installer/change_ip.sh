#!/usr/bin/env bash
ERR_EXIT=-1

old_ip=$1

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

	KUBECTL=$(command -v kubectl)
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

system_service_active() {
    if [[ $# -ne 1 || x"$1" == x"" ]]; then
        return 1
    fi

    local ret
    ret=$($sh_c "systemctl is-active $1")
    if [[ "$ret" == "active" || "$ret" == "activating" ]]; then
        return 0
    fi
    return 1
}

is_k3s(){
	if [ -f /etc/systemd/system/k3s.service ]; then
		return 0
	fi

	return 1
}

precheck_os() {
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

    # try to resolv hostname
    ensure_success $sh_c "hostname -i >/dev/null"

    local ip=$(ping -c 1 "$HOSTNAME" |awk -F '[()]' '/icmp_seq/{print $2}')
    printf "%s\t%s\n\n" "$ip" "$HOSTNAME"

    if [[ x"$ip" == x"" || "$ip" == @("172.17.0.1"|"127.0.0.1"|"127.0.1.1") || ! "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_fatal "incorrect ip for hostname '$HOSTNAME', please check"
    fi

	read -r -p "Are you sure changing this node ip to ${ip}? [yes/no]: " ans </dev/tty

    if [ x"$ans" != x"yes" ]; then
		echo "Please edit /etc/hosts to add the correct node IP"
        echo "exiting..."
        exit
    fi

    local_ip="$ip"
}

is_wsl(){
    wsl=$(uname -a 2>&1)
    if [[ ${wsl} == *WSL* ]]; then
        echo 1
        return
    fi

    echo 0
}


regen_cert_conf(){
	old_IFS=$IFS
	for pem in $1 ; do
		echo -e "[ req ]\ndefault_bits\t= 4096\ndistinguished_name\t= req_distinguished_name\nreq_extensions\t= v3_ext\nprompt\t= no\n[ req_distinguished_name ]" ; 
		IFS="," 

		for att in `openssl x509 -in $pem -text -noout | grep Subject: | cut -d: -f2 ` ;  

			do VALUE=`echo $att | cut -d= -f2-9 `; 
				case $att in 
				\ C\ =*) echo "countryName_default = $VALUE" ;; 
				\ ST\ =*) echo "StateOrProvinceName_default = $VALUE" ;; 
				\ L\ =*) echo "localityName_default = $VALUE";; 
				\ O\ =*) echo "organizationName_default = $VALUE" ;; 
				\ OU\ =*)  echo "organizationUnitName_default = $VALUE" ;; 
				\ CN\ =*)  echo "commonName = $VALUE" ;; 
			esac 
		done

		openssl x509 -in $pem -text | grep -A1 Subject\ Alternative\ Name | tail -1 | xargs echo -e "[ v3_ext ]\nsubjectAltName = "|sed -e 's/IP Address/IP/g'|sed -e "s/$old_ip/$local_ip/g"
	done
	IFS=$old_IFS
}




update_juicefs() {
	$sh_c "systemctl stop juicefs minio minio-operator redis-server"

	local TERMINUS_ROOT="/terminus"
    local fsname="rootfs"

	# update redis
	local redis_root="${TERMINUS_ROOT}/data/redis"
    local redis_conf="${redis_root}/etc/redis.conf"

	# get old ip
	if [ -z "$old_ip" ]; then
		old_ip=$($sh_c "awk '/bind/{print \$NF}' $redis_conf")
	fi

	while [ -z "$old_ip" ]; do
		read -r -p "Cannot find the previous IP, please input: " old_ip </dev/tty
	done

	echo "the previous IP is $old_ip"

	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/hosts"

	ensure_success $sh_c "sed -i 's/bind [0-9.]*/bind $local_ip/g' $redis_conf"
	
	ensure_success $sh_c "systemctl start redis-server"

    # eusure redis is started
    ensure_success $sh_c "( sleep 10 && systemctl --no-pager status redis-server ) || \
    ( systemctl restart redis-server && sleep 3 && systemctl --no-pager status redis-server ) || \
    ( systemctl restart redis-server && sleep 3 && systemctl --no-pager status redis-server )"

    local REDIS_PASSWORD=$($sh_c "awk '/requirepass/{print \$NF}' $redis_conf")
    if [ x"$REDIS_PASSWORD" == x"" ]; then
        echo "no redis password found in $redis_conf"
        exit $ERR_EXIT
    fi

    log_info 'try to connect redis'

    local pong=$(/usr/bin/redis-cli -h "$local_ip" -a "$REDIS_PASSWORD" ping 2>/dev/null)
    if [ x"$pong" != x"PONG" ]; then
        echo "failed to connect redis server: ${local_ip}:6379"
        exit $ERR_EXIT
    fi

	log_info 'update redis IP success'

	# update minio and minio-operator
	local MINIO_ROOT_USER=""
	local MINIO_ROOT_PASSWORD=""
	if [ -f /etc/default/minio ]; then
		log_info 'updating minio'

		ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/default/minio"

		ensure_success $sh_c "systemctl start minio"
		# postpone restart minio-operator, until etcd restarted

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

		log_info 'update minio IP success'

		storage_type="minio"
		MINIO_ROOT_USER="minioadmin"
		MINIO_ROOT_PASSWORD=$(awk -F '=' '/^MINIO_ROOT_PASSWORD/{print $2}' /etc/default/minio)
	fi


	# update juicefs
	local jfs_mountpoint="${TERMINUS_ROOT}/${fsname}"

	log_info 'updating juicefs'
	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/systemd/system/juicefs.service"

	ensure_success $sh_c "systemctl daemon-reload"
	ensure_success $sh_c "systemctl start juicefs"

	if [ "$storage_type" == "minio" ]; then
	    local juicefs_bin="/usr/local/bin/juicefs"
		local bucket="terminus"
		local metadb="redis://:${REDIS_PASSWORD}@${local_ip}:6379/1"

        ensure_success $sh_c "$juicefs_bin config $metadb --bucket http://${local_ip}:9000/${bucket} --access-key $MINIO_ROOT_USER --secret-key $MINIO_ROOT_PASSWORD"
	fi

	ensure_success $sh_c "systemctl --no-pager status juicefs"
    ensure_success $sh_c "sleep 3 && test -d $jfs_mountpoint/.trash"

	log_info 'update juicefs IP success'
}

update_minio_operator(){
	local MINIO_ROOT_PASSWORD=$(awk -F '=' '/^MINIO_ROOT_PASSWORD/{print $2}' /etc/default/minio)
	local MINIO_VOLUMES=$(awk -F '=' '/^MINIO_VOLUMES/{print $2}' /etc/default/minio)

	# re-init minio-operator, only used for uninitialized master node machine
	local ETCDCTL=$(command -v etcdctl)
	local minio_operator_bin="/usr/local/bin/minio-operator"

	# clear minio-operator service
	ensure_success $sh_c "rm -f /etc/default/minio-operator /etc/systemd/system/minio-operator.service"
	ensure_success $sh_c "$ETCDCTL --cacert /etc/ssl/etcd/ssl/ca.pem --cert /etc/ssl/etcd/ssl/node-$HOSTNAME.pem --key /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem del terminus/minio --prefix"

    ensure_success $sh_c "$minio_operator_bin init --address $local_ip --cafile /etc/ssl/etcd/ssl/ca.pem --certfile /etc/ssl/etcd/ssl/node-$HOSTNAME.pem --keyfile /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem --volume $MINIO_VOLUMES --password $MINIO_ROOT_PASSWORD"

	log_info "update minio-operator success"
}

update_k3s_master() {
#	ensure_success $sh_c "$KUBECTL delete node $HOSTNAME"

	ensure_success $sh_c "systemctl stop k3s etcd backup-etcd"
}

update_etcd(){
	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/etcd.env"
	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /usr/local/bin/kube-scripts/etcd-backup.sh"

	# renew etcd cert
	local tmpdir=$(mktemp -d)
	ensure_success $sh_c "mv /etc/ssl/etcd/ssl/* $tmpdir/."
	ensure_success $sh_c "cp $tmpdir/{ca.pem,ca-key.pem} /etc/ssl/etcd/ssl/."
	local confile="$tmpdir/cert.conf"
	ensure_success regen_cert_conf $tmpdir/admin-$HOSTNAME.pem > $confile

	for instance in admin-$HOSTNAME member-$HOSTNAME node-$HOSTNAME; do
		ensure_success $sh_c "openssl req -newkey rsa:2048 -nodes \
             -keyout /etc/ssl/etcd/ssl/${instance}-key.pem \
             -config ${confile} \
             -out /etc/ssl/etcd/ssl/${instance}-cert.csr"

		ensure_success $sh_c "openssl x509 -req \
             -extfile ${confile} \
             -extensions v3_ext \
             -in /etc/ssl/etcd/ssl/${instance}-cert.csr \
             -CA /etc/ssl/etcd/ssl/ca.pem \
             -CAkey /etc/ssl/etcd/ssl/ca-key.pem \
             -CAcreateserial \
             -out /etc/ssl/etcd/ssl/${instance}.pem \
             -days 3650 -sha256"
	done

    ensure_success $sh_c "systemctl daemon-reload"
	ensure_success $sh_c "systemctl start etcd backup-etcd"
}

post_update_k3s_master(){
	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/systemd/system/k3s.service"
	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/systemd/system/k3s.service.env"

    ensure_success $sh_c "systemctl daemon-reload"
	ensure_success $sh_c "systemctl start k3s"
	ensure_success $sh_c "systemctl --no-pager status k3s"

	log_info 'IP changed, the OS will be reloaded in 2 minutes...'
	sleep 120
	# check running pods
	ensure_success $sh_c "$KUBECTL get pods --all-namespaces"

}

update_k8s_master() {
    local KUBEADM=$(command -v kubeadm)

	ensure_success $sh_c "systemctl stop kubelet containerd"

	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/kubernetes/*.yaml"
	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/kubernetes/*.conf"
	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/kubernetes/manifests/*.yaml"
	ensure_success $sh_c "sed -i 's/$old_ip/$local_ip/g' /etc/kubernetes/addons/*.yaml"

	ensure_success $sh_c "rm -f /etc/kubernetes/pki/{apiserver*,front-proxy-client*}"
	ensure_success $sh_c "$KUBEADM init phase certs apiserver --config=/etc/kubernetes/kubeadm-config.yaml"
	ensure_success $sh_c "$KUBEADM init phase certs apiserver-kubelet-client --config=/etc/kubernetes/kubeadm-config.yaml"
	ensure_success $sh_c "$KUBEADM init phase certs front-proxy-client --config=/etc/kubernetes/kubeadm-config.yaml"

	ensure_success $sh_c "kubeadm init phase kubeconfig admin --config=/etc/kubernetes/kubeadm-config.yaml"
	ensure_success $sh_c "cp -f /etc/kubernetes/admin.conf /root/.kube/config"

    ensure_success $sh_c "systemctl daemon-reload"
	ensure_success $sh_c "systemctl start kubelet containerd"

	# restart k8s processes
	$sh_c "killall kube-apiserver" 
	$sh_c "killall kube-scheduler" 
	$sh_c "killall kube-controller-manager" 

	# wait for some time and delete old node
	log_info 'IP changed, the OS will be reloaded in 2 minutes...'
	sleep 120
	ensure_success $sh_c "$KUBECTL get nodes --sort-by=.metadata.creationTimestamp"

	# check running pods
	ensure_success $sh_c "$KUBECTL get pods --all-namespaces"
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

get_bfl_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'tier=bfl' -o jsonpath='{.items[*].status.phase}'"
}

get_settings_status(){
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=settings' -o jsonpath='{.items[*].status.phase}'"
}

get_all_user(){
    $sh_c "${KUBECTL} get user -o jsonpath='{.items[*].metadata.name}'"
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
    status=$(check_together get_appservice_status get_bfl_status get_vault_status get_profile_status get_auth_status get_desktop_status get_settings_status)
    n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        dotn=$(($n % 10))
        dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 0.5

        status=$(check_together get_appservice_status get_bfl_status get_vault_status  get_profile_status get_auth_status get_desktop_status get_settings_status)
        echo -ne "\rPlease waiting          "

    done
    echo
}


main() {
	get_shell_exec
	precheck_os

	local storage_type="s3"
	if is_k3s; then
		if system_service_active "k3s" ; then
			update_k3s_master
		fi 
	fi

    if [[ $(is_wsl) -eq 0 ]]; then
		update_juicefs
	fi
	
	update_etcd

	if is_k3s ; then
		log_info "updating k3s"

		post_update_k3s_master
	else
		log_info "updating k8s"

	    update_k8s_master
	fi 

	if [ "$storage_type" == "minio" ]; then 
		update_minio_operator
	fi

	# check os auto-reloading
    log_info 'Waiting for Terminus reloading ...'
    check_desktop

	for u in $(get_all_user) ; do
		$sh_c "${KUBECTL} rollout restart deploy -n user-space-$u edge-desktop"
		$sh_c "${KUBECTL} rollout restart deploy -n user-space-$u headscale-server"
	done

	$sh_c "killall envoy" 

    check_desktop

	log_info 'Success to change the Terminus IP address!'
}

main
