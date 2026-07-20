---
outline: [2, 3]
description: 了解如何在 Olares 设备上设置 DeerFlow，集成 Ollama 和 Tavily 搜索引擎，实现本地深度研究代理。
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/deerflow.md)为准。
:::

# 使用 DeerFlow 构建本地深度研究代理

DeerFlow 是一个开源框架，能够将简单的研究主题转化为全面、详细的报告。

本指南将介绍如何在 Olares 设备上设置 DeerFlow，并将其与本地 Ollama 模型以及 Tavily 搜索引擎集成，以实现支持网络访问的深度研究。

## 学习目标

在本指南中，你将学习如何：
- 配置 DeerFlow，通过 Ollama 与本地 LLM 通信。
- 配置 Tavily 搜索 API 以访问网络。
- 执行深度研究任务并管理报告。

## Prerequisites

在开始之前，请确保：
- Ollama 已安装并在 Olares 环境中运行。
- 已通过 Ollama 安装至少一个模型。详情请参见 [Ollama](./ollama.md)。
- 你拥有 [Tavily](https://www.tavily.com/) 账户（免费账户即可）。

## 安装 DeerFlow

1. 打开 **Market**，搜索 "DeerFlow"。
   ![安装 DeerFlow](/images/manual/use-cases/deerflow.png#bordered)

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 配置 DeerFlow

DeerFlow 需要 LLM 的连接信息。你可以通过图形界面或命令行编辑 `conf.yaml` 文件来配置。

### 配置 DeerFlow 使用 Ollama

<tabs>
<template #使用图形界面>

1. 打开 Files 应用，导航至 `/Applications/Data/Deerflow/app/`。
2. 找到 `conf.yaml` 文件，并将其下载到本地电脑。
    ![在 Files 中查找 conf.yaml](/images/manual/use-cases/deerflow-conf-yaml-in-files.png#bordered)

3. 使用文本编辑器打开 `conf.yaml` 文件。
4. 修改默认模型设置：
   ```yaml
    BASIC_MODEL:
      base_url:  # 你的 Ollama API 端点（确保包含 /v1 后缀）
      model: # 模型名称
      api_key: # 任意非空字符串
   ```
   例如：
   ```yaml
    BASIC_MODEL:
      base_url: https://39975b9a1.{YOURUSERNAME}.olares.com/v1
      model: "cogito:14b"
      api_key: ollama
   ```
5. 保存文件。
6. 返回 Files 应用，删除原始的 `conf.yaml` 文件，并上传你修改后的版本。

</template>

<template #使用命令行>

你可以直接在主机上通过终端编辑配置文件。
1. 打开 Control Hub，从侧边栏选择 DeerFlow 项目。
2. 导航至 **Deployments** > **deerflow**，然后点击正在运行的 pod。
3. 展开 **deerflow** 容器详情，查看 **Volumes** 部分。
   ![定位 DeerFlow 容器](/images/manual/use-cases/deerflow-locate-containers.png#bordered)

   ![查找 app 文件夹](/images/manual/use-cases/deerflow-app-volume.png#bordered)

4. 复制此路径。
5. 从 Control Hub 打开 Olares 终端，并切换到复制的路径：
   ```bash
   # 替换为实际路径
   cd /olares/rootfs/userspace/pvc-userspace-laresprime-raizlofhiszoin5c/Data/deerflow/app
   ```
6. 使用命令行文本编辑器（如 `nano` 或 `vi`）编辑 `conf.yaml` 文件。例如：
   ```bash
   nano conf.yaml
   ```
7. 修改默认模型设置：
   ```yaml
    BASIC_MODEL:
      base_url:  # 你的 Ollama API 端点（确保包含 /v1 后缀）
      model: # 模型名称
      api_key: # 任意非空字符串
   ```
   例如：
   ```yaml
    BASIC_MODEL:
      base_url: https://39975b9a1.{YOURUSERNAME}.olares.com/v1
      model: "cogito:14b"
      api_key: ollama
   ```
8. 保存更改并退出编辑器。
</template>
</tabs>


### 配置 DeerFlow 使用 Tavily

要启用网络搜索，请将 Tavily API 密钥添加到应用配置中。
1. 在 Control Hub 中，选择 DeerFlow 项目。
2. 在资源列表中点击 **Configmaps**，然后选择 **deerflow-config**。
    ![浏览 DeerFlow 的 configmaps](/images/manual/use-cases/deerflow-configmap.png#bordered)

3. 点击右上角的 <span class="material-symbols-outlined">edit_square</span> 打开编辑器。
4. 在 `data` 部分下添加以下键值对：
   ```yaml
   SEARCH_API: tavily
   TAVILY_API_KEY: tvly-xxx # 你的 Tavily API 密钥
   ```
   ![配置 Tavily](/images/manual/use-cases/deerflow-configure-tavily.png#bordered)

5. 点击 **Confirm** 保存更改。

### 重启 DeerFlow

重启服务以应用新的模型和搜索配置。

1. 在 Control Hub 中，选择 DeerFlow 项目。
2. 在 **Deployments** 下，找到 **deerflow** 并点击 **Restart**。
   ![重启 DeerFlow](/images/manual/use-cases/deerflow-restart.png#bordered)

3. 在确认对话框中，输入 `deerflow` 并点击 **Confirm**。
4. 等待状态图标变为绿色，表示服务已成功重启。

## 运行 DeerFlow

### 运行深度研究任务

1. 从 Olares Launchpad 打开 **DeerFlow**。
2. 点击 **Get Started**，在提示框中输入你的研究主题。
    ![输入研究提示](/images/manual/use-cases/deerflow-enter-prompt.png#bordered)

3. 点击魔棒图标，让 DeerFlow 优化你的提示以获得更好的结果。
4. 启用 **Investigation**。
5. 选择你喜欢的写作风格（例如 **Popular Science**）。
6. 点击 <span class="material-symbols-outlined">arrow_upward</span> 发送请求。

DeerFlow 将生成初步研究计划。如有必要，请查看并编辑此计划，或允许其继续执行。
![生成研究计划](/images/manual/use-cases/deerflow-generate-research-plan.png#bordered)

处理完成后，将显示详细的分析报告。
![查看研究报告](/images/manual/use-cases/deerflow-generate-research-report.png#bordered)

要查看来源和执行的步骤，请点击 **Activities** 标签页。
![查看研究活动](/images/manual/use-cases/deerflow-review-research-activities.png#bordered)

### 编辑并保存报告

:::info 验证引用
AI 模型可能会偶尔生成不准确的引用或"幻觉"链接。请务必手动验证引用部分中的重要来源。
:::

1. 点击右上角的 <span class="material-symbols-outlined">edit</span> 进入编辑模式。
2. 你可以使用 Markdown 调整格式，或选择某个部分并要求 AI 改进或扩展内容。
   ![让 AI 编辑报告](/images/manual/use-cases/deerflow-ask-ai-to-edit.png#bordered)
3. 点击右上角的 <span class="material-symbols-outlined">undo</span> 退出编辑模式。
4. 点击 <span class="material-symbols-outlined">download</span> 将报告保存为 Markdown 文件到本地电脑。

## 添加 MCP 服务器

Model Context Protocol（MCP）通过集成外部工具来扩展 DeerFlow 的功能。例如，添加 Fetch 服务器可以让代理抓取网页内容并将其转换为 Markdown 以供分析。

1. 打开 DeerFlow 应用，点击 <span class="material-symbols-outlined">settings</span> 打开 **Settings** 对话框。
2. 选择 **MCP** 标签页，然后点击 **Add Servers**。
3. 粘贴服务器的 JSON 配置。以下示例添加 fetch 服务器：
   ```json
    {
      "mcpServers": {
        "fetch": {
          "command": "uvx",
          "args": ["mcp-server-fetch"]
        }
      }
    }
   ```
4. 点击 **Add**。服务器将自动启用，并可供研究代理使用。
   ![添加 MCP 服务器](/images/manual/use-cases/deerflow-add-mcp-server.png#bordered)

## 将研究报告转为播客（TTS）

DeerFlow 可以使用文本转语音（TTS）服务（如 Volcengine TTS）将报告转换为 MP3 音频。这需要将 API 凭证添加到应用环境中。

1. 从 [Volcengine](https://console.volcengine.com) 控制台获取你的 **Access Token** 和 **App ID**。
2. 在 Control Hub 中，选择 DeerFlow 项目，然后前往 **Configmaps** > **deerflow-config**。
3. 点击右上角的 **Edit** 图标。
4. 在 `data` 部分下添加以下键：
   ```yaml
   VOLCENGINE_TTS_ACCESS_TOKEN: # 你的 Access Token
   VOLCENGINE_TTS_APPID: # 你的 App ID
   ```
5. 点击 **Confirm** 保存更改。
6. 导航至 **Deployments** > **deerflow** 并点击 **Restart**。

重启后，DeerFlow 将检测这些密钥，播客/TTS 功能将可用。

## FAQ

### DeerFlow 无法生成响应

如果代理无法启动或卡住：
- **检查模型兼容性：** DeerFlow 不支持推理模型（例如 DeepSeek R1）。请切换到标准聊天模型并重试。
- **检查端点配置：** 确保 `conf.yaml` 中的 Ollama API 端点包含 `/v1` 后缀。

### 研究过程中没有网络搜索结果

如果报告内容泛泛且缺乏外部数据：
- **检查模型能力：** 所选 LLM 可能缺乏强大的工具调用能力。请切换到以有效工具使用著称的模型，例如 Qwen 2.5 或 Llama 3.1。
- **验证 API 密钥：** 确保 ConfigMap 中的 `TAVILY_API_KEY` 正确，且账户仍有剩余配额。
