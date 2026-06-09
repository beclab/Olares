---
outline: [2, 3]
description: 了解如何将 OpenClaw 与 WhatsApp 集成。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw 教程, OpenClaw 使用指南, 消息平台集成, WhatsApp 集成
app_version: "1.0.3"
doc_version: "1.0"
doc_updated: "2026-06-09"
---

# 与 WhatsApp 集成

将 OpenClaw 智能体连接到 WhatsApp，实现远程交互。

选择以下任一账号设置方式开始：
- **使用独立账号（推荐）**：购买一张单独的 SIM 卡，为机器人注册一个专用的 WhatsApp 账号。
- **使用个人账号**：如果你仅需将机器人用于个人用途，可以关联自己的 WhatsApp 账号，通过给自己发消息来与机器人对话。

## 学习目标

在本指南中，你将学习如何：
- 使用独立账号或个人账号将 OpenClaw 关联到 WhatsApp。
- 将个人 WhatsApp 号码与独立机器人账号配对。
- 断开并彻底移除 WhatsApp 集成。

## 配置 WhatsApp 渠道

按需选择如下配置方式。

<Tabs>
<template #独立账号>

:::info 前提条件
开始前，请确保你已为机器人注册了一个单独的 WhatsApp 账号，并且已在移动设备上登录该账号。
:::

1. 从桌面打开 OpenClaw CLI。
2. 输入以下命令，启动配置向导：

    ```bash
    openclaw config --section channels
    ```

3. 按如下方式进行配置：

    - **Channel setup**：选择 **Add or update channels**。
    - **Select a channel**：选择 **WhatsApp (QR link)**。
    - **Install WhatsApp plugin**：选择 **Download from ClawHub**。
    - **Link WhatsApp now (QR)**：选择 **Yes**。

    终端中将显示一个二维码。

4. 在移动设备上，打开机器人的 WhatsApp 应用，进入**关联设备**，然后点击**关联设备**以扫描二维码。
5. 在 OpenClaw CLI 中查看消息 `Linked after restart; web session ready`。该消息表示绑定成功。
6. 继续在 OpenClaw CLI 中配置以下设置：

    - **WhatsApp phone setup**：选择 **Separate phone just for OpenClaw**。
    - **WhatsApp DM policy**：选择 **Pairing (recommended)**。
    - **WhatsApp allowFrom (optional pre-allowlist)**：选择 **Unset allowFrom (default)**。
    - 选择 **Finished** 退出。

    OpenClaw 将自动重启使配置生效。等待约 5 到 10 分钟。

7. 重启完成后，使用你的个人 WhatsApp 账号向机器人号码发送一条消息。机器人将回复一条包含配对命令和代码的消息。
8. 返回 OpenClaw CLI，粘贴对话中提供的完整命令（例如，`openclaw pairing approve whatsapp VUTNHSXS`），然后按**回车**。 

    终端中将显示成功消息，表明配对完成。

9. 在 WhatsApp 上再次向机器人发送消息。连接已建立，现在你可以与机器人对话了。
</template>
<template #个人账号>

1. 从桌面打开 OpenClaw CLI。
2. 输入以下命令，启动渠道配置向导：

    ```bash
    openclaw config --section channels
    ```

3. 按如下方式进行配置：

    - **Channel setup**：选择 **Add or update channels**。
    - **Select a channel**：选择 **WhatsApp (QR link)**。
    - **Install WhatsApp plugin**：选择 **Download from ClawHub**。
    - **Link WhatsApp now (QR)**：选择 **Yes**。

    终端中将显示一个二维码。

4. 在移动设备上，打开你的 WhatsApp 应用，进入**关联设备**，然后点击**关联设备**以扫描二维码。
5. 在 OpenClaw CLI 中查看消息 `Linked after restart; web session ready`。该消息表示绑定成功。
6. 继续在 OpenClaw CLI 中配置以下设置：

    - **WhatsApp phone setup**：选择 **This is my personal phone number**。
    - 按提示输入你的个人 WhatsApp 手机号码。
    - 选择 **Finished** 退出。

7. OpenClaw 将自动重启使配置生效。等待约 5 到 10 分钟。
8. 重启完成后，打开 WhatsApp 并给自己发送一条消息。机器人将在同一会话中回复你。
</template>
</Tabs>

:::tip 手动重启
完成配置后，OpenClaw 会自动重启 Gateway。WhatsApp 集成需要此次重启才能完全上线。

如果你等待超过 10 分钟后机器人仍未回复你的消息，请尝试手动重启容器：
1. 从桌面打开控制面板。
2. 找到 `clawdbot` 部署，然后点击右上角的**重启**。
:::

## 断开 WhatsApp

要彻底断开 WhatsApp 与 OpenClaw 的连接，你必须同时从 Control UI 和本地系统文件中移除相关配置。

1. 打开 Control UI，在左侧边栏选择 **Agents**。
2. 在 **Channels** 标签页中，找到 WhatsApp 配置卡片，然后点击 **Logout**。
3. 打开 OpenClaw CLI，然后运行以下命令：

    ```bash
    openclaw config --section channels
    ```

4. 选择 **Remove channel config**，选择 **WhatsApp**，然后选择 **Yes** 删除配置。
5. 选择 **Done** 退出向导。
6. 打开文件管理器，进入**数据** > **clawdbot** > **config** > **credentials**。
7. 删除 `whatsapp` 文件夹，并删除所有独立的 WhatsApp 凭证文件，例如 `whatsapp-pairing.json` 和 `whatsapp-default-allowFrom.json`。
8. 打开控制面板，重启 OpenClaw 容器使更改生效。

## 后续步骤

- [可选：启用网页搜索](openclaw-web-access.md)
- [管理技能和插件](openclaw-skills.md)
