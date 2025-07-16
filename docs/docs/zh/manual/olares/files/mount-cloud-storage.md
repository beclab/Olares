---
description: 了解如何在 Olares 中挂载并访问各类云存储服务。
---

# 挂载与使用云存储

你可以通过 Olares 的**集成**功能轻松挂载云存储服务，并在**文件管理器**应用中直接访问和管理云端文件。

![云存储](/images/zh/manual/olares/files-cloud.png)

## 挂载云存储

要挂载云存储，请在 **LarePass** 或 Olares 的**设置**中连接对应服务：

- **基于 OAuth 的存储服务**：如 Google Drive 和 Dropbox。通过 [LarePass 手机端](../../larepass/integrations.md#通过-oauth-添加云盘)添加对应的集成。
- **基于 API 凭证的存储服务**：如 AWS S3 或腾讯云对象存储（COS）。可通过 [LarePass 手机端](../../larepass/integrations.md#通过-api-密钥添加云盘)或 [Olares 设置](../settings/integrations.md#通过-api-密钥添加云对象存储)中添加对应的集成。

连接成功后，云存储将自动挂载至**文件管理器**应用中的**云存储**目录下。

## 访问云存储

挂载后，你可以像使用本地存储一样访问和管理云端文件：

- **上传 / 下载**文件
- **预览**支持的文件类型
- **重命名**、**移动**或**删除**文件和文件夹

你在文件管理器中的操作将自动同步至对应的云存储服务。

## 卸载云存储

若需取消挂载，可移除对应的集成服务：

- [在 LarePass 中移除集成](../../larepass/integrations.md#断开集成)
- [在 Olares 设置中移除集成](../settings/integrations.md#查看与管理现有集成)

