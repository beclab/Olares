set -o pipefail

PLATFORM=${2:-linux/amd64}
path=""
if [ x"$PLATFORM" == x"linux/arm64" ]; then
    path="arm64/"
fi

cat $1|while read image; do
    echo "if exists $image ... "
    name=$(echo -n "$image"|md5sum|awk '{print $1}')
    checksum="$name.checksum.txt"

    curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/$path$name.tar.gz > /dev/null
    if [ $? -ne 0 ]; then
        code=$(curl -o /dev/null -fsSLI -w "%{http_code}" https://dc3p1870nn3cj.cloudfront.net/$path$name.tar.gz)
        if [ $code -eq 403 ]; then
            set -ex
            docker pull $image
            docker save $image -o $name.tar
            gzip $name.tar

            md5sum $name.tar.gz > $checksum
            backup_file=$(awk '{print $1}' $checksum)
            if [ x"$backup_file"  == x""  ]; then
                echo  "invalid checksum"
                exit 1
            fi

            echo "start to upload [$name.tar.gz]"
            aws s3 cp $name.tar.gz s3://terminus-os-install/$path$name.tar.gz --acl=public-read
            aws s3 cp $name.tar.gz s3://terminus-os-install/backup/$path$backup_file --acl=public-read
            aws s3 cp $checksum s3://terminus-os-install/$path$checksum --acl=public-read
            echo "upload $name completed"
            
            set +ex
        else
            if [ $code -ne 200  ]; then
                echo  "failed to check image"
                exit -1
            fi
        fi
    fi

    

    # re-upload checksum.txt
    curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/$path$checksum > /dev/null
    if [ $? -ne 0 ]; then
        code=$(curl -o /dev/null -fsSLI -w "%{http_code}" https://dc3p1870nn3cj.cloudfront.net/$path$checksum)
        if [ $code -eq 403 ]; then
            set -ex
            docker pull $image
            docker save $image -o $name.tar
            gzip $name.tar

            md5sum $name.tar.gz > $checksum
            backup_file=$(awk '{print $1}' $checksum)
            if [ x"$backup_file"  == x""  ]; then
                echo  "invalid checksum"
                exit 1
            fi

            aws s3 cp $name.tar.gz s3://terminus-os-install/$path$name.tar.gz --acl=public-read
            aws s3 cp $name.tar.gz s3://terminus-os-install/backup/$path$backup_file --acl=public-read
            aws s3 cp $checksum s3://terminus-os-install/$path$checksum --acl=public-read
            echo "upload $name completed"
            set +ex
        else
            if [ $code -ne 200  ]; then
                echo  "failed to check image"
                exit -1
            fi
        fi
    fi

    # upload to tencent cloud cos

    # curl -fsSLI https://cdn.joinolares.cn/$path$name.tar.gz > /dev/null
    # if [ $? -ne 0 ]; then
    #     set -e
    #     docker pull $image
    #     docker save $image -o $name.tar
    #     gzip $name.tar

    #     md5sum $name.tar.gz > $checksum

    #     coscmd upload ./$name.tar.gz /$path$name.tar.gz
    #     coscmd upload ./$checksum /$path$checksum
    #     echo "upload $name to cos completed"

    #     set +e
    # fi

    

    # # re-upload checksum.txt
    # curl -fsSLI https://cdn.joinolares.cn/$path$checksum > /dev/null
    # if [ $? -ne 0 ]; then
    #     set -e
    #     docker pull $image
    #     docker save $image -o $name.tar
    #     gzip $name.tar

    #     md5sum $name.tar.gz > $checksum

    #     coscmd upload ./$name.tar.gz /$path$name.tar.gz
    #     coscmd upload ./$checksum /$path$checksum
    #     echo "upload $name to cos completed"

    #     set +e
    # fi
    

done
