#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))
DEPENDENCIES_MANIFEST=".manifest/dependencies.mf"

if [ ! -d .manifest ]; then
    mkdir -p .manifest
fi

cp $BASE_DIR/../build/manifest/dependencies ${DEPENDENCIES_MANIFEST}
