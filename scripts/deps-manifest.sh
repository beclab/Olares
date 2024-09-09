#!/usr/bin/env bash

PLATFORM=${1:-linux/amd64}

arch="amd64"
if [ x"$PLATFORM" == x"linux/arm64" ]; then
    arch="arm64"
fi

BASE_DIR=$(dirname $(realpath -s $0))

rm -rf $BASE_DIR/../.dependencies
mkdir -p $BASE_DIR/../.dependencies

cp $BASE_DIR/../build/manifest/components $BASE_DIR/../.dependencies/.
cp $BASE_DIR/../build/manifest/pkgs $BASE_DIR/../.dependencies/.
