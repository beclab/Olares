#!/usr/bin/env bash

contains_element() {
  local e match="$1"
  shift
  for e; do [[ "$e" == "$match" ]] && return 0; done
  return 1
}

BASE_DIR=$(dirname $(realpath -s $0))

PACKAGE_MODULE=("apps" "framework" "daemon" "infrastructure" "platform" "vendor")
IMAGE_MANIFEST="$BASE_DIR/../.manifest/images.mf"

IGNORE_IMAGES=(
    "ghcr.io/open-telemetry/opentelemetry-go-instrumentation/autoinstrumentation-go:v0.23.0"
    "bytetrade/autoinstrumentation-apache-httpd:1.0.4-fix3"
    "ghcr.io/open-telemetry/opentelemetry-operator/autoinstrumentation-nodejs:0.53.0"
    "jaegertracing/jaeger:2.3.0"
)

rm -rf $BASE_DIR/../.manifest
mkdir -p $BASE_DIR/../.manifest

TMP_MANIFEST=$(mktemp)
for mod in "${PACKAGE_MODULE[@]}";do
    echo "find images in ${mod} ..."
    ls -A ${mod} | while read app; do
        chart_path="${mod}/${app}"

        if [ -d $chart_path ]; then
            find $chart_path -type f -path '*/.olares/*.yaml' | while read p; do
                bash ${BASE_DIR}/yaml2prop.sh -f $p | while read l;do 
                    if [[ "$l" == *".image = "* || "$l" == "output.containers."*".name"* ]]; then 
                        echo "$l"
                        if [[ $(echo "$l" | awk '{print $3}') == "value" ]]; then
                            echo "ignoring template value"
                            continue
                        fi
                        $img=$(echo "$l" | awk '{print $3}')
                        if contains_element "$img" "${IGNORE_IMAGES[@]}"; then
                            echo "ignoring image in ignore list: $img"
                            continue
                        fi  
                        echo "$l" >> ${TMP_MANIFEST}
                    fi;
                done
            done
        fi
    done
done
awk '{print $3}' ${TMP_MANIFEST} | sort | uniq | grep -v nitro | grep -v orion | grep -v '^nonexisting$' >> ${IMAGE_MANIFEST}

# patch
# fix backup server version
backup_version=$(egrep '{{ \$backupVersion := "(.*)" }}' $BASE_DIR/../framework/backup-server/.olares/config/cluster/deploy/backup_server.yaml | sed 's/{{ \$backupVersion := "\(.*\)" }}/\1/')
if [[ "$OSTYPE" == "darwin"* ]]; then
    bash -c "sed -i '' -e 's/backup-server:vvalue/backup-server:v$backup_version/' ${IMAGE_MANIFEST}"
else
    bash -c "sed -i 's/backup-server:vvalue/backup-server:v$backup_version/' ${IMAGE_MANIFEST}"
fi