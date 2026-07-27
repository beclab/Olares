---
outline: [2, 3]
description: 了解如何通过创建机器人、配置渠道和授权账户，将 OpenClaw 与 Discord 集成。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw 教程, OpenClaw 使用指南, 消息平台集成, Discord 集成
app_version: "1.0.1"
doc_version: "2.0"
doc_updated: "2026-05-27"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/openclaw-integration.md)。
:::

# 将 OpenClaw 与 Discord 集成

将 OpenClaw 智能体连接到 Discord 机器人，实现远程对话。

## 学习目标

在本指南中，你将学习如何：
- 创建 Discord 机器人并生成 API token。
- 将机器人邀请至 Discord 服务器，并授予必要权限。
- 将机器人连接到 OpenClaw，并配置频道访问策略。
- 将你的 Discord 账户与机器人配对，以便与智能体对话。

## 前提条件

- Discord 账号：用于创建机器人应用。
- Discord 服务器：确保你在这个服务器上有添加机器人的权限。

## 步骤 1：创建 Discord 机器人

1. 使用 Discord 账号登录 [Discord Developer Portal](https://discord.com/developers/applications)。
2. 点击**新 APP**。
3. 输入应用名称，同意条款，然后点击**创建**。

    ![创建应用窗口](/images/zh/manual/use-cases/create-app.png#bordered){width=40%}

4. 从左侧边栏，选择**机器人**。
5. 向下滚动至 **Privileged Gateway Intents** 部分，启用以下设置：

    - Presence Intent
    - Server Members Intent
    - Message Content Intent

6. 点击**保存更改**。
7. 向上滚动至**令牌**部分，点击**重置令牌**，然后复制生成的 Discord 机器人 token。后续在 Control UI 中配置频道时需要用到该 token。

    ![重置令牌](/images/zh/manual/use-cases/reset-token.png#bordered)

## 步骤 2：将机器人邀请至服务器

1. 从左侧边栏，选择 **OAuth2**，然后找到 **OAuth2 URL 生成器**部分：

    a. 在**范围**中，选择 **bot** 和 **applications.commands**。

    ![OAuth2 URL Generator](/images/zh/manual/use-cases/oauth21.png#bordered)

    b. 向下滚动至**机器人权限**部分，按下图设置。后续可修改这些设置。

    ![机器人权限](/images/zh/manual/use-cases/bot-permissions1.png#bordered)

2. 复制底部**已生成的 URL**。
3. 将该 URL 粘贴到新的浏览器标签页中，在**添加至服务器**中选择你的 Discord 服务器，点击**继续**，然后点击**授权**。 

    机器人已被授权并添加到你的服务器。

    ![机器人已添加到服务器](/images/manual/use-cases/bot-added.png#bordered)

## 步骤 3：配置频道

运行 OpenClaw 配置向导，连接你的 Discord 机器人。

1. 打开 OpenClaw CLI。
2. 运行以下命令启动配置向导：

    ```bash
    openclaw configure --section channels
    ```

3. 按如下方式进行配置：

    | 配置 | 选项 |
    |:---------|:-------|
    | Where will the Gateway run | Local (this machine) |
    | Channels | Configure/link |
    | Select a channel | Discord (Bot API) |
    | How do you want to provide this Discord bot token | Enter Discord bot token |
    | Enter Discord bot token | 填入[步骤 1](#步骤-1-创建-discord-机器人)中获取的机器人 token |
    | Configure Discord channels access | Yes |
    | Discord channels access | Open (allow all channels) |
    | Select a channel | Finished |
    | Configure DM access policies now?<br>(default: pairing) | Yes |
    | Discord DM policy | Pairing (recommended) |

## 步骤 4：授权你的账户

出于安全考虑，机器人不会与未授权用户对话。你必须将你的 Discord 账户与机器人配对。

1. 打开 Discord，向你的新机器人发送一条私信。机器人将回复一条包含 Pairing Code 的错误消息。
2. 打开 OpenClaw CLI，输入以下命令：

    ```bash
    openclaw pairing approve discord {Your-Pairing-Code}
    ```

3. 获得批准后，你就可以在 Discord 中与智能体对话了。

## 后续步骤

- [可选：启用网页搜索](openclaw-web-access.md)
- [管理技能和插件](openclaw-skills.md)
