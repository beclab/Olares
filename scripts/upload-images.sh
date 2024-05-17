

cat $1|while read image; do
    echo "if exists $image ... "
    name=$(echo -n "$image"|md5sum|awk '{print $1}')
    curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/$name.tar.gz > /dev/null
    if [ $? -ne 0 ]; then
        docker pull $image
        docker save $image -o $name.tar
        gzip $name.tar

        aws s3 cp $name.tar.gz s3://terminus-os-install/$name.tar.gz --acl=public-read
        echo "upload $name completed"
    fi
done
