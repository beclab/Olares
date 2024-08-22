PLATFORM=${1:-linux/amd64}
urlpath=""
arch="amd64"
if [ x"$PLATFORM" == x"linux/arm64" ]; then
    arch="arm64"
fi
if [ x"$PLATFORM" == x"linux/arm64" ]; then
    urlpath="arm64/"
fi
mkdir temp

part=""
CURL_TRY="--connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 "

cat ./dependencies.mf | while IFS= read -r line; do
    [[ -z "$line" ]] && continue

    case "$line" in
        *\[components\]*)
            part="components"
            mkdir -p "${part}"
            continue
            ;;
        *\[pkg\]*)
            part="pkg"
            mkdir -p "${part}"
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
            curl ${CURL_TRY} -L -o ./${part}/${file} ${s1}
            newname=$(echo -n "$file"|md5sum|awk '{print $1}')
            cp ./${part}/${file} ./${part}/../temp/${newname}
        else
            curl ${CURL_TRY} -L -o ./${part}/${s2} ${s1}

            if [ ${s2} == "redis-5.0.14.tar.gz" ]; then
                pushd ${part}
                tar xvf ${s2} && cd redis-5.0.14 && make && make install && cd ..
                rm -rf redis-5.0.14 && mkdir redis-5.0.14 && cp /usr/local/bin/redis* ./redis-5.0.14/
                tar cvf ./redis-5.0.14.tar.gz ./redis-5.0.14/ && rm -rf ./redis-5.0.14/
                newname=$(echo -n "redis-5.0.14.tar.gz"|md5sum|awk '{print $1}')
                cp ./redis-5.0.14.tar.gz ../temp/${newname}
                popd
            else
                newname=$(echo -n "$s2"|md5sum|awk '{print $1}')
                cp ./${part}/${s2} ./${part}/../temp/${newname}
            fi
        fi
    else
        s4=$(echo "$line" | cut -d',' -f4)
        s5=$(echo "$line" | cut -d',' -f5)
        pkgpath="./${part}/${s2}/${arch}"
        mkdir -p ${pkgpath}
        filename=${file}
        if [ ! -z ${s3} ]; then
            filename=${s3}
        fi
        curl ${CURL_TRY} -L -o ${pkgpath}/${filename} ${s1}
        if [ "$s4" == "helm" ]; then
            pushd ${pkgpath}
            tar -zxvf ./${filename} && cp ./linux-${arch}/helm ./ && rm -rf ./linux-${arch} && rm -rf ./${filename}
            if [ ! -z ${s5} ]; then
                newname=$(echo -n "${s5}"|md5sum|awk '{print $1}')
                cp ./helm ../../../../temp/${newname}
            else
                newname=$(echo -n "helm"|md5sum|awk '{print $1}')
                cp ./helm ../../../../temp/${newname}
            fi
            popd
        else
            if [ ! -z ${s5} ]; then
                newname=$(echo -n "${s5}"|md5sum|awk '{print $1}')
                cp ${pkgpath}/${filename} ./${part}/../temp/${newname}
            else 
                newname=$(echo -n "${filename}"|md5sum|awk '{print $1}')
                cp ${pkgpath}/${filename} ./${part}/../temp/${newname}
            fi
        fi
    fi
done

echo "done..."
p=$(pwd)
echo "current dir: ${p}"
echo "file tree:"
tree ./
cd temp
ls | while read -r file; do
    echo "if exists $file ... "
    curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/$urlpath$file > /dev/null
    if [ $? -ne 0 ]; then
        aws s3 cp $file s3://terminus-os-install/$urlpath$file --acl=public-read
        echo "upload $file completed"
    fi
    
done
cd ..

rm -rf ./temp