PLATFORM=${2:-linux/amd64}
path=""
if [ x"$PLATFORM" == x"linux/arm64" ]; then
    path="arm64/"
fi

cat $1|while read image; do
    name=$(echo -n "$image"|md5sum|awk '{print $1}')
    if [ ! -z "$FROM_S3" ]; then
        curl -fsSLI https://cdn.olares.com/$path$name.tar.gz > /dev/null
        if [ $? -ne 0 ]; then
            docker pull $image --platform $PLATFORM
            docker save $image -o $name.tar
            gzip $name.tar
        else
            curl -fsSL https://cdn.olares.com/$path$name.tar.gz -o $name.tar.gz
        fi
    else
        docker pull $image --platform $PLATFORM
        docker save $image -o $name.tar
        gzip $name.tar
    fi
    echo $name
done