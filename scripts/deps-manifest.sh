#!/usr/bin/env bash

PLATFORM=${1:-linux/amd64}

arch="amd64"
if [ x"$PLATFORM" == x"linux/arm64" ]; then
    arch="arm64"
fi

BASE_DIR=$(dirname $(realpath -s $0))
DEPENDENCIES_MANIFEST=".dependencies/dependencies.mf"

rm -rf .dependencies
mkdir -p .dependencies

cp $BASE_DIR/../build/manifest/dependencies.${arch} ${DEPENDENCIES_MANIFEST}
