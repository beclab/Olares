#!/usr/bin/env bash


cdn_url="https://dc3p1870nn3cj.cloudfront.net"

download_checksum(){
    local name=$1
    local checksum=$(curl -SsfL $cdn_url/$name.checksum.txt|awk '{print $1}')

    if [ x"$checksum" == x"" ]; then
        echo "get checksum error, $name"
        exit -1
    fi

    echo $checksum
}

manifest_file=$1

for deps in "components" "pkgs"; do
    while read line; do
        fields=$(echo "$line"|awk -F"," '{print NF}')
        if [[ $fields -lt 4 ]]; then
            echo "format err, $lines"
            exit -1
        fi

        filename=$(echo "$line"|awk -F"," '{print $1}')
        echo "downloading file checksum, $filename"
        path=$(echo "$line"|awk -F"," '{print $2}')
        name=$(echo -n "$filename"|md5sum|awk '{print $1}')

        url_amd64=$cdn_url/$name
        url_arm64=$cdn_url/arm64/$name

        checksum_amd64=$(download_checksum $name)
        checksum_arm64=$(download_checksum arm64/$name)

        echo "$filename,$path,$deps,$url_amd64,$checksum_amd64,$url_arm64,$checksum_arm64" >> $manifest_file
    
    done < $deps

done

for deps in "images.mf"; do
    while read line; do
        filename=$(echo -n "$line"|md5sum|awk '{print $1}')

        echo "downloading file checksum, $line"
        url_amd64=$cdn_url/$name.tar.gz
        url_arm64=$cdn_url/arm64/$name.tar.gz

        checksum_amd64=$(download_checksum $name)
        checksum_arm64=$(download_checksum arm64/$name)

        echo "$filename,$path,$deps,$url_amd64,$checksum_amd64,$url_arm64,$checksum_arm64,$line" >> $manifest_file
    
    done < $deps
done