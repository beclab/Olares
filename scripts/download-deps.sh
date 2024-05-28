arch="amd64"
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
        else
            curl ${CURL_TRY} -L -o ./${part}/${s2} ${s1}
        fi
    else
        s4=$(echo "$line" | cut -d',' -f4)
        pkgpath="./${part}/${s2}/${arch}"
        mkdir -p ${pkgpath}
        filename=${file}
        if [ ! -z ${s3} ]; then
            filename=${s3}
        fi
        curl ${CURL_TRY} -L -o ${pkgpath}/${filename} ${s1}

        if [ "$s4" == "helm" ]; then
            pushd ${pkgpath}
            tar -zxvf ./${filename} && cp ./linux-amd64/helm ./ && rm -rf ./linux-amd64 && rm -rf ./${filename}
            popd
        fi
    fi
done