#!/usr/bin/env bash



BASE_DIR=$(dirname $(realpath -s $0))

PACKAGE_MODULE=("frameworks" "libs" "apps" "third-party")
IMAGE_MANIFEST="$BASE_DIR/../.manifest/images.mf"

rm -rf $BASE_DIR/../.manifest
mkdir -p $BASE_DIR/../.manifest

# copy default base images
cp -f $BASE_DIR/../build/manifest/images ${IMAGE_MANIFEST}
cp -f $BASE_DIR/../build/manifest/images.node.mf $BASE_DIR/../.manifest/images.node.mf

TMP_MANIFEST=$(mktemp)
for mod in "${PACKAGE_MODULE[@]}";do
    echo "find images in ${mod} ..."
    ls ${mod} | while read app; do
        chart_path="${mod}/${app}/config"
        if [ -d $chart_path ]; then
            find $chart_path -type f -name *.yaml | while read p; do
                bash ${BASE_DIR}/yaml2prop.sh -f $p | while read l;do 
                    if [[ "$l" == *".image = "* ]]; then 
                        echo "$l"
                        if [[ $(echo "$l" | awk '{print $3}') == "value" ]]; then
                            echo "ignoring template value"
                            continue
                        fi
                        echo "$l" >> ${TMP_MANIFEST}
                    fi;
                done
            done
        fi
    done
done

awk '{print $3}' ${TMP_MANIFEST} | sort | uniq | grep -v nitro | grep -v orion >> ${IMAGE_MANIFEST}

# patch
# fix backup server version
backup_version=$(egrep '{{ \$backupVersion := "(.*)" }}' $BASE_DIR/../frameworks/backup-server/config/cluster/deploy/backup_server.yaml | sed 's/{{ \$backupVersion := "\(.*\)" }}/\1/')
if [[ "$OSTYPE" == "darwin"* ]]; then
    bash -c "sed -i '' -e 's/backup-server:vvalue/backup-server:v$backup_version/' ${IMAGE_MANIFEST}"
else
    bash -c "sed -i 's/backup-server:vvalue/backup-server:v$backup_version/' ${IMAGE_MANIFEST}"
fi