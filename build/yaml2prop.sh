#!/usr/bin/env bash

FILE=

errorExit () {
    echo; echo "ERROR: $1"; echo
    exit 1
}

usage () {
    cat << END_USAGE

${SCRIPT_NAME} - Convert a YAML formatted file to properties format

Usage: ${SCRIPT_NAME} <options>

-f | --file <name>                : [MANDATORY] Yaml file to process
-h | --help                       : Show this usage

Examples:
========
$ ${SCRIPT_NAME} --file examples/simple.yaml

END_USAGE

    exit 1
}

checkYq () {
    [[ $(yq -V) =~ ( 4.|v4.) ]] || errorExit "Must have yq v4 installed (https://github.com/mikefarah/yq)"
}

processOptions () {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -f | --file)
                FILE="$2"
                shift 2
            ;;
            -h | --help)
                usage
                exit 0
            ;;
            *)
                usage
            ;;
        esac
    done
}

checkFileExists () {
    [ -n "${FILE}" ] || usage
    [ -f "${FILE}" ] || errorExit "File ${FILE} does not exist"
}

removeEmptyArrayAndMap () {
    TMP_FILE=$(mktemp)
    # cp "$1" "$TMP_FILE"
    sed -e 's,\[[[:space:]]*\],,g' -e 's,{[[:space:]]*},,g' -e 's,{{.*}},value,g' "$1" > "$TMP_FILE"
}

processYaml () {
    removeEmptyArrayAndMap $1
    # yq eval '.. | select((tag == "!!map" or tag == "!!seq") | not) | (path | join(".")) + "=" + .' "$TMP_FILE" || errorExit "yq failed"
    cat "$TMP_FILE" | yq -o=props  || errorExit "yq failed"
    rm -f "$TMP_FILE"
}

splitYaml () {
    local newfile=0
    local fileContent=""
    local oldifs=$IFS
    IFS=''
    cat "$FILE" | while read line; do
        if [ "$line" == "---" ]; then
            newfile=0
        fi 

        if [[ "$line" == "apiVersion"* ]]; then
            fileContent=$(mktemp)
            newfile=1
            echo $fileContent
        fi

        if [[ newfile -eq 1 && "x${fileContent}" != "x" ]]; then
            local line_no_spaces="$(echo -e "${line}" | sed -e 's/^[[:space:]]*//')"
            if [[ "$line_no_spaces" != "{{"* ]]; then
                echo "$line" >> $fileContent
            fi
        fi
    done
    IFS=$oldifs
}

main () {
    checkYq
    processOptions "$@"
    checkFileExists
    local files=$(splitYaml)
    echo "$files" | while read f; do
         processYaml "$f"
    done
}

######### Main #########

main "$@"