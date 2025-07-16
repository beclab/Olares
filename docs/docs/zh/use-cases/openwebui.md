---
outline: [2, 3]
description: Open WebUI 使用教程，包括模型管理、语音交互和图像生成功能配置，轻松搭建私有 AI 助手服务。
---

# Open WebUI

Open WebUI 为大语言模型（LLM）提供了直观的管理界面，支持 Ollama 和 OpenAI 兼容的 API。本指南将帮助你在 Olares 中设置和配置 Open WebUI，包括：

* 模型管理
* 语音交互（语音转文字和文字转语音）
* 图像生成功能

:::tip 打开 LarePass 专用网络
确保访问设备已安装 LarePass 并 [启用专用网络](../manual/larepass/private-network.md#在-larepass-中启用专用网络)以确保安全、流畅的访问体验。
:::

## 安装 Ollama 和 Open WebUI
使用 Open WebUI 前需要先安装 Ollama：
* **管理员**：需要同时安装“Ollama”和“Open WebUI”。
* **团队成员**：仅需安装“Open WebUI”，同时确保管理员已安装“Ollama”。

![安装 Ollama 和 Open WebUI](/images/manual/use-cases/install-open-webui.png){width=30%}

:::info Open WebUI 账号
首次使用时需要创建本地 Open WebUI 账号。这个账号仅限用于你的 Olares 环境，不会与任何外部服务连接。
请注意，其他环境中创建的 Open WebUI 账号在这里无法使用，你需要重新创建一个新账号。
:::

## 下载模型
:::tip 如何选择模型？
下载前可以在 [Hugging Face Chatbot Arena Leaderboard](https://huggingface.co/spaces/lmsys/chatbot-arena-leaderboard) 浏览可用模型，并在 [Ollama 模型库](https://ollama.com/library)中确认模型名称。
:::

推荐以下入门级模型（参数量在 13B 及以下）以获得最佳性能：

* `gemma2`：Google 最新推出的高效强大语言模型
* `llama3.2`：Meta 最新的开源视觉语言模型

### 快速下载
1. 在首页点击下拉菜单，输入模型名称（如 `llama3.2`）。
2. 选择 从 Ollama.com 拉取，下载会自动开始。

   ![从首页下载模型](/images/zh/manual/use-cases/openwebui-download-model-quick.png#bordered)
### 通过设置下载
1. 点击左下角的用户名称，选择**管理员面板** > **设置** > **模型**。
2. 在**从 Ollama.com 拉取模型**字段下输入模型名称（如 `llama3.2`）。
3. 点击 <i class="material-symbols-outlined">download</i> 开始下载。

   ![从设置下载模型](/images/zh/manual/use-cases/openwebui-download-model-settings.png#bordered)

## 配置语音功能
### 语音转文字
1. 根据角色安装 Faster Whisper：
   - **管理员**：需要同时安装“Faster Whisper For Cluster”和“Faster Whisper”。
   - **团队成员**：仅需安装“Faster Whisper”，同时确保管理员已安装“Faster Whisper For Cluster”。

   ![Install Faster Whisper](/images/manual/use-cases/install-faster-whisper.png){width=40%}

2. 打开 Open WebUI，进入**管理员面板** > **设置** > **音频**。
3. 选择 OpenAI 作为语音转文字引擎，配置如下：
   - **API 基础 URL**：`http://whisper.whisper-{管理员本地名称}:8000/v1`，例如：`http://whisper.whisper-alice123:8000/v1`。
   - **API 密钥**：输入任意字符
4. 输入模型版本，默认为 `whisper-1`。你可以选择：
   - `tiny.en`
   - `tiny`
   - `base.en`
   - `base`
   - `small.en`
   - `small`
   - `medium.en`
   - `medium`
   - `large-v1`
   - `large-v2`
   - `large-v3`
   - `large`
   - `distil-large-v2`
   - `distil-medium.en`
   - `distil-small.en`
   - `distil-large-v3`
5. 点击**保存**。
6. 配置完成后启动 Faster Whisper。你会看到在 Open WebUI 中配置的模型已自动加载。此时你可以：
   - 直接上传音频开始转录
   - 调整参数，包括：
      - 选择不同的子模型
      - 选择任务类型
      - 配置"温度"设置

   ![配置 Faster Whisper](/images/zh/manual/use-cases/openwebui-faster-whisper.png#bordered)

### 文字转语音
1. 管理员安装 OpenedAI Speech 应用，在集群内启动 OpenAI Speech 服务。
2. 打开 Open WebUI，进入**管理员面板** > **设置** > **音频**。
3. 选择 OpenAI 作为文字转语音引擎，配置如下：
   - **API 基础 URL**：`http://openedaispeech.openedaispeech-{管理员本地名称}:8000/v1`，例如：`http://openedaispeech.openedaispeech-alice123:8000/v1`。
   - **API 密钥**：输入任意字符
4. 点击**保存**。

### 文字转图像
在 Olares 环境中安装了 SD Web UI 共享版后，你可以直接通过 Open WebUI 使用 Stable Diffusion 的强大图像生成功能。

1. 管理员安装 SD Web UI 共享版，在集群里启动 Stable Diffusion 服务。
2. 打开 Open WebUI，进入**管理员面板** > **设置** > **图像**。
3. 选择 **Automatic1111** 作为图像生成引擎，基础 URL 为：`http://sdwebui.sdwebui--{管理员本地名称}:7860`，例如：`http://sdwebui.sdwebui-alice123:7860`。
4. 点击 <i class="material-symbols-outlined">cached</i> 验证连接。
5. 开启**图像生成（实验性）**，选择你偏好的文本生成图像模型检查点。
6. 点击**保存**。