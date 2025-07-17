---
description: Olares Home 目录的内部结构，用于组织镜像、日志、依赖项和版本管理。详细介绍默认安装目录的架构设计。
---
# Olares Home

Olares Home 是 Olares 的默认安装目录，用于存储镜像、日志、依赖组件以及版本管理数据。本文将带你了解 Olares Home 的结构，进一步 Olares 安装过程。

## 路径及设置

默认情况下，Olares Home 位于 `~/.olares`。你可以执行 `olares-cli` 命令时配合 `--base-dir` 选项更改其位置。例如：

```bash
# 指定所有下载组件的存储目录
olares-cli download component --base-dir /custom/path
```

## 目录结构

Olares Home 目录结构如下：

```
./olares
├── images                # 存放下载的镜像文件
│   ├── {image-file}.tar.gz
│   └── {image-file}.tar.gz
├── logs                  # 存放所有日志
├── pkg                   # 存放下载的系统依赖组件
│   ├── cni               # K8s 的网络插件
│   ├── components        # 存放与 K8s 无关的 Olaers 基础软件依赖，
│   │                       如 olaresd/JuiceFS/Redis 等
│   ├── containerd        # CRI 运行时
│   ├── crictl            # CRI 命令行工具
│   ├── etcd              # K8s 的持久化数据库
│   ├── helm              # K8s 的 app 管理命令行工具
│   ├── kube              # 存放 K8s 的核心程序，如 kubeadm/kubelet/k3s
│   └── runc              # OCI 运行时
└── versions              # 存放不同的 Olares 版本
    ├── v1.10.0-20241001
    │   ├── cli
    │   ├── deploy
    │   ├── files
    │   ├── images
    │   ├── logs          # 存放该 Olares 版本相关的日志
    │   │   ├── install.log
    │   │   └── uninstall.log
    │   └── wizard        # 存放 Olares 自带的系统及用户应用 Helm chart
    └── v1.10.0-20240930
        ├── cli
        ├── deploy
        ├── files
        ├── images
        ├── logs
        │   ├── install.log
        │   └── uninstall.log
        └── wizard
```

Olares Home 的结构设计旨在优化文件管理、版本控制和资源共享，其优势包括：
- 不同版本的 Olares 存储在 `versions` 目录下，版本相关的所有文件和日志都集中在对应的子目录中。
- 一台机器上只能运行一个 Olares 实例，避免版本冲突。
- 不同版本会共享 `images` 和 `pkg` 目录下的文件，避免重复下载，节省磁盘空间。

## 了解更多
- [`olares-cli` 命令行参考](../install/cli/olares-cli.md)
- [Olares 安装流程详解](installation-process.md)
- [Olares 环境变量](environment-variables.md)
