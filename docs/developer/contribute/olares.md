# Olares

## Directory structure

```
olares
|-- apps                  # olares built-in apps
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
|-- build                 # olares installer
|   |-- installer
|   |-- manifest
|-- frameworks            # system runtime frameworks
|   |-- app-service
|   |-- backup-server
|   |-- bfl
|   |-- GPU
|   |-- l4-bfl-proxy
|   |-- osnode-init
|   |-- system-server
|   |-- tapr
|-- libs                  # toolkit libs
|   |-- fs-lib
|-- scripts               # scripts for build or package the olares installer
|-- third-party           # third party libs or apps integrated in olares
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

## How to install

```
curl -fsSL https://olares.sh |  bash -
```

## How to build

```
git clone https://github.com/beclab/olares

cd olares

bash scripts/build.sh

```

Run the above scripts, you will get the debug version installer package `install-wizard-debug.tar.gz`

## How to install debug version

```
mkdir -p /path/to/unpack && cd /path/to/unpack

tar zxvf /path/to/olares/install-wizard-debug.tar.gz

make install VERSION=0.0.0-DEBUG

```

## How to uninstall

```bash
olares-cli uninstall --all
```
