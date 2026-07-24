---
outline: deep
description: 在 Olares 设备上设置 DeerFlow 2.0，并将其与本地模型应用配置，实现 AI 驱动的研究和任务处理。
head:
  - - meta
    - name: keywords
      content: Olares, DeerFlow, AI agent, deep research, multi-agent, self-hosted, LLM
doc_version: "1.1"
app_version: "1.0.0"
doc_updated: "2026-07-24"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/deerflow2.md)。
:::

# 设置 DeerFlow 2.0 实现 AI 驱动的研究和任务处理

DeerFlow 是字节跳动开源的智能代理框架，基于 LangGraph 和 LangChain 构建。它通过可扩展的 skill 编排子代理、记忆和沙盒来处理复杂任务。

DeerFlow 2.0 是原版 [DeerFlow](./deerflow.md) 的彻底重写。虽然 1.0 版本是一个深度研究框架，但 2.0 版本是一个通用智能代理平台。

本指南介绍如何在 Olares 上安装 DeerFlow 2.0，并将其与本地模型配置。以 Qwen3.5 27B Q4_K_M（Ollama）为例。

## 学习目标

在本指南中，你将学习如何：
- 在 Olares 上安装 DeerFlow 2.0 并配置本地模型。
- 运行深度研究等任务。

## Prerequisites

- 已从 Market 安装模型应用，且模型已完全下载。

## 安装 DeerFlow 2.0

1. 打开 Market，搜索 "DeerFlow 2.0"。
   ![DeerFlow 2.0](/images/manual/use-cases/deerflow2.png#bordered)

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 配置模型

DeerFlow 2.0 使用 `config.yaml` 文件作为核心配置。要将其连接到本地模型，需要添加一个指向模型应用共享端点的模型条目。

### 获取模型端点和模型名称

1. 从 Launchpad 打开模型应用。页面会显示模型名称（例如 `qwen3.5:27b-q4_K_M`）。记录下来以备后用。
   ![获取模型名称](/images/manual/use-cases/deerflow2-get-model-name.png#bordered)

2. 打开 Settings，然后导航至 **Application** > 你的模型应用（例如 **Qwen3.5 27B Q4_K_M (Ollama)**）。
3. 在 **Shared entrances** 下，选择模型应用以查看端点 URL。

   ![获取共享端点](/images/manual/use-cases/deerflow2-shared-entrance.png#bordered){width=70%}

4. 复制共享端点。例如：
   ```text
   http://94a553e00.shared.olares.com
   ```

### 编辑 config.yaml

1. 打开 Files，导航至 DeerFlow 2.0 应用数据目录：`Application/Data/deerflowv2/config/`。

2. 打开 `config.yaml`，然后点击右上角的 <span class="material-symbols-outlined">edit_square</span> 打开编辑器。

3. 在 `models:` 部分下，使用你复制的共享端点添加模型配置。例如：

   ```yaml
   models:
     - name: qwen3.5:27b-q4_K_M            # 模型的唯一标识符
       display_name: Qwen3.5 27B            # UI 中显示的名称
       use: langchain_openai:ChatOpenAI     # 用于 OpenAI 兼容 API 的 LangChain 类
       model: qwen3.5:27b-q4_K_M           # 模型 ID
       api_key: ollama                      # 使用任意非空文本
       base_url: http://94a553e00.shared.olares.com/v1  # 带 /v1 后缀的共享端点
       supports_thinking: true              # 如果模型支持扩展思考，设为 true
   ```
   ![编辑 config.yaml](/images/manual/use-cases/deerflow2-edit-config-yaml.png#bordered)

4. 点击 <span class="material-symbols-outlined">save</span> 保存更改。

### 重启以应用更改

1. 打开 Control Hub，选择 DeerFlow 2.0 项目。
2. 在 **Deployments** 下，找到后端容器并点击 **Restart**。

   ![重启 DeerFlow 2.0](/images/manual/use-cases/deerflow2-restart.png#bordered)

3. 在确认对话框中，确认重启。
4. 等待状态图标变为绿色。

## 使用 DeerFlow 2.0

模型配置完成后，你就可以开始使用 DeerFlow 2.0 了。

1. 从 Launchpad 打开 DeerFlow 2.0，点击 **Get Started with 2.0**。

2. 首次启动时，创建管理员账号：
    - **Email**：输入管理员账号的电子邮箱地址。
    - **Password**：输入至少包含八个字符的密码。
    - **Confirm Password**：再次输入密码。

   点击 **Create Admin Account** 进入聊天界面。

   ![创建 DeerFlow 管理员账号](/images/manual/use-cases/deerflow2-create-admin-account.png#bordered)

3. 选择你喜欢的执行模式。

   ![选择执行模式](/images/manual/use-cases/deerflow2-select-mode.png#bordered)

   DeerFlow 2.0 提供多种执行模式，控制代理如何处理你的请求，从快速单次回答到多步骤研究（含子代理）。

4. 在聊天框中输入你的提示，或选择一个建议主题作为灵感。

   例如，你可以对某个主题进行深度研究：
   ![深度研究示例](/images/manual/use-cases/deerflow2-research.png#bordered)

   你也可以上传附件，让 DeerFlow 将其作为输入：
   ![上传附件](/images/manual/use-cases/deerflow2-write.png#bordered)

## FAQs

### DeerFlow 2.0 无法生成响应

如果代理无法启动或卡住：

- **检查模型兼容性**：确保你选择的模型已在 `config.yaml` 中正确配置。验证端点 URL 是否正确。
- **检查端点配置**：检查 API 端点是否包含 `/v1` 后缀。

### 如何启用后续建议？

默认情况下，为了降低响应生成后不必要的 GPU 使用率，Olares 上的 DeerFlow 2.0 关闭了后续建议功能。

要启用它：

1. 打开 Control Hub，选择 DeerFlow 2.0 项目。
2. 在 **Deployments** 下，点击 **deerflowv2-frontend** 部署。
3. 点击 <span class="material-symbols-outlined">edit_square</span> 编辑 YAML。
4. 找到 `ENABLE_FOLLOWUP_SUGGESTIONS` 环境变量，将其值改为 `'true'`。
   ![启用后续建议](/images/manual/use-cases/deerflow2-enable-followup-suggestions.png#bordered)

5. 点击 **Confirm** 应用更改。
