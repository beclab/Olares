---
outline: [2, 3]
description: 升级到 Olares 1.12.6 后，将 OnlyOffice 文档迁移到新的共享应用架构。
head:
  - - meta
    - name: keywords
      content: Olares, OnlyOffice, 迁移, 共享应用, Olares 1.12.6
app_version: "1.1.14"
doc_version: "1.0"
doc_updated: "2026-08-04"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/onlyoffice-migration.md)。
:::

# 将 OnlyOffice 迁移到新架构

Olares 1.12.6 更新了共享应用架构，因此无法直接更新已有的 OnlyOffice。本指南介绍如何安装新版本并迁移已有文档。

:::warning
本指南仅适用于升级到 Olares 1.12.6 前安装的 OnlyOffice。如果你是在升级后安装的，请参阅 [OnlyOffice 使用指南](onlyoffice.md)。
:::

## 迁移文档到新应用

1. 备份文档。

   a. 打开文件管理器，进入 **Documents**。

   b. 选择通过 OnlyOffice 上传的文档，将其下载到其他位置。

2. 卸载之前安装的 OnlyOffice 应用。出现提示时，不要勾选**同时删除所有本地数据**，以保留迁移所需的旧版应用数据。

3. 安装新版 OnlyOffice。

   a. 打开应用市场，搜索 "OnlyOffice"。

   b. 点击应用卡片，打开应用详情页。

   c. 查看 **Information** 面板。新版应用的 **Compatibility** 字段显示 `Olares >=1.12.6-0`。

   d. 点击 **Get**，然后点击 **Install**，等待安装完成。

4. 将文档移动到新位置。

   a. 打开文件管理器，进入 **Application** > **Data** > **onlyofficev3** > **documents**。

   b. 将备份的文档移动到此目录。

   c. 从启动台打开 OnlyOffice，确认文件已显示在首页。

现在可以在新版应用中使用原有文档了。

## 清理遗留数据

确认迁移成功且新版应用运行正常后，可以删除旧版数据文件夹以释放磁盘空间。

:::warning
删除旧版数据后无法恢复。继续前，请确认所有文档都能在新版应用中打开，并保留一份单独的备份。
:::

打开文件管理器，进入 **Application** > **Data** > **onlyofficev2**，删除该文件夹。

## 了解更多

- [使用 OnlyOffice 编辑文档](onlyoffice.md)：了解 Olares 应用包含的组件及 Web 客户端的使用方式。
- [共享应用](../manual/olares/market/shared-apps.md)：了解 Olares 共享应用的工作方式。
