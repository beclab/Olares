---
connectionVersion: 1.12.7
connectionLatestPath: /zh/use-cases/presenton
outline:
- 2
- 3
description: 在 Olares 上通过 AI Router 创建可编辑的演示文稿
head:
- - meta
  - name: keywords
    content: Olares, Presenton, AI, 演示文稿, PowerPoint, PDF
app_version: "0.0.8"
doc_version: '1.0'
doc_updated: '2026-10-29'
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/presenton.md)为准。
:::

# 使用 Presenton 创建可编辑的演示文稿

Presenton 根据提示词或文档创建幻灯片。选择设计、检查生成内容后，可导出可编辑的 PowerPoint 文件或 PDF。本 Olares 应用使用 AI Router 的默认对话模型，并将演示文稿数据保存在自己的设备中。

## 学习目标

- 根据提示词生成简短演示文稿。
- 检查并编辑大纲和幻灯片。
- 导出结果，继续编辑或分享。

## 前提条件

- 运行 Olares 1.12.7 或更高版本的 amd64 设备。
- 已连接 [Olares 专用网络](../manual/larepass/private-network.md)的浏览器。
- AI Router 中可用的 **default-chat** 路由。

应用固定使用上游 0.9.12-beta 容器镜像。本应用关闭图片生成，基础流程无需配置单独的图片服务。模板可能使用占位图片，分享前请检查。

## 安装 Presenton

1. 打开应用市场并安装 **Presenton**。
2. 从启动台打开 Presenton。
3. 等待首次启动完成。提示时创建实例账号，然后登录。

应用已预先配置为通过 AI Router 使用 **default-chat**。此流程无需粘贴外部模型密钥。首次使用可能需要下载本地索引资源，之后会重复使用已保存的缓存。

## 创建简短演示文稿

在新建演示文稿页面选择 **Standard**，将幻灯片数量设置为 **1**，语言选择 **English**，再输入具体说明。例如：

> Create one introductory slide for the Portland Library Community Garden. Volunteers meet on Saturdays at 9 AM. Maya Chen coordinates the group. The garden welcomes new volunteers. Use a short title and readable English text.

点击 **Generate presentation**。检查生成的大纲、页数和主题顺序，选择合适的内置模板后，再继续生成完整幻灯片。

![Presenton 中的演示文稿流程](/images/manual/use-cases/presenton-create.png#bordered)

## 检查并编辑

逐页阅读幻灯片。双击文字元素即可编辑，完成后点击元素外部。在顶部点击演示文稿标题可修改名称。导出前请修正姓名、日期和事实。模型可以提供初稿，但可能遗漏细节，或生成不适合当前设计长度的文字。

先用一页幻灯片检查默认模型能否完成流程。较长演示文稿和复杂排版取决于模型返回结构化内容的能力。如果生成较慢，先查看当前任务，不要重复提交相同提示词。

![Presenton 中编辑后的幻灯片](/images/manual/use-cases/presenton-editor.png#bordered)

## 导出并保存结果

点击 **Export**。如需在演示软件中继续编辑，选择 **PPTX**；如需阅读和分享，选择 **PDF**。分享前，打开导出文件并检查排版。

演示文稿记录使用 Olares PostgreSQL。上传文档、素材和导出文件保存在应用的私有数据目录中。下载重要导出文件，并在备份计划中同时包含数据库和应用数据。

## 了解更多

- [Presenton 文档](https://docs.presenton.ai/)
- [Olares Router](./olares-router.md)
