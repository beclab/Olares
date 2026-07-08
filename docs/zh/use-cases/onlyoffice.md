---
outline: [2, 3]
description: 了解如何在升级到 Olares 1.12.6 后，将 OnlyOffice 从旧架构迁移到新的共享应用架构。
head:
  - - meta
    - name: keywords
      content: Olares, OnlyOffice, 迁移, 共享应用, Olares 1.12.6
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-07-08"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/onlyoffice.md)。
:::

# 将 OnlyOffice 迁移到新架构

OnlyOffice 是 Olares 上用于文档编辑与协作的共享应用。Olares 1.12.6 更新了共享应用架构，因此你无法直接更新 OnlyOffice。本指南介绍如何在升级到 Olares 1.12.6 后，将文档迁移到新版 OnlyOffice。

## 开始之前

OnlyOffice 目前仅包含 Document Server。Web 界面是官方的 Node.js 演示客户端 `onlyofficeclient`，仅支持文档上传和在线编辑，暂不支持多人实时协作以及完整的账号和文档管理系统。

## 迁移文档到新应用

1. 备份文档。

   a. 打开**文件管理器**，进入 **Documents**。
   
   b. 选择你通过 OnlyOffice 上传的文档，将其下载到其他位置。

2. 卸载之前安装的 OnlyOffice 应用。在提示时：

   - 不要勾选**同时删除所有本地数据**。
   - 勾选**同时卸载共享服务器（影响所有用户）**。

3. 安装新版 OnlyOffice 应用。

   a. 打开应用市场，搜索 "OnlyOffice"。
   
   b. 点击应用卡片，打开应用详情页。
   
   c. 查看 **Information** 面板。**Compatibility** 字段显示 `Olares >=1.12.6-0` 即为新版本。
   
   d. 点击 **Get**，然后点击 **Install**，等待安装完成。

4. 将文档移动到新的位置。

   a. 打开**文件管理器**，进入 **Application** > **Data** > **onlyofficev3** > **documents**。
   
   b. 将备份的文档移动到这个目录。
   
   c. 从启动台打开 OnlyOffice，确认文件已在首页显示。

你的 OnlyOffice 文档已迁移到新版应用。

## 了解更多

- [共享应用](../manual/olares/market/shared-apps.md)：了解新的共享应用架构。
