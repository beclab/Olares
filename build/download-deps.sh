PLATFORM=${1:-linux/amd64}
line=$2

set -ex
set -o pipefail

os=$(echo "$PLATFORM"|awk -F"/" '{print $1}')
arch=$(echo "$PLATFORM"|awk -F"/" '{print $2}')

fields=$(echo "$line"|awk -F"," '{print NF}')
if [[ $fields -lt 5 ]]; then
    echo "format err, $lines"
    exit -1
fi

filename=$(echo "$line"|awk -F"," '{print $1}')
name=$(echo -n "$filename"|md5sum|awk '{print $1}')

if [ x"$arch" == x"arm64" ]; then
    url=$(echo "$line"|awk -F"," '{print $4}')
else
    url=$(echo "$line"|awk -F"," '{print $3}')
fi

temp_file=$(mktemp)

CURL_TRY="--connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 "

curl -fsSLI ${url} > /dev/null
if [ $? -ne 0 ]; then
    exit -1
fi

curl ${CURL_TRY} -L -o ${temp_file} ${url}
mv ${temp_file} $name


