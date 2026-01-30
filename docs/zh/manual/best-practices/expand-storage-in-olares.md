---
outline: [2, 3]
description: Olares 存储扩展指南，涵盖 SMB 服务器连接、USB 自动挂载与 HDD/SSD 手动挂载设置，助你灵活扩展本地存储空间，高效管理大型 AI 模型与数据资源。
---
# 在 Olares 中扩展存储空间

本文档介绍如何通过不同方式在 Olares 中扩展存储空间。请根据你的使用场景选择最合适的方案：

- **连接 SMB 服务器**：通过网络访问 NAS 或其他计算机上的共享文件。
- **使用 USB 自动挂载**：即插即用的外部存储方式，无需任何 Linux 命令。
- **手动挂载 HDD 或 SSD**：将硬盘挂载到 `/olares/share` 下，作为独立的外部存储，适合存放大型文件。
- **通过 Olares CLI 扩展系统存储**：在基于 LVM 的系统上使用 `disk extend` 命令，将新磁盘合并入系统卷，从而增加系统容量。

本文档介绍如何在 Olares 中扩展存储空间，包括通过 SMB 服务器连接、使用 USB 存储设备自动挂载，以及在 Linux 宿主系统中手动挂载 HDD/SSD。

## 通过 SMB 服务器连接

你可以在Olares 中轻松挂载 SMB（服务器消息块）共享，用来访问和管理共享文件。

1. 在 Olares 网页端，点击**文件管理器** > **外部设备** > **连接服务器**。

2. 输入服务器地址（例如：`//192.168.1.10/shared`），点击**确认**。

3. 连接成功后，可在**文件管理器** > **外部设备**下访问该共享目录。

详情请参阅[挂载 SMB 共享](../olares/files/mount-SMB.md)。

## 通过 USB 存储设备扩展存储

**Olares** 会自动识别并挂载插入的 USB 存储设备。

- 设备插入后将自动挂载，无需命令行操作。

- 你可以在 Olares 网页端或 Larepass 中，点击**文件管理器** > **外部设备**直接访问。

- 直接拔出 USB 设备后，系统会自动完成卸载

- 你也可以在 Olares 网页端手动移除设备。在**文件管理器**中右键点击该 USB 硬盘，选择**卸载**即可断开连接。

## 手动挂载 HDD/SSD

你可以在 **Linux** 宿主系统中手动将 HDD/SSD 挂载到 Olares。

这种方式适合**存放大容量数据**（如 AI 模型）或**长期扩展存储空间**。

### 开始之前

请确保：

- 你拥有当前 Linux 系统的**管理员权限（sudo 权限）**。

- **Olares** 已正确安装并运行。

- 目标硬盘已**格式化**（推荐使用 `ext4` 或 `XFS` 文件系统格式）。

:::tip 挂载路径限制

目前只支持将硬盘挂载在 `/olares/share` 目录下。

未来版本将提供更灵活的挂载方式，敬请期待。
:::

### 识别硬盘

1. 将硬盘插入主机。

2. 执行以下命令，查看系统识别到的硬盘信息：

   ```bash
   fdisk -l
   ```

3. 根据输出内容识别目标硬盘类型与设备名：

    - **NVMe 固态硬盘**：设备名称通常为`/dev/nvme0n1`，`/dev/nvme1n1`。
    - **SATA/机械硬盘**：设备名称通常为`/dev/sda`，`/dev/sdb`。

    每个硬盘下方会列出其分区，如`/dev/nvme1n1p1`，`/dev/nvme1n1p2`或`/dev/sdb1`等。
    ![分区列表](/images/zh/manual/tutorials/expand-storage-partition.png#bordered)

4. 确定你要挂载的目标分区，如`/dev/nvme1n1p1`。

### 临时挂载分区

临时挂载适用于**一次性**或**短期使用**场景（如文件拷贝）。

Linux 或 Olares 重启后，挂载配置将失效。

1. 创建挂载目录：

    ```bash
    sudo mkdir -p /olares/share/<目录名>
    ```

    将`<目录名>`替换为自定义目录名称。

2. 挂载分区：

    ```bash
    sudo mount /dev/<待挂载分区> /olares/share/<目录名>  
    ```

    **示例**：

    ```bash
    sudo mount /dev/nvme1n1p1 /olares/share/hdd0
    ```

3. 验证挂载结果：

    挂载成功后，可在 Olares **文件管理器** > **外部设备**目录下访问该分区内容。

    ![验证挂载结果](/images/zh/manual/tutorials/expand-storage-mount-result-cn.png#bordered)

### 持久挂载分区

如果你希望挂载配置在系统重启后仍保持有效，可通过编辑`/etc/fstab`设置**开机自动挂载**。

1. 执行以下命令获取所有磁盘，并找到目标分区：

    ```bash
    lsblk -f
    ```

    记录以下信息：
    - **FSTYPE**：文件系统类型（如 `ext4`、`xfs`）
    - **UUID**：分区唯一标识符

    ![FSTYPE信息](/images/zh/manual/tutorials/expand-storage-fstype.png#bordered)

2. 创建挂载目录

    ```bash
    sudo mkdir -p /olares/share/<目录名>
    ```

    将`<目录名>`替换为自定义目录名称。

3. 编辑挂载配置文件：

    ```bash
    sudo vi /etc/fstab
    ```

4. 添加挂载配置（推荐使用 **UUID**，以避免设备名变更导致挂载异常）：

    ```
   UUID=<分区UUID> /olares/share/<目录名> <文件系统类型> defaults,nofail 0 0
    ```

    **示例**：

    ```
    UUID=1234-ABCD /olares/share/my_disk ext4 defaults,nofail 0 0
    ```

5. 保存并退出编辑器。

6. 验证配置是否正确（推荐执行）：

    ```bash
    mount -a
    ```
    
    若无报错，表示配置成功。

7. 系统重启后，可在 Olares 的**文件管理器** > **外部设备**目录下，检查是否已自动挂载分区。

    :::warning 警告
    错误的`/etc/fstab`配置可能导致系统无法启动。
    建议先执行`mount -a`，验证无误后再重启。
    :::

### 卸载已挂载的分区

无论是临时挂载还是持久挂载的分区，都可以通过以下方式卸载。

1. 卸载分区：

    ```bash
    sudo umount /olares/share/<目录名>
    ```

    :::tip 注意
    卸载前，请确保没有程序或终端正在访问该目录，否则卸载会失败。
    :::

2. 删除空目录（可选）：

    ```bash
    rm -rf /olares/share/<目录名>
    ```

    :::warning 警告
    删除前，请确认卸载已成功且目录为空。
    :::

    你也可以在 Olares 的**文件管理器**里查看并删除该目录。

## 通过 Olares CLI 扩展系统存储

如果你的 Olares 系统使用基于 LVM 的存储方式，可以使用 `disk` 命令扩展系统存储容量。

手动挂载会将磁盘作为外部存储挂载到 `/olares/share` 目录下。相比之下，`disk extend` 命令用于扩展 Olares 的系统存储空间。扩展完成后，新增磁盘不再显示为独立挂载点。

:::warning 数据丢失警告 
`disk extend` 命令将销毁所选磁盘上的所有数据。 

在继续操作之前，请确保磁盘中没有重要数据或已完成备份。 
:::

### 开始之前

- 将外部硬盘连接到 Olares 主机。
- [SSH](/zh/developer/reference/access-olares-terminal.md) 登录到 Olares 主机终端。

### 识别未挂载的磁盘

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

### 扩展系统存储

1. 确认 Olares 已识别但未挂载该磁盘：

    ```bash
    olares-cli disk list-unmounted
    ```

2. 将检测到的未挂载磁盘加入系统存储：

    ```bash
    olares-cli disk extend
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

### 验证扩展结果

你可以在终端和 UI 界面中验证存储空间是否已增加：

#### 在终端

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

#### 在 UI 界面
从启动台打开仪表盘，确认系统总存储容量已增加。

![Check disk volume in Dashboard](/public/images/zh/manual/tutorials/expand-dashboard-disk.png#bordered)


如需查看完整用法与选项，请参考 [`disk`](/zh/developer/install/cli/disk.md)。