#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))
DEPENDENCIES_MANIFEST=".dependencies/dependencies.mf"

rm -rf .dependencies
mkdir -p .dependencies

cp $BASE_DIR/../build/manifest/dependencies ${DEPENDENCIES_MANIFEST}
