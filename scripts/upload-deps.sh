#!/usr/bin/env bash

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
        if [ x"$line" == x"" ]; then
            continue
        fi
        
        bash ${BASE_DIR}/download-deps.sh $PLATFORM $line
        if [ $? -ne 0 ]; then
            exit -1
        fi

        filename=$(echo "$line"|awk -F"," '{print $1}')
        echo "if exists $filename ... "
        name=$(echo -n "$filename"|md5sum|awk '{print $1}')
        checksum="$name.checksum.txt"
        md5sum $name > $checksum
        backup_file=$(awk '{print $1}' $checksum)
        if [ x"$backup_file"  == x""  ]; then
            echo  "invalid checksum"
            exit 1
        fi

       curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/$path$name > /dev/null
       if [ $? -ne 0 ]; then
            code=$(curl -o /dev/null -fsSLI -w "%{http_code}" https://dc3p1870nn3cj.cloudfront.net/$path$name.tar.gz)
            if [ $code -eq 403 ]; then
                set -ex
                aws s3 cp $name s3://terminus-os-install/$path$name --acl=public-read
                aws s3 cp $name s3://terminus-os-install/backup/$path$backup_file --acl=public-read
                aws s3 cp $checksum s3://terminus-os-install/$path$checksum --acl=public-read
                echo "upload $name to s3 completed"
                set +ex
            else
                if [ $code -ne 200  ]; then
                    echo  "failed to check image"
                    exit -1
                fi
            fi
       fi        

       # upload to tencent cloud cos
    #    curl -fsSLI https://cdn.joinolares.cn/$path$name > /dev/null
    #    if [ $? -ne 0 ]; then
    #         set -ex
    #         coscmd upload ./$name /$path$name
    #         coscmd upload ./$checksum /$path$checksum
    #         echo "upload $name to cos completed"
    #         set +ex
    #    fi        
    done < $deps
done

popd



