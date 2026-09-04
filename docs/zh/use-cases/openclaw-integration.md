---
outline: [2, 3]
description: 了解如何通过创建机器人、配置渠道和授权账户，将 OpenClaw 与 Discord 集成。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw 教程, OpenClaw 使用指南, 消息平台集成, Discord 集成
app_version: "1.0.17"
doc_version: "2.1"
doc_updated: "2026-08-03"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/openclaw-integration.md)为准。
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

1. 打开 **Control UI**，点击左下角的用户账户，然后选择 **Settings**。
2. 从左侧边栏选择 **Channels**，找到 **Discord**，然后点击 **Set up**。
3. 在 **Set up Discord** 窗口中，查看 **How channels work**，向下滚动，然后点击 **Continue**。
4. 按如下方式配置：

   | 设置 | 选项 |
   |:---------|:-------|
   | Discord account | Default (primary) |
   | How do you want to provide this Discord bot token | Enter Discord bot token |
   | Enter Discord bot token | 粘贴[步骤 1](#步骤-1-创建-discord-机器人) 中复制的机器人 token |
   | Configure Discord channels access | Yes |
   | Discord channels access | Open (allow all channels) |
   | Configure DM access policies now?<br>(default: pairing) | Yes |
   | Discord DM access | 查看说明，然后点击 **Continue** |
   | Discord DM policy | Pairing (recommended) |
   | Done. Channels updated | **Continue** |
   | Channel configured | **Finish** |

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
