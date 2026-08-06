---
outline: [2, 3]
description: 从 Olares Space 云端备份中恢复文件，包括查看备份用量和快照。
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, 备份恢复, 系统快照, 数据恢复, 云端备份
---
# 从 Olares Space 备份中恢复数据

使用 Olares Space 查看云端备份并将文件恢复到 Olares。

## 查看备份存储用量

检查备份用量与配额的对比情况。

:::info
对于自托管 Olares 用户，请重点关注备份服务的**存储用量**。这些服务可能会根据使用情况产生费用。更多详情，请见[计费说明](billing.md)。
:::

1. 点击左侧导航栏的 **Backup**。

  ![Olares Space 备份列表](/images/how-to/space/backup_list.png#bordered)

2. 找到 **Backup Usage** 区域。它展示所有备份占用的总存储空间及配额。

## 查看备份详情

查看备份任务的快照，了解已备份的内容。

1. 点击左侧导航栏的 **Backup**。
2. 找到 **Backup List** 区域。你可以在表格中看到所有备份任务。
3. 要查看某个备份任务的快照：

    a. 点击该备份任务的 **View Details**。

    b. 查看备份信息和快照。

    ![Olares Space 快照列表](/images/how-to/space/snapshots_list.png#bordered)

## 从备份恢复

将备份文件恢复到 Olares。

:::info
使用备份和恢复服务前，请了解存储和带宽费用。每个实例包含一定量的免费流量，超出配额部分将产生费用。更多详情，请见[计费说明](billing.md)。
:::

1. 选择要恢复的快照：

   - 要恢复最新快照，点击备份详情页右上角的 **Restore**。
   - 要恢复特定日期的快照，在 **Snapshots** 表格中找到该快照，然后点击该行右侧的 **Restore**。

2. 在 **Restore** 对话框中：

    a. 复制 **Backup Url**。

      ![Olares Space 恢复对话框](/images/how-to/space/backup_restore_dialog.png#bordered){width=70%}

    b. 点击 **Settings > Restore** 链接。这会在新的浏览器标签页中打开 Restore 页面。你也可以直接在 Olares 中导航到 **Settings > Restore**。

3. 在 **Restore** 页面，添加一个恢复任务：

   - 如果页面为空，点击 **Add restore task**，然后选择 **From Olares Space**。
   - 如果页面已有恢复任务，点击右上角的 <i class="material-symbols-outlined">add</i>，然后选择 **From Olares Space**。 
 
  ![添加恢复任务选项](/images/how-to/space/restore_add_task.png#bordered){width=70%}

4. 填写恢复信息：

    a. **Backup URL**：粘贴复制的 **Backup Url**。

    b. **Restore password**：输入创建备份任务时设置的密码。

    c. **Restore location**：选择要恢复到的目录。

    d. **New folder name**：输入存放恢复文件的新文件夹名称。

    ![从 Olares Space 恢复表单](/images/how-to/space/restore_from_olares_space.png#bordered){width=70%}

5. 点击 **Start restore**。
6. 恢复完成后，查看并访问恢复的文件：

    a. 在 **Restore** 页面，点击恢复任务卡片进入恢复详情。

      ![恢复任务完成](/images/how-to/space/restore_complete.png#bordered){width=70%}

    b. 在 **Restore details** 页面，点击 **Open in Files**。文件管理器会打开包含恢复文件的文件夹。

      ![恢复详情页面](/images/how-to/space/restore_details.png#bordered){width=70%}

## 资源

- [备份 Olares 数据](../olares/settings/backup.md)：了解如何在 Olares 中创建备份任务。
