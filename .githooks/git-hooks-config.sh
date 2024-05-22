#!/bin/sh
# This script will config git hook path into specific folder in your project. This script will invoked by maven build.
# @author : Mak Sophea
# @version : 1.0#
#
echo "config git hooksPath to .githooks folder for commit-msg and pre-push"
git config core.hooksPath .githooks