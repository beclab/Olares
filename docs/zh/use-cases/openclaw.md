---
outline: [2, 3]
title: 将 OpenClaw 作为自托管个人 AI 助手运行
description: 在 Olares 上将 OpenClaw 作为自托管个人 AI 助手运行。连接 Discord 或 Slack，同时让助手和数据保留在你的设备上。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, self-hosted ai agent, personal ai agent, local ai agent, openclaw on olares
app_version: "1.0.36"
doc_version: "3.0"
doc_updated: "2026-09-04"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/openclaw.md)为准。
:::

# 将 OpenClaw 作为你的自托管个人 AI 助手运行

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

开始前，你需要准备：

- Discord 账号：用于创建机器人应用。
- Discord 服务器：确保你在这个服务器上有添加机器人的权限。
- 以下模型：

  | 模型类型 | 模型 | 获取方式 |
  | :--- | :--- | :--- |
  | 聊天 | Gemma 4 26B (Ollama) | 从 Market 安装 |

  :::tip 模型提供方
  本教程通过 Ollama API 使用 Gemma 4 26B。如使用其他提供方或本地代理，请参阅 [OpenClaw 关于自定义模型提供方的文档](https://docs.openclaw.ai/concepts/model-providers#providers-via-models-providers-custom%2Fbase-url)。
  :::

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

本教程使用 Gemma 4 26B (Ollama)，这是一款可从 Market 获取的支持工具调用的模型。

:::tip
OpenClaw 需要较大的"上下文窗口"（即 AI 的短期记忆）来处理复杂任务而不会忘记之前的指令。如使用本地模型，建议选择原生支持至少 64K token 上下文窗口的模型。
:::

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-ollama-->

### 步骤 2：运行安装向导

使用交互式向导逐步配置 OpenClaw。

1. 从启动台打开 OpenClaw CLI 应用。
2. 输入以下命令启动交互式安装向导：

    ```bash
    openclaw onboard --classic
    ```

    :::tip 不要单独运行 `openclaw onboard`
    不要单独运行 `openclaw onboard` —— 它会启动一个对话式 TUI，需要在环境变量中配置 API 密钥。由于当前没有配置 API 密钥，这条路走不通。对于交互式向导，请改用 `openclaw onboard --classic`。

    如果不小心进入了 TUI，输入 `/quit` 并按 **Enter** 退出。
    :::

3. 向导会引导你完成一系列步骤。使用方向键移动，按 **Enter** 确认。

    :::tip 关于配置的说明
    为便于快速上手，本教程会跳过向导中的部分高级设置。你可以稍后配置或修改它们。
    :::

    :::tip Custom Provider：Base URL 和 API key
    如果为其他非 Ollama 本地模型选择了 **Custom Provider**，请按以下方式配置连接：
    - **通过 Router**：输入 Router 的 **API Base URL**。例如，`https://<your-olares-id>.laresprime.olares.com/v1`。在 Router 控制台中为 OpenClaw 创建一个 API key，并将其粘贴到 **API Key** 字段中。
    - **直连模型**：输入 Model Console 中显示的 **Base URL**，API key 可填写任意文本。
    :::

    | 配置   | 选项   |
    |:-----------|:---------|
    | Personal-by-default acknowledgment | 选择 **Yes**。  |
    | Help make OpenClaw better | 按需选择。  |
    | Setup mode   | 选择 **QuickStart**。   |
    | Model/auth provider  | 选择 **More**，然后选择 **Ollama**。<br>对于其他本地模型，选择 **Custom Provider**。 |
    | Ollama auth method | 选择 **Ollama**。 |
    | Ollama mode | 选择 **Local only**。 |
    | Ollama base URL  | 删除默认占位文本，然后输入[步骤 1](#步骤-1获取模型连接信息)中复制的 **Base URL**。 |
    | Default model | 选择 **Browse all models**，然后选择已安装的模型 `ollama/gemma4:26b`。 |
    | Test AI access now with a live completion | 选择 **Yes**。<br>出现 `AI access works` 消息表示 OpenClaw 已成功连接到模型。 |
    | Remaining settings (channels, search provider,<br>and skill dependencies) | 选择 **Skip for now**。<br>你可以稍后配置它们。 |

    完成安装向导后，OpenClaw 会自动打开终端用户界面（TUI）。

    ![OpenClaw TUI after setup](/images/manual/use-cases/openclaw-setup-finish-tui2.png#bordered)

4. 输入 `/quit` 并按 **Enter** 退出。

    <!--![OpenClaw TUI first chat](/images/manual/use-cases/openclaw-tui-firstchat.png#bordered)-->

5. 保持 OpenClaw CLI 窗口打开。下一步还需要使用它。

### 步骤 3：配对设备

将 Control UI 与 OpenClaw CLI 配对，以使用图形化管理面板。

<Tabs>
<template #(推荐)-自动配对设备>

1. 从启动台打开 Control UI 应用。**OpenClaw Gateway Dashboard** 将打开。

    会出现 `Auth required` 错误。这是正常现象，表示你尚未提供访问 token。

    ![Gateway dashboard](/images/manual/use-cases/gateway-dashboard5.png#bordered)

2. 找到 **Paste the token from** 这一行，然后复制其中提供的命令 `openclaw gateway auth-token --show`。
3. 返回 OpenClaw CLI 窗口，运行复制的命令，然后复制显示的访问 token。
4. 返回 OpenClaw Gateway Dashboard，找到 **Gateway Token** 字段，粘贴复制的访问 token，然后点击 **Connect**。

    会出现 `Device pairing required` 错误。这是正常现象，表示设备连接正在等待批准。

    ![Device pairing required](/images/manual/use-cases/gateway-device-pairing-required2.png#bordered)

5. 在 `Device pairing required` 错误信息中，找到 `Approve this request:` 这一行，然后复制其下方显示的命令。
6. 返回 OpenClaw CLI 窗口，然后运行复制的命令以授权 Control UI。

    :::tip 关于超时错误的说明
    批准命令有时效限制。如果收到 `unknown requestId` 错误，表示请求已过期。请刷新 Control UI，复制新生成的命令，然后在 OpenClaw CLI 中重新运行。

    ```text
    [openclaw] The CLI command failed.
    [openclaw] Reason: unknown requestId
    ```
    :::

7. 当终端显示批准消息时，返回 Control UI。系统将自动登录并跳转到 **Home** 页面。例如：

    ```text
    Approved 005748253152b66dc0f5f6a801f35617db043f107972f259630a6bd098d5f790 (82e0f4ac-ed44-477e-8c6c-c3d2f4eeedaf)
    ```

    ![Health OK](/images/manual/use-cases/openclaw-connected5.png#bordered)
</template>
<template #(可选)-手动配对设备>

:::tip 何时使用手动配对
自动配对将批准最近的配对请求。如果你有多个待处理请求，需要手动选择要批准的设备，请按本节步骤操作。
:::

1. 从启动台打开 Control UI 应用。**OpenClaw Gateway Dashboard** 将打开。

    会出现 `Auth required` 错误。这是正常现象，表示你尚未提供访问 token。

    ![Gateway dashboard](/images/manual/use-cases/gateway-dashboard5.png#bordered)

2. 找到 **Paste the token from** 这一行，然后复制其中提供的命令 `openclaw gateway auth-token --show`。
3. 返回 OpenClaw CLI 窗口，运行复制的命令，然后复制显示的访问 token。
4. 返回 OpenClaw Gateway Dashboard，找到 **Gateway Token** 字段，粘贴复制的访问 token，然后点击 **Connect**。

    会出现 `Device pairing required` 错误。这是正常现象，表示设备连接正在等待批准。

    ![Device pairing required](/images/manual/use-cases/gateway-device-pairing-required2.png#bordered)

5. 返回 OpenClaw CLI 窗口并输入以下命令：
    ```bash
    openclaw devices list
    ```
6. 在 **Pending** 表格中，找到与你当前设备关联的 **Request** ID。

    :::info
    Request ID 有时效限制。如授权失败，请重新运行 `openclaw devices list` 以获取新的有效 ID。
    :::

    ![View pending device request](/images/manual/use-cases/pending-request.png#bordered)

7. 输入以下命令授权设备：

    ```bash
    openclaw devices approve {RequestID}
    ```

8. 当终端显示批准消息时，返回 Control UI。系统将自动登录并跳转到 **Home** 页面。

    ![Health OK](/images/manual/use-cases/openclaw-connected5.png#bordered)
</template>
</Tabs>

### 步骤 4：个性化 OpenClaw

为使 OpenClaw 机器人更具个性化，强烈建议完成人设设置流程。

该流程通过人设文件确立助手的身份、行为边界和长期记忆。这些文件可确保助手在所有平台和频道上的行为保持一致。

1. 在聊天区域，确保已选中你的模型。

   右下角的指示器（例如 `gemma4:26b · Off`）显示当前模型和推理状态。`Off` 表示推理已禁用，而不是模型不可用。

2. （可选）启用推理。

   默认情况下，推理处于关闭状态以获得更快的响应。要启用或调整推理，请将 **Effort** 滑块拖动到 **Faster** 和 **Smarter** 之间的合适位置。

   ![Model selection panel](/images/manual/use-cases/openclaw-enable-model1.png#bordered)

3. 输入并发送一条简单的消息开始。例如：

    ```text
    Wake up please!
    ```
4. 助手会响应并开始与你对话。你可以在此过程中建立规则、个性特征和偏好。例如：

    ```text
    - Call me Bella. I like simple language without technical jargon and 
    concise bulleted answers.
    - You are John, a witty assistant who uses emojis.
    - Never access my calendar without asking first, and never execute any 
    financial operations.
    ```

5. 与助手对话时，你可以看到它正在将你的偏好写入核心人设文件，例如 `IDENTITY.md`、`USER.md` 和 `SOUL.md`。
6. （可选）如果助手未能更新人设文件，请在聊天中明确要求它执行。

    如果问题仍然存在，请使用以下方法之一解决：
    - **增加上下文窗口**：打开 Files 应用，进入 **Data** > **clawdbot** > **config** > `openclaw.json`，然后将 `contextWindow` 值增加到至少 64K（建议 200K）。

        :::tip
        请注意，较大的上下文窗口会消耗更多显存，因此请选择硬件可支持的值。
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

## 故障排除和常见问题

如需常见错误和行为问题的解决方案，请参阅[常见问题](openclaw-common-issues.md)。

## 了解更多

- [如何在 Discord 中创建服务器](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)