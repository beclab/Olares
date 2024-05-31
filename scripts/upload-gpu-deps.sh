#!/usr/bin/env bash

OS_VERSION=$(. /etc/os-release;echo $ID$VERSION_ID|sed 's/\.//g')
BASE_DIR=$(pwd)

install_apt_depends() {
    sudo apt install aptitude tree -y
}

install_gpu_keyring() {
    if [[ "$OS_VERSION" =~ "ubuntu" ]]; then
        case "$OS_VERSION" in
            ubuntu2404)
                sudo wget https://developer.download.nvidia.com/compute/cuda/repos/$OS_VERSION/x86_64/cuda-keyring_1.1-1_all.deb
                sudo dpkg -i cuda-keyring_1.1-1_all.deb
                ;;
            ubuntu2204|ubuntu2004)
                sudo wget https://developer.download.nvidia.com/compute/cuda/repos/$OS_VERSION/x86_64/cuda-keyring_1.0-1_all.deb
                sudo dpkg -i cuda-keyring_1.0-1_all.deb
                ;;
            *)
                ;;
        esac
    fi
}

download_deps() {
    echo "[Download deps]"
    echo "current path: $BASE_DIR"

    if [ "$OS_VERSION" == "ubuntu2404" ]; then
        echo "[Download deps] Not supported Ubuntu 24.04, EXIT ..."
        return
    fi

    sudo apt-get update
    modules=("cuda-12-1" "nvidia-kernel-open-545" "nvidia-driver-545")
    for mod in "${modules[@]}"; do
        echo "download modules ${mod} ..."
        sudo apt-get -d install ${mod} -y
    done

    distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
    sudo curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | sudo apt-key add -
    sudo curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.list | tee /etc/apt/sources.list.d/libnvidia-container.list

    sudo apt-get update
    sudo apt-get -d install nvidia-container-toolkit -y
    sudo apt-get -d install jq -y

    dir="$BASE_DIR/tmp/"
    mkdir -p $dir/archives && mkdir -p $dir/images
    pushd $dir
        filename="${OS_VERSION}_gpu_deps.tar.gz"
        sudo mv /var/cache/apt/archives/*.deb ./archives/
        

        imgs=("grgalex/nvshare:libnvshare-v0.1-f654c296" "grgalex/nvshare:nvshare-device-plugin-v0.1-f654c296" "grgalex/nvshare:nvshare-scheduler-v0.1-f654c296" "nvcr.io/nvidia/k8s-device-plugin")
        for img in "${imgs[@]}"; do
            echo "pull image ${img} ..."
            sudo docker pull $img
            imgname=$(echo "${img}" |md5sum |awk '{print $1}')
            sudo docker save -o ./images/${imgname}.tar ${img}
        done
        sudo tar -zcvf ../$filename ./archives ./images
    popd


    aws s3 cp ./$filename s3://terminus-os-install/$filename --acl=public-read
    echo "upload $filename completed"
    
}


Main() {
    echo "Current OS Version: $OS_VERSION"
    echo "Current BASE_DIR: $BASE_DIR"

    install_apt_depends
    install_gpu_keyring
    download_deps
}

Main


exit
