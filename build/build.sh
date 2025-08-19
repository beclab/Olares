#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))
rm -rf ${BASE_DIR}/../.dist
DIST_PATH="${BASE_DIR}/../.dist/install-wizard" 
export VERSION=$1
export RELEASE_ID=$2


# vendor replace
if [[ "${REPO_PATH}" != "" && "$REPO_PATH" != "/" ]]; then
    path="vendor${REPO_PATH}"
    echo "replace vendor path: ${path}"
    find ${BASE_DIR}/../$path -type f | while read l; 
    do 
        file=$(awk -F "$path" '{print $1$2}' <<< "$l")  
        if [[ "$file" != ".gitkeep" ]]; then
            echo "replace [$file] with [$l]"
            cp -f "$l" "$file"
        fi
    done
fi


DIST_PATH=${DIST_PATH} bash ${BASE_DIR}/package.sh

bash ${BASE_DIR}/image-manifest.sh
bash ${BASE_DIR}/deps-manifest.sh

mv ${BASE_DIR}/../.dependencies/* ${BASE_DIR}/../.manifest/.
rm -rf ${BASE_DIR}/../.dependencies

set -e
pushd ${BASE_DIR}/../.manifest
bash ${BASE_DIR}/build-manifest.sh ${BASE_DIR}/../.manifest/installation.manifest
python3 ${BASE_DIR}/build-manifest.py ${BASE_DIR}/../.manifest/installation.manifest
popd

pushd $DIST_PATH

rm -rf images
mv ${BASE_DIR}/../.manifest/installation.manifest .
mv ${BASE_DIR}/../.manifest images

if [[ "$OSTYPE" == "darwin"* ]]; then
    TAR=gtar
    SED="sed -i '' -e"
else
    TAR=tar
    SED="sed -i"
fi

if [ ! -z $VERSION ]; then
    sh -c "$SED 's/#__VERSION__/${VERSION}/' wizard/config/settings/templates/terminus_cr.yaml"
    sh -c "$SED 's/#__VERSION__/${VERSION}/' install.sh"
    sh -c "$SED 's/#__VERSION__/${VERSION}/' install.ps1"
    sh -c "$SED 's/#__VERSION__/${VERSION}/' joincluster.sh"
    VERSION="v${VERSION}"
else
    VERSION="debug"
fi

if [ ! -z $RELEASE_ID ]; then
    sh -c "$SED 's/#__RELEASE_ID__/${RELEASE_ID}/' install.sh"
    sh -c "$SED 's/#__RELEASE_ID__/${RELEASE_ID}/' install.ps1"
    sh -c "$SED 's/#__RELEASE_ID__/${RELEASE_ID}/' joincluster.sh"
fi

$TAR --exclude=wizard/tools --exclude=.git -zcvf ${BASE_DIR}/../install-wizard-${VERSION}.tar.gz .

popd
