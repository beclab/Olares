#!/usr/bin/env bash



function run_cmd(){
    echo "$*"
    bash -c "$@"
}

PACKAGE_MODULE=("apps" "framework" "daemon" "infrastructure" "platform" "vendor")

BUILD_TEMPLATE="build/base-package"
DIST=${DIST_PATH:-".dist"}

echo ${DIST}

set -o pipefail
set -e

if [ ! -d ${DIST} ]; then
    mkdir -p ${DIST}
    cp -rf ${BUILD_TEMPLATE}/* ${DIST}/.
    cp -rf ${BUILD_TEMPLATE}/.env ${DIST}/.
    cp -rf ${BUILD_TEMPLATE}/wizard/config/os-chart-template ${DIST}/wizard/config/os-framework
    cp -rf ${BUILD_TEMPLATE}/wizard/config/os-chart-template ${DIST}/wizard/config/os-platform
    rm -rf ${DIST}/wizard/config/os-chart-template
fi

APP_DIST=${DIST}/wizard/config/apps
SETTINGS_DIST=${DIST}/wizard/config/settings/templates
CRD_DIST=${SETTINGS_DIST}/crds
mkdir -p ${APP_DIST}
mkdir -p ${CRD_DIST}

for mod in "${PACKAGE_MODULE[@]}";do
    echo "packaging ${mod} ..."
    SYSTEM_DIST=${DIST}/wizard/config/os-framework/templates
    if [ ${mod} == "platform" ]; then
        SYSTEM_DIST=${DIST}/wizard/config/os-platform/templates
    fi
    DEPLOY_DIST=${SYSTEM_DIST}/deploy
    mkdir -p ${DEPLOY_DIST}

    find ${mod} -type d -name .olares | while read app; do

        # package user app charts to install wizard
        chart_path="${app}/config/user/helm-charts"
        if [ -d ${chart_path} ]; then
            ls ${chart_path} | while read chart; do
                run_cmd "cp -rf ${chart_path}/${chart} ${APP_DIST}"
            done
        fi

        # package cluster crd to install wizard's system chart
        crd_path="${app}/config/cluster/crds"
        if [ -d ${crd_path} ]; then
            ls ${crd_path} | while read crd; do
                run_cmd "cp -rf ${crd_path}/${crd} ${CRD_DIST}"
            done
        fi

        # package cluster deployments to install wizard's system chart
        deploy_path="${app}/config/cluster/deploy"
        if [ -d ${deploy_path} ]; then
            ls ${deploy_path} | while read deploy; do
                run_cmd "cp -rf ${deploy_path}/${deploy} ${DEPLOY_DIST}"
            done
        fi

    done
done

echo "packaging launcher ..."
run_cmd "cp -rf framework/bfl/.olares/config/launcher ${DIST}/wizard/config/"

echo "packaging gpu ..."
run_cmd "cp -rf framework/gpu/.olares/config/gpu ${DIST}/wizard/config/"

echo "packaging completed"