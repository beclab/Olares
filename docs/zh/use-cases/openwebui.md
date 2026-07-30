---
outline: [2, 3]
description: 在 Olares 上自托管 Open WebUI，实现私有的本地 AI 聊天。连接本地模型，让所有会话保留在你自己的设备上。
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, self-hosted AI platform, local LLM, open webui on olares
app_version: "1.0.38"
doc_version: "3.0"
doc_updated: "2026-07-30"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/openwebui.md)。
:::

# 设置 Open WebUI 与本地 AI 对话

Open WebUI 是一个自托管的聊天界面，让你在 Olares 设备上与本地模型进行私密对话。

本指南将带你完成安装 Open WebUI、连接本地模型并开始第一次对话。

## 学习目标

在本指南中，你将学习如何：

- 在 Olares 上安装 Open WebUI。
- 创建管理员账号。
- 将 Open WebUI 连接到本地模型。
- 使用配置好的模型开始聊天会话。

## 前提条件

开始前，你需要：

- 具有足够磁盘空间和内存的 Olares 设备。
- 以下模型：

   | 模型类型 | 模型 | 获取方式 |
   | :--- | :--- | :--- |
   | 聊天 | Qwen3.6-27B (llama.cpp) | 从应用市场安装 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 安装 Open WebUI

1. 打开应用市场，搜索 "Open WebUI"。

   ![Open WebUI](/images/one/open-webui.png#bordered)

2. 点击**获取**，然后点击**安装**。等待安装完成。

## 创建管理员账号

首次启动 Open WebUI 时，你需要创建一个本地管理员账号，用于管理模型和设置。

1. 从启动台打开 Open WebUI。
2. 在欢迎页面中，点击 **Get started**。
3. 输入你的姓名、邮箱和密码，创建管理员账号。

   ![Create account](/images/one/open-webui-create-account.png#bordered)

## 获取模型连接信息

要将 Open WebUI 连接到模型，你需要先从模型控制台收集连接信息。

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

## 在 Open WebUI 中配置连接

连接信息准备就绪后，在 Open WebUI 中将该模型添加为 OpenAI 兼容提供商。

1. 在 Open WebUI 中，点击你的头像图标，选择 **Admin Panel**。
2. 选择 **Settings** 标签页，然后从左侧边栏选择 **Connections**。
3. 在 **Manage OpenAI API Connections** 右侧，点击 <span class="material-symbols-outlined">add</span> 添加新连接。
4. 在 **API Base URL** 字段中，输入你从 Model Console 复制的 **Base URL**。例如，`https://e46e044d.laresprime.olares.com/v1`。

   ![Connection established](/images/manual/use-cases/open-webui-connection-established1.png#bordered)

5. 点击 **Save**。Open WebUI 会自动验证连接。当看到 "OpenAI API settings updated" 消息时，表示连接已建立。

## 开始聊天

模型连接完成后，你就可以使用聊天界面了。

1. 在聊天区域，选择你配置好的模型。

   ![Select model](/images/manual/use-cases/open-webui-chat.png#bordered)

2. 在文本框中输入提示词，然后按 **Enter** 键开始对话。

   ![Chat with LLM](/images/manual/use-cases/open-webui-chat-result.png#bordered)

## 了解更多

- [设置多用户访问](openwebui-multiuser.md)：与 Olares 设备上的其他用户共享 Open WebUI。
- [配置音频](openwebui-audio.md)：启用语音转文字和文字转语音。
- [启用网页搜索](openwebui-search.md)：为聊天增加网页搜索能力。
- [使用知识库](openwebui-knowledge.md)：上传文档并创建用于 RAG 的知识库。
