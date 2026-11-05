---
outline: [2, 3]
description: 在 Hermes Workspace 中起草活动清单、将策划笔记保存在工作区文件、恢复过去的对话，然后连接 Hermes Agent，让对话可以执行操作。
head:
  - - meta
    - name: keywords
      content: "Olares, Hermes Workspace, Router, AI 对话, 工作区文件"
app_version: "0.0.10"
doc_version: "1.0"
doc_updated: "2026-11-05"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/hermes-workspace.md)为准。
:::

# 与 Hermes Workspace 对话、存文件、执行操作

Hermes Workspace 是运行在你的 Olares 上的 AI 工作台。你可以在里边和模型对话，把相关文件保存在同一处。对话和文件在应用重启后仍然可用。

本指南以一个社区图书馆活动为贯穿例子。最后一节可选连接 Hermes Agent，让对话可以执行创建文件、运行命令等操作。

## 学习目标

在本指南中，你将学习如何：

- 在 Hermes Workspace 中起草活动清单。
- 将策划笔记保存在工作区文件中。
- 通过追问恢复已保存的对话。
- 连接 Hermes Agent，让对话可以执行操作。

## 前提条件

- Olares 1.12.7 或更高版本。
- [Router](olares-router.md) 中可用的 `default-chat` 路由。Hermes Workspace 通过该路由与模型对话，因此普通聊天无需额外配置后端。

## 安装 Hermes Workspace

安装时只需要设置一项：你登录时使用的密码。

1. 打开 Market，搜索 "Hermes Workspace"。
2. 点击 **Get**，然后点击 **Install**，等待安装完成。出现提示时，将 **WORKSPACE_PASSWORD** 设置为至少 12 个字符的密码。

## 与你的模型对话

发送一条对话请求，看看 Hermes Workspace 如何与模型对话，然后以回复为基础起草你自己活动的方案。

1. 从 Launchpad 打开 Hermes Workspace，使用安装时设置的密码登录。
2. 在欢迎页点击 **Skip setup**，先从普通聊天开始。
3. 点击 **New Hermes chat**。保持选中 **Default Chat**。
4. 输入以下示例请求并发送：

   ```text
   Create a concise checklist for organizing a small community library event.
   Use three bullet points. Do not call tools or modify files.
   ```

5. 阅读回复。在这个例子中，结果是一份三条要点的清单，涵盖场地和日期、活动环节和宣传。它看起来像一份成品方案，请逐条核对，把与你的活动不符的地方改成自己的信息。

## 在工作区中保存文件

上一步中你已经把清单改成了自己活动的详情。把这些详情保存为工作区文件，回到这个工作区时还能找到。

1. 进入 **Workspace** > **Files**，新建名为 `library-event.md` 的文件。
2. 双击文件，点击右上角 **Edit**，输入策划笔记，然后点击 **Save**。也可以用 **Upload** 上传已有文件。

   ```text
   # Library event plan

   - Date: Saturday, March 14, 2:00-4:00 pm
   - Venue: Community room, Central Library
   - Staff: 2 desk volunteers, 1 setup helper
   - Activities: storytime, book sale, sign-up desk
   - To confirm: guest speaker reply due March 1
   ```

在这里保存文件不会把文件共享给模型。要让模型读取文件，需要在对话中附加它。

## 恢复过去的对话

每段对话都保存为一个会话。重新打开它，从上次中断的地方继续。

1. 进入 **Chat** > **Hermes**，在 **SESSIONS** 下选择要恢复的对话，重新打开其历史记录。
2. 发送追问请求，从上次中断的地方继续，比如让模型根据清单生成活动当天的日程。

## 连接 Hermes Agent

默认情况下，Hermes Workspace 向 `default-chat` 路由背后的模型发送普通对话请求。本节连接 Hermes Agent 后端，让对话也能执行操作。在连接该后端之前，**Jobs**、**Skills**、**MCP** 等侧边栏功能保持锁定状态。

### 配置 Hermes Agent

1. 安装 [Hermes Agent](hermes.md)，并为其配置 Router 的 `default-chat` 路由。
2. 打开 Olares Settings，进入 **Applications** > **Hermes Agent** > **Manage environment variables**：

   a. 启用 **API_SERVER_ENABLED**。

   b. 将 **HERMES_API_SERVER_KEY** 设置为至少 16 个字符的强密钥。妥善保管该密钥。

   c. 点击 **Apply**，等待 Hermes Agent 重启完成。

3. 对于 Hermes Agent 0.22.2，还要在其持久配置中保留这些 API 设置：

   a. 打开 Olares Files，进入 **Data** > **hermesagent** > **home**。

   b. 打开 `.env`，设置 `API_SERVER_ENABLED=true`，并将 `API_SERVER_KEY` 设置为同一个密钥。

4. 从 Launchpad 打开 Hermes CLI，运行：

   ```bash
   hermes config set platforms.api_server.enabled true
   hermes gateway restart
   ```

   如果打开的是聊天界面而不是终端，输入 `/quit` 返回终端，然后再运行命令。

   等 gateway 重启完成后再连接 Hermes Workspace。这样可以避免 Hermes Agent 重启后 API 仍处于禁用状态。

### 连接 Hermes Workspace

1. 打开 Olares Settings，进入 **Applications** > **Hermes Agent** > **Entrances**。
2. 点击 **Hermes Gateway API**，从 **Endpoint** 复制其 URL。对 **Dashboard** 重复同样操作，获取它的 URL。
3. 打开 **Applications** > **Hermes Workspace** > **Manage environment variables**，设置以下字段：

   | 字段 | 值 |
   | :--- | :--- |
   | **HERMES_GATEWAY_URL** | 第 2 步复制的 **Hermes Gateway API** URL |
   | **HERMES_DASHBOARD_URL** | 第 2 步复制的 **Dashboard** URL |
   | **HERMES_API_TOKEN** | 与 Hermes Agent 的 `HERMES_API_SERVER_KEY` 相同的密钥 |

4. 点击 **Apply**，等待应用重启完成。
5. 重新打开 Hermes Workspace。Hermes Agent 后端已连接。
6. 开始新对话，让 Hermes Agent 在自己的工作目录中创建一份小型活动清单，读回文件内容并汇报结果。查看工具执行结果和回复。

   ```text
   Create a small event checklist in your working directory for a community library event.
   Save it as library-event-checklist.md, then read the file back and report what you wrote.
   ```

7. 可选：对话结束后要查找文件，按归属的存储位置查找。

   - **Workspace files**：在 Hermes Workspace 中创建或上传的文件保存在 Files 的 **Data** > **hermesworkspace** > **data** > **files** 下，同时显示在应用的 **Workspace** > **Files** 中。
   - **Agent files**：Hermes Agent 创建的文件在 Files 的 **Data** > **hermesagent** > **home** 下。

## 了解更多

- [Hermes Agent](hermes.md)：连接后提供执行能力的命令行智能体后端。
- [Router](olares-router.md)：`default-chat` 背后的模型路由，包括如何配置本指南使用的路由。
