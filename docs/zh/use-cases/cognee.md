---
connectionVersion: 1.12.7
connectionLatestPath: /zh/use-cases/cognee
outline:
- 2
- 3
description: 在 Olares 上使用 Cognee 构建和浏览文档知识
head:
- - meta
  - name: keywords
    content: Olares, Cognee, 自托管, 知识图谱, 文档, 搜索
app_version: 0.0.8
doc_version: '1.0'
doc_updated: '2026-10-29'
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/cognee.md)为准。
:::

# 使用 Cognee 关联项目知识

Cognee 将文本和文档整理为相互关联的知识，支持搜索和浏览。在 Olares 上运行时，网页界面、上传内容和知识存储都由 Olares 承载。语言任务使用 AI Router 中配置的默认对话模型。

## 学习目标

- 将简短文档添加到数据集。
- 将内容处理为可搜索的知识。
- 提问并检查图谱或源资料。

## 前提条件

- Olares 1.12.7 或更高版本。
- 已连接 [Olares 专用网络](../manual/larepass/private-network.md)的浏览器。
- AI Router 中可用的 **default-chat** 路由。
- 首次下载 Cognee 本地嵌入模型所需的网络连接。

本应用固定使用上游开发版镜像。建议先使用文档副本，保留原文件。

## 安装 Cognee

1. 打开应用市场并安装 **Cognee**。提示时设置至少 12 个字符的初始密码。
2. 从启动台打开 Cognee。
3. 使用 `default_user@example.com` 和安装时设置的密码登录。

小型嵌入模型在 Cognee 内使用 CPU 运行。首次下载可能延迟第一个处理任务，之后会重复使用已保存的模型缓存。

## 添加项目文档

打开 **Brain**，点击 **New brain**，命名为 `community_garden`。名称使用字母、数字和下划线。选择新建的知识库，点击 **Paste text**，添加几条相关事实。例如：

```text
The community garden opens on Saturday mornings.
Maya coordinates the volunteers.
The tool shed is beside the north entrance.
Volunteers meet Maya at the tool shed at 9:00.
```

点击 **Add** 导入文本。等待条目显示 **Ready** 后再提问。将相关文档放在同一个知识库中，便于一起搜索。

![Cognee 中的文档数据集](/images/manual/use-cases/cognee-dataset.png#bordered)

## 搜索和浏览

打开 **Search**，选择数据集并提问：**Where should volunteers meet, and who coordinates them?**

阅读回答，然后打开 **Mindmap** 检查相关图谱或源资料。不同模型生成的关联关系可能不同，使用前请与原文对照。

![Cognee 中相互关联的项目知识](/images/manual/use-cases/cognee-graph.png#bordered)

## 保留知识

上传内容、图数据库文件、向量索引和模型缓存保存在应用的私有数据目录中。关系记录使用 Olares PostgreSQL。迁移或恢复实例时，应同时备份两部分数据。

如果第一个任务较慢，先检查嵌入模型是否仍在下载，不要重复提交相同文档。如果语言生成失败，检查 AI Router 中的 **default-chat** 路由。

## 了解更多

- [Cognee 文档](https://docs.cognee.ai/)
- [Olares Router](./olares-router.md)
