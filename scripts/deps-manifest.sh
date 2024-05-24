#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))
DEPENDENCIES_MANIFEST=".manifest/dependencies.mf"

rm -rf .manifest
mkdir -p .manifest

cp $BASE_DIR/../build/manifest/dependencies ${DEPENDENCIES_MANIFEST}
