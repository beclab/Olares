---
outline: [2, 3]
description: 在 Olares One 主盘上安装 Windows，替换 Olares OS。
head:
  - - meta
    - name: keywords
      content: Olares One, Windows, 主盘, Olares OS, Windows 安装, BIOS
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../one/install-windows-primary-drive.md)为准。
:::

# 在 Olares One 上安装 Windows <Badge type="tip" text="30 min" />

当你希望将设备作为专用 Windows 机器使用时，可以在 Olares One 的主盘上安装 Windows。

此过程将替换 Olares OS。如果你希望保留 Olares OS 并在同一设备上安装 Windows，请使用双启动方案。

:::danger 这会清除 Olares OS
在主盘上安装 Windows 将永久删除 Olares OS、本地账户、已安装的应用、设置以及该驱动器上存储的数据。在继续之前，请备份所有需要保留的数据。
:::

## 学习目标

完成本指南后，你将学会：

- 创建可启动的 Windows 安装 U 盘。
- 从 Windows U 盘启动 Olares One。
- 删除现有的 Olares OS 分区，在主盘上安装 Windows。
- 安装为 Olares One 提供的 Windows 驱动程序。

## 前提条件

**硬件**<br>
- 一个 8 GB 或更大的 U 盘，用于制作 Windows 安装介质。
- 有线键盘和鼠标，连接到 Olares One。
- 显示器，连接到 Olares One。
- 将网线连接到 Olares One，因为 Windows 安装过程中需要有线网络连接。
- 一个 Microsoft 账户，用于完成 Windows 设置过程。

## 步骤 1：创建可启动的 Windows U 盘

1. 从 [Microsoft 官方网站](https://www.microsoft.com/en-us/software-download/windows11)下载 Windows 11 ISO。
2. 下载并安装 [balenaEtcher](https://etcher.balena.io/)。
3. 将 U 盘插入你的电脑。
4. 打开 balenaEtcher，按以下步骤操作：

   a. 点击 **Flash from file**，选择你下载的 ISO 文件。

   b. 点击 **Select target**，选择你的 U 盘。

   c. 点击 **Flash!**，将安装程序写入 U 盘。

   ![balenaEtcher 刷写界面](/images/one/balenaEtcher.png#bordered)

5. 等待刷写和验证完成，然后安全弹出 U 盘。

## 步骤 2：从 Windows U 盘启动

1. 将 Windows U 盘插入 Olares One 的 USB 接口。
2. 打开 Olares One 电源，如果设备已在运行，则重启。
3. 当 Olares logo 出现时，反复按 **Delete** 键进入 BIOS 设置。

   ![BIOS 设置](/images/one/bios-setup-interface.png#bordered)

4. 进入 **Save and Exit** 选项卡。
5. 在 **Boot Override** 下选择 Windows U 盘，然后按 **Enter**。

   Olares One 将重新启动并从 Windows 安装程序启动。

   :::tip
   如果出现 Ventoy 界面，选择 Windows ISO 文件，然后选择 **Boot in normal mode**。
   :::

## 步骤 3：在主盘上安装 Windows

Windows 安装向导将引导你完成安装。

1. 按照屏幕提示操作，直到到达 **Select location to install Windows 11** 界面。
2. 选择主盘上的每个现有分区，然后点击 **Delete**。

   :::danger **不要删除 "Ventoy" 或 "VTOYEFI" 分区**
   这些分区属于你的 U 盘安装介质，不属于主盘。删除它们可能会损坏安装介质。
   :::

3. 主盘上的所有分区删除后，选择生成的 **Unallocated Space**。
4. 点击 **Next** 开始安装 Windows。
5. 等待 Windows 复制文件并重启数次。
6. 当最终设置过程开始时，按照屏幕提示配置 Windows。
7. Windows 桌面出现后，拔出 Windows U 盘。

## 步骤 4：安装 Olares One Windows 驱动

Windows 启动后，安装经过测试的 Olares One 驱动包。该驱动包包含音频、网络、芯片组和 NVIDIA 显卡驱动。

1. 按照[在 Windows 上安装驱动](install-nvidia-driver.md)中的说明下载并安装一体化驱动包。
2. 如果驱动安装程序提示重启，请重启 Windows。
3. Windows 重启后，检查网络、音频、显示和 GPU 设备是否正常工作。
