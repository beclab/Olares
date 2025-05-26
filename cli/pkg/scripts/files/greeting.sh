#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))
source $BASE_DIR/common.sh

ensure_success $sh_c "echo 'test script greetings'"
ensure_success $sh_c "echo lsb_dist=$lsb_dist"
ensure_success $sh_c "echo BASE_DIR=$BASE_DIR"