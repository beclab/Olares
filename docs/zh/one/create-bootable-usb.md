---
outline: [2, 3]
description: 使用最新的 Olares One ISO 创建可启动 USB 驱动器，以便在 Olares One 上重新安装或恢复 Olares OS。
head:
  - - meta
    - name: keywords
      content: Olares One, bootable USB, Olares One ISO, reinstall Olares OS, recovery USB, Balena Etcher
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../one/create-bootable-usb.md)为准。
:::

# 创建 Olares One 可启动 USB 驱动器 <Badge type="tip" text="15 min"/>

当你需要在 Olares One 上重新安装或恢复 Olares OS 时，可以使用最新的 Olares One ISO 创建可启动 USB 驱动器。

如果附带的 USB 驱动器不可用、包含较早的 OS 镜像，或者你想直接使用最新的 OS 镜像重新安装 Olares One，请使用本指南。

:::warning 仅使用 Olares One ISO
创建可启动 USB 驱动器时，请仅使用本指南中链接的 Olares One ISO。

不要使用其他 Olares 安装文档中链接的 Olares ISO 镜像，因为它们适用于通用硬件。如果使用标准自托管 ISO，你的设备可能会被识别为 **Generic** 而非 **Olares One**，部分 Olares One 专属功能可能无法使用。
:::

## 前提条件

- USB 闪存盘：容量 8 GB 或更大。

    :::warning 数据丢失
    创建可启动驱动器时，选定的 USB 驱动器将被擦除。请继续前备份所有重要文件。
    :::

- 电脑：一台 Windows、macOS 或 Linux 电脑用于执行设置。
- 网络连接：稳定的连接，用于下载 ISO 文件和 Balena Etcher。

## 创建可启动 USB 驱动器

1. 下载[最新的官方 Olares One ISO 镜像](https://cdn.olares.com/one/olares-v1.12.6-amd64.iso)到你的电脑。

2. 下载并安装 [**Balena Etcher**](https://etcher.balena.io/)。

3. 将 USB 闪存盘插入电脑。

4. 打开 Balena Etcher 并按以下步骤操作：

   a. 点击 **Flash from file**，选择你下载的 Olares One ISO。

   b. 点击 **Select target**，选择你的 USB 驱动器。

   c. 点击 **Flash!** 将安装程序写入 USB 驱动器。

   ![Balena Etcher 刷写界面](/images/one/balenaEtcher.png#bordered)

5. 等待刷写和验证完成。

6. 安全弹出 USB 驱动器。

## 后续步骤

使用可启动 USB 驱动器在 Olares One 上重新安装 Olares OS。详细说明请参阅[使用可启动 USB 重新安装 Olares OS](create-drive.md)。
