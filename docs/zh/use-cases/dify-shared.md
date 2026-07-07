---
outline: [2, 3]
description: 了解如何在 Olares 1.12.6 中将 Dify Shared 从 v2 架构迁移到新的共享应用架构。
head:
  - - meta
    - name: keywords
      content: Olares, Dify Shared, 迁移, 共享应用, Olares 1.12.6
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-07-07"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/dify-shared.md)。
:::

# 将 Dify Shared 迁移到新架构

Dify Shared 是 Olares 上用于构建 AI 应用、知识库和智能体的共享应用。Olares 1.12.6 更新了共享应用架构，因此你无法直接更新 Dify Shared。本指南介绍如何将数据迁移到新版 Dify Shared。

## 迁移前须知

:::warning 无法保留的数据
迁移后，以下数据无法保留：

- **用户账号**：你需要重新创建所有账号。
- **模型 API Key 和 Dify 系统设置**：你需要重新配置。
- **已处理的知识库 Chunk**：你需要重建所有知识库。
- **应用日志和对话历史**：仅可导入应用配置。
:::

如果多个用户共用 Dify 实例，请确保每位用户在卸载 v2 应用前都已导出自己的数据。

## 导出 Dify 数据

1. 导出应用配置。

   a. 打开 Dify 工作室。
   
   b. 点击应用卡片右下角的扩展按钮，选择**导出 DSL**。
   
   c. 为每个应用保存 `.yml` 文件。

2. 下载知识库中的文档。

   a. 打开文件管理器，进入**缓存** > `<你的设备名称>` > `difyv2` > `volumes` > `app` > `storage` > `upload_files`。
   
   b. 右键文件并下载。
   
   :::tip 根据上传时间识别文件
   该文件夹中的文件名是 Dify 系统内部名称。请根据上传时间、文件格式和大小来识别所需文件。
   :::
   
   如果你的知识库使用外部数据源、Notion 或网站，或者你已备份原始文档，可跳过此步骤。

3. 手动记录模型配置和系统设置。

## 卸载 Dify Shared

打开 Market，进入 **My Olares**，卸载 v2 版 **Dify Shared**。

## 安装新版 Dify Shared

1. 打开 Market，搜索 **Dify**。
2. 点击 **Get**，然后点击 **Install**，等待安装完成。
3. 在应用详情页，查看**信息** > **兼容性**。如果显示为 `Olares >=1.12.6-0`，说明你安装的是新版本。

:::tip 首次启动耗时较长
Dify Shared 首次启动大约需要 10 分钟。请等待设置页面出现后再打开应用。
:::

4. 打开 Dify 并创建管理员账号。

## 导入应用

1. 打开 Dify 工作室，选择**导入 DSL**。
2. 上传之前导出的 `.yml` 文件。
3. 如果 Dify 检测到缺少插件，请安装它们。如果你跳过插件安装或安装失败，应用配置仍然会被导入，但未成功安装的插件节点需要重新配置。
4. 如果应用配置了知识库，请先重建知识库，然后在应用中重新配置。

## 重建知识库

在新版 Dify Shared 中新建知识库，并重新上传文档或重新连接数据源。

## 了解更多

- [共享应用](../manual/olares/market/shared-apps.md)：了解新的共享应用架构。
