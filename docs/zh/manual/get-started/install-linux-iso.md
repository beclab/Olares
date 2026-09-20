---
outline: [2, 3]
description: 通过启动 U 盘在 Linux 物理机上安装 Olares，然后使用 LarePass 完成激活。
head:
  - - meta
    - name: keywords
      content: Olares, ISO 镜像, Linux 安装, 启动盘, Balena Etcher
---

# 在 Linux 设备上通过 ISO 镜像安装 Olares

本文介绍如何通过官方 ISO 镜像在物理机上安装 Olares 系统。

:::warning 在 Olares One 上安装？
此 ISO 仅适用于自托管 x86-64 硬件。如需在 Olares One 上重装系统，请参阅 [Olares One 专用 ISO 指南](/zh/one/create-bootable-usb)，使用专用镜像以保留 Olares One 的专属功能。
:::

## 前提条件

- **Olares 设备**：一台满足 [Linux 系统要求](install-olares.md#linux)的物理机。
- **处理器**：Intel 或 AMD x86-64，不支持 ARM。
<!--@include: ./reusables.md#larepass-prerequisite-->
- **网络**：有线局域网连接。
- **U 盘**：容量至少为 8 GB。
- **操作电脑**：一台用于制作启动盘的 Windows、macOS 或 Linux 电脑。

## 制作启动盘

1. 下载[最新官方 Olares ISO 镜像](https://cdn.olares.cn/olares-v1.12.6-amd64.iso)。
2. 下载并安装 [**Balena Etcher**](https://etcher.balena.io/) 工具。
3. 将 U 盘插入电脑。
4. 打开 Etcher，依次选择：

   ![启动盘](/images/manual/get-started/iso-flash.png#bordered)
    
    a. **镜像文件**：选择 Olares ISO。
    
    b. **目标磁盘**：选择 U 盘。
    
    c. 点击 **Flash** 开始写入安装镜像。

## 从 U 盘启动
1. 将刚刚制作的启动盘插入目标机器。
2. 开机进入 BIOS 设置，并将 USB 启动盘 设置为第一启动项。
3. 保存设置并重启，系统会自动进入 Olares 安装界面。

## 安装 Olares

1. 在安装菜单中选择 **Install Olares to Hard Disk** 并按回车。
2. 安装界面将显示可用磁盘（如 `sda 200G HARDDISK`）。根据提示，输入 `/dev/` 加磁盘名称（如 `/dev/sda`）以选择安装目标盘。 出现格式化风险提示时输入 `yes` 继续。安装过程约需 **4–5 分钟**。
   :::tip 提示
   安装过程中若出现 NVIDIA 显卡驱动相关提示，按回车确认即可。
   :::
3. 出现以下提示时表示安装成功：

   ```shell
   Installation completed successfully!
   ```
此时可移除 U 盘，并按 **Ctrl + Alt + Del** 重启设备。

## 验证安装

重启后进入 Ubuntu 系统：

1. 使用以下默认账户信息登录系统：
    - 账户：`olares`
    - 密码：`olares`

2. 执行以下命令检查安装状态：

    ```bash
    sudo olares-check
    ```
   输出如下表示安装成功：
    
   ```bash
   check list ---------
   check Olaresd: success
   check Containerd: success
   ```

<!--@include: ./install-and-activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md#protect-olares-id-->

<!--@include: ./reusables.md#installation-troubleshooting-tip-->
