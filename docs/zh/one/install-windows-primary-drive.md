---
outline: [2, 3]
description: 在 Olares One 主盘上安装 Windows，替换 Olares OS。
head:
  - - meta
    - name: keywords
      content: Olares One, Windows, 主盘, Olares OS, Windows 安装, BIOS
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../one/install-windows-primary-drive.md)为准。
:::

# 在 Olares One 上安装 Windows <Badge type="tip" text="30 min" />

如果希望将 Olares One 用作专门运行 Windows 的电脑，可以在主硬盘上安装 Windows。

此操作会替换 Olares OS。如果希望在同一台设备上保留 Olares OS 和 Windows，请使用双系统方案。

:::danger 这会清除 Olares OS
在主盘上安装 Windows 将永久删除 Olares OS、本地账户、已安装的应用、设置以及该硬盘上存储的数据。继续前，请备份需要保留的内容。
:::

## 学习目标

完成本指南后，你将学会：

- 创建可启动的 Windows 安装 U 盘。
- 从 Windows U 盘启动 Olares One。
- 删除现有的 Olares OS 分区，在主盘上安装 Windows。
- 安装 Olares One 的 Windows 驱动程序。

## 前提条件

**硬件**<br>
- 一台电脑，用于制作安装 U 盘。
- 一个 8 GB 或更大的 U 盘，用于制作 Windows 安装介质。
- 有线键盘和鼠标，连接到 Olares One。
- 显示器，连接到 Olares One。
- 将网线连接到 Olares One，因为 Windows 安装过程中需要有线网络连接。
- 一个 Microsoft 账户，用于完成 Windows 设置过程。

## 步骤 1：创建可启动的 Windows U 盘

1. 打开 [Microsoft 下载页面](https://www.microsoft.com/zh-cn/software-download/windows11)，找到适用于 x64 设备的 Windows 11 磁盘映像（ISO）下载区域（英文页面标题为 **Download Windows 11 Disk Image (ISO) for x64 devices**），下载 ISO 文件。即使使用 Apple Silicon Mac 制作 U 盘，也要选择 x64 版本，因为 Windows 将安装在 Olares One 上。

   :::warning
   制作安装盘会清空所选 U 盘。请先备份其中的文件，并在写入前核对 U 盘名称和容量。
   :::

2. 根据电脑的操作系统下载工具，然后按照工具的说明，使用 Windows 11 ISO 制作可启动的安装 U 盘。

   - **macOS**：下载 [Wingman](https://github.com/duggan/Wingman/releases)，需要 macOS 13 或更高版本。
   - **Windows**：下载 [Rufus](https://rufus.ie/)。

   <details>
   <summary>在 macOS 上使用 Wingman 的示例</summary>

   1. 打开下载的 `Wingman.dmg`，将 Wingman 拖入“应用程序”。
   2. 打开 Wingman。如果出现 **Wingman needs permission to format USB drives** 提示，点击 **Open System Settings**，在“通用 > 登录项与扩展”的“允许在后台”列表中开启 **Wingman**。如果 macOS 要求验证，请使用触控 ID 或 Mac 登录密码。授权后返回 Wingman，应用会自动检测权限。
   3. 插入 U 盘。点击 **Choose ISO…**，选择已下载的 Windows 11 x64 ISO 文件。
   4. 在 **USB drive** 下选择目标 U 盘。Wingman 显示的是设备名称和容量，名称可能与访达中的卷名不同。请根据这些信息确认目标 U 盘。
   5. 点击 **Create Windows USB**，在确认窗口中再次核对设备名称和容量，然后点击 **Erase and Create**。这会清空该 U 盘上的所有现有数据。
   6. Wingman 会格式化 U 盘、复制安装文件并拆分 `install.wim`。制作期间保持 U 盘连接。等待出现 **Windows 11 USB created. You can unplug it now.** 提示。此时安装 U 盘已自动弹出，可以拔下。

   </details>

## 步骤 2：从 Windows U 盘启动

1. 将 Windows U 盘插入 Olares One 的 USB 接口。
2. 打开 Olares One 电源，如果设备已在运行，则重启。
3. 当 Olares logo 出现时，反复按 **Delete** 键进入 BIOS 设置。

   ![BIOS 设置](/images/one/bios-setup-interface.png#bordered)

4. 进入 **Save & Exit** 选项卡。
5. 在 **Boot Override** 下选择 Windows U 盘，然后按 **Enter**。Olares One 将重启并进入 Windows 安装程序。

   :::tip
   如果出现 Ventoy 界面，选择 Windows ISO 文件，然后选择 **Boot in normal mode**。
   :::

## 步骤 3：在主盘上安装 Windows

Windows 安装向导将引导你完成安装。

1. 按照屏幕提示操作，直到进入 **Select location to install Windows 11** 界面。

   ![Windows 11 安装位置选择界面，显示内置 SSD 和安装 U 盘](/images/one/windows-select-install-location.png)

2. 确认 Olares OS 所在的内置 SSD。

   :::tip 磁盘和分区标识
   **Disk** 后的编号表示磁盘，同一编号下的 **Partition** 表示这块磁盘上的不同分区。**Type** 列中的 **Primary** 表示分区类型。它不代表应该安装 Windows 的目标 SSD。
   :::

   以上图为例：

   | 列表项 | 含义 | 操作 |
   | --- | --- | --- |
   | Disk 0 Partition 1，512.0 MB，System | 内置 SSD 上的启动分区 | 确定替换 Olares OS 后，删除此分区。 |
   | Disk 0 Partition 2，1.9 TB，Unknown | 同一块内置 SSD 上原有的 Olares OS 分区 | 确定替换 Olares OS 后，删除此分区。 |
   | Disk 1 Partition 1: WINDOWS，57.7 GB，Primary | Windows 安装 U 盘 | 保留，不要删除或格式化。 |
   | Disk 2 Unallocated Space，0 B | 没有可用容量的列表项 | 忽略此 0 B 条目。 |

   :::warning 删除前确认自己的磁盘
   图片仅为示例，不代表所有设备的磁盘布局。你的磁盘编号、卷名、容量和分区布局可能不同。

   你需要自行确认目标磁盘并完成备份。只删除准备清空的 SSD 上的分区。如果无法确认目标 SSD，请勿继续操作。
   :::

3. 逐个选择已确认的目标 SSD 上的分区，点击 **Delete Partition**。在此示例中，删除 **Disk 0 Partition 1** 和 **Disk 0 Partition 2**，保留 Disk 1 和 Disk 2 不动。
4. 删除目标 SSD 上的所有分区后，选择同一块磁盘上释放出的未分配空间。在此示例中，选择 **Disk 0 Unallocated Space**（**1.9 TB**）。

   ![删除目标 SSD 分区后选择 Disk 0 Unallocated Space](/images/one/windows-select-unallocated-space.png)

5. 点击 **Next** 继续。
6. 在 **Ready to install** 界面核对安装信息，然后点击 **Install** 开始安装 Windows。
7. 等待 Windows 复制文件并重启数次。
8. 进入 Windows 初始设置流程后，按照屏幕提示完成配置。
9. Windows 桌面出现后，拔出 Windows U 盘。

## 步骤 4：安装 Olares One Windows 驱动

Windows 启动后，安装经过测试的 Olares One 驱动包。该驱动包包含音频、网络、芯片组和 NVIDIA 显卡驱动。

1. 按照[在 Windows 上安装驱动](install-nvidia-driver.md)中的说明下载并安装一体化驱动包。
2. 如果驱动安装程序提示重启，请重启 Windows。
3. Windows 重启后，检查网络、音频、显示和 GPU 设备是否正常工作。
