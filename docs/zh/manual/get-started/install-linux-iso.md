---
outline: [2, 3]
description: 通过官方 ISO 镜像在物理机上安装 Olares 系统，包括系统要求、安装命令和激活流程。
---

# 通过 ISO 镜像安装 Olares

本文介绍如何通过官方 ISO 镜像在物理机上安装 Olares 系统。

::: tip 激活遇到问题？
当前 ISO 镜像使用了国际版网络配置（使用 `olares.com` 域名）。如遇激活问题，请暂时切换至[一键安装脚本](install-linux-script.md)方式安装激活。
:::
<!--@include: ./reusables.md{44,51}-->

## 系统要求

### 必要配置

- **CPU**：4 核及以上，x86-64 架构（AMD 或 Intel）。ARM 芯片目前不支持 ISO 镜像安装。
- **内存**：至少 8 GB 可用内存。
- **存储**：至少 150 GB 的可用 SSD 磁盘空间。
   :::warning 必须使用 SSD
   使用机械硬盘 (HDD) 会导致安装失败。
   :::
- **网络**：需连接至有线局域网。
- **U 盘**：容量 8 GB 或更大。

### 可选硬件

<!--@include: ./gpu-requirements.md{5,}-->

## 制作启动盘

1. 下载[最新官方 Olares ISO 镜像](https://cdn.olares.cn/olares-latest-amd64.iso)。
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

<!--@include: ./reusables.md{38,42}-->