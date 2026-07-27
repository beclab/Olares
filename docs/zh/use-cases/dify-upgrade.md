---
outline: [2, 3]
description: 了解如何在 Olares 1.12.6 中将 Dify 从 v2 架构迁移到新的共享应用架构。
head:
  - - meta
    - name: keywords
      content: Olares, Dify, 迁移, 共享应用, Olares 1.12.6
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-07-08"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/dify-upgrade.md)。
:::

# 将 Dify 迁移到新架构

Dify 是 Olares 上用于构建 AI 应用、知识库和智能体的共享应用。Olares 1.12.6 更新了共享应用架构，因此你无法直接更新 Dify。本指南介绍如何将数据迁移到新版 Dify。

## 迁移前须知

知识库依赖于由特定 Dify 版本生成的向量索引。这些索引无法跨版本复用，而新版 Dify 使用了更高效的检索引擎。因此，你无法直接恢复旧应用数据，也无法复用旧的知识库索引。

你必须手动迁移数据：
- 将应用配置导出为 DSL 文件。日志和对话等运行时数据不会被保留。
- 导出知识库中的原始文档。已处理的知识库 Chunk 不会被保留。

安装新版 Dify 应用后，还需要：
- 重新创建用户账号和权限。
- 重新输入模型 API Key 和系统偏好设置。
- 重新上传文档以重建每个知识库。

## 导出 Dify 数据

如果多个用户共用 Dify 实例，请确保每位用户在卸载 v2 应用前都已导出自己的数据。

1. 导出应用配置。

   a. 打开 Dify，进入 **Studio** 页面。
   
   b. 将鼠标悬停在目标应用卡片上，点击 <i class="material-symbols-outlined">more_horiz</i>，然后选择 **Export DSL**。`.yml` 文件会自动下载。
   
   c. 对每个要迁移的应用重复上述操作。

2. 下载知识库中的文档。

   如果你的知识库使用外部数据源、Notion 或网站，或者你已备份原始文档，可跳过此步骤。

   a. 打开**文件**，进入 **Cache** > **difyv2** > **volumes** > **app** > **storage** > **upload_files**。
   
   b. 下载文件。
   
   :::tip 根据上传时间识别文件
   该文件夹中的文件名是 Dify 内部名称。请根据上传时间、文件格式和大小来识别所需文件。
   :::

3. 手动记录模型配置和系统设置。

## 卸载 Dify

打开 **Market**，进入 **My Olares**，卸载 Dify 应用。不要选择 **Also remove all local data**。

## 安装新版 Dify 应用

1. 打开 **Market**，搜索 "Dify"。
2. 点击应用卡片，打开应用详情页。
3. 查看 **Information** 面板。**Compatibility** 字段显示 `Olares >=1.12.6-0` 即为新版本。
4. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 导入应用

1. 打开 Dify 并创建管理员账号。
2. 进入 **Studio** 页面。
3. 对每个之前导出的应用：

   a. 点击 **CREATE APP** 卡片上的 **Import DSL file**。
   
   b. 选择该应用的 `.yml` 文件并上传。

4. 如果 Dify 检测到缺少插件，请安装它们。

   如果跳过某个插件或安装失败，应用配置仍会被导入，但你需要重新配置相关插件节点。

5. 如果应用使用了知识库，请先重建知识库，然后在应用中重新配置。
6. 进入 **Studio** 页面，确认导入的应用已显示。

## 重建知识库

1. 在新版 Dify 中新建知识库，并重新上传文档或重新连接数据源。
2. 打开一个使用了知识库的应用，测试一次查询以确认检索正常。

## 了解更多

- [共享应用](../manual/olares/market/shared-apps.md)：了解新的共享应用架构。
