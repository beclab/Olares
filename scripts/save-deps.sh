BASE_DIR=$(dirname $(realpath -s $0))

fileprefix="deps"
name=$(md5sum dependencies.mf |awk '{print $1}')
echo "READY to find this deps ${fileprefix}-${name}.tar.gz"

curl -fsSLI https://dc3p1870nn3cj.cloudfront.net/${fileprefix}-${name}.tar.gz > /dev/null
if [ $? -eq 0 ]; then
    curl -fsSL https://dc3p1870nn3cj.cloudfront.net/${fileprefix}-${name}.tar.gz -o ${fileprefix}-${name}.tar.gz
    tar -zxvf ${fileprefix}-${name}.tar.gz
    rm -rf ${fileprefix}-${name}.tar.gz
else
    bash ${BASE_DIR}/download-deps.sh
fi