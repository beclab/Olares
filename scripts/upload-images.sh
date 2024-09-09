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
        set -e
        docker pull $image
        docker save $image -o $name.tar
        gzip $name.tar

        md5sum $name.tar.gz > $checksum

        aws s3 cp $name.tar.gz s3://terminus-os-install/$path$name.tar.gz --acl=public-read
        aws s3 cp $checksum s3://terminus-os-install/$path$checksum --acl=public-read
        echo "upload $name completed"
        set +e
    fi

    # re-upload checksum.txt
    curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/$path$checksum > /dev/null
    if [ $? -ne 0 ]; then
        set -e
        docker pull $image
        docker save $image -o $name.tar
        gzip $name.tar

        md5sum $name.tar.gz > $checksum

        aws s3 cp $name.tar.gz s3://terminus-os-install/$path$name.tar.gz --acl=public-read
        aws s3 cp $checksum s3://terminus-os-install/$path$checksum --acl=public-read
        echo "upload $name completed"
        set +e
    fi

done
