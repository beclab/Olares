---
description: 在 PVE 虚拟化平台上使用 ISO 镜像安装配置 Olares 的完整步骤，包括系统要求、虚拟机配置、安装和激活过程。
---
# 使用 ISO 镜像在 PVE 上安装 Olares
你可以使用 ISO 镜像将 Olares 直接安装在 Proxmox 虚拟环境（PVE）上。本指南将带你了解:如何下载 Olares ISO、在 PVE 中配置必要参数、完成安装和激活流程。

::: warning 不适用于生产环境
该部署方式当前仍有功能限制，建议仅用于开发或测试环境。
:::

<!--@include: ./reusables.md{39,45}-->

## 系统要求
请确保设备满足以下配置要求：

- CPU：4 核及以上
- 内存：不少于 8GB 可用内存
- 存储：不少于 200GB 的可用磁盘空间，需使用 SSD 硬盘安装。
- 支持的系统版本：PVE 8.2.2
## 下载 Olares ISO 镜像
下载官方 Olares ISO 镜像。

## 配置 PVE 虚拟机
在 PVE 运行 Olares 时，请确保虚拟机满足以下配置要求。你可以在**创建新虚拟机**时应用这些设置，也可以**调整已有虚拟机**以符合要求。

### 虚拟机所需配置

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

下图为 PVE 中虚拟机硬件的示例配置。

![PVE Hardware](/images/developer/install/pve-hardware.png#bordered)

:::warning 注意
使用 HDD（机械硬盘）可能会导致安装失败。
:::

:::info 版本兼容性
虽然以上版本已经过验证，但其他版本也可能正常运行 Olares。根据你的环境可能需要进行调整。如果你在这些平台安装时遇到任何问题，欢迎在 [GitHub](https://github.com/beclab/Olares/issues/new) 上提问。
:::

## 安装 Olares

1. 启动虚拟机。
2. 在启动菜单中选择 **Install Olares to Hard Disk**。
3. 在 Olares System Installer 界面，可用的磁盘会被列出。输入第一个磁盘的盘符作为目标磁盘。

    - 随后会出现警告：

    ```text
    WARNING: This will DESTROY all data on <your target disk>
    ```

    出现 `Continue? (yes/no):` 提示时，输入 `yes` 以继续。

:::tip 
安装过程中会显示安装日志。如果需要安装显卡驱动，会出现安装进度条，并可能显示相关警告，例如：

    ```bash
   WARNING: nvidia-installer was forced to guess the X Iibrary path 'usr/lib'and X module path ...
    ```

按**回车键**忽略警告即可。

4. 安装完成后，你会看到以下信息：

    ```
    Installation completed successfully!
    ```

    按回车键后，按组合键 **CTRL + ALT + DEL** 来重启虚拟机。

## 验证安装

虚拟机重启后，将进入 Ubuntu 系统。

1. 使用默认账号登陆 Ubuntu：

- 用户名：`olares`
- 密码：`olares`

2. 运行以下命令确认 Olares 是否安装成功：`sudo olares-check`

如果运行结果如下，则说明安装成功：

```
...
check Olaresd:  success
check Containerd:  success
```

## 激活 Olares

请按以下步骤激活 Olares：
![ISO 激活](/images/manual/larepass/iso-activate.png#bordered)


1. 打开 LarePass 应用。
2. 点击**发现附近的 Olares**，应用将显示你的 Olares 设备。
3. 点击**立即安装**，完成剩余安装过程。
4. 点击**立即激活**以激活设备并初始化系统。
5. 按提示设置 Olares 登录密码。
   ![ISO 激活-2](/images/manual/larepass/iso-activate-2.png#bordered)

完成后，你即可通过个人 URL 和凭证访问 Olares。

:::tip 注意
此操作需要你的手机和 PVE 主机处于同一网络。
:::

## 登录 Olares

1. 在浏览器输入个人 URL，（如`https://desktop.{olares-id}.olares.cn`），按任意键继续。

2. 在登录界面，输入 Olares 密码。

   ![Log in](/images/manual/get-started/log-in.png#bordered)
3. 系统会要求完成双重验证。你可以选择在 LarePass 上确认登录，或手动输入 6 位验证码。
   
   ![确认登录](/images/manual/larepass/confirm-login.png#bordered)

   ::: info
   验证码有时效限制，请在过期前完成输入。如果验证码过期，需要重新生成。
   :::

登录后你就会看到 Olares 桌面。🎉

<!--@include: ./reusables.md{33,37}-->