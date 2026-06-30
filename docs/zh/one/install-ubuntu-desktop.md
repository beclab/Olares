---
outline: [2,3]
description: 通过在 Olares One 的主 SSD 上替换现有的 Olares OS 来重新安装 Ubuntu Desktop。
head:
  - - meta
    - name: keywords
      content: Olares One, Ubuntu Desktop, NVMe SSD, 操作系统安装, 全新安装, 图形化设置
---

# 在 Olares One 上安装 Ubuntu Desktop <Badge type="tip" text="30 min" />

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../one/install-ubuntu-desktop.md)为准。
:::

将 Olares One 主 NVMe SSD 上预装的 Olares OS 替换为 Ubuntu Desktop 的全新安装。

:::danger 这将抹掉 Olares OS
在主驱动器上安装 Ubuntu Desktop 将永久删除 Olares OS、本地账户、已安装的应用、设置以及存储在该驱动器上的数据。在继续之前备份你需要的所有内容。
:::

## 学习目标

完成本指南后，你将学会：
- 创建可启动的 Ubuntu Desktop 安装 U 盘。
- 从安装 U 盘启动 Olares One。
- 覆盖主驱动器并安装 Ubuntu Desktop。

## 前提条件

**硬件**
- Olares One 内部安装的主 NVMe M.2 SSD。
- 一个 U 盘（8 GB 或更大）用于安装介质。
- 有线键盘和鼠标。
- 连接到 Olares One 的显示器。

## 步骤 1：创建可启动的 Ubuntu Desktop U 盘

1. 从[官方 Ubuntu 网站](https://ubuntu.com/download/desktop)下载 Ubuntu Desktop ISO。
2. 下载并安装 [balenaEtcher](https://etcher.balena.io/)。
3. 将 U 盘插入你的计算机。
4. 打开 balenaEtcher 并按照以下步骤操作：

   a. 点击 **从文件刷入** 并选择下载的 ISO。

   b. 点击 **选择目标** 并选择你的 U 盘。

   c. 点击 **刷入！** 将安装程序写入 U 盘。

   ![balenaEtcher 刷入界面](/images/one/balenaEtcher.png#bordered)

5. 等待刷入和验证过程完成，然后安全弹出 U 盘。

## 步骤 2：从 Ubuntu Desktop U 盘启动

1. 将 Ubuntu Desktop U 盘插入 Olares One 的 USB 端口。
2. 打开 Olares One 电源，或如果当前正在运行则重启它。
3. 当 Olares 标志出现时，反复按 **Delete** 键进入 BIOS 设置。

   ![BIOS 设置菜单](/images/one/bios-setup-interface.png#bordered)

4. 进入 **保存并退出** 标签页。
5. 在 **启动覆盖** 部分，从列表中选择你的 U 盘，然后按 **Enter**。

    ![在 BIOS 启动菜单中选择 Ubuntu U 盘](/images/one/select-ubuntu-usb-in-bios3.png#bordered)

    系统从 U 盘启动进入 Ubuntu Desktop 图形化启动菜单。

## 步骤 3：安装并配置 Ubuntu Desktop

图形化安装向导将引导你完成替换主驱动器上的 Olares OS 的过程。

1. 在 GNU GRUB 中，选择 **试用或安装 Ubuntu**。等待初始加载序列完成，语言选择屏幕出现。

   ![Ubuntu 安装类型](/images/one/ubuntu-install-desktop.png#bordered)

2. 按照屏幕上的提示完成标准设置。
3. 当看到 Ubuntu 已安装并准备就绪时，点击 **立即重启**。
4. 移除安装 U 盘，并在提示时按 **Enter**。系统重新启动并进入全新的 Ubuntu Desktop 环境。
5. 使用你在配置期间设置的账户凭据登录。

## 资源

- [在 Olares One 上安装 Ubuntu Server](install-ubuntu-server.md)
- [Ubuntu Desktop 文档](https://documentation.ubuntu.com/desktop/en/latest/)
