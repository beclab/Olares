BASE_DIR=$(dirname $(realpath -s $0))
PLATFORM=${1:-linux/amd64}

path=""
if [ x"$PLATFORM" == x"linux/arm64" ]; then
    path="arm64/"
fi

