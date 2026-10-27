---
description: 了解如何在 Olares 中挂载并访问各类云存储服务。
head:
  - - meta
    - name: keywords
      content: Olares, 文件管理器, 云存储, Google Drive, AWS S3, Dropbox, 挂载云盘
---

# 挂载与使用云存储

你可以将 Google Drive、Dropbox、AWS S3、腾讯云 COS 等云存储挂载到 Olares，并在**文件管理器**应用中直接访问和管理云端文件。

![云存储](/images/zh/manual/olares/files-cloud.png)

## 连接云存储服务

云存储通过**集成**功能连接，可在 LarePass 手机端或 Olares**设置**中完成，具体步骤取决于服务类型。

### 通过 OAuth 连接（Google Drive、Dropbox）

基于 OAuth 登录的服务需在 LarePass 手机端完成授权：

1. 在手机上打开 LarePass。
2. 进入**设置** > **LarePass 设置** > **集成**，点击右上角 <i class="material-symbols-outlined">add</i>。
3. 选择 **Google Drive** 或 **Dropbox**。
4. 按提示登录并授权。

### 通过 API 密钥连接（AWS S3、腾讯云 COS）

AWS S3、腾讯云 COS 等服务需在 Olares**设置**中使用 Access Key & Secret Key 手动配置：

1. 从 Dock 或启动台打开**设置**，进入**集成** > **关联您的账户与数据**。
2. 点击右上角**添加账户**。
3. 选择 **AWS S3** 或 **Tencent COS**，点击**确认**。
4. 在弹出的对话框中输入 Access Key、Secret Key、Region 和 Bucket name。
5. 点击**下一步**。凭证验证通过后将显示成功提示。

你也可以在 LarePass 中添加此类服务：进入**设置** > **LarePass 设置** > **集成**，点击右上角 <i class="material-symbols-outlined">add</i>，选择服务并输入凭证。

:::tip 需要在 LarePass 中操作的集成
OAuth 类型的集成以及 Olares Space 需在 **LarePass** 应用中完成连接。
:::

连接成功后，云存储将出现在文件管理器的**云存储**目录下。

## 访问云存储

挂载后，你可以像使用本地存储一样访问和管理云端文件：

- **上传 / 下载**文件
- **预览**支持的文件类型
- **重命名**、**移动**或**删除**文件和文件夹

你在文件管理器中的操作将自动同步至对应的云存储服务。

## 卸载云存储

卸载云存储即移除对应的集成服务：

- **在 LarePass 中**：进入**设置** > **LarePass 设置** > **集成**，点击要移除的集成，点击右上角 <i class="material-symbols-outlined">more_horiz</i>，选择**删除**。
- **在 Olares 设置中**：进入**集成** > **关联您的账户与数据**，点击集成卡片，在**账户设置**页面点击**删除**。
