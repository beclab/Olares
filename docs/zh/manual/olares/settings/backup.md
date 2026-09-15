---
description: 了解如何在 Olares 中备份和恢复文件与应用，支持本地存储、Olares Space 以及 AWS S3、腾讯云 COS 等云服务。
head:
  - - meta
    - name: keywords
      content: Olares, 备份, 恢复, 增量备份, 快照, 备份计划, Olares Space, AWS S3, 腾讯云 COS
---

# 备份与恢复 Olares 数据

Olares 提供灵活的备份方案，可对**指定文件夹**与 **Wise 应用**执行全量和增量备份，并支持设置自动备份计划。你可以将数据备份到本地或网络存储，并从本地路径、Olares Space 或 AWS S3、腾讯云 COS 等云服务的快照中恢复文件或应用数据。

## 备份数据

### 添加备份任务

添加备份任务步骤如下：

1. 进入**设置** › **备份**，点击**添加备份**。
2. 选择**备份文件**或**备份应用**。
3. 在**添加备份任务**页面，配置备份基础信息如下：

    | 配置项 | 说明 |
    |:--|:--|
    | **备份位置** | - **本地路径**：建议选择外接设备，例如 U 盘、SMB 网络目录或外部硬盘。<br> - **网络存储**：支持 Olares Space、AWS S3、腾讯云 COS。在弹窗中点击**添加账户**即可添加存储账户，详见[连接云存储](../files/mount-cloud-storage.md)。 |
    | **Region**（仅限 Olares Space 存储） | 如果备份位置是 Olares Space，系统会自动选择对应的存储区域。 |
    | **备份路径**（备份文件时显示） | 指定要备份的文件夹路径。 |
    | **选择应用**（备份应用时显示） | 从下拉菜单中选择需要备份的应用，目前仅支持 Wise。 |
    | **备份名称** | 输入便于识别的任务名称，建议包含用途及时间戳。 |

    :::warning 注意  
    请在开始备份前确保以下事项：

    - 所选存储空间的可用容量足以存储该任务的备份数据。
    - 已成功订阅对应存储服务且账户处于正常状态。
    - 对目标存储路径拥有读写权限（例如挂载的 SMB 网络目录等）。
    :::

4. 设置备份计划与安全项目：
   - **快照频率**：选择备份执行的频率，支持按**日/周/月**备份。
   - **快照时间**：指定具体运行时间。
   - **备份密码**：为备份文件加密，保护隐私。
5. 点击 **提交**。系统首次将执行全量备份，后续按计划进行增量备份。

    :::warning 注意  
    备份过程中请勿关闭主机或重启备份服务。  
    :::

### 管理备份任务

创建成功后，备份任务将显示在备份首页。点击任务右侧 **>** 操作按钮可进入详情页进行管理：

| 操作 | 说明 |
|:--|:--|
| **管理** | - **编辑**：修改快照频率与执行时间<br>- **暂停**：暂停执行备份任务 <br>- **删除**：删除任务及其所有快照 |
| **立即快照** | 立刻手动执行一次备份 |

### 查看快照

在任务详情页底部，查看每次快照的**时间**与**状态**。点击右侧 **>** 操作按钮，可查看以下快照详情信息。

- **创建时间**：该快照的执行时间。
- **大小**：快照占用空间大小。
- **状态**：快照执行状态。
- **备份类型**：显示该快照是全量备份还是增量备份。

## 从备份中恢复数据

你可以通过已有的备份快照，将文件恢复至指定目录，或恢复应用数据。目前应用恢复仅支持 **Wise**。

### 从本地路径恢复

1. 进入**设置** › **还原**，点击**添加恢复**。
2. 选择**从本地路径恢复**。
3. 选择本地备份路径。请确保路径选择至备份任务目录层级，例如，如果备份位置为 `/documents`，备份名称为 `demo`，请确保选择路径为 `/documents/olares-backups/demo-xxxx`。
4. 输入备份密码。
5. 点击**查询快照**，获取可用快照列表。
6. 在目标快照后点击**恢复**以加载快照。
7. 如果是恢复**文件**，请指定恢复位置和目标文件夹，然后点击**开始恢复**。  
   如果是恢复 **Wise 应用**，无需指定路径，直接点击**开始恢复**即可。

### 从 Olares Space 恢复

:::info
使用备份和恢复服务前，请了解存储和带宽费用。每个实例包含一定量的免费流量，超出配额部分将产生费用。对于自托管 Olares 用户，请重点关注备份的存储用量。更多详情，请见[计费说明](../../space/billing.md)。
:::

1. 使用 LarePass 应用扫码登录 [Olares Space](https://www.olares.com/space)。
2. 点击左侧导航栏的 **Backup**。**Backup Usage** 区域显示所有备份占用的总存储空间及配额。

   ![Olares Space 备份列表](/images/how-to/space/backup_list.png#bordered)

3. 找到目标备份任务，点击 **View Details** 查看其快照。

   ![Olares Space 快照列表](/images/how-to/space/snapshots_list.png#bordered)

4. 选择要恢复的快照：
   - 要恢复最新快照，点击备份详情页右上角的 **Restore**。
   - 要恢复特定日期的快照，在 **Snapshots** 表格中找到该快照，然后点击该行右侧的 **Restore**。
5. 在 **Restore** 对话框中，复制 **Backup Url**。

   ![Olares Space 恢复对话框](/images/how-to/space/backup_restore_dialog.png#bordered){width=70%}

6. 切换到 Olares。进入**设置** › **还原**，添加恢复任务：
   - 如果页面为空，点击**添加恢复任务**，然后选择**从 Olares Space 链接恢复**。
   - 如果页面已有恢复任务，点击右上角的 <i class="material-symbols-outlined">add</i>，然后选择**从 Olares Space 链接恢复**。

   ![添加恢复任务选项](/images/how-to/space/restore_add_task.png#bordered){width=70%}

7. 填写恢复信息：
   - **Backup URL**：粘贴复制的 **Backup Url**。
   - **Restore password**：输入创建备份任务时设置的密码。
   - **Restore location**：选择要恢复到的目录。
   - **New folder name**：输入存放恢复文件的新文件夹名称。

   ![从 Olares Space 恢复表单](/images/how-to/space/restore_from_olares_space.png#bordered){width=70%}

8. 点击 **Start restore**。
9. 恢复完成后，在 **Restore** 页面点击恢复任务卡片，进入恢复详情。

   ![恢复任务完成](/images/how-to/space/restore_complete.png#bordered){width=70%}

10. 在 **Restore details** 页面，点击 **Open in Files**。文件管理器会打开包含恢复文件的文件夹。

    ![恢复详情页面](/images/how-to/space/restore_details.png#bordered){width=70%}

### 从 AWS S3 或腾讯云 COS 恢复

1. 获取备份文件夹链接：
   - **AWS S3**：访问 [AWS S3 控制台](https://console.aws.amazon.com/s3)，进入存储桶，在 `olares-backups` 目录中找到目标备份文件夹。选中后生成该文件夹的**预签名链接**。操作方法可参考 [AWS 文档](https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/userguide/ShareObjectPreSignedURL.html)。
   - **腾讯云 COS**：访问[腾讯云 COS 控制台](https://console.cloud.tencent.com/cos) > **存储桶** > **文件列表**，在 `olares-backups` 目录中找到目标备份文件夹。点击文件夹右侧**详情**，在底部复制**临时访问链接**。操作方法可参考[腾讯云文档](https://cloud.tencent.com/document/product/436/68284)。
2. 进入**设置** › **还原**，点击**添加恢复**，选择对应的云服务恢复方式。
3. 将获取的链接粘贴至 **Backup URL** 输入框。
4. 输入备份密码。
5. 点击**查询快照**，获取可用快照列表。
6. 选取目标快照并点击**恢复**以加载快照。
7. 如果是恢复**文件**，请指定恢复位置和目标文件夹，然后点击**开始恢复**。  
   如果是恢复 **Wise 应用**，无需指定路径，直接点击**开始恢复**即可。

### 管理恢复任务

还原任务创建成功后将显示在还原页面的任务列表中。点击右侧 **>** 按钮进入可查看详细状态。可执行的操作包括：

- **取消还原任务**：在还原过程中点击**取消**，以中断并取消还原任务。
- **查看文件或应用**：还原完成后，可直接点击**打开应用**或**打开文件夹**。
