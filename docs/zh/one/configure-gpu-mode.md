---
outline: [2, 3]
description: 在运行 Ubuntu 的 Olares One 上配置 GPU 模式。
head:
  - - meta
    - name: keywords
      content: Olares One, Ubuntu, GPU 模式, 性能模式, 安静模式, NVIDIA, nvidia-powerd, ioport
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../one/configure-gpu-mode.md)为准。
:::

# 在 Ubuntu 上配置 GPU 模式

按照本指南，在运行 Ubuntu 的 Olares One 上配置 GPU 模式。

## 前提条件

- Olares One 已开机并运行 Ubuntu。
- 可以直接或通过 SSH 访问 Olares One 上的 Ubuntu 终端，并使用 `sudo` 执行命令。

## 步骤 1：安装 NVIDIA 驱动

NVIDIA 驱动负责让 Ubuntu 与独立 GPU 通信。启用 NVIDIA 电源管理和切换 GPU 模式前，需要先安装驱动。

:::info 使用 NVIDIA .run 安装程序
`.run` 安装程序会安装 NVIDIA 驱动，并包含电源管理所需的 `nvidia-powerd` 服务。通过 `apt` 安装的驱动不包含该服务。
:::

1. 在 Olares One 的 Ubuntu 终端中，下载 NVIDIA 驱动安装程序：

   ```bash
   wget https://us.download.nvidia.com/XFree86/Linux-x86_64/595.84/NVIDIA-Linux-x86_64-595.84.run
   ```

   :::info 驱动版本
   本指南使用驱动版本 `595.84`，已在 Olares One 上测试通过。其他版本未经过测试，`nvidia-powerd` 服务在部分版本上可能无法启动。
   :::

2. 添加可执行权限：

   ```bash
   chmod +x ./NVIDIA-Linux-x86_64-595.84.run
   ```

3. 运行安装程序：

   ```bash
   sudo ./NVIDIA-Linux-x86_64-595.84.run
   ```

4. 按照安装程序中的提示操作，直到安装完成。

## 步骤 2：启用 NVIDIA 电源管理

`nvidia-powerd` 服务负责在高负载期间协调 CPU 和 GPU 的功耗分配。启用该服务后，NVIDIA 驱动可以应用系统支持的电源管理策略。

1. 启用并启动 `nvidia-powerd` 服务：

   ```bash
   sudo systemctl enable --now nvidia-powerd
   ```

2. 检查服务状态：

   ```bash
   sudo systemctl status nvidia-powerd
   ```

   服务状态应显示为 `active (running)`。

   :::info
   由于 Olares One 没有电池，你可能会看到 `ERROR! JPAC is not created/already destroyed, ignoring request.`。如果服务状态为 `active (running)`，可以忽略这条消息。
   :::

## 步骤 3：设置 GPU 模式

启用 NVIDIA 电源管理后，可以使用硬件 I/O 端口访问工具 `ioport`，将 GPU 设置为性能模式（`3`）或安静模式（`0`）。

1. 安装 `ioport`：

   ```bash
   sudo apt update && sudo apt install -y ioport
   ```

2. 根据需要选择一种模式，执行对应的命令。

   :::warning 仅适用于 Olares One
   以下命令会访问 Olares One 的嵌入式控制器（EC）。请仅在 Olares One 上使用，并原样复制所选模式的命令。
   :::

   启用性能模式（`3`）：

   ```bash
   sudo outb 0x302 0x60 && sudo outb 0x303 3
   ```

   启用安静模式（`0`）：

   ```bash
   sudo outb 0x302 0x60 && sudo outb 0x303 0
   ```

3. 检查 GPU 功耗上限是否与所选模式一致：

   ```bash
   nvidia-smi
   ```

   在 `Pwr:Usage/Cap` 列中查看 `Cap` 的值。性能模式（`3`）应为 `150W`，安静模式（`0`）应为 `50W`。

   :::info 性能模式可动态提升至 175W
   性能模式下，GPU 空闲时 `nvidia-smi` 显示的功耗上限为 `150W`。满负荷运行时，该上限会动态提升到 `175W`。
   :::

   **示例输出：**

   ```text
   +-----------------------------------------------------------------------------------------+
   | NVIDIA-SMI 595.84                 Driver Version: 595.84         CUDA Version: 13.2     |
   +-----------------------------------------+------------------------+----------------------+
   | GPU  Name                 Persistence-M | Bus-Id        Disp.A | Volatile Uncorr. ECC |
   | Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
   |                                         |                      |               MIG M. |
   |=========================================+========================+======================|
   |   0  NVIDIA GeForce RTX 5090 ...  Off |   00000000:02:00.0 Off |                  N/A |
   | N/A   41C    P8               7W /  150W |   9706MiB /  24463MiB |      0%      Default |
   |                                         |                      |                    N/A |
   +-----------------------------------------+------------------------+----------------------+
   ```

## 资源

- [在 Olares One 上安装 Ubuntu Server](install-ubuntu-server.md)
- [在 Olares One 上安装 Ubuntu Desktop](install-ubuntu-desktop.md)
- [管理 BIOS 和 EC](update-firmware.md)
