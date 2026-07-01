---
outline: [2, 3] 
description: 了解如何在 Olares One 上备份和恢复文件及应用数据。
head:
  - - meta
    - name: keywords
      content: 备份, 恢复
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/backup-restore.md)。
:::

# 备份和恢复数据 <Badge type="tip" text="15 min" />

Olares One 提供内置的备份功能来保护重要文件和应用数据。你可以创建完整备份和增量备份，将它们存储在本地或网络存储上，并在需要时从任何可用的快照恢复数据。

## 学习目标

完成本教程后，你将学会如何：
- 为文件夹和受支持的应用创建备份任务。
- 配置备份位置、计划和密码保护。
- 管理现有的备份任务。
- 将文件恢复到特定目录或从快照恢复应用数据。

## 备份你的数据

备份任务定义了要备份什么、存储在哪里以及何时运行。

### 创建备份任务

1. 进入 **Settings** > **Backup**。
2. 点击 **Add backup task**。如果提示，选择 **Back up files** 或 **Back up apps** 继续。
3. 在 Add backup task 页面，配置以下设置：

    | 字段                  | 说明                                                                                                                                                                                                                                                                                                                                                                                          |
    |------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
    | **Backup location**    | <ul><li>**Local path**：选择外部设备，如 USB 驱动器、SMB <br>共享或外置硬盘。</li><br><li>**Olares Space**：通过 LarePass 扫描 **Settings** > 你的头像 > **Olares<br> Space** 中的二维码。<br></li><li> **AWS S3 或 Tencent COS**：在对话框中点击 **Add account**，或进入 <br>**Settings** > **Integrations** > **Link your accounts & data**。</li></ul> |
    | **Region**             | 仅云存储。选择你的存储桶所在区域。                                                                                                                                                                                                                                                                                                                                                |
    | **Backup path**        | 仅文件备份。浏览并选择要备份的特定目录。                                                                                                                                                                                                                                                                                                                               |
    | **Select application** | 仅应用备份。选择要备份的应用。目前仅支持 Wise。                                                                                                                                                                                                                                                                                                                   |
    | **Backup name**        | 输入一个可识别的任务名称。建议包含用途和时间戳。                                                                                                                                                                                                                                                                                                                                              |
    | **Snapshot frequency** | 选择 **Every day**、**Every week** 或 **Every month**。                                                                                                                                                                                                                                                                                                                                            |
    | **Run backup at**      | 设置备份运行的具体时间。                                                                                                                                                                                                                                                                                                                                                  |
    | **Backup password**    | 设置密码以加密你的快照。                                                                                                                                                                                                                                                                                                                                                            |
    | **Confirm password**   | 重新输入你设置的密码。                                                                                                                                                                                                                                                                                                                                                                       |
    
    ![Add backup task](/images/one/settings-add-backup.png#bordered)
4. 点击 **Submit** 创建并启动任务。
    - 首次运行将是完整备份。
    - 后续运行将是增量备份（仅保存更改）。

### 管理备份任务

创建后，任务将显示在列表中。点击任务旁边的 <i class="material-symbols-outlined">chevron_right</i> 箭头查看详情。

| 操作           | 说明                                                                                                                                                                     |
|------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Manage**       | <ul><li> **Edit**：修改快照频率和时间。<br></li><li> **Pause**：暂停任务。<br></li><li> **Delete**：删除任务及所有关联的快照。</li></ul> |
| **Snapshot now** | 立即运行备份。                                                                                                                                                       |

![管理备份任务](/images/one/settings-manage-backup-task.png#bordered){width=85% style="display:block;margin-left:0;margin-right:auto;"}

### 查看快照记录

在备份详情页面底部，你可以查看此任务的快照：

| 字段             | 说明                                 |
|-------------------|---------------------------------------------|
| **Creation time** | 快照创建时间。              |
| **Size**          | 快照数据的总大小。            |
| **Status**        | 快照的执行状态。       |
| **Backup type**   | 是完整备份还是增量备份。 |

![查看快照](/images/one/settings-snapshots.png#bordered){width=70% style="display:block;margin-left:0;margin-right:auto;"}

## 恢复数据

你可以将文件恢复到特定文件夹，或使用任何有效的快照恢复应用数据。

### 创建恢复任务

1. 进入 **Settings** > **Restore**。
2. 点击 **Add restore task**。
3. 选择与你备份存储位置匹配的方法。

### 从本地路径恢复

此方法用于恢复存储在连接到 Olares 的 USB 驱动器或 SMB 共享上的备份。

1. 选择本地备份路径。路径必须指向特定的备份任务文件夹。
   
    如果备份名称为 `demo`，备份位置为 `/documents`，则路径应为：
    ```text
    /documents/olares-backups/demo-xxxx
    ```
2. 输入你的备份密码。
3. 点击 **Query snapshots** 加载可用的快照。
4. 点击目标快照旁边的 **Restore**。
5. 开始恢复：
    - 文件：选择恢复位置，然后点击 **Start restore**。
    - Wise：直接点击 **Start restore**（无需目标路径）。

### 从 Olares Space 恢复

如果你备份到了 Olares Space，请使用此方法。你需要 LarePass 移动应用。

1. 打开 LarePass 并扫描登录 [Olares Space](https://space.olares.com)。
2. 在 **Backup** 页面，找到备份任务并点击 **View Details**。
3. 获取快照 URL：
   - 点击右上角 **Restore** 获取最新快照 URL，或
   - 选择特定快照并点击旁边的 **Restore**。
4. 复制 URL 并粘贴到 From Space URL 页面的 **Backup URL** 字段中。
5. 输入你的备份密码。
6. 开始恢复：
    - 文件：选择恢复位置和目标文件夹，然后点击 **Start restore**。
    - Wise：直接点击 **Start restore**。

### 从云存储恢复（AWS S3 / Tencent COS）

此方法用于恢复存储在 AWS S3 或 Tencent COS 上的备份。

1. 打开你的云存储控制台并找到 `olares-backups` 目录。
2. 选择目标备份文件夹并生成一个**预签名 URL**。
    - 对于 AWS S3，请参阅 [AWS S3 文档](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ShareObjectPreSignedURL.html)了解预签名 URL。
    - 对于 Tencent COS，请按照[腾讯云文档](https://cloud.tencent.com/document/product/436/68284)生成临时访问 URL。
3. 复制生成的 URL。
4. 将其粘贴到 From AWS S3 URL 或 From Tencent COS URL 页面的 **Backup URL** 字段中。
5. 输入你的备份密码。
6. 点击 **Query snapshots** 加载可用的快照。
7. 点击目标快照旁边的 **Restore**。
8. 开始恢复：
    - 文件：选择恢复位置和目标文件夹，然后点击 **Start restore**。
    - Wise：直接点击 **Start restore**。

## 监控恢复任务

创建后，恢复任务将显示在 Restore 页面的任务列表中。点击任务旁边的 <i class="material-symbols-outlined">chevron_right</i> 箭头查看详情并管理恢复任务。

- Cancel：在运行过程中停止进程。
- View data：当状态显示为 `Completed` 时，点击 **Open folder** 或 **Open app** 访问恢复的数据。
