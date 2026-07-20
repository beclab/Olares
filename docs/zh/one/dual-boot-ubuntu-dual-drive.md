---
outline: [2,3]
description: 在 Olares One 的第二块 SSD 上安装 Ubuntu，并设置与 Olares OS 的双系统启动。
head:
  - - meta
    - name: keywords
      content: 双系统启动, Ubuntu, NVMe SSD, GRUB, Olares One
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../one/dual-boot-ubuntu-dual-drive.md)为准。
:::

# 在第二块 SSD 上双系统启动 Ubuntu <Badge type="tip" text="40 min" />

在第二块 NVMe SSD 上安装 Ubuntu，为开发、测试或备用系统创建独立环境，而不会影响 Olares OS。

双硬盘方案在物理层面隔离了两个系统。这确保了 Olares OS 的稳定性和安全性，同时提供原生启动任一操作系统的灵活性。

## 学习目标

阅读本指南后，你将学会如何：
- 制作可启动的 Ubuntu U 盘。
- 在第二块 SSD 上安装 Ubuntu。
- 配置 GRUB 以同时检测 Olares OS 和 Ubuntu。
- 在启动时切换两个系统。

## 前提条件

**硬件**
- Olares One 已物理安装第二块 NVMe M.2 SSD。
- 一个 U 盘（8 GB 或更大）用于制作 Ubuntu 安装介质。
- 有线键盘和鼠标。
- 显示器已连接到 Olares One。

## 步骤 1：制作可启动的 Ubuntu U 盘

1. 从 [Ubuntu 官方网站](https://ubuntu.com/download/server) 下载 Ubuntu ISO（26.04 LTS 或更新版本）。你可以选择 Server 或 Desktop 版本。
2. 下载并安装 [balenaEtcher](https://etcher.balena.io/)。
3. 将 U 盘插入你的电脑。
4. 打开 balenaEtcher，按以下步骤操作：

   a. 点击 **Flash from file**，选择你下载的 ISO 文件。

   b. 点击 **Select target**，选择你的 U 盘。

   c. 点击 **Flash!**，将安装镜像写入 U 盘。

   ![balenaEtcher 刷写界面](/images/one/balenaEtcher.png#bordered)

5. 等待刷写和验证完成，然后安全弹出 U 盘。

## 步骤 2：从 Ubuntu U 盘启动

1. 将 Ubuntu U 盘插入 Olares One。
2. 打开 Olares One 电源，或如果已在运行则重启。
3. 当 Olares 标志出现时，反复按 **Delete** 键进入 BIOS 设置。

   ![BIOS 设置菜单](/images/one/bios-setup-interface.png#bordered)

4. 进入 **Save & Exit** 选项卡，在 **Boot Override** 下从列表中选择你的 U 盘，然后按 **Enter**。

   ![在 BIOS 启动菜单中选择 Ubuntu U 盘](/images/one/select-ubuntu-usb-in-bios2.png#bordered)

   系统将重启并从 U 盘启动进入 Ubuntu 安装界面。

## 步骤 3：在第二块 SSD 上安装 Ubuntu

以下步骤以 Ubuntu Server 26.04 为例。Desktop 版本的流程类似。

1. 在 GNU GRUB 中，选择 **Try or Install Ubuntu Server**。等待初始加载序列完成，语言选择界面出现。

   ![Ubuntu 安装类型](/images/one/ubuntu-install-type.png#bordered)

2. 选择你的语言，然后按 **Enter**。

   ![Ubuntu 语言选择](/images/one/ubuntu-language.png#bordered)

3. 保持默认键盘布局 English (US)，然后按 **Enter**。
4. 在 **Choose the type of installation** 界面，选择 **Ubuntu Server**，然后按 **Enter**。
5. 在 **Network configuration** 界面，暂时跳过网络配置，选择底部的 **Continue without network**，然后按 **Enter**。

   :::tip
   连接到网络会触发补丁和依赖项的自动后台下载。这可能会显著延迟安装，并可能因网络波动导致安装程序挂起。跳过此步骤可确保从纯净 ISO 镜像进行快速、完全本地的安装。
   :::

   ![Ubuntu 网络配置](/images/one/ubuntu-network.png#bordered)

6. 在 **Proxy configuration** 界面，除非你的环境需要代理，否则留空，然后按 **Enter**。
7. 在 **Ubuntu archive mirror configuration** 界面，保持默认的 Ubuntu 存档镜像 URL，忽略 "no network" 警告，然后按 **Enter**。
8. 在 **Guided storage configuration** 界面：

   a. 确保已选择 **Use an entire disk**。

   b. 在下方的下拉列表中，确认已选中目标磁盘。例如，本场景中的 **FORESEE** 磁盘。

   c. 向下导航到 **Set up this disk as an LVM group**，使用 **Space** 键取消选择。

   :::tip
   禁用 LVM 可强制安装程序自动创建稳定、简单的标准 ext4 分区。这消除了多操作系统环境中未来 GRUB 引导加载器冲突或错误的风险。
   :::

   d. 导航到页面底部，选择 **Done**，然后按 **Enter**。

   ![Ubuntu 引导存储配置](/images/one/ubuntu-guided-storage.png)

9. 在 **Storage configuration** 摘要界面，确认以下详细信息，然后按 **Enter**：

   - 在 **FILE SYSTEM SUMMARY** 下，确保系统已在目标磁盘上自动分配 `/boot/efi` (fat32) 分区和 `/` (ext4) 标准分区。
   - 在 **USED DEVICES** 下，确保只有你的目标磁盘处于 "to be formatted" 状态。

   ![Ubuntu 存储配置摘要](/images/one/ubuntu-storage-summary.png)

10. 在 **Confirm destructive action** 窗口中，选择 **Continue**，然后按 **Enter** 开始格式化。
11. 在 **Profile configuration** 界面，设置你的账户，然后按 **Enter**。
12. 在 **Upgrade to Ubuntu Pro** 界面，选择 **Skip Ubuntu Pro setup for now**，然后按 **Enter**。
13. 在 **SSH configuration** 界面，选择 **Install OpenSSH server** 以允许后续连接网络后进行远程终端管理，然后按 **Enter**。
14. 系统将开始部署。等待顶部横幅显示 **Installation complete**。

   ![Ubuntu 安装完成](/images/one/ubuntu-install-complete.png#bordered)

15. 选择底部的 **Reboot Now**，然后按 **Enter**。
16. 出现提示时，拔出安装 U 盘并按 **Enter**。系统将自动重启。

## 步骤 4：修改 BIOS 启动顺序

重启后，系统默认启动进入 Olares OS。这是因为安装程序将其引导加载器 (GRUB) 放置在了主磁盘的 EFI 分区中，而主板仍将原来的 Olares 硬盘识别为主启动设备。

手动将 **Boot Option #1** 更新为新硬盘，以强制主板加载新生成的 GRUB 菜单。该菜单能够成功识别 Ubuntu 和 Olares OS。

1. 重启 Olares One，反复按 **Delete** 键进入 BIOS 设置。
2. 进入 **Boot** 选项卡，找到 **Boot Option Priorities** 部分。
3. 将 **Boot Option #1** 改为指向新安装的硬盘：

   a. 导航到 **Boot Option #1**，然后按 **Enter**。

   b. 在弹出窗口中，选择新安装的硬盘，然后按 **Enter**。

   ![在 BIOS 中修改启动顺序](/images/one/ubuntu-boot-order.png)

4. 按 **F10**，然后选择 **Yes** 保存并退出 BIOS。系统将自动重启。

## 步骤 5：在 Olares OS 和 Ubuntu 之间切换

重启后，**GNU GRUB** 双系统启动菜单将自动出现。

1. 选择要启动的操作系统。系统将在 10 秒后自动执行高亮条目。

   - **启动 Ubuntu**：选择 **Ubuntu**。
   - **启动 Olares OS**：选择包含 Olares 的条目，如 **Ubuntu 24.04.3 LTS (24.04) (on /dev/mapper/olares-vg-root)**。

   ![GRUB 双系统菜单](/images/one/grub-dual-os-ubuntu.png#bordered)

2. 在已登录状态下切换到另一个操作系统，在终端运行 `sudo reboot`，出现提示时输入密码。当 **GNU GRUB** 菜单出现时，选择你想要启动的系统。

   :::info
   在终端输入密码时，出于安全考虑字符不会显示。确保输入正确的密码后按 **Enter**。
   :::

## 相关资源

- [在第二块 SSD 上双系统启动 Windows](dual-boot-dual-drive.md)
- [Ubuntu Server 文档](https://ubuntu.com/server/docs)
- [GRUB 手册](https://www.gnu.org/software/grub/manual/grub/html_node/)
