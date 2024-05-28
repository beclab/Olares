#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))
echo "Push Deps to S3 base_dir: ${BASE_DIR}"

if [ ! -d ".dependencies" ]; then
    exit 1
fi

pushd ${BASE_DIR}/../.dependencies
fileprefix="deps"
name=$(md5sum dependencies.mf |awk '{print $1}')
echo "filename: ${fileprefix}-${name}.tar.gz"
curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/${fileprefix}-$name.tar.gz > /dev/null
if [ $? -eq 0 ]; then
    echo "dependencies file ${fileprefix}-${name}.tar.gz exists, STOP..."
    exit 1
fi

bash ${BASE_DIR}/download-deps.sh

name=$(md5sum dependencies.mf |awk '{print $1}')
echo "filename: ${fileprefix}-${name}.tar.gz"
curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/${fileprefix}-$name.tar.gz > /dev/null
if [ $? -ne 0 ]; then
    echo "dependencies file ${fileprefix}-${name}.tar.gz not found, prepare to upload to S3"
    tar -czf ./$name.tar.gz ./components ./pkg ./dependencies.mf && cp ./$name.tar.gz ../
    popd
    rm -rf ./.dependencies/
    aws s3 cp $name.tar.gz s3://terminus-os-install/${fileprefix}-$name.tar.gz --acl=public-read
    echo "upload $name completed"
else
    echo "dependencies file ${fileprefix}-${name}.tar.gz exists, EXIT..."
fi