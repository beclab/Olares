---
outline: [2, 3]
description: 在 Olares 上通过 Hermes Workspace 和 Router 起草活动清单，并集中保存相关文件。
head:
  - - meta
    - name: keywords
      content: "Olares, Hermes Workspace, Router, AI 对话, 工作区文件"
app_version: "0.0.7"
doc_version: "1.0"
doc_updated: "2026-10-30"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/hermes-workspace.md)为准。
:::

# 使用 Hermes Workspace 策划图书馆活动

Hermes Workspace 将 AI 对话和工作区文件集中在浏览器中。在 Olares 上，本应用通过 `default-chat` 连接 Router，并将对话和文件保存在应用存储中。

## 学习目标

本指南将介绍如何：

- 通过 Router 起草简短的活动清单。
- 将策划笔记保存在工作区文件中。
- 重新打开已保存的对话。

## 前提条件

- Olares 1.12.7 或更高版本。
- [Router](olares-router.md) 中可用的 `default-chat` 路由。

## 安装 Hermes Workspace

1. 打开应用市场，搜索“Hermes Workspace”。

   <!-- ![Hermes Workspace](/images/manual/use-cases/hermes-workspace.png#bordered) -->

2. 点击**获取**，然后点击**安装**，等待安装完成。按提示将 **WORKSPACE_PASSWORD** 设置为至少 12 个字符的密码。

## 起草活动清单

1. 从启动台打开 Hermes Workspace，使用安装时设置的密码登录。
2. 点击 **New Hermes chat**，保留 **Default Chat** 选项。
3. 输入并发送以下示例请求：

   ```text
   Create a concise checklist for organizing a small community library event.
   Use three bullet points in English. Do not call tools or modify files.
   ```

4. 阅读回复，根据实际活动调整日期、场地、人员和活动安排。模型输出取决于 Router 中配置的模型。

![Hermes Workspace 中保存的对话](/images/manual/use-cases/hermes-workspace-chat.jpg#bordered)

## 保存相关笔记

打开 **Workspace**，点击 **New file**，将文件命名为 `library-event.md`。打开文件，添加策划笔记并保存修改。也可以通过 **Upload** 添加已有文件。

文件保存在此应用的工作区内。添加文件不会自动将内容发送给模型。需要将支持的文件用于对话时，请使用聊天附件按钮。

## 重新打开对话

在 **SESSIONS** 下选择对话，即可查看历史记录。应用重启后，对话、工作区文件和本地设置仍然保留。

## 高级智能体功能

本应用提供基础 Router 对话和工作区文件管理。工具调用、定时任务、技能、记忆和编排功能需要另行安装并配置 Hermes Agent 网关和仪表盘，基础 Router 连接不会启用这些功能。

智能体应用的使用方式请参阅 [Hermes Agent](hermes.md)。连接外部网关属于高级配置，不在本指南范围内。
