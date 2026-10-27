---
outline: [2, 3]
description: 了解如何使用 Lares 和 Olares CLI Agent Skills，通过自然语言管理 Olares。
head:
  - - meta
    - name: keywords
      content: Olares One, Lares, Router, Agent Skills, 自然语言, AI 智能体, Olares CLI
---

# 通过自然语言管理 Olares

Lares 是 Olares 内置的 AI 助手。借助 Router 和已连接的模型，它能理解你的自然语言请求，并通过 Olares CLI Agent Skills 将其转化为真实的设备管理操作。例如，你可以让 Lares 检查系统状态、安装应用、管理文件或排查问题。

本指南将带你完成第一次 Lares 会话。你将检查环境是否就绪、开始对话，并尝试几个常见任务。

## 学习目标

完成本教程后，你将学会：

- 为 Lares 准备环境。
- 开始与 Lares 的第一次对话。
- 通过自然语言管理 Olares。

## 步骤 1：准备环境

准备工作取决于你的起点。请在下表中找到属于你的情况。

| 起点 | 预装应用 | 下一步 |
| --- | --- | --- |
| Olares One v1.12.7 出厂镜像<br>（新设备） | <ul><li>Lares</li><li>Router</li><li><nobr>Qwen3.8-27B (llama.cpp)</nobr></li></ul> | 打开 Lares 即可开始 |
| <ul><li>自托管 Olares v1.12.7<br>（全新安装或升级）</li><li><nobr>Olares One 升级至 v1.12.7</nobr></li></ul> | 无 | 从 Market 安装 Router、Lares 和 Qwen3.8-27B (llama.cpp) |

:::warning 重要：一次只运行一个任务
在 Olares One 上运行 Qwen3.8-27B (llama.cpp) 时，我们建议一次只运行一个请求，以保证 100K 上下文窗口和模型精度下的最佳体验。
:::

## 步骤 2：开始第一次 Lares 对话

1. 从启动台打开 Lares。
2. 保留默认工作区，或选择其他工作区。
3. 保留默认的写入权限，或选择只读或完全访问。
4. 确认已选中 Qwen3.8-27B 模型。

   ![Lares 聊天界面](/images/one/lares-chat1.png#bordered)

5. 发送你的第一句 `Hello`。消息发送成功后，你就可以通过与 Lares 对话来管理 Olares 了。

   ![Lares 回复](/images/one/lares-chat-response1.png#bordered)

## 步骤 3：尝试常见任务

环境已就绪，来试试这些常见任务，从快速状态查看到完整的应用部署。

### 查看设备配置

先问一个简单问题：

```text
I'm new to Olares. Check this device's configuration first.
```

![在 Lares 中查看设备配置](/images/one/onboard-scenario-question2.png#bordered)

### 从 Market 安装应用

让 Lares 帮你安装一个应用：

```text
Install NocoDB from the Olares Market and tell me when it's ready.
```

![在 Lares 中安装应用](/images/one/onboard-scenario-install4.png#bordered)

### 部署应用到 Olares

进阶任务：让 Lares 从 GitHub 仓库部署一个项目。以下示例使用金融应用 "Wealthfolio"。

```text
Deploy this app to Olares: https://github.com/wealthfolio/wealthfolio
and make sure it has a desktop icon.
```

Lares 会检查源应用、准备 Olares 应用 chart 并更新所需的清单文件。根据应用不同，这可能需要几分钟。完成后，Lares 会告诉你如何验证结果。

![在 Lares 中部署应用](/images/one/onboard-scenario-porting3.png#bordered)

然后，你就可以在启动台和 **My Olares** 中找到该应用。

![部署完成的应用](/images/one/onboard-scenario-ported1.png#bordered)

## 不止 Lares

Lares 是推荐入口，但不是使用 Olares CLI Agent Skills 的唯一方式。你也可以在 Olares 上的其他智能体应用中使用同样的技能，或通过 Codex、Cursor 等本地智能体，用自然语言管理你的 Olares 设备。

## 了解更多

- [Olares、Olares CLI 与 AI 智能体：关系解读](https://www.olares.com/blog/olares-cli-ai-agents-explained)：Olares、Olares CLI 和 AI 智能体应用之间的关系。