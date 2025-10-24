---
description: 在 PVE 虚拟化平台上使用 ISO 镜像安装配置 Olares 的完整步骤，包括系统要求、虚拟机配置、安装和激活过程。
---
# 使用 ISO 镜像在 PVE 上安装 Olares
你可以使用 ISO 镜像将 Olares 直接安装在 Proxmox 虚拟环境（PVE）上。本指南将带你了解：如何下载 Olares ISO、在 PVE 中配置必要参数、完成安装和激活流程。

::: warning 不适用于生产环境
该部署方式当前仍有功能限制，建议仅用于开发或测试环境。
:::

:::info 安装遇到问题？
如果安装过程中遇到问题，[可以提交 GitHub Issue](https://github.com/beclab/Olares/issues/new)。提交时请提供以下信息：

- 使用的平台或环境（如 Ubuntu、PVE 等）。
- 安装方式（脚本安装或 ISO 镜像）。
- 详细的错误信息（包括日志、错误提示或截图）。
:::

## 系统要求
请确保设备满足以下配置要求：

- CPU：4 核及以上，X86-64 架构（AMD 或 Intel）。ARM 芯片目前不支持 ISO 镜像安装。
- 内存：不少于 8GB 可用内存
- 存储：不少于 200GB 的可用磁盘空间，需使用 SSD 硬盘安装。使用 HDD（机械硬盘）可能会导致安装失败。
- 支持的系统版本：PVE 8.2.2

::: tip 激活遇到问题？
当前 ISO 镜像使用了国际版网络配置（使用 `olares.com` 域名）。如遇激活问题，请暂时切换至[一键安装脚本](install-pve-script.md)方式安装激活。
:::

## 下载 Olares ISO 镜像
点击[此处](https://cdn.olares.cn/olares-latest-amd64-cn.iso)下载最新官方 Olares ISO 镜像。

## 配置 PVE 虚拟机
在 PVE 运行 Olares 时，请确保虚拟机满足以下配置要求。你可以在**创建新虚拟机**时应用这些设置，也可以**调整已有虚拟机**以符合要求。

- 操作系统:
  - `ISO 镜像`：选择 Olares 官方镜像文件。
- 系统:
  - `BIOS`：选择 OVMF (UEFI)。
  - `EFI 存储`：选择一个存储位置（如本地 LVM 或目录），用于保存 UEFI 固件变量。
  - `预注册密钥`：**取消勾选**以禁用安全启动。
- 磁盘:
  - `磁盘大小 (GiB)`：不少于 200GB
- CPU:
  - `核心`：4核及以上
- 内存：
  - `内存 (MiB)`：不少于 8GB

下图为 PVE 中虚拟机硬件的示例配置。

![PVE Hardware](/images/developer/install/pve-hardware.png#bordered)

:::info 版本兼容性
虽然以上版本已经过验证，但其他版本也可能正常运行 Olares。根据你的环境可能需要进行调整。如果你在这些平台安装时遇到任何问题，欢迎在 [GitHub](https://github.com/beclab/Olares/issues/new) 上提问。
:::

## 安装 Olares

虚拟机创建完成后，按照以下步骤在 PVE 上安装 ISO。

1. 启动虚拟机。
2. 在启动菜单中选择 **Install Olares to Hard Disk**。
3. 在 Olares System Installer 界面，会显示可用磁盘列表（例如，`sda 200G QEMU HARDDISK`）。输入`/dev/`加上第一个磁盘的名称（例如，`/dev/sda`）来选择目标磁盘。当屏幕上出现警告时，输入`yes`继续即可。

   :::tip 注意
   安装过程中，可能会出现与 NVIDIA 显卡驱动相关的警告。如果出现此类警告，按**回车键**忽略即可。
   :::

4. 安装完成后，你会看到以下信息：

    ```
    Installation completed successfully!
    ```
    按**回车键**后，在 Proxmox Web 界面，选择**重启**以重启虚拟机。

## 验证安装

虚拟机重启后，将进入 Ubuntu 系统。

1. 使用默认账号登陆 Ubuntu：

     - 用户名：`olares`
    - 密码：`olares`

2. 运行以下命令确认 Olares 是否安装成功：
     ```bash
     sudo olares-check
     ```
   如果运行结果如下，则说明安装成功：

    ```
    ...
    check Olaresd:  success
    check Containerd:  success
    ```

<!--@include: ./install-and-activate-olares.md{4,16}-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{33,37}-->