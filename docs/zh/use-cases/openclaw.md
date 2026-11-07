---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/openclaw
outline: [2, 3]
title: 将 OpenClaw 作为自托管个人 AI 助手运行
description: 在 Olares 上将 OpenClaw 作为自托管个人 AI 助手运行。连接 Discord 或 Slack，同时让助手和数据保留在你的设备上。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, self-hosted ai agent, personal ai agent, local ai agent, openclaw on olares
app_version: "1.0.57"
doc_version: "4.0"
doc_updated: "2026-11-05"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/openclaw.md)为准。
:::

# 将 OpenClaw 作为你的自托管个人 AI 助手运行

<VersionRouteSelect />

OpenClaw 是一款专为本地设备设计的个人 AI 助手。它可以直接接入 Discord、Slack 等消息应用，让你在这些应用中与其交互。

它相当于一位"始终在线"的助手，能够执行搜索和发送文档、管理日历、浏览网页等实际任务。

## 学习目标

在本指南中，你将学习如何：
- 安装并初始化 OpenClaw 环境。
- 将 OpenClaw 与 Discord 等频道集成。
- 可选：启用网页搜索功能。
- 管理技能和插件。
- 使用 OpenClaw 助手管理 Olares。
- 可选：启用沙箱以实现安全代码执行。

## 前提条件

开始前，你需要：

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- Discord 账号：用于创建机器人应用。
- Discord 服务器：确保你在这个服务器上有添加机器人的权限。
- 以下模型：

  | 模型类型 | 模型 | 获取方式 |
  | :--- | :--- | :--- |
  | 聊天 | Qwen3.8-27B (llama.cpp) | 从 Market 安装 |

  :::tip 模型提供方
  本教程通过 Router 的 OpenAI 兼容 API 使用 Qwen3.8-27B (llama.cpp)。如使用其他提供方或本地代理，请参阅 [OpenClaw 关于自定义模型提供方的文档](https://docs.openclaw.ai/concepts/model-providers#providers-via-models-providers-custom%2Fbase-url)。
  :::

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 升级说明

如果你正在升级现有的 OpenClaw 安装，请在继续之前查看版本特定的更改和故障排除步骤。更多信息，请参阅[升级 OpenClaw](openclaw-upgrade.md)。

## 安装 OpenClaw

1. 在 Olares Market 中搜索 "OpenClaw"。

    ![在应用市场中搜索 OpenClaw](/images/zh/manual/use-cases/find-openclaw1.png#bordered){width=90%}

2. 点击**获取**，然后点击**安装**。安装完成后，启动台会出现两个快捷方式：
    - **OpenClaw CLI**：命令行界面
    - **Control UI**：图形化管理面板

    ![OpenClaw 入口](/images/manual/use-cases/openclaw-entry-points1.png#bordered){width=30%}

:::tip 运行多个 OpenClaw 实例
Olares 支持应用克隆。如果你希望同时运行多个独立的 AI 助手处理不同任务，可以克隆 OpenClaw 应用。更多信息，请参阅[克隆应用](../manual/olares/market/clone-apps.md)。
:::

## 初始化 OpenClaw

快速完成助手的初始化配置。

### 步骤 1：获取模型连接信息

本教程使用 Qwen3.8-27B (llama.cpp)，这是一款可从 Market 获取的支持工具调用的模型。

:::tip
OpenClaw 需要较大的"上下文窗口"（即 AI 的短期记忆）来处理复杂任务而不会忘记之前的指令。如使用本地模型，建议选择原生支持至少 64K token 上下文窗口的模型。
:::

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### 步骤 2：配置模型

安装 OpenClaw 后，Olares 会启动并管理 Gateway。使用模型配置向导将 OpenClaw 连接到 Router。

1. 从启动台打开 OpenClaw CLI 应用。
2. 运行以下命令：

    ```bash
    openclaw configure --section model
    ```

3. 使用方向键移动，按 **Enter** 确认。按下表配置连接：

    | 配置 | 选项 |
    |:-----|:-----|
    | Model/auth provider | 选择 **More…**，然后选择 **Custom Provider**。 |
    | API Base URL | 粘贴步骤 1 中从 Router 复制的 Base URL，保留 `/v1`。 |
    | How do you want to provide this API key? | 如果出现此提示，选择 **Paste API key now**。 |
    | API Key (leave blank if not required) | 填写 `olares`。 |
    | Endpoint compatibility | 选择 **OpenAI-compatible**。 |
    | Model ID | 填写 `default-chat`。等待出现 `Verification successful`，再继续下一步。 |
    | Endpoint ID | 填写 `router`。模型引用为 `router/default-chat`。 |
    | Model alias (optional) | 留空。 |
    | Does this model support image input? | Qwen3.8-27B (llama.cpp) 选择 **Yes**。使用其他模型时，以 Router 显示的能力为准。 |

    OpenClaw 的模型客户端要求密钥字段有值，因此填写占位值 `olares`。Router 会使用这个 Olares 应用的身份进行认证。

4. 出现 `Configuration updated` 后，打开终端用户界面（TUI）：

    ```bash
    openclaw tui
    ```

    确认顶部显示 `ws://127.0.0.1:18789`。发送一条消息，等待助手回复。

    ![OpenClaw TUI 已连接到 Router](/images/manual/use-cases/router-client-connect-openclaw.png#bordered)

5. 输入 `/quit` 并按 **Enter** 退出 TUI。保持 OpenClaw CLI 窗口打开，下一步还需要使用它配对设备。

### 步骤 3：配对设备

将浏览器与 Gateway 配对，以使用 Control UI。

1. 从启动台打开 Control UI 应用。首次连接时，页面会要求提供 Gateway 令牌。
2. 在 OpenClaw CLI 窗口中运行：

    ```bash
    openclaw gateway auth-token --show
    ```

3. 将显示的令牌复制到 Control UI 的 Gateway 密钥字段中，然后连接。
4. Control UI 会提示批准此浏览器。复制页面上显示的 `openclaw devices approve` 命令，在 OpenClaw CLI 窗口中运行。

    :::tip 配对请求过期
    如果命令返回 `unknown requestId`，请刷新 Control UI，再运行页面上新显示的批准命令。
    :::

5. 批准后，Control UI 会自动连接并打开主聊天页面。发送一条消息，等待助手回复。

:::details 手动选择待处理请求
如果有多个待处理的浏览器请求，请在 OpenClaw CLI 中运行 `openclaw devices list`。确认请求与你的浏览器对应后，再批准该请求：

```bash
openclaw devices approve <request-id>
```
:::

### 步骤 4：个性化 OpenClaw

为使 OpenClaw 机器人更具个性化，强烈建议完成人设设置流程。

该流程通过人设文件确立助手的身份、行为边界和长期记忆。这些文件可确保助手在所有平台和频道上的行为保持一致。

1. 在聊天区域，确保已选中你的模型。

   右下角的指示器（例如 `default-chat · Off`）显示当前模型和推理状态。`Off` 表示推理已禁用，而不是模型不可用。

2. （可选）启用推理。

   默认情况下，推理处于关闭状态以获得更快的响应。要启用或调整推理，请将 **Effort** 滑块拖动到 **Faster** 和 **Smarter** 之间的合适位置。

   ![Model selection panel](/images/manual/use-cases/openclaw-enable-model1.png#bordered)

3. 输入并发送一条简单的消息开始。例如：

    ```text
    Wake up please!
    ```
4. 助手会做出回应，并开始通过提问了解你。你可以在此过程中建立规则、个性特征和偏好。例如：

    ```text
    - Call me Bella. I like simple language without technical jargon and 
    concise bulleted answers.
    - You are John, a witty assistant who uses emojis.
    - Never access my calendar without asking first, and never execute any 
    financial operations.
    ```

5. 与助手对话时，助手会将你的偏好写入核心人设文件，例如 `IDENTITY.md`、`USER.md` 和 `SOUL.md`。
6. （可选）如果助手未能更新人设文件，请在聊天中明确要求它执行。

    如果问题仍然存在，请使用以下方法之一解决：
    - **检查上下文窗口**：先[在 Router 中查看实际配置的上下文大小](/zh/manual/best-practices/connect-ai-apps.md#check-context-window)。在 Files 中打开 **Data** > **clawdbot** > **config** > `openclaw.json`，将 `contextWindow` 设为引擎支持的值。例如，**Engine args** 显示 `-c 104448` 时，填写值不能超过 `104448`。

        :::tip
        修改 OpenClaw 的 `contextWindow` 不会扩大引擎的上下文容量。如需更大的上下文，先在 Router 中将引擎配置为模型和硬件支持的值，再更新客户端。
        :::

    - **更换模型**：切换到具有更好工具调用和指令遵循能力的模型。

7. 继续对话，直到助手收集到足够的信息。
8. 验证人设文件是否已成功更新：

    a. 从启动台打开 Files 应用。

    b. 进入 **Data** > **clawdbot** > **config** > **workspace**。

    c. 检查 `.md` 文件的修改时间，以确定哪些文件最近更新过，例如 `USER.md`、`IDENTITY.md` 和 `SOUL.md`。

    ![Persona files generated by OpenClaw](/images/manual/use-cases/openclaw-persona-files2.png#bordered)

    d. （可选）下载文件，在支持的文本编辑器中查看，并验证其中包含你新建的规则，例如你的名字、语言风格和限制。

    e. 如果存在临时的 `BOOTSTRAP.md` 文件，请删除它。

    :::tip 修改人设设置
    如需将来更改这些设置，请使用以下方法之一：
    - 在聊天中要求助手更新其规则。
    - 从此文件夹下载 `.md` 文件，在文本编辑器中编辑后重新上传，以覆盖旧文件。
    :::

## 后续步骤

1. [与 Discord 集成](openclaw-integration.md)，实现与助手的远程对话。
2. [启用网页搜索](openclaw-web-access.md)，使助手能够访问实时互联网信息。
3. [安装技能和插件](openclaw-skills.md)，进一步扩展助手的能力。
4. [使用 Openclaw 移动客户端连接](openclaw-mobile.md)，通过 LarePass VPN 扫码或手动配对。

## 故障排除和常见问题

如需常见错误和行为问题的解决方案，请参阅[常见问题](openclaw-common-issues.md)。

## 了解更多

- [如何在 Discord 中创建服务器](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)
