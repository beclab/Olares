---
outline: [2, 4]
description: 在 Olares 上安装 Open WebUI，并将其连接到本地模型后端。你可以使用 Ollama 从 Ollama Registry 拉取模型，也可以使用预配置的模型应用。
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, 本地 LLM, AI 聊天机器人, Ollama
app_version: "1.0.29"
doc_version: "2.0"
doc_updated: "2026-05-13"
---

# 设置 Open WebUI 与本地 AI 对话

Open WebUI 为 Olares 设备上的本地模型提供易用的聊天界面。默认情况下，Open WebUI 不内置任何模型，因此需要先将其连接到模型后端。

本指南介绍两种配置方式：使用 Ollama 从 Ollama 模型库（Registry）直接拉取模型，或使用专用的预配置模型应用，例如 Qwen3.5 27B Q4_K_M (Ollama)。

## 学习目标

在本指南中，你将学习如何：

- 在 Olares 上安装 Open WebUI。
- 创建管理员账号。
- 使用 Ollama 或专用模型应用配置模型后端。
- 将模型连接到 Open WebUI 并开始聊天。

## 前提条件

- 拥有足够磁盘空间和内存的 Olares 设备。
- 拥有从应用市场安装共享应用的管理员权限。

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

## 配置模型后端

Open WebUI 需要后端模型来生成回复。可从以下选项中选择一种方式配置模型后端。

:::tip 多模型使用建议
Ollama（选项 A）更灵活，但在单个 Ollama 实例中同时托管多个模型可能会导致资源调度冲突。

如果你需要使用多个模型，为了获得更好的性能和稳定性，建议安装独立的模型应用（选项 B），而不是使用通用的 Ollama 应用。
:::

### 选项 A：使用 Ollama

使用 Ollama 可以从 Ollama 模型库拉取并切换不同模型，适合需要更高灵活性的场景。

#### 安装 Ollama

1. 打开应用市场，搜索 "Ollama"。
   ![Ollama](/images/manual/use-cases/ollama.png#bordered)

2. 点击**获取**，然后点击**安装**。等待安装完成。

#### 下载模型

安装 Ollama 后，你可以直接通过 Open WebUI 界面拉取模型。

:::tip 先浏览可用模型
访问 [Ollama Library](https://ollama.com) 浏览可用模型，并在下载前获取准确的模型名称。模型名称必须完全匹配，才能成功拉取。
:::

<Tabs>
<template #通过-Open-WebUI-主页>

1. 打开 Open WebUI。
2. 点击聊天页面顶部的模型下拉菜单，然后在搜索框中输入模型名称。例如：`llama3.2`。
3. 点击 **Pull "llama3.2" from Ollama.com**。下载会自动开始。

   ![Download from homepage](/images/one/open-webui-download-from-homepage.png#bordered)

</template>
<template #通过-Open-WebUI-设置>

1. 打开 Open WebUI。
2. 点击你的头像图标，选择 **Admin Panel**。
3. 前往 **Settings** > **Models**。
4. 点击右上角的 **Manage**，打开 **Manage Models** 窗口。
5. 在 **Pull a model from Ollama.com** 下，输入模型名称。例如：`llama3.2`。

   ![Download from settings](/images/one/open-webui-download-from-settings1.png#bordered)

6. 点击 <i class="material-symbols-outlined">download</i>。下载会自动开始。

</template>
</Tabs>

:::tip 下载时间
模型大小从 2 GB 到 20+ GB 不等。下载时间取决于你的网络速度。
:::

#### 验证连接

Open WebUI 会自动检测并连接到本地 Ollama。下载的模型出现在聊天页面顶部的模型下拉列表中时，说明连接成功。

### 选项 B：使用模型应用

模型应用会将特定模型与预配置设置打包在一起。如果你希望获得开箱即用的模型，而不想手动管理 Ollama 模型库，推荐使用此选项。

#### 安装模型应用

1. 打开应用市场，搜索你需要的模型。
2. 点击**获取**，然后点击**安装**。等待安装完成。

   ![Install model app](/images/one/qwen3.5-27b.png#bordered)

#### 下载模型

打开刚刚安装的模型应用。模型下载会自动开始。

![Downloading model](/images/one/qwen3.5-27b-downloading.png#bordered)

看到完成页面后，模型即可使用。

![Model downloaded](/images/one/qwen3.5-27b-downloaded.png#bordered)

#### 获取模型应用端点

为了让 Open WebUI 与该模型通信，需要获取模型应用的共享端点 URL。

1. 打开 Olares 设置，前往**应用** > **[Model App Name]**。
2. 在**共享入口** 中，选择该模型，并查看其端点 URL。

   ![Get shared endpoint](/images/one/qwen3.5-27b-shared-entrance.png#bordered){width=70%}

3. 复制端点 URL。例如：`http://94a553e00.shared.olares.com`。

:::tip 为什么不使用模型页面显示的 URL？
模型应用页面显示的 URL 与当前用户绑定，并依赖浏览器前端调用。如果你的设备和 Olares 不在同一局域网内，这些调用可能会触发 Olares 登录，也可能遇到跨域限制（CORS）。为避免这些问题出现，需使用共享端点 URL。
:::

#### 将模型应用连接到 Open WebUI

现在回到 Open WebUI，使用刚刚复制的端点 URL 连接模型。

1. 在 Open WebUI 中，点击你的头像图标，选择 **Admin Panel**。
2. 前往 **Settings** > **Connections**。
3. 在 **Manage Ollama API Connections** 右侧，点击 <span class="material-symbols-outlined">add</span> 添加新连接。
4. 在 **URL** 字段中，粘贴之前复制的模型应用端点 URL。
5. 点击 **Save**。

   Open WebUI 会自动验证连接。看到 "Ollama API settings updated" 消息时，说明连接已建立。

   ![Connection established](/images/one/open-webui-connection-established.png#bordered)

## 开始聊天

模型连接完成后，你就可以使用聊天界面了。

1. 在模型下拉列表中，选择已配置的模型。

   ![Select model](/images/one/open-webui-qwen3.5-27b.png#bordered)

2. 在文本框中输入提示词，然后按回车键开始对话。

   ![Chat with LLM](/images/one/open-webui-chat1.png#bordered)

## 了解更多

- [设置多用户访问](openwebui-multiuser.md)：与 Olares 设备上的其他用户共享 Open WebUI。
- [配置音频](openwebui-audio.md)：启用语音转文字和文字转语音。
- [启用网页搜索](openwebui-search.md)：为聊天增加网页搜索能力。
- [使用知识库](openwebui-knowledge.md)：上传文档并创建用于 RAG 的知识库。
