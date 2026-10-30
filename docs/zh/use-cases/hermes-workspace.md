---
outline: [2, 3]
description: 在 Olares 上通过 Hermes Workspace 和 Router 起草活动清单，并集中保存相关文件。
head:
  - - meta
    - name: keywords
      content: "Olares, Hermes Workspace, Router, AI 对话, 工作区文件"
app_version: "0.0.9"
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

## 连接 Hermes Agent 执行工具任务

安装并配置 [Hermes Agent](hermes.md)，使用 Router `default-chat`。在其应用环境变量中启用 **API_SERVER_ENABLED**，并将 **HERMES_API_SERVER_KEY** 设置为至少 16 个字符的强密钥。请妥善保管该密钥。

使用 Hermes Agent 0.22.2 时，还需要将 API 设置保存在其持久配置中。在文件管理器中打开 **Data** > **hermesagent** > **home** > **.env**，设置 `API_SERVER_ENABLED=true`，并将 `API_SERVER_KEY` 设为同一密钥。保留其他配置项，妥善保护该文件。然后在 Hermes CLI 终端中运行：

```bash
hermes config set platforms.api_server.enabled true
hermes gateway restart
```

等待网关重启完成后再连接 Workspace，避免 Agent 重启后 API 仍处于停用状态。

1. 在设置中打开**应用** > **Hermes Agent** > **入口**，获取当前实例的 **Hermes Gateway API** 和 **Dashboard** 地址。
2. 打开**应用** > **Hermes Workspace** > **环境变量**，填写以下字段：

   | 字段 | 值 |
   | :--- | :--- |
   | **HERMES_GATEWAY_URL** | Hermes Gateway API 入口地址 |
   | **HERMES_DASHBOARD_URL** | Dashboard 入口地址 |
   | **HERMES_API_TOKEN** | 与 Hermes Agent 的 HERMES_API_SERVER_KEY 相同的值 |

3. 等待应用完成配置更新，再打开 Hermes Workspace。连接后端后，即可使用智能体技能和定时任务。
4. 新建对话，让智能体在自己的工作目录中创建一份简短的活动清单，读回文件并汇报结果。结合工具返回内容检查回复。

Hermes Agent 在自己的容器中执行工具。其文件和持久记忆与浏览器工作区文件、Workspace 的 **Memory** 笔记分别存储。请通过 Hermes Agent 的存储目录和 Dashboard 检查智能体生成的文件与记忆。

连接后的后端支持工具任务、技能管理、持久记忆和定时任务。定时任务可以在无人操作时调用模型。检查试用任务的输出后，请将其暂停。

委派的子任务可能在父对话结束当前轮次后才完成并返回结果。若尚未出现最终报告，可以要求父对话继续。完整的 Workspace 多智能体编排和 MCP 管理依赖更多后端能力，不在本指南范围内。
