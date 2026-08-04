---
outline: deep
description: 学习如何在 Olares 中挂载 NAS 共享文件夹，并将照片导入 Immich 作为外部库。
head:
  - - meta
    - name: keywords
      content: Olares, Immich, photo backup, self-hosted photos, photo management, face recognition, smart search
app_version: "1.0.15"
doc_version: "1.0"
doc_updated: "2026-04-09"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/immich-import-from-nas.md)。
:::

# 将 NAS 中的照片导入 Immich

如果你有存储在 NAS 设备上的照片，可以在 Olares 中挂载 NAS 共享文件夹，并将其导入 Immich 作为外部库。

:::info
本教程以 Synology NAS 为例。其他 NAS 品牌的过程可能有所不同。
:::

## Prerequisites

- Immich 已更新至 V1.0.15 或更高版本。
- Olares 设备和 NAS 在同一本地网络上。
- 共享文件夹配置：
  - 共享文件夹配置为允许通过本地网络进行读/写访问（SMB）。
  - 共享文件夹的**在"网络邻居"中隐藏此共享文件夹**选项未选中。

## 步骤 1：将 NAS 共享文件夹挂载到 Olares

1. 打开 Files，点击 **External**，然后点击 **Connect to server**。

   ![在 Files 中连接到服务器](/images/manual/use-cases/immich-connect-server.png#bordered)

2. 在 **Connect to server** 窗口中：

    a. 以 SMB 格式输入 NAS IP 地址（例如 `//192.168.50.156/`），然后点击 **Confirm**。

    b. 输入你的 NAS 用户名和密码，然后点击 **Confirm**。

      ![在 Files 中连接到服务器](/images/manual/use-cases/immich-nas-login.png#bordered){width=60%}

    c. 选择要挂载的文件夹（本例中为 `/CZ-test`），然后点击 **Confirm**。

      ![在 Files 中连接到服务器](/images/manual/use-cases/immich-nas-select-share.png#bordered){width=60%}

      连接后，共享文件夹将出现在 **External** 目录中。

      ![NAS 挂载到 Files](/images/manual/use-cases/immich-nas-mounted.png#bordered)

## 步骤 2：将文件夹添加到 Immich 外部库

1. 打开 Immich，点击右上角的用户头像，然后选择 **Administration**。
2. 点击左侧边栏中的 **External Libraries**。
3. 创建一个新库或使用现有库。
4. 在 **Folders** 区域中，点击 **Add**。
5. 输入导入路径。路径格式为 `/external_storage/` 后跟你挂载的目录名称。在本例中，它是：

   ```text
   /external_storage/CZ-test
   ```
   ![NAS 挂载到 Files](/images/manual/use-cases/immich-add-nas-folder.png#bordered)

6. 点击 **Add**。
7. 点击右上角的 **Scan** 开始扫描。

    扫描完成后，NAS 中的照片将出现在照片时间线中。

    :::tip 扫描大型文件夹
    如果文件夹包含大量文件，扫描可能需要一段时间，并消耗大量 NAS 磁盘 I/O。你可以前往 **Administration** > **Job Queues** 并暂停其中的一些任务以加快处理速度。
    :::
