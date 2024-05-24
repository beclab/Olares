#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))
mfp=".manifest"
mfn="${mfp}/dependencies.mf"
arch="amd64"
fileprefix="deps"
part=""

if [ ! -d ".manifest" ]; then
    exit 1
fi

name=$(md5sum ${mfn} |awk '{print $1}')
echo "filename: ${fileprefix}-${name}.tar.gz"
curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/${fileprefix}-$name.tar.gz > /dev/null
if [ $? -eq 0 ]; then
    echo "dependencies file ${fileprefix}-${name}.tar.gz exists, STOP..."
    exit 1
fi

cat ${mfn} | while IFS= read -r line; do
    [[ -z "$line" ]] && continue

    case "$line" in
        *\[components\]*)
            part="components"
            mkdir -p "${mfp}/${part}"
            continue
            ;;
        *\[pkg\]*)
            part="pkg"
            mkdir -p "${mfp}/${part}"
            continue
            ;;
        *)
            ;;
    esac

    if [ -z "$part" ]; then
        exit 1
    fi

    s1=$(echo "$line" | cut -d',' -f1)
    s2=$(echo "$line" | cut -d',' -f2)
    s3=$(echo "$line" | cut -d',' -f3)

    file=$(echo "$s1" | rev | cut -d'/' -f1 | rev)

    if [ "$part" == "components" ]; then
        if [ -z "$s2" ]; then
            curl -L -o ./${mfp}/${part}/${file} ${s1}
        else
            curl -L -o ./${mfp}/${part}/${s2} ${s1}
        fi
    else
        s4=$(echo "$line" | cut -d',' -f4)
        pkgpath="./${mfp}/${part}/${s2}/${arch}"
        mkdir -p ${pkgpath}
        filename=${file}
        if [ ! -z ${s3} ]; then
            filename=${s3}
        fi
        curl -L -o ${pkgpath}/${filename} ${s1}

        if [ "$s4" == "helm" ]; then
            pushd ${pkgpath}
            tar -zxvf ./${filename} && cp ./linux-amd64/helm ./ && rm -rf ./linux-amd64 && rm -rf ./${filename}
            popd
        fi
    fi
done

name=$(md5sum ${mfn} |awk '{print $1}')
echo "filename: ${fileprefix}-${name}.tar.gz"
curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/${fileprefix}-$name.tar.gz > /dev/null
if [ $? -ne 0 ]; then
    echo "dependencies file ${fileprefix}-${name}.tar.gz not found, prepare to upload to S3"
    cd ./${mfp} && tar -czf ./$name.tar.gz ./components ./pkg && cp ./$name.tar.gz ../
    aws s3 cp $name.tar.gz s3://terminus-os-install/${fileprefix}-$name.tar.gz --acl=public-read
    echo "upload $name completed"
else
    echo "dependencies file ${fileprefix}-${name}.tar.gz exists, EXIT..."
fi