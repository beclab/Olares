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

## 查看当前 CUDA 版本

每个 NVIDIA 驱动版本都捆绑一个特定版本的 CUDA 运行时。

在 Olares 宿主机上运行以下命令，查看当前的驱动版本和 CUDA 版本：

```bash
nvidia-smi
```

示例输出：

```bash
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 590.44.01              Driver Version: 590.44.01      CUDA Version: 13.1     |
+-----------------------------------------+------------------------+----------------------+
| GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
| Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
|                                         |                        |               MIG M. |
|=========================================+========================+======================|
|   0  NVIDIA GeForce RTX 4060 Ti     Off |   00000000:01:00.0 Off |                  N/A |
|  0%   41C    P8              8W /  165W |   11256MiB /  16380MiB |      0%      Default |
|                                         |                        |                  N/A |
+-----------------------------------------+------------------------+----------------------+

+-----------------------------------------------------------------------------------------+
| Processes:                                                                              |
|  GPU   GI   CI              PID   Type   Process name                        GPU Memory |
|        ID   ID                                                               Usage      |
|=========================================================================================|
|    0   N/A  N/A           60935      C   ./koboldcpp                             242MiB |
```

:::tip
如果你只知道目标 CUDA 版本，可以在 [NVIDIA CUDA 发行说明](https://docs.nvidia.com/cuda/cuda-toolkit-release-notes/index.html#cuda-major-component-versions)中查找对应的驱动版本。
:::

## 安装特定版本的 CUDA

### 步骤 1：下载驱动 runfile

1. 访问 [NVIDIA 驱动下载页面](https://www.nvidia.com/en-us/drivers/)。
2. 选择你的 GPU 产品类型、系列和型号，并将操作系统选为 **Linux 64-bit**。
3. 点击 **搜索**，记录结果中显示的驱动版本号（例如 `575.64.05`）。
4. 在 Olares 宿主机上，运行以下命令下载 runfile。将 `575.64.05` 替换为你查到的驱动版本号：

    ::: code-group
    ```bash [curl]
    VERSION=575.64.05
    curl -sSOL https://us.download.nvidia.com/XFree86/Linux-x86_64/${VERSION}/NVIDIA-Linux-x86_64-${VERSION}.run
    ```
    ```bash [wget]
    VERSION=575.64.05
    wget https://us.download.nvidia.com/XFree86/Linux-x86_64/${VERSION}/NVIDIA-Linux-x86_64-${VERSION}.run
    ```
    :::

5. 赋予 runfile 可执行权限，以便下一步运行：

    ```bash
    chmod +x NVIDIA-Linux-x86_64-575.64.05.run
    ```

### 步骤 2：执行安装

使用 root 权限运行 runfile，然后重启宿主机：

```bash
sudo ./NVIDIA-Linux-x86_64-575.64.05.run
sudo reboot now
```

::: warning 必须重启
安装驱动后必须重启宿主机，更改才能生效。
:::

### 步骤 3：更新 Olares GPU 状态

宿主机重启后，执行以下命令以更新 Olares 中该节点的 CUDA 和驱动版本信息：

```bash
olares-cli gpu enable
```

### 步骤 4：确认安装成功

运行以下命令检查新的 CUDA 版本是否生效：

```bash
nvidia-smi
```

安装成功后，输出中会显示已安装的驱动版本和 CUDA 版本。例如：

```bash
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 575.64.05              Driver Version: 575.64.05      CUDA Version: 12.9     |
+-----------------------------------------+------------------------+----------------------+
```
