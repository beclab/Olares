# Olares

## 目录结构

```
olares
|-- apps                  # 系统应用
|   |-- agent
|   |-- analytic
|   |-- market
|   |-- market-server
|   |-- argo
|   |-- desktop
|   |-- devbox
|   |-- vault
|   |-- files
|   |-- knowledge
|   |-- nitro
|   |-- notifications
|   |-- profile
|   |-- rss
|   |-- search
|   |-- settings
|   |-- system-apps
|   |-- wise
|   |-- wizard
|-- build                 # Olares installer
|   |-- installer
|   |-- manifest
|-- frameworks            # 系统运行时组件
|   |-- app-service
|   |-- backup-server
|   |-- bfl
|   |-- GPU
|   |-- l4-bfl-proxy
|   |-- osnode-init
|   |-- system-server
|   |-- tapr
|-- libs                  # 工具包库
|   |-- fs-lib
|-- scripts               # 用于构建或打包 olares 安装程序的脚本
|-- third-party           # Olares 中集成的第三方库或应用程序
|   |-- authelia
|   |-- headscale
|   |-- infisical
|   |-- juicefs
|   |-- ks-console
|   |-- ks-installer
|   |-- kube-state-metrics
|   |-- notification-mananger
|   |-- predixy
|   |-- redis-cluster-operator
|   |-- seafile-server
|   |-- seahub
|   |-- tailscale
```

## 如何安装

```
curl -fsSL https://olares.sh |  bash -
```

## 如何构建

```
git clone https://github.com/beclab/olares

cd olares

bash scripts/build.sh

```

运行以上脚本，你将获得 debug 版本安装包 `install-wizard-debug.tar.gz`。

## How to install debug version

```
mkdir -p /path/to/unpack && cd /path/to/unpack

tar zxvf /path/to/olares/install-wizard-debug.tar.gz

make install VERSION=0.0.0-DEBUG

```

## 如何卸载

```bash
olares-cli uninstall --all
```
