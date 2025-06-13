#!/usr/bin/env bash



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

repaire_crd_terminus() {
    local patch
    export KUBECONFIG=/root/.kube/config  # for ubuntu
	KUBECTL=$(command -v kubectl)

    if [ ! -z "${AWS_SESSION_TOKEN_SETUP}" ]; then
        patch='[{"op":"add","path":"/metadata/annotations/bytetrade.io~1s3-sts","value":"'"$AWS_SESSION_TOKEN_SETUP"'"},{"op":"add","path":"/metadata/annotations/bytetrade.io~1s3-ak","value":"'"$AWS_ACCESS_KEY_ID_SETUP"'"},{"op":"add","path":"/metadata/annotations/bytetrade.io~1s3-sk","value":"'"$AWS_SECRET_ACCESS_KEY_SETUP"'"},{"op":"add","path":"/metadata/annotations/bytetrade.io~1cluster-id","value":"'"$CLUSTER_ID"'"}]'
        $sh_c "${KUBECTL} patch terminus.sys.bytetrade.io terminus --type='json' -p='$patch'"
    fi
}


get_shell_exec

juicefs_bin="/usr/local/bin/juicefs"
ip=$(ping -c 1 "$HOSTNAME" |awk -F '[()]' '/icmp_seq/{print $2}')
pwd=$($sh_c "awk '/requirepass/{print \$NF}' /olares/data/redis/etc/redis.conf")


$sh_c "${juicefs_bin} config redis://:${pwd}@${ip}:6379/1 --access-key ${AWS_ACCESS_KEY_ID_SETUP} --secret-key ${AWS_SECRET_ACCESS_KEY_SETUP} --session-token ${AWS_SESSION_TOKEN_SETUP}"

# update terminus cr
repaire_crd_terminus