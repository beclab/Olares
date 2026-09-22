---
outline: [2, 3]
description: 通过启动 U 盘在 NVIDIA DGX Spark 上安装 Olares，然后激活并配置 GPU 显存。
head:
  - - meta
    - name: keywords
      content: Olares, DGX Spark, ISO 镜像, NVIDIA, GPU 显存
---

# 在 Spark 上通过 ISO 镜像安装 Olares

本文介绍如何使用官方 ISO 镜像在 NVIDIA DGX Spark 上安装 Olares。

## 前提条件

- **Olares 设备**：NVIDIA DGX Spark。
<!--@include: ./reusables.md#larepass-prerequisite-->
- **外设**：已连接到 DGX Spark 的显示器和键盘。
- **U 盘**：容量至少为 8 GB。
- **操作电脑**：一台用于制作启动盘的 Windows、macOS 或 Linux 电脑。
- **网络**：建议使用网线连接 DGX Spark 和路由器。

## 制作启动盘

1. 下载[适用于 DGX Spark 的官方 Olares ISO 镜像](https://cdn.olares.cn/spark/olares-v1.12.6-arm64.iso)。
2. 下载并安装 [**Balena Etcher**](https://etcher.balena.io/) 工具。
3. 将 U 盘插入电脑。
4. 打开 Balena Etcher，依次选择：

   ![启动盘](/images/manual/get-started/iso-flash.png#bordered)

   a. **镜像文件**：选择 Olares ISO。

   b. **目标磁盘**：选择 U 盘。

   c. 点击 **Flash** 开始写入安装镜像。

## 从 U 盘启动

1. 将制作好的启动盘插入 DGX Spark。
2. 重启 DGX Spark，然后立即反复按 **Delete** 键进入 BIOS 设置。
   <!-- ![BIOS 设置](/images/one/bios-setup.png#bordered) -->

3. 导航到 **Boot** 选项卡，将 **Boot Option #1** 设置为 U 盘，然后按 **Enter**。
   <!-- ![设置启动项](/images/one/bios-set-boot-option.png#bordered) -->

4. 按 **F4** 保存并重启。系统将自动进入 Olares 安装界面。

## 安装 Olares

1. 在安装菜单中选择 **Install Olares to Hard Disk** 并按回车。
   ![Olares 安装界面](/images/one/olares-installer.png#bordered)

2. 安装界面将显示可用磁盘（如 `sda 200G HARDDISK`）。根据提示，输入 `/dev/` 加磁盘名称（如 `/dev/sda`）以选择安装目标盘。安装过程约需 **4–5 分钟**。
   :::tip 提示
   安装过程中若出现 NVIDIA 显卡驱动相关提示，按 **Enter** 确认即可。
   :::
3. 出现以下提示时表示安装成功：

   ```shell
   Installation completed successfully!
   ```

4. 移除 U 盘，然后手动关闭 DGX Spark 再重新开机。
   :::warning 重要
   如果跳过此步骤，激活过程将失败。
   :::

5. 为避免启动延迟，打开 DGX Spark 并立即反复按 **Delete** 键进入 BIOS 设置。将内部硬盘设置为 **Boot Option #1**。

## 完成安装并激活 Olares

### 连接到 DGX Spark

<tabs>
<template #通过有线局域网设置>

1. 确保 DGX Spark 通过网线连接到路由器。
2. 打开 LarePass。如果还没有 Olares ID，点击**创建账号**并按提示完成设置。在激活页面点击**发现附近的 Olares**。
3. 从列表中选择目标 Olares 实例。

</template>
<template #通过无线网络设置>

1. 打开 LarePass。如果还没有 Olares ID，点击**创建账号**并按提示完成设置。在激活页面点击**发现附近的 Olares**。
2. 点击底部的**蓝牙网络设置**。
3. 从蓝牙列表中选择你的设备，点击**网络设置**。
4. 按照提示将 DGX Spark 连接到你手机当前使用的 Wi-Fi 网络。
5. 连接成功后，返回主屏幕并再次点击**发现附近的 Olares**以找到你的设备。

</template>
</tabs>

### 激活 Olares

1. 在 LarePass 应用中，在你刚找到的设备上点击**立即安装**。
2. 安装完成后，点击**立即激活**。
3. 在**选择反向代理**对话框中，选择距离你地理位置较近的节点。安装程序将随后为 Olares 配置 HTTPS 证书和 DNS。
   :::tip 提示
   你可以在之后 Olares 的[更改反向代理](../olares/settings/change-frp.md)页面调整此设置。
   :::
4. 选择 Olares 语言。Olares 支持英语、简体中文、德语、西班牙语、意大利语、法语和日语。
   :::info
   这里选择的是 Olares 语言，不会改变当前 LarePass 的界面语言。后续激活流程仍使用 LarePass 当前语言。激活完成后，Olares 桌面会使用此处选择的语言。
   :::
5. 按照屏幕提示设置 Olares 的登录密码，然后点击**完成**。

   ![ISO 激活-2](/images/manual/larepass/iso-activate-4.png#bordered)

激活完成后，LarePass 将显示你的 Olares 设备桌面地址，例如 `https://desktop.marvin123.olares.cn`。

<!--@include: ./log-in-to-olares.md-->

## 配置 AI 应用的 GPU 显存

DGX Spark 采用统一内存架构，CPU 和 GPU 共享 128 GB 的 LPDDR5x 内存。与传统 GPU 拥有独立显存不同，DGX Spark 不区分系统内存和 GPU 显存。

在 DGX Spark 上，Olares 默认使用**容量分片**模式进行 GPU 资源管理。当你安装 AI 应用时，Olares 会自动分配最低所需的内存，以确保应用能够正常启动和运行。

如需调整某个应用的内存分配，可以手动修改：

1. 从 Olares 打开**设置**，然后进入 **GPU** 页面。
2. 在**分配显存**区域，找到目标应用。

   ![显存切片](/images/zh/manual/get-started/install-spark-memory-slicing.png#bordered){width=70%}

3. 点击显存值旁边的 <i class="material-symbols-outlined">edit_square</i>。
4. 在**编辑显存分配**对话框中，输入所需的显存大小（GB），然后点击**确认**。

<!--@include: ./reusables.md#protect-olares-id-->

<!--@include: ./reusables.md#installation-troubleshooting-tip-->
