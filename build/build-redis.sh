#!/usr/bin/env bash

PLATFORM=${1:-linux/amd64}
SUFFIX=$2

if [[ "$SUFFIX" != "" ]]; then
    SUFFIX="-${SUFFIX}"
fi


os=$(echo "$PLATFORM"|awk -F"/" '{print $1}')
arch=$(echo "$PLATFORM"|awk -F"/" '{print $2}')

set -o pipefail
set -xe

curl -Lo redis-5.0.14.tar.gz https://download.redis.io/releases/redis-5.0.14.tar.gz 

tar zxvf redis-5.0.14.tar.gz && \
cd redis-5.0.14 && \
make && \
make install && \
cd ..

rm -rf redis-5.0.14 && \
mkdir redis-5.0.14 && \
cp /usr/local/bin/redis* ./redis-5.0.14/

tar czvf ./redis-5.0.14.tar.gz ./redis-5.0.14/ && \
aws s3 cp redis-5.0.14.tar.gz s3://terminus-os-install/redis-5.0.14_${os}_${arch}${SUFFIX}.tar.gz --acl=public-read
