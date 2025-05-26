#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))
rm -rf ${BASE_DIR}/../.dist
DIST_PATH="${BASE_DIR}/../.dist/install-wizard" 
VERSION=$1

DIST_PATH=${DIST_PATH} bash ${BASE_DIR}/package.sh

bash ${BASE_DIR}/image-manifest.sh
bash ${BASE_DIR}/deps-manifest.sh

mv ${BASE_DIR}/../.dependencies/* ${BASE_DIR}/../.manifest/.
rm -rf ${BASE_DIR}/../.dependencies

set -e
pushd ${BASE_DIR}/../.manifest
bash ${BASE_DIR}/build-manifest.sh ${BASE_DIR}/../.manifest/installation.manifest
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


$TAR --exclude=wizard/tools --exclude=.git -zcvf ${BASE_DIR}/../install-wizard-${VERSION}.tar.gz .

popd
