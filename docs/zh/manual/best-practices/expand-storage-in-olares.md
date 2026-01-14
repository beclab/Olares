---
outline: [2, 3]
description: Olares 存储扩展指南，涵盖 SMB 服务器连接、USB 自动挂载与 HDD/SSD 手动挂载设置，助你灵活扩展本地存储空间，高效管理大型 AI 模型与数据资源。
---
# 在 Olares 中扩展存储空间

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

- 断开 USB 设备时，系统会自动将其卸载。

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

### 列出未挂载的磁盘

运行以下命令列出已连接但尚未挂载的磁盘：

```bash
olares-cli disk list-unmounted
```

### 扩展系统存储

运行以下命令将检测到的未挂载磁盘纳入系统存储：

```bash
olares-cli disk extend
```

如需查看完整用法与选项，请参考 [`disk`](/zh/developer/install/cli/disk.md)。