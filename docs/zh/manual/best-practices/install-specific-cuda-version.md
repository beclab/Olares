---
outline: [2, 3]
description: 了解如何在 Olares 宿主机上安装特定版本的 NVIDIA CUDA 驱动，以满足不同场景下的版本需求。
---

# 安装特定版本的 CUDA

在 Olares 上运行基于 NVIDIA 显卡的应用时，宿主机和应用容器都需要安装 CUDA 驱动。虽然这两个版本通常需要匹配，但大部分情况下，即使容器内的 CUDA 版本高于宿主机，应用也能正常运行。

为了支持最前沿的 AI 应用，Olares 官方只维护最新版本的 CUDA。但在以下场景中，你可能需要安装其他版本：

- 某些应用或 AI 模型依赖特定的 CUDA 或驱动版本才能运行。
- 你希望固定版本以保持稳定性，避免自动升级。
- 最新驱动与你的工作负载存在兼容性问题。

在这些情况下，可以选择在宿主机上手动安装特定版本的 CUDA 驱动。

## 前提条件

开始前，请确保你的环境满足以下要求：

- 已安装并启用了 GPU 支持的 Olares
- 兼容的 NVIDIA GPU
- 对 Olares 宿主机的 root 或 sudo 权限

## 步骤 1：下载驱动 runfile

1. 访问 [NVIDIA 驱动下载页面](https://www.nvidia.com/en-us/drivers/)。
2. 选择你的 GPU 产品类型、系列和型号，并将操作系统选为 **Linux 64-bit**。
3. 点击 **搜索**，记录结果中显示的驱动版本号（例如 `575.64.05`）。
4. 在 Olares 宿主机上，将以下命令中的版本号替换为你查到的版本号，然后执行以下载 runfile：

```bash
curl -sSOL https://us.download.nvidia.com/XFree86/Linux-x86_64/575.64.05/NVIDIA-Linux-x86_64-575.64.05.run
chmod +x NVIDIA-Linux-x86_64-575.64.05.run
```

## 步骤 2：执行安装

使用 root 权限运行 runfile，然后重启宿主机：

```bash
sudo ./NVIDIA-Linux-x86_64-575.64.05.run
sudo reboot now
```

::: warning 必须重启
安装驱动后必须重启宿主机，更改才能生效。
:::

## 步骤 3：更新 Olares GPU 状态

宿主机重启后，执行以下命令以更新 Olares 中该节点的 CUDA 和驱动版本信息：

```bash
olares-cli gpu enable
```

::: tip 无需先执行禁用
直接运行 `olares-cli gpu enable` 即可，不需要先执行 `olares-cli gpu disable`。
:::

## 步骤 4：确认安装成功

运行以下命令检查新的 CUDA 版本是否生效：

```bash
nvidia-smi
```

安装成功后，输出中会显示已安装的驱动版本和 CUDA 版本。例如：

```
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 575.64.05              Driver Version: 575.64.05      CUDA Version: 12.9     |
+-----------------------------------------+------------------------+----------------------+
```
