---
outline: [2, 3]
description: 了解如何在 Olares One 上手动挂载外接 SSD，用于临时或永久存储扩展。
head:
  - - meta
    - name: keywords
      content: Olares One, 外接 SSD, 扩展存储
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../one/expand-storage-external-ssd.md)为准。
:::

# 通过外接 SSD 扩展存储 <Badge type="tip" text="30 min" />

你可以手动将大容量外接 SSD 挂载到 Olares One 的特定系统路径上。

这种方式适用于长期存储扩展，例如下载更多或更大的本地 AI 模型。

:::warning HDD 支持
本指南适用于 SSD。机械硬盘 (HDD) 尚未在 Olares One 上测试。
:::
:::info 挂载路径
目前仅支持挂载到 `/olares/share` 目录下。

挂载灵活性将在未来版本中改进。
:::

## 前提条件
**硬件**
- Olares One 已完成设置并正在运行。
- 外接 SSD 已连接到 Olares One。

**SSH 访问**
- [SSH 访问 Olares One](access-terminal-ssh.md)。

**经验**
- 熟悉基本的终端命令和命令行界面 (CLI)。

## 步骤 1：识别硬盘

1. 通过 SSH 或 Control Hub 连接到 Olares One 终端。

2. 运行以下命令查看已检测到的硬盘：

   ```bash
   sudo fdisk -l
   ```

3. 从输出中识别你的目标硬盘。每块硬盘在 **Device** 列下会列出其分区，如 `/dev/nvme1n1p1`、`/dev/nvme1n1p2` 或 `/dev/sdb1`。

   ![分区列表](/images/manual/tutorials/expand-storage-partition.png#bordered)

4. 记下你打算挂载的目标分区。例如：`/dev/nvme1n1p1`。

## 步骤 2：挂载分区
### 选项 A：临时挂载分区

临时挂载适用于一次性任务，如数据迁移。设备重启后配置将丢失。

1. 创建挂载点目录：

   ```bash
   sudo mkdir -p /olares/share/<directory_name>
   ```

   将 `<directory_name>` 替换为自定义名称。

2. 将分区挂载到该目录：

   ```bash
   sudo mount /dev/<partition> /olares/share/<directory_name>    
   ```

   例如：

   ```bash
   sudo mount /dev/nvme1n1p1 /olares/share/hdd0
   ```

3. 在 Files 中导航到 **External** 目录以验证挂载。你应该能看到新的文件夹内容。

   ![检查挂载结果](/images/manual/tutorials/expand-storage-mount-result-en.png#bordered)

### 选项 B：永久挂载分区
对于长期使用，你需要通过 `/etc/fstab` 文件配置系统开机时自动挂载该硬盘。

1. 获取 UUID。
   :::tip 使用 UUID 识别设备
   使用 UUID 比使用设备名称（如 `/dev/sdb1`）更安全，因为设备名称在你将硬盘插入不同端口时可能会改变。
   :::
   a. 运行以下命令：
    ```bash
    lsblk -f
    ```
   b. 记下以下信息：
   - **FSTYPE**：文件系统类型（如 `ext4`、`xfs`）。
   - **UUID**：分区的唯一标识符。

   ![检查挂载结果](/images/manual/tutorials/expand-storage-fstype.png#bordered)

2. 创建挂载目录：

   ```bash
   sudo mkdir -p /olares/share/<directory_name>
   ```

   将 `<directory_name>` 替换为自定义名称。

3. 打开配置文件。

   ```bash
   sudo vi /etc/fstab
   ```

4. 添加挂载条目。在文件末尾使用以下格式添加新行：

   ```bash
   UUID=<UUID> /olares/share/<directory_name> <FSTYPE> defaults,nofail 0 0
   ```

   例如：

   ```bash
   UUID=1234-ABCD /olares/share/my_disk ext4 defaults,nofail 0 0
   ```

5. 按 `Esc`，输入 `:wq`，然后按 **Enter** 保存更改并退出编辑器。

6. 运行以下命令验证配置。

   ```bash
   mount -a
   ```
   :::tip 防止启动失败
   错误的 `/etc/fstab` 配置可能导致系统无法启动。

   强烈建议在重启前先运行 `mount -a` 来验证配置。
   :::
   如果没有出现错误，则设置成功。

7. 重启后，确认硬盘已在 **External** 目录中自动挂载。

## 步骤 3：卸载分区
:::warning 不可逆操作
卸载前确保没有程序或终端正在访问该目录。
:::

要安全移除硬盘或删除挂载点配置：
1. 卸载分区：

   ```bash
   sudo umount /olares/share/<directory_name>
   ```

2. （可选）如果不再需要该文件夹，删除目录：

   ```bash
   rm -rf /olares/share/<directory_name>
   ```

## 相关资源
- [在 Olares 中管理文件](../manual/olares/files/index.md)
- [通过 USB 设备扩展存储](expand-storage-usb-drive.md)
- [使用内置 SSD 扩展存储](expand-storage-internal-ssd.md)
