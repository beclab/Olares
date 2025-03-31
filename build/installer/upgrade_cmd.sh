#!/usr/bin/env bash




# Upgrading will be executed in  app-service container based on kubesphere/kubectl:v1.22.9
# By default, the tool packages will be installed via apt during the docker build

# env:
# BASE_DIR


function command_exists() {
	command -v "$@" > /dev/null 2>&1
}

function get_shell_exec(){
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

function get_bfl_api_port(){
    local username=$1
    $sh_c "${KUBECTL} get svc bfl -n user-space-${username} -o jsonpath='{.spec.ports[0].nodePort}'"
}

# function get_docs_port(){
#     local username=$1
#     $sh_c "${KUBECTL} get svc swagger-ui -n user-space-${username} -o jsonpath='{.spec.ports[0].nodePort}'"
# }

function get_desktop_port(){
    local username=$1
    $sh_c "${KUBECTL} get svc edge-desktop -n user-space-${username} -o jsonpath='{.spec.ports[0].nodePort}'"
}

function get_user_password(){
    local username=$1
    $sh_c "${KUBECTL} get user ${username} -o jsonpath='{.spec.password}'"
}

function get_user_email(){
    local username=$1
    $sh_c "${KUBECTL} get user ${username} -o jsonpath='{.spec.email}'"
}


function ensure_success() {
    "$@"
    local ret=$?

    if [ $ret -ne 0 ]; then
        echo "Fatal error, command: '$@'"
        exit $ret
    fi

    return $ret
}

function validate_user(){
    local username=$1
    $sh_c "${KUBECTL} get ns user-space-${username} > /dev/null" 
    local ret=$?
    if [ $ret -ne 0 ]; then
        echo "no"
    else
        echo "yes"
    fi
}

function get_bfl_node(){
    local username=$1
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'tier=bfl' -o jsonpath='{.items[*].spec.nodeName}'"
}

function get_bfl_url() {
    local username=$1
    local user_bfl_port=$(get_bfl_api_port ${username})

    bfl_ip=$(curl -s http://checkip.dyndns.org/ | grep -o "[[:digit:].]\+")
    echo "http://$bfl_ip:${user_bfl_port}/bfl/apidocs.json"
}

function get_userspace_dir(){
    local username=$1
    local space_dir=$2
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'tier=bfl' -o \
    jsonpath='{range .items[0].spec.volumes[*]}{.name}{\" \"}{.persistentVolumeClaim.claimName}{\"\\n\"}{end}'" | \
    while read pvc; do
        local pvc_data=($pvc)
        if [ ${#pvc_data[@]} -gt 1 ]; then
            if [ "x${pvc_data[0]}" == "x${space_dir}" ]; then
                local USERSPACE_PVC="${pvc_data[1]}"
                local pv=$($sh_c "${KUBECTL} get pvc -n user-space-${username} ${pvc_data[1]} -o jsonpath='{.spec.volumeName}'")
                local pv_path=$($sh_c "${KUBECTL} get pv ${pv} -o jsonpath='{.spec.hostPath.path}'")
                local USERSPACE_PV_PATH="${pv_path}"

                echo "${USERSPACE_PVC} ${USERSPACE_PV_PATH} ${pv}"
                break
            fi
        fi
    done 
}

function get_bfl_rand16(){
    local username=$1
    local prefix=$2

    $sh_c "${KUBECTL} get sts -n user-space-${username} bfl -o jsonpath='{.metadata.annotations.${prefix}_rand16}'"
}

function gen_app_key_secret(){
    local app=$1
    local key="bytetrade_${app}_${RANDOM}"
    local t=$(date +%s)
    local secret=$(echo -n "${key}|${t}"|md5sum|cut -d" " -f1)

    echo "${key} ${secret:0:16}"
}

function get_app_key_secret(){
    local username=$1
    local app=$2

    local ks=$($sh_c "${KUBECTL} get appperm ${app} -n user-system-${username} -o jsonpath='{.spec.key} {.spec.secret}'")

    if [ "x${ks}" == "x" ]; then
        ks=$(gen_app_key_secret "${app}")
    fi

    echo "${ks}"
}


function get_app_settings(){
    local username=$1
    local apps=("vault" "desktop" "message" "wise" "search" "appstore" "notification" "dashboard" "settings" "studio" "profile" "agent" "files")
    for a in ${apps[@]};do
        ks=($(get_app_key_secret "$username" "$a"))
        echo '
  '${a}':
    appKey: '${ks[0]}'    
    appSecret: "'${ks[1]}'"    
        '
    done
}



function gen_bfl_values(){
    local username=$1
    local user_bfl_port=$(get_bfl_api_port ${username})

    echo "Try to find the current bfl pv ..."
    local pvc_path=($(get_userspace_dir ${username} "userspace-dir"))
    local appcache_pvc_path=($(get_userspace_dir ${username} "appcache-dir"))
    local dbdata_pvc_path=($(get_userspace_dir ${username} "dbdata-dir"))

    local userspace_rand16=$(get_userspace_dir ${username} "userspace")
    local appcache_rand16=$(get_userspace_dir ${username} "Cache")
    local dbdata_rand16=$(get_userspace_dir ${username} "dbdata")

    echo '
bfl:
  nodeport: '${user_bfl_port}'
  username: '${username}'

  userspace_rand16: '${userspace_rand16}'
  userspace_pv: '${pvc_path[2]}'
  userspace_pvc: '${pvc_path[0]}'

  appcache_rand16: '${appcache_rand16}'
  appcache_pv: '${appcache_pvc_path[2]}'
  appcache_pvc: '${appcache_pvc_path[0]}'

  dbdata_rand16: '${dbdata_rand16}'
  dbdata_pv: '${dbdata_pvc_path[2]}'
  dbdata_pvc: '${dbdata_pvc_path[0]}'
  ' > ${BASE_DIR}/wizard/config/launcher/values.yaml
}


function gen_settings_values(){
    local username=$1
    # local userpwd="$(get_user_password ${username})"
    # local useremail="$(get_user_email ${username})"

    echo '
namespace:
  name: user-space-'${username}'
  role: admin

user:
  name: '${username}'
    ' > ${BASE_DIR}/wizard/config/settings/values.yaml
}

function gen_app_values(){
    local username=$1

    local bfl_node=$(get_bfl_node ${username})
    local bfl_doc_url=$(get_bfl_url ${username})
    local desktop_ports=$(get_desktop_port ${username})
#    local docs_ports=$(get_docs_port ${username})

    echo "Try to find pv ..."
    local pvc_path=($(get_userspace_dir ${username} "userspace-dir"))
    local appcache_pvc_path=($(get_userspace_dir ${username} "appcache-dir"))
    local dbdata_pvc_path=($(get_userspace_dir ${username} "dbdata-dir"))

    local app_perm_settings=$(get_app_settings ${username})
    cat ${BASE_DIR}/wizard/config/launcher/values.yaml > ${BASE_DIR}/wizard/config/apps/values.yaml
    cat << EOF >> ${BASE_DIR}/wizard/config/apps/values.yaml
  url: '${bfl_doc_url}'
  nodeName: ${bfl_node}
pvc:
  userspace: ${pvc_path[0]}
userspace:
  appCache: ${appcache_pvc_path[1]}
  dbdata: ${dbdata_pvc_path[1]}
  userData: ${pvc_path[1]}/Home
  appData: ${pvc_path[1]}/Data

desktop:
  nodeport: ${desktop_ports}
os:
  ${app_perm_settings}
EOF
}

function close_apps(){
    local username=$1
    local app_list=(
        "vault-deployment"
    )


    for app in ${app_list[@]} ; do
        $sh_c "${KUBECTL} scale deployment ${app} -n user-space-${username} --replicas=0"
    done
}

repeat(){
    for i in $(seq 1 $1); do
        echo -n $2
    done
}

function get_appservice_pod(){
    $sh_c "${KUBECTL} get pod  -n os-system -l 'tier=app-service' -o jsonpath='{.items[*].metadata.name}'"
}

function get_appservice_status(){
    $sh_c "${KUBECTL} get pod  -n os-system -l 'tier=app-service' -o jsonpath='{.items[*].status.phase}'"
}

function get_desktop_status(){
    local username=$1
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=edge-desktop' -o jsonpath='{.items[*].status.phase}'"
}

function get_vault_status(){
    local username=$1
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=vault' -o jsonpath='{.items[*].status.phase}'"
}


function get_bfl_status(){
    local username=$1
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'tier=bfl' -o jsonpath='{.items[*].status.phase}'"
}

function get_fileserver_status(){
    $sh_c "${KUBECTL} get pod  -n os-system -l 'app=files' -o jsonpath='{.items[*].status.phase}'"
}

function get_filefe_status(){
    local username=$1
    $sh_c "${KUBECTL} get pod  -n user-space-${username} -l 'app=files' -o jsonpath='{.items[*].status.phase}'"
}

function check_fileserver(){
    local status=$(get_fileserver_status)
    local n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        local dotn=$(($n % 10))
        local dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for file-server starting ${dot}"
        sleep 0.5

        status=$(get_fileserver_status)
        echo -ne "\rWaiting for file-server starting          "

    done
    echo
}

function check_appservice(){
    local status=$(get_appservice_status)
    local n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        local dotn=$(($n % 10))
        local dot=$(repeat $dotn '>')

        echo -ne "\rWaiting for app-service starting ${dot}"
        sleep 0.5

        status=$(get_appservice_status)
        echo -ne "\rWaiting for app-service starting          "

    done
    echo
}

function check_filesfe(){
    local username=$1
    local status=$(get_filefe_status ${username})
    local n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        local dotn=$(($n % 10))
        local dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 0.5

        status=$(get_filefe_status ${username})
        echo -ne "\rPlease waiting          "

    done
    echo
}

function check_bfl(){
    local username=$1
    local status=$(get_bfl_status ${username})
    local n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        local dotn=$(($n % 10))
        local dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 0.5

        status=$(get_bfl_status ${username})
        echo -ne "\rPlease waiting          "

    done
    echo
}

function check_desktop(){
    local username=$1
    local status=$(get_desktop_status ${username})
    local n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        local dotn=$(($n % 10))
        local dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 0.5

        status=$(get_desktop_status ${username})
        echo -ne "\rPlease waiting          "

    done
    echo
}

function check_vault(){
    local username=$1
    local status=$(get_vault_status ${username})
    local n=0
    while [ "x${status}" != "xRunning" ]; do
        n=$(expr $n + 1)
        local dotn=$(($n % 10))
        local dot=$(repeat $dotn '>')

        echo -ne "\rPlease waiting ${dot}"
        sleep 0.5

        status=$(get_vault_status ${username})
        echo -ne "\rPlease waiting          "

    done
    echo
}

function check_all(){
    local pods=$@
    for p in ${pods[@]}; do
        local n=$(echo "${p}"|awk -F"@" '{print $1}')
        local ns=$(echo "${p}"|awk -F"@" '{print $2}')
        local s=$($sh_c "${KUBECTL} get pod  -n ${ns} -l 'app=${n}' -o jsonpath='{.items[*].status.phase}'")
        echo -ne "\rPlease wait: ${p}"
        while [ "x${s}" != "xRunning" ];do
            echo -ne "\rPlease wait: ${p}"

            s=$($sh_c "${KUBECTL} get pod  -n ${ns} -l 'app=${n}' -o jsonpath='{.items[*].status.phase}'")
        done
        echo
    done
}

function upgrade_ksapi(){
    local users=$@
    local current_version="beclab/ks-apiserver:v3.3.0-ext-3"
    local image=$($sh_c "${KUBECTL} get deploy ks-apiserver -n kubesphere-system -o jsonpath='{.spec.template.spec.containers[0].image}'")
    if [ "x${image}" != "x${current_version}" ]; then
        echo "upgrade ks-apiserver and restore token ..."

        secret=$(echo -n "ks_redis_${RANDOM}"|md5sum|cut -d" " -f1)
        $sh_c "${KUBECTL} -n kubesphere-system create secret generic redis-secret --from-literal=auth=${secret:0:12}"

        local old_jwt=$($sh_c "${KUBECTL} get configmap  kubesphere-config -n kubesphere-system -o jsonpath='{.data.kubesphere\.yaml}'|grep jwtSecret|awk -F':' '{print \$2}'")
        sed -i -e "s/__jwtkey__/${old_jwt}/" ${BASE_DIR}/deploy/cm-kubesphere-config.yaml

        $sh_c "${KUBECTL} apply -f ${BASE_DIR}/deploy/redis-deploy.yaml"
        $sh_c "${KUBECTL} apply -f ${BASE_DIR}/deploy/cm-kubesphere-config.yaml"
        check_all "redis@kubesphere-system"

        $sh_c "${KUBECTL} -n kubesphere-system set image deployment/ks-apiserver ks-apiserver=beclab/ks-apiserver:v3.3.0-ext-3"
        $sh_c "${KUBECTL} patch deploy ks-apiserver -n kubesphere-system --patch-file=${BASE_DIR}/deploy/ks-apiserver-patch.yaml"

        check_all "ks-apiserver@kubesphere-system"

        for username in ${users[@]}; do
            $sh_c "${KUBECTL} rollout restart deploy authelia-backend -n user-system-${username}"

            check_all "authelia-backend@user-system-${username}"
        done
    fi
}

function upgrade_jfs(){
    local users=$@
    local JFS_VERSION="11.1.1"
    local current_jfs_version=$(/usr/local/bin/juicefs --version|awk '{print $3}'|awk -F'+' '{print $1}')

    if [ "x${JFS_VERSION}" != "x${current_jfs_version}" ]; then
        echo "upgrade JuiceFS ..."
        local juicefs_bin="/usr/local/bin/juicefs"
        ensure_success $sh_c "curl ${CURL_TRY} -kLO https://github.com/beclab/juicefs-ext/releases/download/v${JFS_VERSION}/juicefs-v${JFS_VERSION}-linux-amd64.tar.gz"
        ensure_success $sh_c "tar -zxf juicefs-v${JFS_VERSION}-linux-amd64.tar.gz"
        ensure_success $sh_c "chmod +x juicefs"

        ensure_success $sh_c "systemctl stop juicefs"
        ensure_success $sh_c "mv juicefs ${juicefs_bin}"
        ensure_success $sh_c "rm -f /tmp/JuiceFS-IPC.sock"
        ensure_success $sh_c "systemctl start juicefs"

        echo "restart pods ... "

        ensure_success $sh_c "${KUBECTL} rollout restart sts app-service -n os-system"

        local tf=$(mktemp)
        ensure_success $sh_c "${KUBECTL} get deployment -A -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.namespace} {.spec.template.spec.volumes}{\"\n\"}{end}' | grep '/olares/rootfs'" > $tf
        while read dep; do
            local depinfo=($dep)
            ensure_success $sh_c "${KUBECTL} rollout restart deployment ${depinfo[0]} -n ${depinfo[1]}"
        done < $tf

        for user in ${users[@]}; do
            ensure_success $sh_c "${KUBECTL} rollout restart sts bfl -n user-space-${user}"
        done
        
        sleep 10  # waiting for restarting to begin
    fi
}


function upgrade_terminus(){
    HELM=$(command -v helm)
    KUBECTL=$(command -v kubectl)

    # find sudo 
    get_shell_exec

    # fetch user list
    local users=()
    local admin_user=""
    local tf=$(mktemp)
    ensure_success $sh_c "${KUBECTL} get user -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.annotations.bytetrade\.io\/owner-role}{\"\n\"}{end}'" > $tf
    while read userdata; do
        local userinfo=($userdata)
        local valid=$(validate_user "${userinfo[0]}")
        if [ "x-${valid}" == "x-yes" ]; then
            if [ "x-${userinfo[1]}" == "x-platform-admin" ]; then 
                admin_user="${userinfo[0]}"
            fi

            i=${#users[@]}
            users[$i]=${userinfo[0]}
        fi
    done < $tf

    if [ "x${admin_user}" == "x" ]; then
        echo "Admin user not found. Upgrading failed." >&2
        exit -1
    fi

    # upgrade_jfs ${users[@]}
    local selfhosted=$($sh_c "${KUBECTL} get terminus terminus -o jsonpath='{.spec.settings.selfhosted}'")
    local domainname=$($sh_c "${KUBECTL} get terminus terminus -o jsonpath='{.spec.settings.domainName}'")
    sed -i "s/#__DOMAIN_NAME__/${domainname}/" ${BASE_DIR}/wizard/config/settings/templates/terminus_cr.yaml
    sed -i "s/#__SELFHOSTED__/${selfhosted}/" ${BASE_DIR}/wizard/config/settings/templates/terminus_cr.yaml

    echo "Upgrading olares system components ... "
    gen_settings_values ${admin_user}
    ensure_success $sh_c "${HELM} upgrade -i settings ${BASE_DIR}/wizard/config/settings -n default --reuse-values"

    # patch
    ensure_success $sh_c "${KUBECTL} apply -f ${BASE_DIR}/deploy/patch-globalrole-workspace-manager.yaml"
    ensure_success $sh_c "$KUBECTL apply -f ${BASE_DIR}/deploy/patch-notification-manager.yaml"

    # clear apps values.yaml
    cat /dev/null > ${BASE_DIR}/wizard/config/apps/values.yaml
    cat /dev/null > ${BASE_DIR}/wizard/config/launcher/values.yaml
    local appservice_pod=$(get_appservice_pod)
    local copy_charts=("launcher" "apps")
    for cc in ${copy_charts[@]}; do
        ensure_success $sh_c "${KUBECTL} cp ${BASE_DIR}/wizard/config/${cc} os-system/${appservice_pod}:/userapps"
    done

    local ks_redis_pwd=$($sh_c "${KUBECTL} get secret -n kubesphere-system redis-secret -o jsonpath='{.data.auth}' |base64 -d")
    for user in ${users[@]}; do
        echo "Upgrading user ${user} ... "
        gen_bfl_values ${user}

        # gen bfl app key and secret
        bfl_ks=($(get_app_key_secret ${user} "bfl"))

        # install launcher , and init pv
        ensure_success $sh_c "${HELM} upgrade -i launcher-${user} ${BASE_DIR}/wizard/config/launcher -n user-space-${user} --set bfl.appKey=${bfl_ks[0]} --set bfl.appSecret=${bfl_ks[1]} -f ${BASE_DIR}/wizard/config/launcher/values.yaml --reuse-values"

        gen_app_values ${user}
        close_apps ${user}

        for appdir in "${BASE_DIR}/wizard/config/apps"/*/; do
          if [ -d "$appdir" ]; then
            releasename=$(basename "$appdir")

            # ignore wizard
            # FIXME: unintitialized user's wizard should be upgrade
            if [ x"${releasename}" == x"wizard" ]; then 
                continue
            fi

            if [ "$user" != "$admin_user" ];then
                releasename=${releasename}-${user}
            fi
            ensure_success $sh_c "${HELM} upgrade -i ${releasename} ${appdir} -n user-space-${user} --reuse-values --set kubesphere.redis_password=${ks_redis_pwd} -f ${BASE_DIR}/wizard/config/apps/values.yaml"
          fi
        done

    done

    # upgrade app service in the last. keep app service online longer
    local terminus_is_cloud_version=$($sh_c "${KUBECTL} get cm -n os-system backup-config -o jsonpath='{.data.terminus-is-cloud-version}'")
    local backup_cluster_bucket=$($sh_c "${KUBECTL} get cm -n os-system backup-config -o jsonpath='{.data.backup-cluster-bucket}'")
    local backup_key_prefix=$($sh_c "${KUBECTL} get cm -n os-system backup-config -o jsonpath='{.data.backup-key-prefix}'")
    local backup_secret=$($sh_c "${KUBECTL} get cm -n os-system backup-config -o jsonpath='{.data.backup-secret}'")
    local backup_server_data=$($sh_c "${KUBECTL} get cm -n os-system backup-config -o jsonpath='{.data.backup-server-data}'")

    ensure_success $sh_c "${HELM} upgrade -i system ${BASE_DIR}/wizard/config/system -n os-system --reuse-values \
        --set kubesphere.redis_password=${ks_redis_pwd} --set backup.bucket=\"${backup_cluster_bucket}\" \
        --set backup.key_prefix=\"${backup_key_prefix}\" --set backup.is_cloud_version=\"${terminus_is_cloud_version}\" \
        --set backup.sync_secret=\"${backup_secret}\""

    echo 'Waiting for App-Service ...'
    sleep 2 # wait for controller reconiling
    check_appservice
    echo

    echo 'Waiting for Vault ...'
    check_vault ${admin_user}
    echo

    echo 'Starting BFL ...'
    check_bfl ${admin_user}
    echo

    echo 'Starting files ...'
    check_fileserver
    check_filesfe ${admin_user}
    echo

    echo 'Starting Desktop ...'
    check_desktop ${admin_user}
    echo

}


echo "Start to upgrade olares ... "

upgrade_terminus

echo -e "\e[91m Success to upgrade olares.\e[0m Open your new desktop in the browser and have fun !"
