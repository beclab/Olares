#!/usr/bin/env bash

BASE_DIR=$(dirname $(realpath -s $0))

rm -rf $BASE_DIR/../.dependencies
mkdir -p $BASE_DIR/../.dependencies
manifest=$BASE_DIR/../.dependencies/components

# bash 3.2 (in macos) don't support associative array, so we use this ugly hack
# to get the key for a given value
function get_key(){
    if [ "$1" == "id" ]; then
        echo 0
    elif [ "$1" == "name" ]; then
        echo 1
    elif [ "$1" == "amd64" ]; then
        echo 2
    elif [ "$1" == "arm64" ]; then
        echo 3
    fi
}

# vendor replace
if [[ "${REPO_PATH}" != "" && "$REPO_PATH" != "/" ]]; then
    path="vendor${REPO_PATH}"
    echo "replace vendor path: ${path}"
    find ${BASE_DIR}/../$path -type f | while read l; 
    do 
        file=$(awk -F "$path" '{print $1$2}' <<< "$l")  
        if [[ "$file" != ".gitkeep" ]]; then
            echo "replace [$file] with [$l]"
            cp -f "$l" "$file"
        fi
    done
fi


find $BASE_DIR/../ -type f -name Olares.yaml | while read f; do
  if [[ "$f" == *"/vendor/"* ]]; then
      echo "skip vendor file $f"
      continue
  fi
  echo "Processing $f"
  declare -a bins
  IFS=
  while read l;do 
    if [[ "$l" == "output.binaries."* ]]; then
        kv=${l#output.binaries.}
        key=$(awk -F' = ' '{print $1}' <<< "$kv")
        value=$(awk -F' = ' '{print $2}' <<< "$kv")

        idx=$(awk -F'.' '{print $1}' <<< "$key")
        field=$(awk -F'.' '{print $2}' <<< "$key")

        old_field=${bins[$idx]}
        if [[ "$old_field" == "" ]]; then
          old_field="$field=$value"
        else
          old_field="$old_field|$field=$value"
        fi

        bins[$idx]=$old_field
    fi
  done <<< $(bash ${BASE_DIR}/yaml2prop.sh -f $f)

  
  for bin in "${bins[@]}"; do
    bobj=$(tr '|' '\n' <<< $bin)
    declare -a com
    while read bl; do
        k=$(awk -F'=' '{print $1}' <<< "$bl")
        v=$(awk -F'=' '{print $2}' <<< "$bl")
        k=$(get_key $k)
        com[$k]=$v
    done <<< "$bobj"
        echo "key: ${com[@]}"


    name_path=${com[$(get_key "name")]}
    n=$(awk -F"," '{print NF}' <<< ${name_path})
    if [[ $n -gt 1 ]]; then
        name_path="$(awk -F"," '{print $1}' <<< ${name_path}),$(awk -F"," '{print $2}' <<< ${name_path})"
    else
        name_path="${name_path},pkg/components"
    fi

    amd64=$(get_key "amd64")
    arm64=$(get_key "arm64")
    id=$(get_key "id")
    echo "${name_path},${com[$amd64]},${com[$arm64]},${com[$id]}" >> ${manifest}

    unset com
  done

  unset bins
done

sed -i "s/#__VERSION__/${VERSION}/g" ${manifest}

if [[ ! -z "$RELEASE_ID" ]]; then
    RELEASE_ID_SUFFIX=".$RELEASE_ID"
else
    RELEASE_ID_SUFFIX=""
fi
sed -i "s/#__RELEASE_ID_SUFFIX__/${RELEASE_ID_SUFFIX}/g" ${manifest}

path="${REPO_PATH:-/}"
sed -i "s|#__REPO_PATH__|${path}|g" ${manifest}
