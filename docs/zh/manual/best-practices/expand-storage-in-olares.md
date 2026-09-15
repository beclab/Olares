---
outline: [2, 3]
description: 在基于 LVM 的 Olares 系统上，使用 Olares CLI 将新磁盘合并入系统卷，扩展系统存储容量。
head:
  - - meta
    - name: keywords
      content: Olares, 扩展系统存储, LVM, disk extend, olares-cli
---
# 扩展 Olares 系统存储

如果你的 Olares 系统使用基于 LVM 的存储方式，可以使用 `olares-cli disk` 命令扩展系统存储容量。扩展完成后，新增磁盘将合并入系统卷，不再显示为独立挂载点。

其他存储方案请参阅[挂载 SMB 共享](../olares/files/mount-SMB.md)、[挂载本地磁盘](../mount-local-disk.md)或[使用 U 盘](../use-usb-drive.md)。

:::warning 数据丢失警告 
`disk extend` 命令将销毁所选磁盘上的所有数据。 

在继续操作之前，请确保磁盘中没有重要数据或已完成备份。 
:::

## 开始之前

- 将外部硬盘连接到 Olares 主机。
- SSH 登录到 Olares 主机终端。

## 识别未挂载的磁盘

列出主机上的所有块设备：

```bash
lsblk | grep -v loop
```
通过磁盘容量判断新接入的磁盘，并确认该磁盘当前没有挂载点。请勿选择包含 `/` 或 `/boot` 的磁盘。

**示例输出**：

```text
NAME        MAJ:MIN RM   SIZE RO TYPE MOUNTPOINTS
sda           8:0    0 931.5G  0 disk
├─sda1        8:1    0   512M  0 part /boot
└─sda2        8:2    0   931G  0 part /
nvme1n1     259:3    0 931.5G  0 disk
```
示例中的 `sda` 是系统盘，挂载点为 `/` 和 `/boot`， `nvme1n1` 是新连接的磁盘。

## 扩展系统存储

1. 确认 Olares 已识别但未挂载该磁盘：

    ```bash
    olares-cli disk list-unmounted
    ```

2. 将检测到的未挂载磁盘加入系统存储：

    ```bash
    sudo olares-cli disk extend
    ```
3. 当命令行提示确认时，输入 `YES` 继续。

    ```text
    WARNING: This will DESTROY all data on /dev/<device>
    Type 'YES' to continue, CTRL+C to abort:
    ```

    **示例输出**：
    ```text
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

## 验证扩展结果

你可以在终端和 UI 界面中验证存储空间是否已增加：

### 在终端

- 检查数据存储位置 `/olares` 目录的大小以确认扩容成功：

    ```bash
    df -h /olares
    ```

    **示例输出**：
    ```text
    Filesystem                  Size   Used  Avail Use% Mounted on
    /dev/mapper/olares--vg-root 1.8T   285G   1.4T  17% /olares
    ```

- 查看新硬盘是否已合并入 `olares--vg-data` 卷：

    ```bash
    lsblk | grep -v loop
    ```
    **示例输出**：
    ```text
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

### 在 UI 界面
从启动台打开仪表盘，确认系统总存储容量已增加。

![Check disk volume in Dashboard](/images/zh/manual/tutorials/expand-dashboard-disk.png#bordered)


如需查看完整用法与选项，请参考 `disk` 命令文档。