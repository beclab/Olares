---
outline: [2, 3]
description: 了解如何使用 NVMe M.2 SSD 扩展 Olares One 的存储空间。
head:
  - - meta
    - name: keywords
      content: Olares One, NVMe SSD, 扩展存储, LVM, olares-cli
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../one/expand-storage-internal-ssd.md)。
:::

# 使用 NVMe M.2 SSD 扩展存储 <Badge type="tip" text="15 min" />

如果你在 Olares One 中安装了第二块内置 NVMe SSD，可以使用 `olares-cli` 将其合并到主系统存储中。

与作为独立文件夹挂载的外接硬盘不同，此方法使用逻辑卷管理 (LVM) 来无缝扩展你的根文件系统。


## 开始之前
:::warning 一次性不可逆操作
- 这是一次性操作。SSD 将成为系统卷的组成部分，之后无法轻易分离。
- 如果在扩展后物理拆除此 SSD，Olares 文件系统将不完整，导致操作系统崩溃或无法启动。你需要手动重新安装 Olares OS。
  :::

Olares One 主板包含两个 PCIe SSD 插槽：
* 插槽 1 (PCIe 4.0)：预装 2TB 系统 SSD 占用。
* 插槽 2 (PCIe 5.0)：可用于扩展。

你可以使用第二个插槽来扩展当前的系统存储。

## 前提条件
**硬件**
- Olares One 已完成设置并正在运行。
- 第二块 NVMe M.2 SSD 已物理安装在 Olares One 中。

**SSH 访问**
- [SSH 访问 Olares One](access-terminal-ssh.md)。

**经验**
- 熟悉基本的终端命令和命令行界面 (CLI)。

## 步骤 1：识别未挂载的硬盘

1. 通过 SSH 或 Control Hub 连接到 Olares One 终端。

2. 列出主机上的块设备：

   ```bash
   lsblk | grep -v loop
   ```

3. 检查大小和挂载点以识别新硬盘。

   示例输出：

   ```text
   NAME        MAJ:MIN RM   SIZE RO TYPE MOUNTPOINTS
   sda           8:0    0 931.5G  0 disk
   ├─sda1        8:1    0   512M  0 part /boot
   └─sda2        8:2    0   931G  0 part /
   nvme1n1     259:3    0 931.5G  0 disk
   ```
在本示例中，`sda` 是挂载在 `/` 和 `/boot` 的系统盘，而 `nvme1n1` 是新的、未挂载的 SSD。

## 步骤 2：扩展系统存储

1. 验证 Olares 是否识别到未挂载的硬盘：

   ```bash
   olares-cli disk list-unmounted
   ```

2. 运行扩展命令：

   ```bash
   sudo olares-cli disk extend
   ```

3. 当命令提示确认时，输入 `YES` 继续。
   ```bash
   WARNING: This will DESTROY all data on /dev/<device>
   Type 'YES' to continue, CTRL+C to abort:
   ```

   示例输出：
   ```bash
   Selected volume group to extend: olares-vg
   Selected logical volume to extend: data
   Selected unmounted device to use: /dev/nvme0n1
   Extending logical volume data in volume group olares-vg using device /dev/nvme0n1
   WARNING: This will DESTROY all data on /dev/nvme0n1
   Type 'YES' to continue, CTRL+C to abort: YES
   Selected device /dev/nvme0n1 has existing partitions. Cleaning up...
   Deleting existing partitions on device /dev/nvme0n1...
   Creating partition on device /dev/nvme0n1...
   Creating physical volume on device /dev/nvme0n1...
   Extending volume group olares-vg with logic volume data on device /dev/nvme0n1...
   Disk extension completed successfully.

   id  LV    VG         LSize    Mountpoints
   1   data  olares-vg  <3.63t   /var,/olares
   2   root  olares-vg  100.00g  /
   3   swap  olares-vg  1.00g
   ...
   ```
## 步骤 3：验证扩展

你可以通过 Control Hub (UI) 或命令行验证存储扩展。

<tabs>
<template #Via-Dashboard-(UI)>

1. 打开 Dashboard。
2. 检查 Disk 部分，确认总系统存储容量已增加。

   ![在 Dashboard 中检查磁盘容量](/images/manual/tutorials/expand-dashboard-disk.png#bordered)
</template>

<template #Via-command-line>

1. 检查 `/olares` 目录的总大小：

   ```bash
   df -h /olares
   ```

   示例输出：
   ```text
   Filesystem                  Size   Used  Avail Use% Mounted on
   /dev/mapper/olares--vg-root 1.8T   285G   1.4T  17% /olares
   ```
   **Size** 列现在应反映合并后的容量。
2. 确认磁盘结构：
   ```bash
   lsblk | grep -v loop
   ```
   示例输出：

   ```bash
   NAME                MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
   nvme0n1             259:0    0  1.9T  0 disk
   └─nvme0n1p1         259:2    0  1.9T  0 part
     └─olares--vg-data 252:2    0  3.6T  0 lvm  /olares /var
   nvme1n1             259:3    0  1.9T  0 disk
   ├─nvme1n1p1         259:4    0  512M  0 part /boot/efi
   └─nvme1n1p2         259:5    0  1.9T  0 part
     ├─olares--vg-root 252:1    0  100G  0 lvm  /
     └─olares--vg-swap 252:0    0    1G  0 lvm  [SWAP]
   ```
你应该能看到新硬盘（如 `nvme0n1`）列在 `olares--vg-data` 组下，与主数据分区共享相同的挂载点。

</template>
</tabs>


## 相关资源
- [通过 USB 设备扩展存储](expand-storage-usb-drive.md)
- [使用外接 SSD 扩展存储](expand-storage-external-ssd.md)
- [`olares-cli disk`](../developer/install/cli/disk.md)。
