---
outline: [2,3]
description: 通过在 Olares One 的主 SSD 上替换现有的 Olares OS 来重新安装 Ubuntu Server。
head:
  - - meta
    - name: keywords
      content: Olares One, Ubuntu Server, NVMe SSD, 操作系统安装, 全新安装
---

# 在 Olares One 上安装 Ubuntu Server <Badge type="tip" text="25 min" />

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/install-ubuntu-server.md)。
:::

将 Olares One 主 NVMe SSD 上预装的 Olares OS 替换为 Ubuntu Server 的全新安装。

:::danger 这将抹掉 Olares OS
在主驱动器上安装 Ubuntu Server 将永久删除 Olares OS、本地账户、已安装的应用、设置以及存储在该驱动器上的数据。在继续之前备份你需要的所有内容。
:::

## 学习目标

完成本指南后，你将学会：
- 创建可启动的 Ubuntu Server 安装 U 盘。
- 从安装 U 盘启动 Olares One。
- 覆盖主驱动器并安装 Ubuntu Server。

## 前提条件

**硬件**
- Olares One 内部安装的主 NVMe M.2 SSD。
- 一个 U 盘（8 GB 或更大）用于安装介质。
- 有线键盘和鼠标。
- 连接到 Olares One 的显示器。

## 步骤 1：创建可启动的 Ubuntu Server U 盘

1. 从[官方 Ubuntu 网站](https://ubuntu.com/download/server)下载 Ubuntu Server ISO，版本 26.04 LTS 或更高。
2. 下载并安装 [balenaEtcher](https://etcher.balena.io/)。
3. 将 U 盘插入你的计算机。
4. 打开 balenaEtcher 并按照以下步骤操作：

   a. 点击 **从文件刷入** 并选择下载的 ISO。

   b. 点击 **选择目标** 并选择你的 U 盘。

   c. 点击 **刷入！** 将安装程序写入 U 盘。

   ![balenaEtcher 刷入界面](/images/one/balenaEtcher.png#bordered)

5. 等待刷入和验证过程完成，然后安全弹出 U 盘。

## 步骤 2：从 Ubuntu Server U 盘启动

1. 将 Ubuntu Server U 盘插入 Olares One 的 USB 端口。
2. 打开 Olares One 电源，或如果当前正在运行则重启它。
3. 当 Olares 标志出现时，反复按 **Delete** 键进入 BIOS 设置。

   ![BIOS 设置菜单](/images/one/bios-setup-interface.png#bordered)

4. 进入 **保存并退出** 标签页。
5. 在 **启动覆盖** 部分，从列表中选择你的 U 盘，然后按 **Enter**。

    ![在 BIOS 启动菜单中选择 Ubuntu U 盘](/images/one/select-ubuntu-usb-in-bios3.png#bordered)

    系统重新启动并从 U 盘进入 Ubuntu Server 安装界面。

## 步骤 3：安装并配置 Ubuntu Server

基于文本的安装向导将引导你完成替换主驱动器上的 Olares OS 的过程。

1. 在 GNU GRUB 中，选择 **试用或安装 Ubuntu Server**。等待初始加载序列完成，语言选择屏幕出现。

   ![Ubuntu 安装类型](/images/one/ubuntu-install-type.png#bordered)

2. 选择你的语言，然后按 **Enter**。

   ![Ubuntu 语言选择](/images/one/ubuntu-language.png#bordered)

3. 保持默认的英文 US 键盘布局，然后按 **Enter**。
4. 在 **选择安装类型** 屏幕上，选择 **Ubuntu Server**，然后按 **Enter**。
5. 在 **网络配置** 屏幕上，暂时跳过网络配置，选择底部的 **继续而不连接网络**，然后按 **Enter**。

    :::tip
    连接到网络会触发补丁和依赖项的自动后台下载。这可能会显著延迟安装，并可能由于网络波动导致安装程序挂起。跳过此步骤可确保从纯 ISO 镜像进行快速、完全本地的安装。
    :::

   ![Ubuntu 网络配置](/images/one/ubuntu-network.png#bordered)

6. 在 **代理配置** 屏幕上，除非你的环境需要，否则留空，然后按 **Enter**。
7. 在 **Ubuntu 存档镜像配置** 屏幕上，保持默认的 Ubuntu 存档镜像 URL，忽略 "无网络" 警告，然后按 **Enter**。
8. 在 **引导式存储配置** 屏幕上：

   a. 确保 **使用整个磁盘** 已选中，并且包含 Olares OS 的主磁盘显示在下方的下拉列表中。

   b. 向下找到 **将此磁盘设置为 LVM 组** 并使用 **Space** 键清除选择。

      :::tip
      禁用 LVM 会强制安装程序自动创建稳定、简单的标准 ext4 分区。这消除了不必要的复杂性，并确保干净、传统的分区布局。
      :::

   c. 转到页面底部，选择 **完成**，然后按 **Enter**。

   ![Ubuntu 引导式存储配置](/images/one/ubuntu-guided-storage1.png)

9. 在 **存储配置** 摘要屏幕上验证详细信息。在 **已使用设备** 下，确保主驱动器处于 "将被格式化" 状态，然后按 **Enter**。

    ![Ubuntu 存储配置摘要](/images/one/ubuntu-storage-summary1.png)

10. 在 **确认破坏性操作** 窗口中，选择 **继续**，然后按 **Enter**。
11. 在 **配置文件** 屏幕上，设置你的账户凭据，选择底部的 **完成**，然后按 **Enter**。
12. 在 **升级到 Ubuntu Pro** 屏幕上，选择 **暂时跳过 Ubuntu Pro 设置**，然后按 **Enter**。
13. 在 **SSH 配置** 屏幕上，选择 **安装 OpenSSH 服务器** 以允许稍后连接到网络时进行远程终端管理，然后按 **Enter**。
14. 系统开始部署。等待顶部横幅显示 **安装完成**。

      ![Ubuntu 安装完成](/images/one/ubuntu-install-complete.png#bordered)

15. 选择底部的 **立即重启**，然后按 **Enter**。
16. 移除安装 U 盘，并在提示时按 **Enter**。系统自动重新启动进入全新的 Ubuntu Server 环境。

      ![Ubuntu 启动](/images/one/ubuntu-launch.png#bordered)

17. 使用你之前设置的账户凭据登录。

## 资源

- [在 Olares One 上安装 Ubuntu Desktop](install-ubuntu-desktop.md)
- [Ubuntu Server 文档](https://ubuntu.com/server/docs)
