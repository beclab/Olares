#!/usr/bin/env bash

set -ex
set -o pipefail

BASE_DIR=$(dirname $(realpath -s $0))
echo "Push Deps to S3 base_dir: ${BASE_DIR}"

if [ ! -d "$BASE_DIR/../.dependencies" ]; then
    exit 1
fi

PLATFORM=${1:-linux/amd64}

path=""
if [ x"$PLATFORM" == x"linux/arm64" ]; then
    path="arm64/"
fi

pushd $BASE_DIR/../.dependencies

for deps in "components" "pkgs"; do
    while read line; do
        bash ${BASE_DIR}/download-deps.sh $PLATFORM $line
        filename=$(echo "$line"|awk -F"," '{print $1}')
        echo "if exists $filename ... "
        name=$(echo -n "$filename"|md5sum|awk '{print $1}')
        checksum="$name.checksum.txt"
        md5sum $name > $checksum

#        curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/$path$name > /dev/null
#        if [ $? -ne 0 ]; then
            aws s3 cp $name s3://terminus-os-install/$path$name --acl=public-read
            aws s3 cp $checksum s3://terminus-os-install/$path$checksum --acl=public-read
            echo "upload $name completed"
#        fi        
    done < $deps
done

popd



