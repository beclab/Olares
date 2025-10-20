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
    manifest="$name.manifest.json"

    curl -fsSLI https://cdn.olares.com/$path$name.tar.gz > /dev/null
    if [ $? -ne 0 ]; then
        code=$(curl -o /dev/null -fsSLI -w "%{http_code}" https://cdn.olares.com/$path$name.tar.gz)
        if [ $code -eq 403 ]; then
            set -ex
            skopeo copy --insecure-policy docker://$image oci-archive:$name.tar
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
    curl -fsSLI https://cdn.olares.com/$path$checksum > /dev/null
    if [ $? -ne 0 ]; then
        code=$(curl -o /dev/null -fsSLI -w "%{http_code}" https://cdn.olares.com/$path$checksum)
        if [ $code -eq 403 ]; then
            set -ex
            skopeo copy --insecure-policy docker://$image oci-archive:$name.tar
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
                echo  "failed to check image checksum"
                exit -1
            fi
        fi
    fi

    # upload manifest.json
    curl -fsSLI https://cdn.olares.com/$path$manifest > /dev/null
    if [ $? -ne 0 ]; then   
        code=$(curl -o /dev/null -fsSLI -w "%{http_code}" https://cdn.olares.com/$path$manifest)
        if [ $code -eq 403 ]; then
            set -ex
            BASE_DIR=$(dirname $(realpath -s $0))
            python3 $BASE_DIR/get-manifest.py $image -o $manifest

            aws s3 cp $manifest s3://terminus-os-install/$path$manifest --acl=public-read
            echo "upload $name manifest completed"
            set +ex
        else
            if [ $code -ne 200  ]; then
                echo  "failed to check image manifest"
                exit -1
            fi
        fi
    fi
done
